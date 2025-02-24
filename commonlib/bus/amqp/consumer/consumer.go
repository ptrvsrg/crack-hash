package consumer

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/goccy/go-json"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/streadway/amqp"

	commonamqp "github.com/ptrvsrg/crack-hash/commonlib/bus/amqp"
)

type (
	Handler[T any] func(ctx context.Context, data T, delivery amqp.Delivery) error

	Config struct {
		Unmarshal func(data []byte, v any) error
		Queue     string
		Consumer  string
		AutoAck   bool
		Exclusive bool
		NoLocal   bool
		NoWait    bool
		Args      map[string]any
	}

	Consumer interface {
		Subscribe(ctx context.Context) error
	}

	consumer[T any] struct {
		ch        *commonamqp.Channel
		handler   Handler[T]
		config    Config
		unmarshal func(data []byte, v any) error
		logger    zerolog.Logger
		wg        sync.WaitGroup
		errChan   chan error
	}
)

func New[T any](ch *commonamqp.Channel, handler Handler[T], cfg Config) Consumer {
	if handler == nil {
		handler = func(context.Context, T, amqp.Delivery) error { return nil }
	}

	if cfg.Unmarshal == nil {
		cfg.Unmarshal = json.Unmarshal
	}

	c := &consumer[T]{
		ch:        ch,
		handler:   handler,
		config:    cfg,
		unmarshal: cfg.Unmarshal,
		errChan:   make(chan error, 1),
		logger: log.With().
			Str("component", "amqp-consumer").
			Type("type", *new(T)).
			Str("queue", cfg.Queue).
			Logger(),
	}

	return c
}

func (c *consumer[T]) connect(_ context.Context) (<-chan amqp.Delivery, error) {
	msgCh, err := c.ch.Consume(
		c.config.Queue,
		c.config.Consumer,
		c.config.AutoAck,
		c.config.Exclusive,
		c.config.NoLocal,
		c.config.NoWait,
		c.config.Args,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to consume queue: %w", err)
	}

	return msgCh, nil
}

func (c *consumer[T]) subscribe(ctx context.Context) error {
	var (
		msgCh <-chan amqp.Delivery
		err   error
	)

	for {
		if msgCh, err = c.connect(ctx); err != nil {
			c.logger.Error().Err(err).Msg("failed to connect consumer to AMQP")
			time.Sleep(10 * time.Second)
			continue
		}
		break
	}

	c.logger.Info().Msg("consumer connected")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("consumer stopped")
			return nil

		case d, ok := <-msgCh:
			if !ok {
				if c.ch.IsClosed() {
					return nil
				}

				c.logger.Info().Msg("try to reconnect consumer")

				c.wg.Add(1)
				go func() {
					defer c.wg.Done()
					if err := c.subscribe(ctx); err != nil {
						c.errChan <- err
					}
				}()

				return nil
			}

			c.logger.Info().Bytes("body", d.Body).Msg("got new event")

			data := *new(T)
			if err := c.unmarshal(d.Body, &data); err != nil {
				c.logger.Error().Err(err).Msg("failed to unmarshal event")
				continue
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						c.logger.Error().Msgf("catch panic: %v\n%s", r, string(debug.Stack()))
					}
				}()

				if err := c.handler(ctx, data, d); err != nil {
					c.logger.Error().Err(err).Msg("failed to consume event")
				}
			}()
		}
	}
}

func (c *consumer[T]) Subscribe(ctx context.Context) error {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		if err := c.subscribe(ctx); err != nil {
			c.errChan <- err
		}
	}()

	// Wait for all goroutines to finish and propagate errors
	go func() {
		c.wg.Wait()
		close(c.errChan)
	}()

	// Collect and return the first error, if any
	errs := make([]error, 0)
	for err := range c.errChan {
		errs = append(errs, err)
	}

	return errors.Join(errs...)
}

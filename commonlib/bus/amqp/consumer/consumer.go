package consumer

import (
	"context"
	"runtime/debug"
	"sync"

	"github.com/goccy/go-json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

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
		Subscribe(ctx context.Context)
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

func (c *consumer[T]) connect(ctx context.Context) <-chan amqp.Delivery {
	return c.ch.Consume(
		ctx,
		c.config.Queue,
		c.config.Consumer,
		c.config.AutoAck,
		c.config.Exclusive,
		c.config.NoLocal,
		c.config.NoWait,
		c.config.Args,
	)
}

func (c *consumer[T]) Subscribe(ctx context.Context) {
	msgCh := c.connect(ctx)
	c.logger.Info().Msg("consumer connected")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info().Msg("consumer stopped")
			return

		case d, ok := <-msgCh:
			if !ok {
				if c.ch.IsClosed() {
					return
				}

				c.logger.Info().Msg("consumer closed, try to reconnect")
				msgCh = c.connect(ctx)
				continue
			}

			c.logger.Info().Bytes("body", d.Body).Msg("got new event")

			data := *new(T)
			if err := c.unmarshal(d.Body, &data); err != nil {
				c.logger.Error().Err(err).Msg("failed to unmarshal event")
				continue
			}

			// catch panic
			go func() {
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

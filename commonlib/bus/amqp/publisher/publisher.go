package publisher

import (
	"context"
	"errors"
	"fmt"
	"github.com/goccy/go-json"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"sync"
	"time"

	amqp2 "github.com/ptrvsrg/crack-hash/commonlib/bus/amqp"
)

const (
	Transient  DeliveryMode = 1
	Persistent DeliveryMode = 2
)

type (
	DeliveryMode uint8

	Config struct {
		Exchange    string
		RoutingKey  string
		Marshal     func(v any) ([]byte, error)
		ContentType string
	}

	Publisher[T any] interface {
		SendMessage(ctx context.Context, message *T, mode DeliveryMode, mandatory, immediate bool) error
	}

	publisher[T any] struct {
		ch          *amqp2.Channel
		config      Config
		marshal     func(v any) ([]byte, error)
		contentType string
		logger      zerolog.Logger
		isConnected bool
		muConn      sync.Mutex
	}
)

func New[T any](ch *amqp2.Channel, config Config) Publisher[T] {
	if config.Marshal == nil {
		config.Marshal = json.Marshal
	}

	if config.ContentType == "" {
		config.ContentType = "application/json"
	}

	pub := &publisher[T]{
		config:      config,
		ch:          ch,
		marshal:     config.Marshal,
		contentType: config.ContentType,
		logger: log.With().
			Str("component", "amqp-publisher").
			Type("type", *new(T)).
			Str("exchange", config.Exchange).
			Str("routing-key", config.RoutingKey).
			Logger(),
	}

	return pub
}

func (p *publisher[T]) connect(_ context.Context) error {
	p.muConn.Lock()
	defer p.muConn.Unlock()

	if p.isConnected {
		return nil
	}

	p.isConnected = true

	return nil
}

func (p *publisher[T]) SendMessage(
	ctx context.Context, message *T, mode DeliveryMode, mandatory, immediate bool,
) error {
	p.logger.Debug().Msg("send message")

	body, err := p.marshal(message)
	if err != nil {
		p.logger.Error().Err(err).Stack().Msg("failed to marshal message")
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ampqMsg := p.buildMessage(body, mode)

	for i := 0; i < 3; i++ {
		sendErr := p.sendMessage(ctx, mandatory, immediate, ampqMsg)
		if sendErr == nil {
			break
		}

		p.logger.Error().Err(sendErr).Stack().Msg("failed to publish a message")
		err = errors.Join(err, sendErr)

		time.Sleep(1 * time.Second)
	}

	if err != nil {
		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func (p *publisher[T]) sendMessage(ctx context.Context, mandatory, immediate bool, ampqMsg *amqp.Publishing) error {
	if !p.isConnected {
		if err := p.connect(ctx); err != nil {
			return fmt.Errorf("failed to connect: %w", err)
		}
	}

	if err := p.ch.PublishWithContext(
		ctx,
		p.config.Exchange,
		p.config.RoutingKey,
		mandatory,
		immediate,
		*ampqMsg,
	); err != nil {
		p.muConn.Lock()
		p.isConnected = false
		p.muConn.Unlock()

		return fmt.Errorf("failed to publish a message: %w", err)
	}

	return nil
}

func (p *publisher[T]) buildMessage(body []byte, mode DeliveryMode) *amqp.Publishing {
	return &amqp.Publishing{
		DeliveryMode: uint8(mode),
		ContentType:  p.contentType,
		Body:         body,
	}
}

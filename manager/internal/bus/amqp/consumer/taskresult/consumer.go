package taskresult

import (
	"context"
	"errors"
	"fmt"

	amqp1 "github.com/rabbitmq/amqp091-go"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp"
	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/consumer"
	"github.com/ptrvsrg/crack-hash/manager/config"
	"github.com/ptrvsrg/crack-hash/manager/internal/service/domain"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
)

func NewConsumer(ch *amqp.Channel, cfg config.AMQPConsumerConfig, svc domain.HashCrackTask) consumer.Consumer {
	return consumer.New(
		ch, handle(svc),
		consumer.Config{
			Queue:     cfg.Queue,
			Consumer:  "",
			AutoAck:   false,
			Exclusive: false,
			NoLocal:   false,
			NoWait:    false,
		},
	)
}

func handle(svc domain.HashCrackTask) consumer.Handler[message.HashCrackTaskResult] {
	return func(ctx context.Context, msg message.HashCrackTaskResult, delivery amqp1.Delivery) error {
		err := svc.SaveResultSubtask(ctx, &msg)
		if err != nil && !errors.Is(err, domain.ErrTaskNotFound) && !errors.Is(err, domain.ErrInvalidRequestID) {
			if err := delivery.Reject(true); err != nil {
				return fmt.Errorf("failed to reject message: %w", err)
			}

			return fmt.Errorf("failed to save result task: %w", err)
		}

		if err := delivery.Ack(false); err != nil {
			return fmt.Errorf("failed to ack message: %w", err)
		}

		return nil
	}
}

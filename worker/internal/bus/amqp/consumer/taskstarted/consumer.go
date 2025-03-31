package taskstarted

import (
	"context"
	"fmt"

	amqp1 "github.com/rabbitmq/amqp091-go"

	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp"
	"github.com/ptrvsrg/crack-hash/commonlib/bus/amqp/consumer"
	"github.com/ptrvsrg/crack-hash/manager/pkg/message"
	"github.com/ptrvsrg/crack-hash/worker/config"
	"github.com/ptrvsrg/crack-hash/worker/internal/service/domain"
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

func handle(svc domain.HashCrackTask) consumer.Handler[message.HashCrackTaskStarted] {
	return func(ctx context.Context, msg message.HashCrackTaskStarted, delivery amqp1.Delivery) error {
		if err := svc.ExecuteTask(ctx, &msg); err != nil {
			if err := delivery.Reject(true); err != nil {
				return fmt.Errorf("failed to reject message: %w", err)
			}

			return fmt.Errorf("failed to execute task: %w", err)
		}

		if err := delivery.Ack(false); err != nil {
			return fmt.Errorf("failed to ack message: %w", err)
		}

		return nil
	}
}

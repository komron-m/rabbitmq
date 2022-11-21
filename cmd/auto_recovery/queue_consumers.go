package main

import (
	"context"
	"github.com/komron-m/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
)

func firstConsumer(_ context.Context, msg amqp.Delivery) {
	log.Println("firstConsumer body:", string(msg.Body))
	if err := msg.Ack(false); err != nil {
		log.Fatal("ack on first consumer failed", err.Error())
	}
}

func secondConsumer(_ context.Context, msg amqp.Delivery) {
	log.Println("secondConsumer body:", string(msg.Body))
	if err := msg.Ack(false); err != nil {
		log.Fatal("ack on second consumer failed", err.Error())
	}
}

func getQueueConsumerParams() []rabbitmq.QueueConsumerParams {
	deadLetterExchange := "test.dead_letter_exchange"

	return []rabbitmq.QueueConsumerParams{
		// dead letter exchange
		{
			ExchangeName: deadLetterExchange,
			ExchangeType: amqp.ExchangeTopic,
			ExchangeArgs: nil,
			QueueName:    "test.dead_letter_queue",
			QueueArgs: amqp.Table{
				"x-queue-mode": "lazy",
			},
			RoutingKeys:  []string{"#"},
			Consumer:     nil, // TODO: implement your consumer for `poisoned` messages
			ConsumerArgs: nil,
		},
		// exchange 1
		{
			ExchangeName: "test.exchange_1",
			ExchangeType: amqp.ExchangeTopic,
			ExchangeArgs: nil,
			QueueName:    "test.queue_1",
			QueueArgs: amqp.Table{
				"x-queue-type":           "quorum", // see: https://www.rabbitmq.com/quorum-queues.html#use-cases
				"x-delivery-limit":       3,
				"x-dead-letter-exchange": deadLetterExchange,
			},
			RoutingKeys:  []string{"#"},
			Consumer:     rabbitmq.ConsumerFunc(firstConsumer),
			ConsumerArgs: nil,
		},
		// exchange 2
		{
			ExchangeName: "test.exchange_2",
			ExchangeType: amqp.ExchangeTopic,
			ExchangeArgs: nil,
			QueueName:    "test.queue_2",
			QueueArgs: amqp.Table{
				"x-queue-type":           "quorum",
				"x-delivery-limit":       3,
				"x-dead-letter-exchange": deadLetterExchange,
			},
			RoutingKeys:  []string{"#"},
			Consumer:     rabbitmq.ConsumerFunc(secondConsumer),
			ConsumerArgs: nil,
		},
	}
}

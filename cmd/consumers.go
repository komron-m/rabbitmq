package main

import (
	"bytes"
	"context"
	"log"

	"github.com/google/uuid"
	"github.com/komron-m/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func simpleLogger(ctx context.Context, msg amqp.Delivery) {
	log.Println("EXCHANGE", msg.Exchange)
	log.Println("ROUTING_KEY", msg.RoutingKey)
	log.Println("MSG_BODY", string(msg.Body))

	log.Println(bytes.Repeat([]byte("-"), 64))

	if err := msg.Ack(false); err != nil {
		log.Fatal(err)
	}
}

func makeConsumers() []rabbitmq.AMQPConsumer {
	primary := makeConsumer(
		"main",
		"main_consumer_1",
		amqp.Table{
			"x-queue-type":     "quorum",
			"x-delivery-limit": 3,
		},
		[]string{
			"topic_1",
			"topic_2",
		},
		rabbitmq.ConsumerFunc(simpleLogger),
	)

	secondary := makeConsumer(
		"main",
		"main_consumer_2",
		amqp.Table{
			"x-queue-type":     "quorum",
			"x-delivery-limit": 3,
		},
		[]string{
			"topic_2",
			"topic_3",
		},
		rabbitmq.ConsumerFunc(simpleLogger),
	)

	var consumers []rabbitmq.AMQPConsumer
	consumers = append(consumers, primary, secondary)
	return consumers
}

func makeConsumer(
	exchangeName string,
	queueName string,
	queueArgs amqp.Table,
	routingKeys []string,
	consumer rabbitmq.IConsumer,
) rabbitmq.AMQPConsumer {
	return rabbitmq.AMQPConsumer{
		ExchangeParams: rabbitmq.ExchangeParams{
			Name:       exchangeName,
			Type:       amqp.ExchangeTopic,
			Durable:    true,
			AutoDelete: false,
			Internal:   false,
			Nowait:     false,
			Args:       nil,
		},
		QueueParams: rabbitmq.QueueParams{
			Name:       queueName,
			Durable:    true,
			AutoDelete: false,
			Exclusive:  false,
			Nowait:     false,
			Args:       queueArgs,
		},
		QueueBindParams: rabbitmq.QueueBindParams{
			Nowait: false,
			Args:   nil,
		},
		ConsumerParams: rabbitmq.ConsumerParams{
			RoutingKeys: routingKeys,
			ConsumerID:  uuid.NewString(),
			AutoAck:     false,
			Exclusive:   false,
			NoLocal:     false,
			Nowait:      false,
			Args:        nil,
		},
		IConsumer: consumer,
	}
}

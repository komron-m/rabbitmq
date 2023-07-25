package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

// An IConsumer handles an amqp message (delivery - in terms of github.com/rabbitmq/amqp091-go library).
//
// Underlying implementations are responsible for either `Acknowledge` or `Reject` of the message.
//
// HINT: one can consume messages the same way as http request handling via http.Handler
// i.e. building chain of middlewares for: panic-handling, logging, tracing and error-handling ...etc.
type IConsumer interface {
	Consume(ctx context.Context, msg amqp.Delivery)
}

type ConsumerFunc func(ctx context.Context, msg amqp.Delivery)

func (f ConsumerFunc) Consume(ctx context.Context, msg amqp.Delivery) {
	f(ctx, msg)
}

// AMQPConsumer collects all parameters required for creating a queue and consuming messages
type AMQPConsumer struct {
	ExchangeParams
	QueueParams
	QueueBindParams
	ConsumerParams

	IConsumer
}

// ExchangeParams and all other underlying structs created for compacting amqp091-go `function` parameters
type ExchangeParams struct {
	Name       string
	Type       string
	Durable    bool
	AutoDelete bool
	Internal   bool
	Nowait     bool
	Args       amqp.Table
}

type QueueParams struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	Nowait     bool
	Args       amqp.Table
}

type QueueBindParams struct {
	Nowait bool
	Args   amqp.Table
}

type ConsumerParams struct {
	RoutingKeys []string
	ConsumerID  string
	AutoAck     bool
	Exclusive   bool
	NoLocal     bool
	Nowait      bool
	Args        amqp.Table
}

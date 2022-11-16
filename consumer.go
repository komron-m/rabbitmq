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
// i.e. building chain of middlewares for: panic-handling, logging and tracing, error-handling ...etc.
type IConsumer interface {
	Consume(context.Context, amqp.Delivery)
}

type ConsumerFunc func(context.Context, amqp.Delivery)

func (f ConsumerFunc) Consume(ctx context.Context, msg amqp.Delivery) {
	f(ctx, msg)
}

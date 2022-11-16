package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ErrPublishNotAcked struct {
	ExchangeName string
	RoutingKey   string
	Msg          amqp.Publishing
}

func (e *ErrPublishNotAcked) Error() string {
	return fmt.Sprintf("publish not acknowledged by server, exchange: %s, routing_key: %s", e.ExchangeName, e.RoutingKey)
}

func NewErrPublishNotAcked(exName, key string, msg amqp.Publishing) *ErrPublishNotAcked {
	return &ErrPublishNotAcked{
		ExchangeName: exName,
		RoutingKey:   key,
		Msg:          msg,
	}
}

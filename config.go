package rabbitmq

import amqp "github.com/rabbitmq/amqp091-go"

type Config struct {
	// params for setting up dial and new connection
	DialParams DialParams
	// https://www.rabbitmq.com/docs/consumer-prefetch
	QOSParams QOSParams

	// this callback will be invoked whenever network failure happens or node shuts down
	NetworkErrCallback func(*amqp.Error)

	PublisherConfirmEnabled bool

	PublisherConfirmNowait bool
}

type DialParams struct {
	User       string
	Password   string
	Host       string
	Port       string
	AMQPConfig amqp.Config
}

type QOSParams struct {
	PrefetchCount int
	PrefetchSize  int
	Global        bool
}

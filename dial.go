package rabbitmq

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Dial a handy wrapper for base "github.com/rabbitmq/amqp091-go" DialConfig function
func Dial(params DialParams) (*amqp.Connection, error) {
	proto := "amqp"
	if params.AMQPConfig.TLSClientConfig != nil {
		proto = "amqps"
	}
	url := fmt.Sprintf("%s://%s:%s@%s:%s/", proto, params.User, params.Password, params.Host, params.Port)
	return amqp.DialConfig(url, params.AMQPConfig)
}

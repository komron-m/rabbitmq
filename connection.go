package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
)

type ConnectionConfig struct {
	User     string
	Password string
	Host     string
	Port     string
	Qos      int
	amqpCfg  amqp.Config
}

func DialConfig(cfg ConnectionConfig) (*amqp.Connection, error) {
	proto := "amqp"
	if cfg.amqpCfg.TLSClientConfig != nil {
		proto = "amqps"
	}
	url := fmt.Sprintf("%s://%s:%s@%s:%s/", proto, cfg.User, cfg.Password, cfg.Host, cfg.Port)
	return amqp.DialConfig(url, cfg.amqpCfg)
}

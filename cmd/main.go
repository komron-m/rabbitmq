package main

import (
	"log"
	"os"
	"time"

	"github.com/komron-m/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	client, err := rabbitmq.NewClient(rabbitmq.ClientConfig{
		NetworkErrCallback: func(a *amqp.Error) {
			log.Println("NetworkErrCallback:", a.Error())
		},

		AutoRecoveryInterval: time.Second * 10,

		AutoRecoveryErrCallback: func(err error) bool {
			log.Println("AutoRecoveryErrCallback:", err.Error())
			return true
		},

		ConsumerAutoRecoveryErrCallback: func(consumer rabbitmq.AMQPConsumer, err error) {
			log.Println("ConsumerAutoRecoveryErrCallback:", err.Error())
			log.Println("ConsumerAutoRecoveryErrCallback.consumer:", consumer.ConsumerID)
		},

		DialCfg: rabbitmq.DialConfig{
			User:     "rabbitmq",
			Password: "secret",
			Host:     "localhost",
			Port:     "5672",
			AMQPConfig: amqp.Config{
				Vhost: "/",
			},
		},

		PublisherConfirmEnabled: true,
		PublisherConfirmNowait:  false,
		ConsumerQos:             2,
		ConsumerPrefetchSize:    0,
		ConsumerGlobal:          false,
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	consumers := makeConsumers()
	for _, consumer := range consumers {
		if err := client.Consume(consumer); err != nil {
			log.Fatal(err)
		}
	}

	signal := make(chan os.Signal)
	<-signal
}

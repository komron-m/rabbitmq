package main

import (
	"github.com/komron-m/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

func logNetworkError(a *amqp.Error) {
	log.Println("NetworkErrCallback:", a.Error())
}

func handleHouseKeepingError(err error) bool {
	log.Println("AutoRecoveryErrCallback:", err.Error())
	return true
}

func main() {
	cfg := rabbitmq.ClientConfig{
		ConnectionConfig: rabbitmq.ConnectionConfig{
			User:                    "rabbitmq1",
			Password:                "secret",
			Host:                    "localhost",
			Port:                    "5672",
			PublisherConfirmEnabled: true,
			Qos:                     2,
		},
		ConsumerParams:     getQueueConsumerParams(),
		NetworkErrCallback: logNetworkError,

		AutoRecoveryEnabled:     true,
		AutoRecoveryInterval:    time.Minute,
		AutoRecoveryErrCallback: handleHouseKeepingError,
	}

	client, err := rabbitmq.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	log.Println("Press CTRL+C to break application")

	ch := make(<-chan struct{})
	<-ch
}

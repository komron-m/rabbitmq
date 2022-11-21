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
		// provide connection params
		ConnectionConfig: rabbitmq.ConnectionConfig{
			User:                    "rabbitmq1",
			Password:                "secret",
			Host:                    "localhost",
			Port:                    "5672",
			PublisherConfirmEnabled: true,
			Qos:                     2,
		},

		// provide list of all your consumers you want to be initialized and restored after reconnection
		ConsumerParams: getQueueConsumerParams(),

		// provide a callback for your internal usage, ex: logging about network error
		NetworkErrCallback: logNetworkError,

		//
		AutoRecoveryEnabled: true,

		// initial connection and first reconnection will be instant,
		// if reconnection fails wait for given interval
		AutoRecoveryInterval: time.Minute,

		// during reconnection there might be more errors
		// so through this callback you can log error and decide whether reconnection should go on
		// by returning boolean as an indicator
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

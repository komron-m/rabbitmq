package main

import (
	"github.com/komron-m/rabbitmq"
	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"time"
)

func logNetworkError(a *amqp.Error, config rabbitmq.ConnectionConfig) {
	log.Println("NetworkErrCallback:", a.Error(), "host port", config.Host, "::", config.Port)
}

func handleHouseKeepingError(err error, config rabbitmq.ConnectionConfig) bool {
	log.Println("HousekeepingErrCallback:", err.Error(), "host port", config.Host, "::", config.Port)
	return true
}

func main() {
	cfg := rabbitmq.ClientConfig{
		ConnConfigs:        getConnectionConfigs(),
		ConsumerParams:     getQueueConsumerParams(),
		NetworkErrCallback: logNetworkError,

		AutoRecoveryEnabled:     true,
		AutoRecoveryInterval:    time.Minute,
		HousekeepingErrCallback: handleHouseKeepingError,
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

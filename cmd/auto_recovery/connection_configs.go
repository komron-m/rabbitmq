package main

import "github.com/komron-m/rabbitmq"

func getConnectionConfigs() []rabbitmq.ConnectionConfig {
	return []rabbitmq.ConnectionConfig{
		{
			User:                    "rabbitmq1",
			Password:                "secret",
			Host:                    "localhost",
			Port:                    "5672",
			PublisherConfirmEnabled: true,
			Qos:                     2,
		},
		{
			User:                    "rabbitmq1",
			Password:                "secret",
			Host:                    "localhost",
			Port:                    "5674",
			PublisherConfirmEnabled: true,
			Qos:                     2,
		},
		{
			User:                    "rabbitmq1",
			Password:                "secret",
			Host:                    "localhost",
			Port:                    "5676",
			PublisherConfirmEnabled: true,
			Qos:                     2,
		},
		{
			User:                    "rabbitmq1",
			Password:                "secret",
			Host:                    "localhost",
			Port:                    "5678",
			PublisherConfirmEnabled: true,
			Qos:                     2,
		},
		{
			User:                    "rabbitmq1",
			Password:                "secret",
			Host:                    "localhost",
			Port:                    "5680",
			PublisherConfirmEnabled: true,
			Qos:                     2,
		},
	}
}

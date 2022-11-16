package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type ClientConfig struct {
	// list of all connection configurations of rabbitmq nodes to connect to
	ConnConfigs []ConnectionConfig

	// list of all consumers to build during initial start-up
	// note that only these consumers will be restored after connection recovery
	ConsumerParams []QueueConsumerParams

	// this callback will be invoked whenever network failure happens or node shuts down
	NetworkErrCallback func(*amqp.Error, ConnectionConfig)

	// if false then below params not required
	AutoRecoveryEnabled bool

	// note that this interval is taken into account when on reconnecting multiple times in row
	AutoRecoveryInterval time.Duration

	// this callback will be invoked whenever connection retry fails or resource freeing returns error
	// returning value as bool should indicate whether to keep retrying recovery or not
	HousekeepingErrCallback func(error, ConnectionConfig) bool
}

type Client struct {
	*client

	activeConnectionIndex int
	cfg                   ClientConfig
	conn                  *amqp.Connection
}

func NewClient(cfg ClientConfig) (*Client, error) {
	if len(cfg.ConnConfigs) == 0 {
		return nil, fmt.Errorf("provide at least one rabbitmq.ConnectionConfig")
	}

	c := new(Client)
	c.cfg = cfg
	c.activeConnectionIndex = 0
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	// initially takes first configuration
	connCfg := c.cfg.ConnConfigs[c.activeConnectionIndex]
	conn, err := DialConfig(connCfg)
	if err != nil {
		return err
	}

	// it's recommended to separate publisher and consumer channels in order to avoid heavy control-flows
	// see also: https://www.rabbitmq.com/channels.html#flow-control

	// create publisher channel
	publisherChannel, err := conn.Channel()
	if err != nil {
		return err
	}
	if connCfg.PublisherConfirmEnabled {
		if err := publisherChannel.Confirm(false); err != nil {
			return err
		}
	}

	// create consumer channel
	consumerChannel, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := consumerChannel.Qos(connCfg.Qos, 0, false); err != nil {
		return err
	}

	c.client = &client{
		publisherChannel: publisherChannel,
		consumerChannel:  consumerChannel,
	}
	c.conn = conn

	// register and start consuming messages
	var registerConsumerErr error
	for _, consumerParam := range c.cfg.ConsumerParams {
		registerConsumerErr = c.client.RegisterQueueConsumer(consumerParam)
	}
	if registerConsumerErr != nil {
		return registerConsumerErr
	}

	go func() {
		errNotifyChan := c.conn.NotifyClose(make(chan *amqp.Error))

		for connectionErr := range errNotifyChan {
			c.cfg.NetworkErrCallback(connectionErr, connCfg)

			if c.cfg.AutoRecoveryEnabled {
				c.reconnect()
			}
		}
	}()
	return nil
}

func (c *Client) reconnect() {
	if err := c.Close(); err != nil {
		shouldContinue := c.cfg.HousekeepingErrCallback(err, c.cfg.ConnConfigs[c.activeConnectionIndex])
		if !shouldContinue {
			return
		}
	}

	// round-robin over connection configurations at try to connect to new node each time error happens
	c.activeConnectionIndex += 1
	c.activeConnectionIndex = c.activeConnectionIndex % len(c.cfg.ConnConfigs)

	if err := c.connect(); err != nil {
		shouldContinue := c.cfg.HousekeepingErrCallback(err, c.cfg.ConnConfigs[c.activeConnectionIndex])
		if shouldContinue {
			time.AfterFunc(c.cfg.AutoRecoveryInterval, func() {
				c.reconnect()
			})
		}
	}
}

func (c *Client) Close() (err error) {
	if !c.conn.IsClosed() {
		err = c.client.Close()
		err = c.conn.Close()
	}
	return
}

package rabbitmq

import (
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type ClientConfig struct {
	// connection configuration of rabbitmq
	ConnectionConfig ConnectionConfig

	// list of all consumers to build during initial start-up
	// note that only these consumers will be restored after connection recovery
	ConsumerParams []QueueConsumerParams

	// this callback will be invoked whenever network failure happens or node shuts down
	NetworkErrCallback func(*amqp.Error)

	// if false then below params not required
	AutoRecoveryEnabled bool

	// note that this interval is taken into account when on reconnecting multiple times in row
	AutoRecoveryInterval time.Duration

	// this callback will be invoked whenever connection retry fails or resource freeing returns error
	// returning value as bool should indicate whether to keep retrying recovery or not
	AutoRecoveryErrCallback func(error) bool
}

type Client struct {
	*client

	cfg  ClientConfig
	conn *amqp.Connection
}

func NewClient(cfg ClientConfig) (*Client, error) {
	c := new(Client)
	c.cfg = cfg
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	// initially takes first configuration
	conn, err := DialConfig(c.cfg.ConnectionConfig)
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
	if c.cfg.ConnectionConfig.PublisherConfirmEnabled {
		if err := publisherChannel.Confirm(false); err != nil {
			return err
		}
	}

	// create consumer channel
	consumerChannel, err := conn.Channel()
	if err != nil {
		return err
	}
	if err := consumerChannel.Qos(c.cfg.ConnectionConfig.Qos, 0, false); err != nil {
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
			c.cfg.NetworkErrCallback(connectionErr)

			if c.cfg.AutoRecoveryEnabled {
				c.reconnect()
			}
		}
	}()
	return nil
}

func (c *Client) reconnect() {
	if err := c.Close(); err != nil {
		shouldContinue := c.cfg.AutoRecoveryErrCallback(err)
		if !shouldContinue {
			return
		}
	}

	// try to connect
	if err := c.connect(); err != nil {
		shouldContinue := c.cfg.AutoRecoveryErrCallback(err)
		if shouldContinue {
			// when instant reconnect fails, try again after some time
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

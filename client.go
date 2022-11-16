package rabbitmq

import (
	"fmt"
	amqp "github.com/rabbitmq/amqp091-go"
	"time"
)

type ClientConfig struct {
	ConnConfigs          []ConnectionConfig
	NetworkErrorCallback func(*amqp.Error, ConnectionConfig)

	AutoRecoveryEnabled          bool
	AutoRecoveryInterval         time.Duration
	CloseConnectionErrorCallback func(error, ConnectionConfig)
	AutoReconnectErrorCallback   func(error, ConnectionConfig)
}

type Client struct {
	*client

	cfg                   ClientConfig
	consumerParams        []QueueConsumerParams
	activeConnectionIndex int

	conn *amqp.Connection
}

func NewClient(cfg ClientConfig, consumerParams []QueueConsumerParams) (*Client, error) {
	if len(cfg.ConnConfigs) == 0 {
		return nil, fmt.Errorf("provide at least one rabbitmq.ConnectionConfig")
	}

	c := new(Client)
	c.cfg = cfg
	c.consumerParams = consumerParams
	c.activeConnectionIndex = 0
	if err := c.connect(); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) connect() error {
	connCfg := c.cfg.ConnConfigs[c.activeConnectionIndex]
	conn, err := DialConfig(connCfg)
	if err != nil {
		return err
	}

	publisherChannel, err := conn.Channel()
	if err != nil {
		return err
	}
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

	var registerConsumerErr error
	for _, consumerParam := range c.consumerParams {
		registerConsumerErr = c.client.RegisterQueueConsumer(consumerParam)
	}
	if registerConsumerErr != nil {
		return registerConsumerErr
	}

	go func() {
		errNotifyChan := c.conn.NotifyClose(make(chan *amqp.Error))
		for connectionErr := range errNotifyChan {
			c.cfg.NetworkErrorCallback(connectionErr, connCfg)

			if c.cfg.AutoRecoveryEnabled {
				c.reconnect()
			}
		}
	}()
	return nil
}

func (c *Client) reconnect() {
	if err := c.Close(); err != nil {
		c.cfg.CloseConnectionErrorCallback(err, c.cfg.ConnConfigs[c.activeConnectionIndex])
	}

	c.activeConnectionIndex += 1
	c.activeConnectionIndex = c.activeConnectionIndex % len(c.cfg.ConnConfigs)

	if err := c.connect(); err != nil {
		c.cfg.AutoReconnectErrorCallback(err, c.cfg.ConnConfigs[c.activeConnectionIndex])
		time.AfterFunc(c.cfg.AutoRecoveryInterval, func() {
			c.reconnect()
		})
	}
}

func (c *Client) Close() (err error) {
	err = c.client.Close()

	err = c.conn.Close()

	return
}

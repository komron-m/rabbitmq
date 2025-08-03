package rabbitmq

import (
	"context"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Client struct {
	cfg       Config
	consumers []ConsumerParams

	connection    *amqp.Connection
	publisherChan *amqp.Channel
	consumerChan  *amqp.Channel

	mx sync.RWMutex
	wg sync.WaitGroup
}

func NewClient(cfg Config) (*Client, error) {
	client := Client{
		cfg: cfg,
	}

	if err := client.connect(); err != nil {
		return nil, err
	}

	return &client, nil
}

func (c *Client) connect() error {
	conn, err := Dial(c.cfg.DialParams)
	if err != nil {
		return err
	}

	// it's recommended to separate publisher and consumer channels in order to avoid heavy control-flows
	// https://www.rabbitmq.com/channels.html#flow-control
	publisherChannel, err := conn.Channel()
	if err != nil {
		return err
	}
	if c.cfg.PublisherConfirmEnabled {
		if err := publisherChannel.Confirm(c.cfg.PublisherConfirmNowait); err != nil {
			return err
		}
	}

	// create new channel for consumer
	consumerChannel, err := conn.Channel()
	if err != nil {
		return err
	}

	qosParams := c.cfg.QOSParams
	if err := consumerChannel.Qos(qosParams.PrefetchCount, qosParams.PrefetchSize, qosParams.Global); err != nil {
		return err
	}

	// reassign
	c.connection = conn
	c.publisherChan = publisherChannel
	c.consumerChan = consumerChannel

	go func() {
		errNotifyChan := c.connection.NotifyClose(make(chan *amqp.Error))

		for connectionErr := range errNotifyChan {
			c.cfg.NetworkErrCallback(connectionErr)
			break
		}
	}()

	return nil
}

// Publish wrapper for amqp091-go PublishWithContext
func (c *Client) Publish(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return c.publisherChan.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}

// PublishConfirm publishes message and waits for confirmation
// make sure `PublisherConfirmEnabled` is true in configuration
func (c *Client) PublishConfirm(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	defConfirm, err := c.publisherChan.PublishWithDeferredConfirm(exchange, key, mandatory, immediate, msg)
	if err != nil {
		return err
	}

	acked, err := defConfirm.WaitContext(ctx)
	if err != nil {
		return err
	}

	if !acked {
		return NewErrPublishNotAcked(exchange, key, msg)
	}

	return nil
}

// Consume makes all necessary housekeeping stuff for creating an exchange, queue, queue-binding
// then starts new goroutine (if IConsumer is set) which starts receiving messages from rabbitmq-server
// safe for concurrent usage
func (c *Client) Consume(consumer ConsumerParams) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	c.consumers = append(c.consumers, consumer)

	return c.consume(consumer)
}

func (c *Client) consume(consumer ConsumerParams) error {
	if err := c.consumerChan.ExchangeDeclare(
		consumer.ExchangeParams.Name,
		consumer.ExchangeParams.Type,
		consumer.ExchangeParams.Durable,
		consumer.ExchangeParams.AutoDelete,
		consumer.ExchangeParams.Internal,
		consumer.ExchangeParams.Nowait,
		consumer.ExchangeParams.Args,
	); err != nil {
		return err
	}

	queue, err := c.consumerChan.QueueDeclare(
		consumer.QueueParams.Name,
		consumer.QueueParams.Durable,
		consumer.QueueParams.AutoDelete,
		consumer.QueueParams.Exclusive,
		consumer.QueueParams.Nowait,
		consumer.QueueParams.Args,
	)
	if err != nil {
		return err
	}

	for _, routingKey := range consumer.RoutingKeys {
		// The default exchange is a `direct type` with no name (empty string) pre-declared by the broker.
		// It has one special property that makes it very useful for simple applications:
		// every queue that is created is automatically bound to it with a routing event which is the same as the queue name.
		if consumer.ExchangeParams.Name != "" {
			if err := c.consumerChan.QueueBind(
				queue.Name,
				routingKey,
				consumer.ExchangeParams.Name,
				consumer.QueueBindParams.Nowait,
				consumer.QueueBindParams.Args,
			); err != nil {
				return err
			}
		}
	}

	if consumer.IConsumer != nil {
		deliveries, err := c.consumerChan.Consume(
			consumer.QueueParams.Name,
			consumer.RoutingParams.ConsumerID,
			consumer.RoutingParams.AutoAck,
			consumer.RoutingParams.Exclusive,
			consumer.RoutingParams.NoLocal,
			consumer.RoutingParams.Nowait,
			consumer.RoutingParams.Args,
		)
		if err != nil {
			return err
		}

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()

			for msg := range deliveries {
				consumer.Consume(context.Background(), msg)
			}
		}()
	}

	return nil
}

func (c *Client) Close() error {
	var err error

	if !c.connection.IsClosed() {
		err = c.publisherChan.Close()

		for _, consumer := range c.consumers {
			err = c.consumerChan.Cancel(consumer.ConsumerID, false)
		}
		c.wg.Wait()

		err = c.consumerChan.Close()

		err = c.connection.Close()
	}

	return err
}

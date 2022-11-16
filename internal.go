package rabbitmq

import (
	"context"
	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"
	"sync"
)

type QueueConsumerParams struct {
	ExchangeName string
	ExchangeType string
	ExchangeArgs amqp.Table
	QueueName    string
	QueueArgs    amqp.Table
	RoutingKeys  []string

	Consumer     IConsumer
	ConsumerArgs amqp.Table
}

type client struct {
	publisherChannel *amqp.Channel
	consumerChannel  *amqp.Channel

	// internal stuff
	consumerIDS []string
	wg          sync.WaitGroup
	mx          sync.Mutex
}

func (c *client) Publish(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	return c.publisherChannel.PublishWithContext(ctx, exchange, key, mandatory, immediate, msg)
}

func (c *client) PublishSync(ctx context.Context, exchange, key string, mandatory, immediate bool, msg amqp.Publishing) error {
	defConfirm, err := c.publisherChannel.PublishWithDeferredConfirmWithContext(ctx, exchange, key, mandatory, immediate, msg)
	if err != nil {
		return err
	}
	success := defConfirm.Wait()
	if !success {
		return NewErrPublishNotAcked(exchange, key, msg)
	}
	return nil
}

func (c *client) RegisterQueueConsumer(p QueueConsumerParams) error {
	c.mx.Lock()
	defer c.mx.Unlock()

	if err := c.consumerChannel.ExchangeDeclare(
		p.ExchangeName,
		p.ExchangeType,
		true,
		false,
		false,
		false,
		p.ExchangeArgs,
	); err != nil {
		return err
	}

	durable := p.QueueName != ""
	queue, err := c.consumerChannel.QueueDeclare(
		p.QueueName,
		durable,
		false,
		false,
		false,
		p.QueueArgs,
	)
	if err != nil {
		return err
	}

	for _, key := range p.RoutingKeys {
		// The default exchange is a `direct type` with no name (empty string) pre-declared by the broker.
		// It has one special property that makes it very useful for simple applications:
		// every queue that is created is automatically bound to it with a routing event which is the same as the queue name.
		if p.ExchangeName != "" {
			if err := c.consumerChannel.QueueBind(
				queue.Name,
				key,
				p.ExchangeName,
				false,
				nil,
			); err != nil {
				return err
			}
		}
	}

	if p.Consumer != nil {
		consumerID := uuid.NewString()
		c.consumerIDS = append(c.consumerIDS, consumerID)
		deliveries, err := c.consumerChannel.Consume(
			p.QueueName,
			consumerID,
			false,
			false,
			false,
			false,
			p.ConsumerArgs,
		)
		if err != nil {
			return err
		}

		c.wg.Add(1)
		go func() {
			defer c.wg.Done()

			for msg := range deliveries {
				p.Consumer.Consume(context.Background(), msg)
			}
		}()
	}

	return nil
}

func (c *client) Close() (err error) {
	c.mx.Lock()
	defer c.mx.Unlock()

	// close publisher channel
	err = c.publisherChannel.Close()

	// stop consumers
	for _, consumerID := range c.consumerIDS {
		err = c.consumerChannel.Cancel(consumerID, false)
	}
	c.wg.Wait()

	// close consumer channel
	err = c.consumerChannel.Close()

	// clean internals
	c.consumerIDS = []string{}

	return
}

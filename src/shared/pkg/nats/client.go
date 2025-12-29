package nats

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

// New creates a new NatsClient and connects to the NATS server at the given URL
func New(url string) (*NatsClient, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, errors.New("could not connect to NATs server at specified URL")
	}

	js, err := nc.JetStream(nats.PublishAsyncMaxPending(256)) // enable JetStream with async publishing (buffer size of 256)
	if err != nil {
		return nil, errors.New("could not connect to NATS JetStream server")
	}

	// add a users stream to JetStream with custom configuration
	usersStreamConfig := createStreamConfig("users", []string{"users.>"})

	_, err = js.AddStream(usersStreamConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create users stream: %v", err)
	}

	// add more streams when needed

	log.Println("Successfully connected to NATS JetStream server")

	return &NatsClient{
		nc: nc,
		js: js,
	}, nil
}

// Publish a message to a subject
func (c *NatsClient) Publish(subject string, data []byte) error {
	_, err := c.js.Publish(subject, data)
	return err
}

// Subscribe subscribes to a subject with a queue group (this enables load balancing) and a durable name (to maintain state)
func (c *NatsClient) Subscribe(subject, queue, durable string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.js.QueueSubscribe(subject, queue, handler, nats.Durable(durable), nats.ManualAck())
}

// Close closes the NATS connection
func (c *NatsClient) Close() {
	if c.nc != nil {
		err := c.nc.Drain() // gracefully close the connection
		if err != nil {
			c.nc.Close()
		}
	}
}

func createStreamConfig(name string, subjects []string) *nats.StreamConfig {
	return &nats.StreamConfig{
		Name:      name,
		Subjects:  subjects,
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour, // Keep for 7 days
		MaxBytes:  100 * 1024 * 1024,  // Or until 100 MB is reached
	}
}

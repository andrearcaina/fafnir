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

	log.Println("Successfully connected to NATS JetStream server")

	return &NatsClient{
		nc: nc,
		js: js,
	}, nil
}

// AddStream with the given name and subjects (creates a new stream if it doesn't exist already)
func (c *NatsClient) AddStream(name string, subjects []string) (*nats.StreamInfo, error) {
	streamConfig := &nats.StreamConfig{
		Name:      name,
		Subjects:  subjects,
		Storage:   nats.FileStorage,
		Retention: nats.LimitsPolicy,
		MaxAge:    7 * 24 * time.Hour, // keep for 7 days
		MaxBytes:  100 * 1024 * 1024,  // or until 100 MB is reached
	}

	streamInfo, err := c.js.AddStream(streamConfig)
	if err != nil {
		// ensures idempotency - if the stream already exists, return its info (no error)
		if err == nats.ErrStreamNameAlreadyInUse {
			streamInfo, _ := c.js.StreamInfo(name)
			log.Printf("Stream %s already exists", name)
			return streamInfo, nil
		}

		return nil, fmt.Errorf("could not create %s stream: %v", name, err)
	}

	log.Printf("Stream %s created successfully", name)

	return streamInfo, nil
}

// Publish a message to a subject
func (c *NatsClient) Publish(subject string, data []byte) (*nats.PubAck, error) {
	return c.js.Publish(subject, data)
}

// QueueSubscribeSubscribe to a subject with a queue group (this enables load balancing) and a durable name (to maintain state)
func (c *NatsClient) QueueSubscribe(subject, queue, durable string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.js.QueueSubscribe(subject, queue, handler, nats.Durable(durable), nats.ManualAck())
}

// Close the NATS connection
func (c *NatsClient) Close() {
	if c.nc != nil {
		err := c.nc.Drain() // gracefully close the connection
		if err != nil {
			c.nc.Close()
		}
	}
}

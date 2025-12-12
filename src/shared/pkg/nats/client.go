package nats

import (
	"errors"
	"log"

	"github.com/nats-io/nats.go"
)

type NatsClient struct {
	conn *nats.Conn
}

// New creates a new NatsClient and connects to the NATS server at the given URL
func New(url string) (*NatsClient, error) {
	conn, err := nats.Connect(url)
	if err != nil {
		return nil, errors.New("could not connect to NATs server at specified URL")
	}

	log.Println("Successfully connected to NATS server")
	return &NatsClient{conn: conn}, nil
}

// Publish a message to a subject
func (c *NatsClient) Publish(subject string, data []byte) error {
	return c.conn.Publish(subject, data)
}

// Subscribe to a subject with a message handler
func (c *NatsClient) Subscribe(subject string, handler nats.MsgHandler) (*nats.Subscription, error) {
	return c.conn.Subscribe(subject, handler)
}

func (c *NatsClient) Close() {
	if c.conn != nil {
		err := c.conn.Drain() // gracefully close the connection
		if err != nil {
			c.conn.Close()
		}
	}
}

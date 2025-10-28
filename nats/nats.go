package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
)

type NATSClient = nats.Conn

func NewNATSClient(natsURL string) (*NATSClient, error) {
	opts := []nats.Option{
		nats.Name("CollaborativeEditor-Sessions"),
		nats.Timeout(10 * time.Second),
		nats.ReconnectWait(2 * time.Second),
		nats.MaxReconnects(5),
	}

	conn, err := nats.Connect(natsURL, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect NATS: %w", err)
	}

	return (*NATSClient)(conn), nil
}

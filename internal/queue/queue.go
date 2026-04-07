package queue

import "context"

type Message struct {
	ID      string
	Payload []byte
}

type Queue interface {
	Enqueue(ctx context.Context, queue string, payload []byte) error
	Dequeue(ctx context.Context, queue string) (*Message, error)
	Acknowledge(ctx context.Context, queue string, id string) error
}

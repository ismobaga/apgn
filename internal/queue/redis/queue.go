package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"

	"github.com/ismobaga/apgn/internal/queue"
)

// RedisQueue implements queue.Queue using Redis LPUSH/BRPOP.
// Inflight messages are tracked in a separate processing list.
type RedisQueue struct {
	client *goredis.Client
}

func New(client *goredis.Client) *RedisQueue {
	return &RedisQueue{client: client}
}

func processingKey(q string) string {
	return q + ":processing"
}

func (r *RedisQueue) Enqueue(ctx context.Context, q string, payload []byte) error {
	return r.client.LPush(ctx, q, payload).Err()
}

// Dequeue blocks for up to 5 seconds waiting for a message.
// The message ID is a UUID that identifies it in the processing list.
func (r *RedisQueue) Dequeue(ctx context.Context, q string) (*queue.Message, error) {
	result, err := r.client.BRPop(ctx, 5*time.Second, q).Result()
	if err == goredis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("redis brpop: %w", err)
	}

	// result[0] = queue name, result[1] = value
	payload := []byte(result[1])
	id := uuid.New().String()

	// Store in processing list for acknowledgement tracking
	if err := r.client.HSet(ctx, processingKey(q), id, payload).Err(); err != nil {
		return nil, fmt.Errorf("redis hset processing: %w", err)
	}

	return &queue.Message{ID: id, Payload: payload}, nil
}

func (r *RedisQueue) Acknowledge(ctx context.Context, q string, id string) error {
	return r.client.HDel(ctx, processingKey(q), id).Err()
}

package queue

import (
	"context"
	"encoding/json"

	goredis "github.com/redis/go-redis/v9"
)

type RedisStreamPublisher struct {
	client *goredis.Client
	stream string
}

func NewRedisStreamPublisher(client *goredis.Client, stream string) *RedisStreamPublisher {
	return &RedisStreamPublisher{client: client, stream: stream}
}

func (r *RedisStreamPublisher) PublishSignal(event SignalEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}
	_, err = r.client.XAdd(context.Background(), &goredis.XAddArgs{
		Stream: r.stream,
		Values: map[string]any{
			"event": string(payload),
		},
	}).Result()
	return err
}

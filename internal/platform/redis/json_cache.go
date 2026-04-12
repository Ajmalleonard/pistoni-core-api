package redis

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type JSONCache struct {
	client *Connection
	ttl    time.Duration
}

func NewJSONCache(conn Connection, ttl time.Duration) *JSONCache {
	if ttl <= 0 {
		ttl = 2 * time.Minute
	}
	return &JSONCache{client: &conn, ttl: ttl}
}

func (c *JSONCache) Enabled() bool {
	return c != nil && c.client != nil && c.client.Enabled && c.client.Client != nil
}

func (c *JSONCache) Get(ctx context.Context, key string, target any) (bool, error) {
	if !c.Enabled() {
		return false, nil
	}
	raw, err := c.client.Client.Get(ctx, key).Bytes()
	if err != nil {
		if strings.Contains(err.Error(), "redis: nil") {
			return false, nil
		}
		return false, err
	}
	if err := json.Unmarshal(raw, target); err != nil {
		return false, err
	}
	return true, nil
}

func (c *JSONCache) Set(ctx context.Context, key string, value any) error {
	if !c.Enabled() {
		return nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Client.Set(ctx, key, raw, c.ttl).Err()
}

func (c *JSONCache) Delete(ctx context.Context, keys ...string) error {
	if !c.Enabled() || len(keys) == 0 {
		return nil
	}
	return c.client.Client.Del(ctx, keys...).Err()
}

func (c *JSONCache) DeletePrefix(ctx context.Context, prefix string) error {
	if !c.Enabled() || prefix == "" {
		return nil
	}

	var cursor uint64
	for {
		keys, next, err := c.client.Client.Scan(ctx, cursor, prefix+"*", 100).Result()
		if err != nil {
			return err
		}
		if len(keys) > 0 {
			if err := c.client.Client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
		}
		cursor = next
		if cursor == 0 {
			return nil
		}
	}
}

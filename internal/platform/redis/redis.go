package redis

import (
	"context"
	"fmt"
	"strings"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"pistoni/core/internal/config"
)

type Connection struct {
	Client   *goredis.Client
	Enabled  bool
	Reason   string
	Endpoint string
}

func Connect(cfg config.Config) Connection {
	if !hasRedisConfig(cfg) {
		return Connection{Enabled: false, Reason: "missing redis configuration"}
	}

	opt, endpoint, err := optionsFromConfig(cfg)
	if err != nil {
		return Connection{Enabled: false, Reason: err.Error()}
	}

	client := goredis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return Connection{Enabled: false, Reason: err.Error(), Endpoint: endpoint}
	}

	return Connection{Client: client, Enabled: true, Endpoint: endpoint}
}

func hasRedisConfig(cfg config.Config) bool {
	return strings.TrimSpace(cfg.RedisURL) != "" || strings.TrimSpace(cfg.RedisHost) != ""
}

func optionsFromConfig(cfg config.Config) (*goredis.Options, string, error) {
	if u := strings.TrimSpace(cfg.RedisURL); u != "" {
		opt, err := goredis.ParseURL(u)
		if err != nil {
			return nil, "", fmt.Errorf("invalid REDIS_URL: %w", err)
		}
		return opt, u, nil
	}

	host := strings.TrimSpace(cfg.RedisHost)
	if host == "" {
		return nil, "", fmt.Errorf("REDIS_HOST is required when REDIS_URL is empty")
	}

	port := cfg.RedisPort
	if port == 0 {
		port = 6379
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	return &goredis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}, addr, nil
}

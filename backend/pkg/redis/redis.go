package redis

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/monitoring-system/backend/config"
	redis "github.com/redis/go-redis/v9"
)

func NewRedisClient(cfg *config.Config) *redis.Client {
	addr := fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort)
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️ Redis ping failed (%s): %v. Redis features will operate in fallback mode if unavailable.", addr, err)
	} else {
		log.Printf("✅ Connected to Redis at %s", addr)
	}

	return client
}

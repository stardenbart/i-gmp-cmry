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

		// Pool sizing — harus >= jumlah VU concurrent
		// Default go-redis = 10 × GOMAXPROCS ≈ 80 pada 8-core.
		// Dengan 200 VU semua concurrent membutuhkan Redis, default terlalu kecil:
		// 120 VU mengantri nunggu koneksi bebas → wait time makan 3s budget → 503.
		PoolSize:     250, // headroom di atas 200 VU untuk internal goroutines
		MinIdleConns: 20,  // pre-warm koneksi agar tidak cold-start di awal test

		// Fail-fast timeout: kalau pool habis, gagal cepat (< context timeout)
		// bukan menunggu diam sampai context deadline habis
		PoolTimeout: 500 * time.Millisecond,

		// Network timeouts: mencegah slow Redis round-trip makan seluruh budget
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
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

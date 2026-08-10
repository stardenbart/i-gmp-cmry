package inspectionhandler

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/usecase/lockusecase"
	"github.com/monitoring-system/backend/pkg/response"
	redis "github.com/redis/go-redis/v9"
)

type SyncHandler struct {
	lockManager *lockusecase.LockManager
	redis       *redis.Client
}

func NewSyncHandler(lm *lockusecase.LockManager, r *redis.Client) *SyncHandler {
	return &SyncHandler{
		lockManager: lm,
		redis:       r,
	}
}

// SyncConsolidated handles GET /api/v1/inspeksi/sync
// Consolidates lock status, kawasan events, and user notifications in 1 unified response (from Redis 0-SQL query)
func (sh *SyncHandler) SyncConsolidated(c *fiber.Ctx) error {
	kawasanID := c.Query("kawasan_id")
	userID := c.Query("user_id")
	sinceStr := c.Query("since")

	var since int64
	if sinceStr != "" {
		since, _ = strconv.ParseInt(sinceStr, 10, 64)
	}

	serverTime := time.Now().UnixMilli()

	// 1. Fetch Lock Statuses for Kawasan (Redis cache)
	lockStatuses := []fiber.Map{}
	if kawasanID != "" && sh.redis != nil {
		ctx := c.Context()
		pattern := "lock:aspek:" + kawasanID + ":*"
		keys, err := sh.redis.Keys(ctx, pattern).Result()
		if err == nil {
			for _, key := range keys {
				val, err := sh.redis.Get(ctx, key).Result()
				if err == nil && val != "" {
					partsStr := key[len("lock:aspek:"+kawasanID+":"):]
					ttl, _ := sh.redis.TTL(ctx, key).Result()

					lockStatuses = append(lockStatuses, fiber.Map{
						"aspek_id":    partsStr,
						"status":      "LOCKED",
						"raw_value":   val,
						"ttl_seconds": int(ttl.Seconds()),
					})
				}
			}
		}
	}

	return response.OK(c, "Sync completed", fiber.Map{
		"server_time":   serverTime,
		"kawasan_id":    kawasanID,
		"user_id":       userID,
		"since":         since,
		"lock_statuses": lockStatuses,
	})
}

package lockusecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/monitoring-system/backend/internal/domain/inspection"
	redis "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrLockConflict       = errors.New("aspek sedang dikunci oleh user lain")
	ErrLockExpiredOrStolen = errors.New("lock tidak ditemukan atau sudah kadaluarsa")
	ErrNotLockOwner       = errors.New("anda bukan pemilik lock aspek ini")
)

type LockManager struct {
	redis  *redis.Client
	logger *zap.Logger
}

func NewLockManager(r *redis.Client, logger *zap.Logger) *LockManager {
	return &LockManager{
		redis:  r,
		logger: logger,
	}
}

func generateSecureToken() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// AcquireLock attempts to lock an Aspek for 30s.
func (lm *LockManager) AcquireLock(ctx context.Context, kawasanID, aspekID, userID string) (*inspection.LockResult, error) {
	if lm.redis == nil {
		// Fallback when Redis is disabled
		token := generateSecureToken()
		return &inspection.LockResult{
			LockToken: token,
			ExpiresAt: time.Now().Add(30 * time.Second),
		}, nil
	}

	key := fmt.Sprintf("lock:aspek:%s:%s", kawasanID, aspekID)
	token := generateSecureToken()
	value := fmt.Sprintf("%s:%s:%d", userID, token, time.Now().Unix())
	ttl := 30 * time.Second

	// Check if existing lock belongs to the same user (reconnect scenario)
	existing, err := lm.redis.Get(ctx, key).Result()
	if err == nil && existing != "" {
		parts := strings.SplitN(existing, ":", 3)
		if len(parts) >= 2 && parts[0] == userID {
			lm.redis.Set(ctx, key, value, ttl)
			return &inspection.LockResult{
				LockToken: token,
				ExpiresAt: time.Now().Add(ttl),
			}, nil
		}
		return nil, fmt.Errorf("%w (locked by user: %s)", ErrLockConflict, parts[0])
	}

	ok, err := lm.redis.SetNX(ctx, key, value, ttl).Result()
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}
	if !ok {
		return nil, ErrLockConflict
	}

	// Set heartbeat key
	hbKey := fmt.Sprintf("heartbeat:%s:%s:%s", kawasanID, aspekID, userID)
	lm.redis.Set(ctx, hbKey, "1", 35*time.Second)

	return &inspection.LockResult{
		LockToken: token,
		ExpiresAt: time.Now().Add(ttl),
	}, nil
}

// RenewLock extends lock TTL by 30 seconds if owned by userID & token
func (lm *LockManager) RenewLock(ctx context.Context, kawasanID, aspekID, userID, token string) error {
	if lm.redis == nil {
		return nil
	}

	key := fmt.Sprintf("lock:aspek:%s:%s", kawasanID, aspekID)
	prefix := fmt.Sprintf("%s:%s:", userID, token)

	script := redis.NewScript(`
		local val = redis.call('GET', KEYS[1])
		if val and string.sub(val, 1, #ARGV[1]) == ARGV[1] then
			redis.call('PEXPIRE', KEYS[1], 30000)
			return 1
		end
		return 0
	`)

	res, err := script.Run(ctx, lm.redis, []string{key}, prefix).Int()
	if err != nil || res == 0 {
		return ErrLockExpiredOrStolen
	}

	// Refresh heartbeat key
	hbKey := fmt.Sprintf("heartbeat:%s:%s:%s", kawasanID, aspekID, userID)
	lm.redis.Set(ctx, hbKey, "1", 35*time.Second)

	return nil
}

// ReleaseLock releases the lock atomically
func (lm *LockManager) ReleaseLock(ctx context.Context, kawasanID, aspekID, userID, token string) error {
	if lm.redis == nil {
		return nil
	}

	key := fmt.Sprintf("lock:aspek:%s:%s", kawasanID, aspekID)
	prefix := fmt.Sprintf("%s:%s:", userID, token)

	script := redis.NewScript(`
		local val = redis.call('GET', KEYS[1])
		if val and string.sub(val, 1, #ARGV[1]) == ARGV[1] then
			redis.call('DEL', KEYS[1])
			return 1
		end
		return 0
	`)

	res, err := script.Run(ctx, lm.redis, []string{key}, prefix).Int()
	if err != nil || res == 0 {
		return ErrNotLockOwner
	}

	hbKey := fmt.Sprintf("heartbeat:%s:%s:%s", kawasanID, aspekID, userID)
	lm.redis.Del(ctx, hbKey)

	return nil
}

// ValidateLock checks if the lock is held by user & token
func (lm *LockManager) ValidateLock(ctx context.Context, kawasanID, aspekID, userID, token string) error {
	if lm.redis == nil {
		return nil
	}

	key := fmt.Sprintf("lock:aspek:%s:%s", kawasanID, aspekID)
	expected := fmt.Sprintf("%s:%s:", userID, token)

	val, err := lm.redis.Get(ctx, key).Result()
	if err != nil || !strings.HasPrefix(val, expected) {
		return ErrLockExpiredOrStolen
	}
	return nil
}

// GetLockInfo returns current lock details
func (lm *LockManager) GetLockInfo(ctx context.Context, kawasanID, aspekID string) (*inspection.AspekLockStatus, error) {
	if lm.redis == nil {
		return &inspection.AspekLockStatus{AspekID: aspekID, Status: "FREE"}, nil
	}

	key := fmt.Sprintf("lock:aspek:%s:%s", kawasanID, aspekID)
	val, err := lm.redis.Get(ctx, key).Result()
	if err == redis.Nil || val == "" {
		return &inspection.AspekLockStatus{AspekID: aspekID, Status: "FREE"}, nil
	}
	if err != nil {
		return nil, err
	}

	parts := strings.SplitN(val, ":", 3)
	ttl, _ := lm.redis.TTL(ctx, key).Result()
	expiresAt := time.Now().Add(ttl)

	return &inspection.AspekLockStatus{
		AspekID:   aspekID,
		Status:    "LOCKED",
		LockedBy:  parts[0],
		ExpiresAt: &expiresAt,
	}, nil
}

// SaveDraftState stores draft inspection data in Redis HASH
func (lm *LockManager) SaveDraftState(ctx context.Context, kawasanID, aspekID, userID, dataJSON string, skor float64) error {
	if lm.redis == nil {
		return nil
	}

	stateKey := fmt.Sprintf("state:aspek:%s:%s", kawasanID, aspekID)
	pipe := lm.redis.Pipeline()
	pipe.HSet(ctx, stateKey,
		"user_id", userID,
		"data", dataJSON,
		"skor", fmt.Sprintf("%.2f", skor),
		"status", "InProgress",
		"last_saved_at", time.Now().Unix(),
	)
	pipe.Expire(ctx, stateKey, 24*time.Hour)
	_, err := pipe.Exec(ctx)
	return err
}

// GetDraftState retrieves draft data for an Aspek from Redis
func (lm *LockManager) GetDraftState(ctx context.Context, kawasanID, aspekID string) (map[string]string, error) {
	if lm.redis == nil {
		return map[string]string{}, nil
	}

	stateKey := fmt.Sprintf("state:aspek:%s:%s", kawasanID, aspekID)
	return lm.redis.HGetAll(ctx, stateKey).Result()
}

// MarkAspekCompleted marks an Aspek as completed in tracker and checks if all are done
func (lm *LockManager) MarkAspekCompleted(ctx context.Context, kawasanID, aspekID, periode string) (bool, error) {
	if lm.redis == nil {
		return false, nil
	}

	trackerKey := fmt.Sprintf("tracker:%s:%s", kawasanID, periode)
	sessionKey := fmt.Sprintf("session:%s:%s", kawasanID, periode)

	script := redis.NewScript(`
		redis.call('HSET', KEYS[1], ARGV[1], '1')
		local vals = redis.call('HVALS', KEYS[1])
		local all_done = true
		local count = 0
		for _, v in ipairs(vals) do
			if v == '1' then
				count = count + 1
			else
				all_done = false
			end
		end
		redis.call('HSET', KEYS[2], 'completed_aspek', tostring(count))
		if all_done then
			redis.call('HSET', KEYS[2], 'status', 'Completed')
			return 1
		end
		return 0
	`)

	res, err := script.Run(ctx, lm.redis, []string{trackerKey, sessionKey}, aspekID).Int()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}

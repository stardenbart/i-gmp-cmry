package idgen

import (
	"crypto/rand"
	"fmt"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
)

// counter is a naive in-memory counter for sequential ID generation within the same second.
// For production with multiple instances, replace with a DB sequence or Redis counter.
var (
	mu      sync.Mutex
	counter = make(map[string]int)
)

// Generate creates a formatted ID: PREFIX-YYYYMMDD-SEQ (e.g., DEPT-20240101-001).
// seq is auto-incremented per prefix per application run.
func Generate(prefix string) string {
	return GenerateRandom(prefix)
}

// GenerateRandom creates a formatted ID with random suffix for high-throughput scenarios
// where sequential IDs may collide (e.g., activity logs). Format: PREFIX-YYYYMMDD-RANDOM
func GenerateRandom(prefix string) string {
	mu.Lock()
	defer mu.Unlock()

	// Use YYMMDD (6 chars) instead of YYYYMMDD (8 chars) to ensure
	// the total ID length remains <= 20 characters (e.g., ALOG-260717-e4163cc9)
	today := time.Now().Format("060102")
	randomBytes := make([]byte, 4)
	rand.Read(randomBytes)
	randomHex := fmt.Sprintf("%x", randomBytes)

	return fmt.Sprintf("%s-%s-%s", prefix, today, randomHex)
}

// GenerateSequential queries DB for the highest existing ID with prefix and returns next padded ID.
// Example: PREFIX="DEPT", digits=3 -> DEPT-008
func GenerateSequential(db *gorm.DB, tableName, columnName, prefix string, digits int) string {
	mu.Lock()
	defer mu.Unlock()

	var lastID string
	// Find latest ID matching prefix pattern
	err := db.Table(tableName).
		Select(fmt.Sprintf("\"%s\"", columnName)).
		Where(fmt.Sprintf("\"%s\" LIKE ?", columnName), prefix+"-%").
		Order(fmt.Sprintf("LENGTH(\"%s\") DESC, \"%s\" DESC", columnName, columnName)).
		Limit(1).
		Scan(&lastID).Error

	nextNum := 1
	if err == nil && lastID != "" {
		parts := strings.Split(lastID, "-")
		if len(parts) > 0 {
			numStr := parts[len(parts)-1]
			var n int
			if _, parseErr := fmt.Sscanf(numStr, "%d", &n); parseErr == nil && n > 0 {
				nextNum = n + 1
			}
		}
	}

	if nextNum == 1 {
		var count int64
		db.Table(tableName).Count(&count)
		if count > 0 {
			nextNum = int(count) + 1
		}
	}

	formatStr := fmt.Sprintf("%%s-%%0%dd", digits)
	return fmt.Sprintf(formatStr, prefix, nextNum)
}

// Predefined prefix constants — aligned with ERD table names.
const (
	PrefixDepartment    = "DEPT"
	PrefixArea          = "AREA"
	PrefixKawasan       = "KWS"
	PrefixDetailKawasan = "DKWS"
	PrefixUser          = "USR"
	PrefixRole          = "ROLE"
	PrefixModule        = "MOD"
	PrefixPermission    = "PERM"
	PrefixRolePerm      = "RP"
	PrefixPICMap        = "PIC"
	PrefixAspek         = "ASPK"
	PrefixDetail        = "DTL"
	PrefixUraian        = "URN"
	PrefixInspection    = "INSP"
	PrefixResult        = "RES"
	PrefixIssue         = "ISS"
	PrefixIssuePhoto    = "ISPH"
	PrefixLoginLog      = "LLOG"
	PrefixActivityLog   = "ALOG"
	PrefixSession       = "SES"
	PrefixRefreshToken  = "RTK"
)

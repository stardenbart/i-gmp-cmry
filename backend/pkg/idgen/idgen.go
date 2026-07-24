package idgen

import (
	"crypto/rand"
	"fmt"
	"sync"
	"time"
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
)

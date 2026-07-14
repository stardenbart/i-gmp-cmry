package idgen

import (
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
	mu.Lock()
	defer mu.Unlock()

	today := time.Now().Format("20060102")
	key := prefix + today

	counter[key]++
	seq := counter[key]

	return fmt.Sprintf("%s-%s-%03d", prefix, today, seq)
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
)

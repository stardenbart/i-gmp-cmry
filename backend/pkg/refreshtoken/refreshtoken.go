// Package refreshtoken issues and hashes opaque refresh/CSRF tokens.
//
// Refresh tokens are intentionally NOT JWTs: a JWT can't be revoked before
// its own expiry without a denylist, which defeats the point of a short
// access-token TTL. An opaque random token whose SHA-256 hash is stored in
// the database can be revoked (or its whole rotation family killed on reuse
// detection) with a single row update — see authusecase.Refresh.
package refreshtoken

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

// Generate creates a new 256-bit random token. It returns the raw value (sent
// to the client exactly once, inside an httpOnly cookie) and its SHA-256 hex
// digest (the only form ever persisted — a leaked database row can't be
// replayed as a cookie value).
func Generate() (raw string, hash string) {
	raw = randomHex(32)
	return raw, Hash(raw)
}

// NewCSRFToken creates a random token for the double-submit CSRF cookie. It
// doesn't need to be hashed/stored server-side — its only job is to be a
// value an attacker's cross-site page cannot read (see middleware.CSRFMiddleware).
func NewCSRFToken() string {
	return randomHex(32)
}

// Hash returns the SHA-256 hex digest of a raw token, for comparison against
// stored hashes.
func Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func randomHex(numBytes int) string {
	b := make([]byte, numBytes)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read only fails if the OS entropy source is
		// unavailable, which is unrecoverable for anything security
		// sensitive — fail loudly rather than hand out a predictable token.
		panic("refreshtoken: failed to read random bytes: " + err.Error())
	}
	return hex.EncodeToString(b)
}

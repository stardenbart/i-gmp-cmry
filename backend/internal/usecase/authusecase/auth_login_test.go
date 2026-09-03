package authusecase

import (
	"testing"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/password"
	"github.com/monitoring-system/backend/pkg/refreshtoken"
)

type loginTestLogRepo struct {
	entries []logdomain.LoginLog
}

func (r *loginTestLogRepo) FindAll(int, int, string, string) ([]logdomain.LoginLog, int64, error) {
	return r.entries, int64(len(r.entries)), nil
}
func (r *loginTestLogRepo) FindByID(string) (*logdomain.LoginLog, error) { return nil, nil }
func (r *loginTestLogRepo) Create(entry *logdomain.LoginLog) error {
	r.entries = append(r.entries, *entry)
	return nil
}
func (r *loginTestLogRepo) UpdateLogoutAt(string, time.Time) error { return nil }

// fakeRefreshTokenRepo is a minimal in-memory stand-in good enough to
// exercise rotation and reuse detection without a real database.
type fakeRefreshTokenRepo struct {
	byHash map[string]*authdomain.RefreshToken
}

func newFakeRefreshTokenRepo() *fakeRefreshTokenRepo {
	return &fakeRefreshTokenRepo{byHash: make(map[string]*authdomain.RefreshToken)}
}

func (r *fakeRefreshTokenRepo) Create(rt *authdomain.RefreshToken) error {
	cp := *rt
	r.byHash[rt.TokenHash] = &cp
	return nil
}

func (r *fakeRefreshTokenRepo) FindByHash(hash string) (*authdomain.RefreshToken, error) {
	rt, ok := r.byHash[hash]
	if !ok {
		return nil, nil
	}
	cp := *rt
	return &cp, nil
}

func (r *fakeRefreshTokenRepo) RevokeByID(id string, replacedBy *string) error {
	for _, rt := range r.byHash {
		if rt.ID == id {
			now := time.Now()
			rt.RevokedAt = &now
			rt.ReplacedBy = replacedBy
		}
	}
	return nil
}

func (r *fakeRefreshTokenRepo) RevokeFamily(familyID string) error {
	now := time.Now()
	for _, rt := range r.byHash {
		if rt.FamilyID == familyID && rt.RevokedAt == nil {
			rt.RevokedAt = &now
		}
	}
	return nil
}

func (r *fakeRefreshTokenRepo) RevokeAllForUser(userID string) error {
	now := time.Now()
	for _, rt := range r.byHash {
		if rt.UserID == userID && rt.RevokedAt == nil {
			rt.RevokedAt = &now
		}
	}
	return nil
}

func newTestUseCase(userRepo authdomain.UserRepository, logs logdomain.LoginLogRepository, refreshRepo authdomain.RefreshTokenRepository) authdomain.AuthUseCase {
	return NewAuthUseCase(userRepo, logs, refreshRepo, jwt.New("test-secret", time.Hour), 7*24*time.Hour)
}

func TestLoginAcceptsUsernameOrEmail(t *testing.T) {
	hash, err := password.Hash("SecurePass123")
	if err != nil {
		t.Fatal(err)
	}
	user := &authdomain.User{
		UserID:       "USR-1",
		Username:     "auditor",
		Email:        "auditor@example.com",
		PasswordHash: hash,
		RoleID:       "ROLE-1",
		UserStatus:   authdomain.UserStatusActive,
	}

	for _, identifier := range []string{"auditor", "AUDITOR@EXAMPLE.COM"} {
		t.Run(identifier, func(t *testing.T) {
			logs := &loginTestLogRepo{}
			uc := newTestUseCase(&resetTestUserRepo{user: user}, logs, newFakeRefreshTokenRepo())
			result, err := uc.Login(&authdomain.LoginRequest{
				Username: identifier,
				Password: "SecurePass123",
			}, "127.0.0.1", "test")
			if err != nil {
				t.Fatalf("Login() error = %v", err)
			}
			if result.User.UserID != user.UserID || result.AccessToken == "" || result.RefreshToken == "" || result.CSRFToken == "" {
				t.Fatal("login did not return the expected user, access token, refresh token, and CSRF token")
			}
			if len(logs.entries) != 1 || logs.entries[0].LoginStatus != logdomain.LoginStatusSuccess {
				t.Fatal("successful login was not recorded")
			}
		})
	}
}

func TestRefreshRotatesToken(t *testing.T) {
	hash, err := password.Hash("SecurePass123")
	if err != nil {
		t.Fatal(err)
	}
	user := &authdomain.User{
		UserID:       "USR-1",
		Username:     "auditor",
		Email:        "auditor@example.com",
		PasswordHash: hash,
		RoleID:       "ROLE-1",
		UserStatus:   authdomain.UserStatusActive,
	}
	refreshRepo := newFakeRefreshTokenRepo()
	uc := newTestUseCase(&resetTestUserRepo{user: user}, &loginTestLogRepo{}, refreshRepo)

	login, err := uc.Login(&authdomain.LoginRequest{Username: "auditor", Password: "SecurePass123"}, "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	refreshed, err := uc.Refresh(login.RefreshToken, "127.0.0.1", "test")
	if err != nil {
		t.Fatalf("Refresh() error = %v", err)
	}
	if refreshed.RefreshToken == login.RefreshToken {
		t.Fatal("Refresh() must issue a brand-new refresh token, not reuse the old one")
	}
	if refreshed.User.UserID != user.UserID {
		t.Fatal("Refresh() returned the wrong user")
	}

	// The original (now-rotated) token must no longer work.
	if _, err := uc.Refresh(login.RefreshToken, "127.0.0.1", "test"); err == nil {
		t.Fatal("a rotated-away refresh token must be rejected")
	}

	// Reuse of the already-rotated token must revoke the WHOLE family,
	// including the token that replaced it — the standard reuse-detection
	// response to a suspected stolen token.
	if _, err := uc.Refresh(refreshed.RefreshToken, "127.0.0.1", "test"); err == nil {
		t.Fatal("refresh token family must be revoked after reuse of a rotated token")
	}
}

func TestRefreshRejectsUnknownToken(t *testing.T) {
	uc := newTestUseCase(&resetTestUserRepo{}, &loginTestLogRepo{}, newFakeRefreshTokenRepo())
	if _, err := uc.Refresh(refreshtoken.NewCSRFToken(), "127.0.0.1", "test"); err == nil {
		t.Fatal("an unknown refresh token must be rejected")
	}
}

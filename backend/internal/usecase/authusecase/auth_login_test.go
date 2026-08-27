package authusecase

import (
	"testing"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/password"
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
			uc := NewAuthUseCase(&resetTestUserRepo{user: user}, logs, jwt.New("test-secret", 1))
			result, err := uc.Login(&authdomain.LoginRequest{
				Username: identifier,
				Password: "SecurePass123",
			}, "127.0.0.1", "test")
			if err != nil {
				t.Fatalf("Login() error = %v", err)
			}
			if result.User.UserID != user.UserID || result.Token == "" {
				t.Fatal("login did not return the expected user and token")
			}
			if len(logs.entries) != 1 || logs.entries[0].LoginStatus != logdomain.LoginStatusSuccess {
				t.Fatal("successful login was not recorded")
			}
		})
	}
}

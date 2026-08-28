package authusecase

import (
	"errors"
	"strings"
	"testing"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/password"
)

type resetTestUserRepo struct {
	user        *authdomain.User
	updateCount int
}

func (r *resetTestUserRepo) FindAll(int, int, string, string, string, string) ([]authdomain.User, int64, error) {
	return nil, 0, nil
}
func (r *resetTestUserRepo) FindByID(string) (*authdomain.User, error) { return r.user, nil }
func (r *resetTestUserRepo) FindByUsername(username string) (*authdomain.User, error) {
	if r.user != nil && username == r.user.Username {
		return r.user, nil
	}
	return nil, errors.New("not found")
}
func (r *resetTestUserRepo) FindByEmail(email string) (*authdomain.User, error) {
	if r.user != nil && strings.EqualFold(email, r.user.Email) {
		return r.user, nil
	}
	return nil, authdomain.ErrEmailNotRegistered
}
func (r *resetTestUserRepo) Create(*authdomain.User) error { return nil }
func (r *resetTestUserRepo) Update(user *authdomain.User) error {
	r.user = user
	r.updateCount++
	return nil
}
func (r *resetTestUserRepo) Delete(string) error { return nil }

type resetTestOTPRepo struct {
	hash          string
	expiresAt     time.Time
	consumeResult bool
	consumeCount  int
}

func (r *resetTestOTPRepo) ReplaceForUser(_ string, hash string, expiresAt time.Time) error {
	r.hash = hash
	r.expiresAt = expiresAt
	return nil
}
func (r *resetTestOTPRepo) ConsumeValid(_ string, _ string, _ time.Time, _ int) (bool, error) {
	r.consumeCount++
	return r.consumeResult, nil
}

type resetTestSettingRepo struct {
	template string
}

func (resetTestSettingRepo) FindAll(string) ([]masterdomain.Setting, error) { return nil, nil }

func (r resetTestSettingRepo) FindByKey(string, string) (*masterdomain.Setting, error) {
	if r.template != "" {
		return &masterdomain.Setting{SettingValue: r.template}, nil
	}
	return nil, errors.New("not found")
}
func (resetTestSettingRepo) Update(string, string, string, string) error { return nil }

type resetTestMailer struct {
	data     map[string]string
	template string
}

func (m *resetTestMailer) Send([]string, string, string) error                      { return nil }
func (m *resetTestMailer) SendTemplate([]string, string, string, interface{}) error { return nil }
func (m *resetTestMailer) SendForPlant(string, []string, string, string) error      { return nil }
func (m *resetTestMailer) SendTemplateForPlant(_ string, _ []string, _ string, template string, data interface{}) error {
	m.data = data.(map[string]string)
	m.template = template
	return nil
}

func TestForgotPasswordCreatesOTPWithoutChangingPassword(t *testing.T) {
	originalHash, err := password.Hash("OriginalPass123")
	if err != nil {
		t.Fatal(err)
	}
	users := &resetTestUserRepo{user: &authdomain.User{UserID: "USR-1", Email: "user@example.com", PasswordHash: originalHash}}
	otps := &resetTestOTPRepo{}
	mailer := &resetTestMailer{}
	uc := NewUserUseCase(users, mailer, resetTestSettingRepo{}, nil, otps)

	before := time.Now()
	if err := uc.ForgotPassword(&authdomain.ForgotPasswordRequest{Email: "user@example.com"}); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if users.updateCount != 0 || users.user.PasswordHash != originalHash {
		t.Fatal("requesting an OTP must not change the current password")
	}
	if mailer.data["OTP"] == "" || len(mailer.data["OTP"]) != passwordResetOTPLength {
		t.Fatalf("expected a %d-digit OTP in email data", passwordResetOTPLength)
	}
	if !password.Compare(mailer.data["OTP"], otps.hash) {
		t.Fatal("stored OTP hash does not match the emailed OTP")
	}
	if otps.expiresAt.Before(before.Add(passwordResetOTPExpiry - time.Second)) {
		t.Fatal("OTP expiry is shorter than expected")
	}
}

func TestForgotPasswordFallsBackFromTemplateWithoutRequiredOTPVariables(t *testing.T) {
	users := &resetTestUserRepo{user: &authdomain.User{UserID: "USR-1", Email: "user@example.com"}}
	mailer := &resetTestMailer{}
	uc := NewUserUseCase(
		users,
		mailer,
		resetTestSettingRepo{template: "<p>Template lama tanpa kode reset</p>"},
		nil,
		&resetTestOTPRepo{},
	)

	if err := uc.ForgotPassword(&authdomain.ForgotPasswordRequest{Email: "user@example.com"}); err != nil {
		t.Fatalf("ForgotPassword() error = %v", err)
	}
	if mailer.template != mail.TmplForgotPassword {
		t.Fatal("invalid configured template must fall back to the built-in OTP template")
	}
}

func TestForgotPasswordRejectsUnregisteredEmail(t *testing.T) {
	users := &resetTestUserRepo{}
	otps := &resetTestOTPRepo{}
	mailer := &resetTestMailer{}
	uc := NewUserUseCase(users, mailer, resetTestSettingRepo{}, nil, otps)

	err := uc.ForgotPassword(&authdomain.ForgotPasswordRequest{Email: "unknown@example.com"})
	if !errors.Is(err, authdomain.ErrEmailNotRegistered) {
		t.Fatalf("ForgotPassword() error = %v, want ErrEmailNotRegistered", err)
	}
	if otps.hash != "" || mailer.data != nil {
		t.Fatal("unregistered email must not create or send an OTP")
	}
}

func TestResetPasswordWithOTPChangesPasswordAfterValidOTP(t *testing.T) {
	originalHash, _ := password.Hash("OriginalPass123")
	users := &resetTestUserRepo{user: &authdomain.User{UserID: "USR-1", Email: "user@example.com", PasswordHash: originalHash}}
	otps := &resetTestOTPRepo{consumeResult: true}
	uc := NewUserUseCase(users, &resetTestMailer{}, resetTestSettingRepo{}, nil, otps)

	err := uc.ResetPasswordWithOTP(&authdomain.ResetPasswordWithOTPRequest{
		Email:           "user@example.com",
		OTP:             "123456",
		NewPassword:     "NewPassword123",
		ConfirmPassword: "NewPassword123",
	})
	if err != nil {
		t.Fatalf("ResetPasswordWithOTP() error = %v", err)
	}
	if otps.consumeCount != 1 || users.updateCount != 1 {
		t.Fatal("valid OTP must be consumed exactly once and update the user")
	}
	if !password.Compare("NewPassword123", users.user.PasswordHash) {
		t.Fatal("new password was not stored")
	}
}

func TestResetPasswordWithOTPRejectsMismatchedConfirmation(t *testing.T) {
	users := &resetTestUserRepo{user: &authdomain.User{UserID: "USR-1", Email: "user@example.com"}}
	otps := &resetTestOTPRepo{consumeResult: true}
	uc := NewUserUseCase(users, &resetTestMailer{}, resetTestSettingRepo{}, nil, otps)

	err := uc.ResetPasswordWithOTP(&authdomain.ResetPasswordWithOTPRequest{
		Email:           "user@example.com",
		OTP:             "123456",
		NewPassword:     "NewPassword123",
		ConfirmPassword: "Different123",
	})
	if err == nil {
		t.Fatal("expected mismatched confirmation to be rejected")
	}
	if otps.consumeCount != 0 || users.updateCount != 0 {
		t.Fatal("mismatched password must not consume OTP or update user")
	}
}

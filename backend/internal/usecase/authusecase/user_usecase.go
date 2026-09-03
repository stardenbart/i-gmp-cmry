package authusecase

import (
	"crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	masterdomain "github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/internal/domain/pic"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/mail"
	"github.com/monitoring-system/backend/pkg/password"
)

type userUseCase struct {
	userRepo    authdomain.UserRepository
	mailer      mail.Mailer
	settingRepo masterdomain.SettingRepository
	picRepo     pic.PICMappingRepository
	otpRepo     authdomain.PasswordResetOTPRepository
	refreshRepo authdomain.RefreshTokenRepository
}

const (
	passwordResetOTPLength      = 6
	passwordResetOTPExpiry      = 10 * time.Minute
	passwordResetOTPMaxAttempts = 5
)

// NewUserUseCase creates a new UserUseCase implementation. refreshRepo may be
// nil (e.g. in tests that don't exercise a password-changing path) — every
// call site below treats a nil refreshRepo as "skip revocation" rather than
// panicking, since revocation is defense-in-depth on top of the password
// change itself, not the primary effect.
func NewUserUseCase(userRepo authdomain.UserRepository, mailer mail.Mailer, settingRepo masterdomain.SettingRepository, picRepo pic.PICMappingRepository, otpRepo authdomain.PasswordResetOTPRepository, refreshRepo authdomain.RefreshTokenRepository) authdomain.UserUseCase {
	return &userUseCase{userRepo: userRepo, mailer: mailer, settingRepo: settingRepo, picRepo: picRepo, otpRepo: otpRepo, refreshRepo: refreshRepo}
}

// revokeSessions best-effort revokes every refresh token for a user. Called
// whenever a password changes (self-service change, admin reset, or OTP
// reset) so a session/refresh token stolen before the change doesn't stay
// valid for the rest of its TTL (up to 7 days) after the legitimate owner
// has changed their credentials specifically to shut that access out.
func (uc *userUseCase) revokeSessions(userID string) {
	if uc.refreshRepo == nil {
		return
	}
	_ = uc.refreshRepo.RevokeAllForUser(userID)
}

func (uc *userUseCase) GetAll(page, limit int, search, roleID, deptID, plantID string) ([]authdomain.User, int64, error) {
	return uc.userRepo.FindAll(page, limit, search, roleID, deptID, plantID)
}

func (uc *userUseCase) GetByID(id string) (*authdomain.User, error) {
	return uc.userRepo.FindByID(id)
}

func (uc *userUseCase) Create(req *authdomain.CreateUserRequest) (*authdomain.User, error) {
	// Check username uniqueness
	existing, _ := uc.userRepo.FindByUsername(req.Username)
	if existing != nil {
		return nil, errors.New("username already exists")
	}

	hashed, err := password.Hash(req.Password)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &authdomain.User{
		UserID:       idgen.Generate(idgen.PrefixUser),
		DepartmentID: req.DepartmentID,
		RoleID:       req.RoleID,
		PlantID:      req.PlantID,
		Username:     req.Username,
		FullName:     req.FullName,
		Email:        req.Email,
		PasswordHash: hashed,
		UserStatus:   authdomain.UserStatusActive,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Create PIC mappings if any kawasan was selected.
	// KategoriPIC is optional; admin can map a kawasan without a category.
	if len(req.PICKawasanIDs) > 0 && uc.picRepo != nil {
		_ = uc.upsertPICMappings(user.UserID, req.PICKawasanIDs, req.PICKategori)
	}

	return user, nil
}

func (uc *userUseCase) Update(id string, req *authdomain.UpdateUserRequest) (*authdomain.User, error) {
	user, err := uc.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if req.DepartmentID != "" {
		user.DepartmentID = req.DepartmentID
	}
	if req.RoleID != "" {
		user.RoleID = req.RoleID
	}
	if req.PlantID != nil {
		user.PlantID = req.PlantID
	}
	if req.FullName != "" {
		user.FullName = req.FullName
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.UserStatus != "" {
		user.UserStatus = req.UserStatus
	}
	if err := uc.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Update PIC mappings.
	// Two independent signals from the request:
	//   1. PICKawasanIDs (slice, can be nil if not sent) → source of truth for kawasan scope.
	//      When the slice is present (even if empty), we wipe the old mappings and
	//      rebuild from this list. An empty slice means "remove all mappings".
	//   2. PICKategori (*string, nil if not sent) → only applied in-place when
	//      PICKawasanIDs is NOT sent, so editing other fields (name/email/status)
	//      alone never disturbs the kawasan mapping.
	if uc.picRepo != nil {
		switch {
		case req.PICKawasanIDs != nil:
			kategori := ""
			if req.PICKategori != nil {
				kategori = *req.PICKategori
			}
			_ = uc.upsertPICMappings(id, req.PICKawasanIDs, kategori)
		case req.PICKategori != nil:
			// Only kategori changed; update existing rows in-place.
			if existing, err := uc.picRepo.FindByUserID(id); err == nil {
				for i := range existing {
					m := existing[i]
					if m.KategoriPIC == *req.PICKategori {
						continue
					}
					m.KategoriPIC = *req.PICKategori
					_ = uc.picRepo.Update(&m)
				}
			}
		}
	}

	return user, nil
}

// upsertPICMappings replaces all PIC_Mapping rows for the given user with
// one row per kawasanID. Pass an empty slice to remove all mappings.
// KategoriPIC may be empty when the admin did not choose a category.
func (uc *userUseCase) upsertPICMappings(userID string, kawasanIDs []string, kategori string) error {
	if uc.picRepo == nil {
		return nil
	}

	// Wipe existing mappings for this user.
	if existing, err := uc.picRepo.FindByUserID(userID); err == nil {
		for _, m := range existing {
			_ = uc.picRepo.Delete(m.PICMapID)
		}
	}

	// Recreate from the source-of-truth slice (may be empty → no rows).
	for _, kawasanID := range kawasanIDs {
		if kawasanID == "" {
			continue
		}
		picMap := &pic.PICMapping{
			PICMapID:    idgen.Generate(idgen.PrefixPICMap),
			AreaID:      nil, // Deprecated: kawasan is the join key used by issue scoping.
			KawasanID:   kawasanID,
			UserID:      userID,
			KategoriPIC: kategori,
		}
		_ = uc.picRepo.Create(picMap)
	}
	return nil
}

func (uc *userUseCase) Delete(id string) error {
	_, err := uc.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	return uc.userRepo.Delete(id)
}

func (uc *userUseCase) ChangePassword(id string, req *authdomain.ChangePasswordRequest) error {
	user, err := uc.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	if !password.Compare(req.OldPassword, user.PasswordHash) {
		return errors.New("old password is incorrect")
	}
	hashed, err := password.Hash(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}
	user.PasswordHash = hashed
	if err := uc.userRepo.Update(user); err != nil {
		return err
	}
	uc.revokeSessions(user.UserID)
	return nil
}

func (uc *userUseCase) AdminResetPassword(id string, req *authdomain.AdminResetPasswordRequest) error {
	user, err := uc.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	hashed, err := password.Hash(req.NewPassword)
	if err != nil {
		return errors.New("failed to hash new password")
	}
	user.PasswordHash = hashed
	if err := uc.userRepo.Update(user); err != nil {
		return err
	}
	uc.revokeSessions(user.UserID)
	return nil
}

// ForgotPassword creates a short-lived OTP without changing the current password.
func (uc *userUseCase) ForgotPassword(req *authdomain.ForgotPasswordRequest) error {
	user, err := uc.userRepo.FindByEmail(strings.TrimSpace(req.Email))
	if errors.Is(err, authdomain.ErrEmailNotRegistered) {
		return authdomain.ErrEmailNotRegistered
	}
	if err != nil {
		return errors.New("failed to validate reset email")
	}
	if user == nil {
		return authdomain.ErrEmailNotRegistered
	}
	if uc.otpRepo == nil {
		return errors.New("password reset service is unavailable")
	}

	otp, err := generateNumericOTP(passwordResetOTPLength)
	if err != nil {
		return errors.New("failed to generate reset OTP")
	}
	otpHash, err := password.Hash(otp)
	if err != nil {
		return errors.New("failed to protect reset OTP")
	}
	if err := uc.otpRepo.ReplaceForUser(user.UserID, otpHash, time.Now().Add(passwordResetOTPExpiry)); err != nil {
		return errors.New("failed to create reset OTP")
	}

	tmpl := mail.TmplForgotPassword
	plantID := ""
	if user.PlantID != nil {
		plantID = *user.PlantID
	}
	if s, findErr := uc.settingRepo.FindByKey(masterdomain.SettingKeyEmailTemplateForgotPass, plantID); findErr == nil && isValidPasswordResetOTPTemplate(s.SettingValue) {
		tmpl = s.SettingValue
	}

	return uc.mailer.SendTemplateForPlant(
		plantID,
		[]string{user.Email},
		"[Monitoring Audit] Kode OTP Reset Password",
		tmpl,
		map[string]string{
			"FullName":         user.FullName,
			"Username":         user.Username,
			"OTP":              otp,
			"OTPExpiryMinutes": "10",
			// Backward compatibility for templates saved before the OTP flow.
			"TempPassword": otp,
		},
	)
}

func isValidPasswordResetOTPTemplate(value string) bool {
	return strings.Contains(value, "{{.OTP}}") && strings.Contains(value, "{{.OTPExpiryMinutes}}")
}

func (uc *userUseCase) ResetPasswordWithOTP(req *authdomain.ResetPasswordWithOTPRequest) error {
	if req.NewPassword != req.ConfirmPassword {
		return errors.New("konfirmasi password tidak sama")
	}
	user, err := uc.userRepo.FindByEmail(strings.TrimSpace(req.Email))
	if err != nil || user == nil || uc.otpRepo == nil {
		return errors.New("OTP tidak valid atau sudah kedaluwarsa")
	}

	hashedPassword, err := password.Hash(req.NewPassword)
	if err != nil {
		return errors.New("gagal memproses password baru")
	}
	valid, err := uc.otpRepo.ConsumeValid(user.UserID, strings.TrimSpace(req.OTP), time.Now(), passwordResetOTPMaxAttempts)
	if err != nil {
		return errors.New("gagal memverifikasi OTP")
	}
	if !valid {
		return errors.New("OTP tidak valid atau sudah kedaluwarsa")
	}

	user.PasswordHash = hashedPassword
	if err := uc.userRepo.Update(user); err != nil {
		return errors.New("gagal menyimpan password baru")
	}
	uc.revokeSessions(user.UserID)
	return nil
}

func generateNumericOTP(n int) (string, error) {
	const digits = "0123456789"
	result := make([]byte, n)
	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		result[i] = digits[idx.Int64()]
	}
	return string(result), nil
}

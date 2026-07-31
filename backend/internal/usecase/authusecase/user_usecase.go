package authusecase

import (
	"crypto/rand"
	"errors"
	"math/big"

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
}

// NewUserUseCase creates a new UserUseCase implementation.
func NewUserUseCase(userRepo authdomain.UserRepository, mailer mail.Mailer, settingRepo masterdomain.SettingRepository, picRepo pic.PICMappingRepository) authdomain.UserUseCase {
	return &userUseCase{userRepo: userRepo, mailer: mailer, settingRepo: settingRepo, picRepo: picRepo}
}

func (uc *userUseCase) GetAll(page, limit int, search, roleID, deptID string) ([]authdomain.User, int64, error) {
	return uc.userRepo.FindAll(page, limit, search, roleID, deptID)
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
			AreaID:      "", // Deprecated: kawasan is the join key used by issue scoping.
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
	return uc.userRepo.Update(user)
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
	return uc.userRepo.Update(user)
}

// ForgotPassword generates a temp password and sends it via email.
func (uc *userUseCase) ForgotPassword(req *authdomain.ForgotPasswordRequest) error {
	user, err := uc.userRepo.FindByEmail(req.Email)
	if err != nil || user == nil {
		// Return generic message to avoid email enumeration
		return nil
	}

	// Generate 10-char random temp password
	tempPass, err := generateTempPassword(10)
	if err != nil {
		return errors.New("failed to generate temporary password")
	}

	hashed, err := password.Hash(tempPass)
	if err != nil {
		return errors.New("failed to hash temporary password")
	}
	user.PasswordHash = hashed
	if err := uc.userRepo.Update(user); err != nil {
		return errors.New("failed to update password")
	}

	// Send email asynchronously — non-blocking
	go func() {
		// Fetch dynamic template from database, fallback to static if not found
		tmpl := mail.TmplForgotPassword
		if s, err := uc.settingRepo.FindByKey(masterdomain.SettingKeyEmailTemplateForgotPass); err == nil && s.SettingValue != "" {
			tmpl = s.SettingValue
		}

		_ = uc.mailer.SendTemplate(
			[]string{user.Email},
			"[Monitoring Audit] Reset Password Anda",
			tmpl,
			map[string]string{
				"FullName":     user.FullName,
				"Username":     user.Username,
				"TempPassword": tempPass,
			},
		)
	}()

	return nil
}

// generateTempPassword returns a random alphanumeric string of the given length.
func generateTempPassword(n int) (string, error) {
	const chars = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghjkmnpqrstuvwxyz23456789"
	result := make([]byte, n)
	for i := range result {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(chars))))
		if err != nil {
			return "", err
		}
		result[i] = chars[idx.Int64()]
	}
	return string(result), nil
}

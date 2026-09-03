package authusecase

import (
	"errors"
	"strings"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/password"
	"github.com/monitoring-system/backend/pkg/refreshtoken"
)

type authUseCase struct {
	userRepo        authdomain.UserRepository
	loginLogRepo    logdomain.LoginLogRepository
	refreshRepo     authdomain.RefreshTokenRepository
	jwtManager      *jwt.Manager
	refreshTokenTTL time.Duration
}

// NewAuthUseCase creates a new AuthUseCase implementation.
func NewAuthUseCase(
	userRepo authdomain.UserRepository,
	loginLogRepo logdomain.LoginLogRepository,
	refreshRepo authdomain.RefreshTokenRepository,
	jwtManager *jwt.Manager,
	refreshTokenTTL time.Duration,
) authdomain.AuthUseCase {
	return &authUseCase{
		userRepo:        userRepo,
		loginLogRepo:    loginLogRepo,
		refreshRepo:     refreshRepo,
		jwtManager:      jwtManager,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (uc *authUseCase) Login(req *authdomain.LoginRequest, ip, device string) (*authdomain.LoginResponse, error) {
	identifier := strings.TrimSpace(req.Username)
	user, err := uc.userRepo.FindByUsername(identifier)
	if err != nil || user == nil {
		user, err = uc.userRepo.FindByEmail(identifier)
	}
	loginStatus := logdomain.LoginStatusSuccess

	if err != nil || user == nil {
		loginStatus = logdomain.LoginStatusFailed
		uc.recordLogin("", ip, device, loginStatus)
		return nil, errors.New("invalid username/email or password")
	}

	if !password.Compare(req.Password, user.PasswordHash) {
		loginStatus = logdomain.LoginStatusFailed
		uc.recordLogin(user.UserID, ip, device, loginStatus)
		return nil, errors.New("invalid username/email or password")
	}

	if user.UserStatus != authdomain.UserStatusActive {
		return nil, errors.New("account is not active")
	}

	accessToken, accessExpiresAt, err := uc.jwtManager.Generate(user.UserID, user.Username, user.RoleID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	familyID := idgen.Generate(idgen.PrefixRefreshToken)
	rawRefresh, refreshExpiresAt, err := uc.issueRefreshToken(user.UserID, familyID, ip, device)
	if err != nil {
		return nil, err
	}

	uc.recordLogin(user.UserID, ip, device, loginStatus)

	return &authdomain.LoginResponse{
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshToken:     rawRefresh,
		RefreshExpiresAt: refreshExpiresAt,
		CSRFToken:        refreshtoken.NewCSRFToken(),
		User:             toUserInfo(user),
	}, nil
}

// Refresh rotates a still-valid refresh token into a fresh access + refresh
// pair. Rotation always issues a brand-new refresh token and revokes the old
// one, even though the old one hadn't expired yet — this bounds how long a
// stolen-but-not-yet-used refresh token stays viable, and lets reuse of an
// already-rotated token be detected below (a strong signal of theft, since a
// legitimate client never presents the same refresh token twice).
func (uc *authUseCase) Refresh(rawRefreshToken, ip, device string) (*authdomain.LoginResponse, error) {
	if rawRefreshToken == "" {
		return nil, errors.New("refresh token is required")
	}

	stored, err := uc.refreshRepo.FindByHash(refreshtoken.Hash(rawRefreshToken))
	if err != nil || stored == nil {
		return nil, errors.New("invalid refresh token")
	}
	if stored.RevokedAt != nil {
		// This exact token was already rotated away (or logged out) once
		// before — presenting it again means either a client bug or someone
		// replaying a token they stole earlier. Kill every token in the
		// family so both the thief and the legitimate client are forced
		// back to a fresh login, where the legitimate owner will notice.
		_ = uc.refreshRepo.RevokeFamily(stored.FamilyID)
		return nil, errors.New("invalid refresh token")
	}
	if time.Now().After(stored.ExpiresAt) {
		return nil, errors.New("refresh token expired")
	}

	user, err := uc.userRepo.FindByID(stored.UserID)
	if err != nil || user == nil {
		return nil, errors.New("invalid refresh token")
	}
	if user.UserStatus != authdomain.UserStatusActive {
		return nil, errors.New("account is not active")
	}

	accessToken, accessExpiresAt, err := uc.jwtManager.Generate(user.UserID, user.Username, user.RoleID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	rawRefresh, refreshExpiresAt, newID, err := uc.issueRefreshTokenWithID(user.UserID, stored.FamilyID, ip, device)
	if err != nil {
		return nil, err
	}
	_ = uc.refreshRepo.RevokeByID(stored.ID, &newID)

	return &authdomain.LoginResponse{
		AccessToken:      accessToken,
		AccessExpiresAt:  accessExpiresAt,
		RefreshToken:     rawRefresh,
		RefreshExpiresAt: refreshExpiresAt,
		CSRFToken:        refreshtoken.NewCSRFToken(),
		User:             toUserInfo(user),
	}, nil
}

func (uc *authUseCase) Logout(userID, loginLogID, rawRefreshToken string) error {
	if rawRefreshToken != "" {
		if stored, err := uc.refreshRepo.FindByHash(refreshtoken.Hash(rawRefreshToken)); err == nil && stored != nil && stored.RevokedAt == nil {
			_ = uc.refreshRepo.RevokeByID(stored.ID, nil)
		}
	}
	now := time.Now()
	return uc.loginLogRepo.UpdateLogoutAt(loginLogID, now)
}

func (uc *authUseCase) Me(userID string) (*authdomain.UserInfo, error) {
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	info := toUserInfo(user)
	return &info, nil
}

func (uc *authUseCase) issueRefreshToken(userID, familyID, ip, device string) (raw string, expiresAt time.Time, err error) {
	raw, expiresAt, _, err = uc.issueRefreshTokenWithID(userID, familyID, ip, device)
	return raw, expiresAt, err
}

func (uc *authUseCase) issueRefreshTokenWithID(userID, familyID, ip, device string) (raw string, expiresAt time.Time, id string, err error) {
	raw, hash := refreshtoken.Generate()
	id = idgen.Generate(idgen.PrefixRefreshToken)
	expiresAt = time.Now().Add(uc.refreshTokenTTL)
	err = uc.refreshRepo.Create(&authdomain.RefreshToken{
		ID:         id,
		UserID:     userID,
		FamilyID:   familyID,
		TokenHash:  hash,
		ExpiresAt:  expiresAt,
		IPAddress:  ip,
		DeviceInfo: device,
	})
	if err != nil {
		return "", time.Time{}, "", errors.New("failed to create session")
	}
	return raw, expiresAt, id, nil
}

func (uc *authUseCase) recordLogin(userID, ip, device string, status logdomain.LoginStatus) {
	var uid *string
	if userID != "" {
		uid = &userID
	}
	logEntry := &logdomain.LoginLog{
		LoginLogID:  idgen.Generate(idgen.PrefixLoginLog),
		UserID:      uid,
		LoginAt:     time.Now(),
		IPAddress:   ip,
		DeviceInfo:  device,
		LoginStatus: status,
	}
	_ = uc.loginLogRepo.Create(logEntry) // non-blocking; ignore error
}

func toUserInfo(user *authdomain.User) authdomain.UserInfo {
	return authdomain.UserInfo{
		UserID:       user.UserID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		DepartmentID: user.DepartmentID,
		RoleID:       user.RoleID,
		PlantID:      user.PlantID,
		UserStatus:   user.UserStatus,
	}
}

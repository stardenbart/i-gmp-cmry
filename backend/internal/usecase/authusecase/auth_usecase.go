package authusecase

import (
	"errors"
	"time"

	authdomain "github.com/monitoring-system/backend/internal/domain/auth"
	logdomain "github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/pkg/idgen"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/password"
)

type authUseCase struct {
	userRepo     authdomain.UserRepository
	loginLogRepo logdomain.LoginLogRepository
	jwtManager   *jwt.Manager
}

// NewAuthUseCase creates a new AuthUseCase implementation.
func NewAuthUseCase(
	userRepo authdomain.UserRepository,
	loginLogRepo logdomain.LoginLogRepository,
	jwtManager *jwt.Manager,
) authdomain.AuthUseCase {
	return &authUseCase{
		userRepo:     userRepo,
		loginLogRepo: loginLogRepo,
		jwtManager:   jwtManager,
	}
}

func (uc *authUseCase) Login(req *authdomain.LoginRequest, ip, device string) (*authdomain.LoginResponse, error) {
	user, err := uc.userRepo.FindByUsername(req.Username)
	loginStatus := logdomain.LoginStatusSuccess

	if err != nil || user == nil {
		loginStatus = logdomain.LoginStatusFailed
		uc.recordLogin("", ip, device, loginStatus)
		return nil, errors.New("invalid username or password")
	}

	if !password.Compare(req.Password, user.PasswordHash) {
		loginStatus = logdomain.LoginStatusFailed
		uc.recordLogin(user.UserID, ip, device, loginStatus)
		return nil, errors.New("invalid username or password")
	}

	if user.UserStatus != authdomain.UserStatusActive {
		return nil, errors.New("account is not active")
	}

	token, expiredAt, err := uc.jwtManager.Generate(user.UserID, user.Username, user.RoleID)
	if err != nil {
		return nil, errors.New("failed to generate token")
	}

	uc.recordLogin(user.UserID, ip, device, loginStatus)

	return &authdomain.LoginResponse{
		Token:     token,
		ExpiredAt: expiredAt,
		User: authdomain.UserInfo{
			UserID:       user.UserID,
			Username:     user.Username,
			FullName:     user.FullName,
			Email:        user.Email,
			DepartmentID: user.DepartmentID,
			RoleID:       user.RoleID,
			PlantID:      user.PlantID,
			UserStatus:   user.UserStatus,
		},
	}, nil
}

func (uc *authUseCase) Logout(userID, loginLogID string) error {
	now := time.Now()
	return uc.loginLogRepo.UpdateLogoutAt(loginLogID, now)
}

func (uc *authUseCase) Me(userID string) (*authdomain.UserInfo, error) {
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}
	return &authdomain.UserInfo{
		UserID:       user.UserID,
		Username:     user.Username,
		FullName:     user.FullName,
		Email:        user.Email,
		DepartmentID: user.DepartmentID,
		RoleID:       user.RoleID,
		PlantID:      user.PlantID,
		UserStatus:   user.UserStatus,
	}, nil
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

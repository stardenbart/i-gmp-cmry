package auth

import "time"

// UserStatus defines allowed values for the UserStatus field.
type UserStatus string

const (
	UserStatusActive    UserStatus = "Active"
	UserStatusInactive  UserStatus = "Inactive"
	UserStatusSuspended UserStatus = "Suspended"
)

// User represents the Users table.
type User struct {
	UserID           string     `gorm:"column:UserID;primaryKey" json:"user_id"`
	DepartmentID     string     `gorm:"column:DepartmentID;not null" json:"department_id"`
	RoleID           string     `gorm:"column:RoleID;not null" json:"role_id"`
	Username         string     `gorm:"column:Username;uniqueIndex;not null" json:"username"`
	FullName         string     `gorm:"column:FullName;not null" json:"full_name"`
	Email            string     `gorm:"column:Email;uniqueIndex;not null" json:"email"`
	PasswordHash     string     `gorm:"column:PasswordHash;not null" json:"-"`
	UserStatus       UserStatus `gorm:"column:UserStatus;not null;default:Active" json:"user_status"`
	UserCreatedAt    time.Time  `gorm:"column:UserCreatedAt;autoCreateTime" json:"created_at"`
	UserUpdatedAt    time.Time  `gorm:"column:UserUpdatedAt;autoUpdateTime" json:"updated_at"`

	// Relations (preload when needed)
	Role *Role `gorm:"foreignKey:RoleID;references:RoleID" json:"role,omitempty"`
}

func (User) TableName() string { return "Users" }

// ─── DTOs ──────────────────────────────────────────────────────────────────

type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required,min=6"`
}

type LoginResponse struct {
	Token     string    `json:"token"`
	ExpiredAt time.Time `json:"expired_at"`
	User      UserInfo  `json:"user"`
}

type UserInfo struct {
	UserID       string     `json:"user_id"`
	Username     string     `json:"username"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email"`
	DepartmentID string     `json:"department_id"`
	RoleID       string     `json:"role_id"`
	UserStatus   UserStatus `json:"user_status"`
}

type CreateUserRequest struct {
	DepartmentID string `json:"department_id" validate:"required"`
	RoleID       string `json:"role_id" validate:"required"`
	Username     string `json:"username" validate:"required,min=3,max=50"`
	FullName     string `json:"full_name" validate:"required"`
	Email        string `json:"email" validate:"required,email"`
	Password     string `json:"password" validate:"required,min=8"`
}

type UpdateUserRequest struct {
	DepartmentID string     `json:"department_id"`
	RoleID       string     `json:"role_id"`
	FullName     string     `json:"full_name"`
	Email        string     `json:"email" validate:"omitempty,email"`
	UserStatus   UserStatus `json:"user_status" validate:"omitempty,oneof=Active Inactive Suspended"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type AdminResetPasswordRequest struct {
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ─── Repository Interface ──────────────────────────────────────────────────

type UserRepository interface {
	FindAll(page, limit int, search, roleID, deptID string) ([]User, int64, error)
	FindByID(id string) (*User, error)
	FindByUsername(username string) (*User, error)
	FindByEmail(email string) (*User, error)
	Create(u *User) error
	Update(u *User) error
	Delete(id string) error
}

// ─── UseCase Interface ─────────────────────────────────────────────────────

type UserUseCase interface {
	GetAll(page, limit int, search, roleID, deptID string) ([]User, int64, error)
	GetByID(id string) (*User, error)
	Create(req *CreateUserRequest) (*User, error)
	Update(id string, req *UpdateUserRequest) (*User, error)
	Delete(id string) error
	ChangePassword(id string, req *ChangePasswordRequest) error
	AdminResetPassword(id string, req *AdminResetPasswordRequest) error
	ForgotPassword(req *ForgotPasswordRequest) error
}

// ─── Auth UseCase Interface ────────────────────────────────────────────────

type AuthUseCase interface {
	Login(req *LoginRequest, ip, device string) (*LoginResponse, error)
	Logout(userID, loginLogID string) error
	Me(userID string) (*UserInfo, error)
}

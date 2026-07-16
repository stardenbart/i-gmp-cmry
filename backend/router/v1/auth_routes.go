package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/auth"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/internal/usecase/authusecase"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/mail"
	"gorm.io/gorm"
)

// RegisterAuthRoutes wires auth dependencies and mounts routes.
func RegisterAuthRoutes(rg fiber.Router, db *gorm.DB, mailer mail.Mailer, jwtManager *jwt.Manager, log *logger.Logger, actLogUC logging.ActivityLogUseCase) {
	// ── Wire dependencies ──────────────────────────────────────────────
	userRepo := authrepo.NewUserRepository(db)
	loginLogRepo := authrepo.NewLoginLogRepository(db)
	rolePermRepo := authrepo.NewRolePermissionRepository(db)
	settingRepo := masterrepo.NewSettingRepository(db)

	authUC := authusecase.NewAuthUseCase(userRepo, loginLogRepo, jwtManager)
	userUC := authusecase.NewUserUseCase(userRepo, mailer, settingRepo)
	rolePermUC := authusecase.NewRolePermissionUseCase(rolePermRepo)

	authHandler := auth.NewAuthHandler(authUC)
	userHandler := auth.NewUserHandler(userUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)

	// ── Public routes ──────────────────────────────────────────────────
	authGroup := rg.Group("/auth")
	{
		authGroup.Post("/login", authHandler.Login)
		authGroup.Post("/forgot-password", userHandler.ForgotPassword)
	}

	// ── Protected routes ───────────────────────────────────────────────
	protected := rg.Group("/", authMW, actLogMW)
	{
		protected.Post("/auth/logout", authHandler.Logout)
		protected.Get("/auth/me", authHandler.Me)

		// User management — requires permission check
		users := protected.Group("/users")
		{
			users.Get("", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "READ"), userHandler.GetAll)
			users.Post("", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "CREATE"), userHandler.Create)
			users.Get("/:id", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "READ"), userHandler.GetByID)
			users.Put("/:id", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "UPDATE"), userHandler.Update)
			users.Delete("/:id", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "DELETE"), userHandler.Delete)
			users.Put("/:id/change-password", authMW, userHandler.ChangePassword)
			users.Put("/:id/reset-password", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "UPDATE"), userHandler.ResetPassword)
		}
	}
}

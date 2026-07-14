package v1

import (
	"github.com/gin-gonic/gin"
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
func RegisterAuthRoutes(rg *gin.RouterGroup, db *gorm.DB, mailer mail.Mailer, jwtManager *jwt.Manager, log *logger.Logger) {
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

	// ── Public routes ──────────────────────────────────────────────────
	authGroup := rg.Group("/auth")
	{
		authGroup.POST("/login", authHandler.Login)
		authGroup.POST("/forgot-password", userHandler.ForgotPassword)
	}

	// ── Protected routes ───────────────────────────────────────────────
	protected := rg.Group("/", authMW)
	{
		protected.POST("/auth/logout", authHandler.Logout)
		protected.GET("/auth/me", authHandler.Me)

		// User management — requires permission check
		users := protected.Group("/users")
		{
			users.GET("", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "READ"), userHandler.GetAll)
			users.POST("", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "CREATE"), userHandler.Create)
			users.GET("/:id", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "READ"), userHandler.GetByID)
			users.PUT("/:id", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "UPDATE"), userHandler.Update)
			users.DELETE("/:id", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "DELETE"), userHandler.Delete)
			users.PUT("/:id/change-password", authMW, userHandler.ChangePassword)
			users.PUT("/:id/reset-password", middleware.PermissionMiddleware(rolePermUC, "MOD-USR", "UPDATE"), userHandler.ResetPassword)
		}
	}
}

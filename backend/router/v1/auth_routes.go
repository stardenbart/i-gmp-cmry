package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/logging"
	"github.com/monitoring-system/backend/internal/handler/auth"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/authrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/masterrepo"
	"github.com/monitoring-system/backend/internal/infrastructure/persistence/picrepo"
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
	passwordResetOTPRepo := authrepo.NewPasswordResetOTPRepository(db)
	loginLogRepo := authrepo.NewLoginLogRepository(db)
	rolePermRepo := authrepo.NewRolePermissionRepository(db)
	userPermRepo := authrepo.NewUserPermissionRepository(db)
	settingRepo := masterrepo.NewSettingRepository(db)
	picMappingRepo := picrepo.NewPICMappingRepository(db)

	authUC := authusecase.NewAuthUseCase(userRepo, loginLogRepo, jwtManager)
	userUC := authusecase.NewUserUseCase(userRepo, mailer, settingRepo, picMappingRepo, passwordResetOTPRepo)
	rolePermUC := authusecase.NewRolePermissionUseCase(rolePermRepo)
	userPermUC := authusecase.NewUserPermissionUseCase(userPermRepo)

	// User filter
	userFilterRepo := authrepo.NewUserFilterRepository(db)
	userFilterUC := authusecase.NewUserFilterUseCase(userFilterRepo)

	authHandler := auth.NewAuthHandler(authUC)
	userHandler := auth.NewUserHandler(userUC)
	userPermHandler := auth.NewUserPermissionHandler(userPermUC)
	userFilterHandler := auth.NewUserFilterHandler(userFilterUC)

	authMW := middleware.AuthMiddleware(jwtManager)
	actLogMW := middleware.ActivityLogMiddleware(actLogUC)
	plantScopeMW := middleware.PlantScopeMiddleware(userRepo)

	// ── Public routes ──────────────────────────────────────────────────
	authGroup := rg.Group("/auth")
	{
		authGroup.Post("/login", middleware.AuthRateLimiter(), authHandler.Login)
		authGroup.Post("/forgot-password", middleware.AuthRateLimiter(), userHandler.ForgotPassword)
		authGroup.Post("/reset-password", middleware.AuthRateLimiter(), userHandler.ResetPasswordWithOTP)
	}

	// ── Protected routes ───────────────────────────────────────────────
	protected := rg.Group("/", authMW, actLogMW)
	{
		protected.Post("/auth/logout", authHandler.Logout)
		protected.Get("/auth/me", authHandler.Me)

		// User management — requires permission check
		users := protected.Group("/users", plantScopeMW)
		{
			users.Get("", middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-USR", "READ"), userHandler.GetAll)
			users.Post("", middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-USR", "CREATE"), userHandler.Create)
			// Filter route — same permission as list (before /:id)
			users.Get("/filter", middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-USR", "READ"), userFilterHandler.GetFiltered)
			users.Get("/:id", func(c *fiber.Ctx) error {
				userID := middleware.GetUserID(c)
				id := c.Params("id")
				if userID == id {
					return c.Next()
				}
				return middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-USR", "READ")(c)
			}, userHandler.GetByID)
			users.Put("/:id", func(c *fiber.Ctx) error {
				userID := middleware.GetUserID(c)
				id := c.Params("id")
				if userID == id {
					return c.Next()
				}
				return middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-USR", "UPDATE")(c)
			}, userHandler.Update)
			users.Delete("/:id", middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-USR", "DELETE"), userHandler.Delete)
			users.Put("/:id/change-password", authMW, userHandler.ChangePassword)
			users.Put("/:id/reset-password", middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-USR", "UPDATE"), userHandler.ResetPassword)

			// User Permissions Overrides
			users.Get("/:id/permissions", func(c *fiber.Ctx) error {
				userID := middleware.GetUserID(c)
				id := c.Params("id")
				if userID == id {
					return c.Next()
				}
				return middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-USR", "READ")(c)
			}, userPermHandler.GetByUserID)
			users.Put("/:id/permissions", middleware.PermissionMiddleware(rolePermUC, userPermUC, "MOD-USR", "UPDATE"), userPermHandler.SetPermissions)
		}
	}
}

package v1

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/handler/searchhandler"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/jwt"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/opensearch"
)

// RegisterSearchRoutes wires OpenSearch dependencies and mounts routes.
func RegisterSearchRoutes(rg fiber.Router, osClient *opensearch.Client, jwtManager *jwt.Manager, log *logger.Logger) {
	searchH := searchhandler.NewSearchHandler(osClient)

	authMW := middleware.AuthMiddleware(jwtManager)
	search := rg.Group("/search", authMW)
	{
		search.Get("/inspections", searchH.SearchInspections)
		search.Get("/issues", searchH.SearchIssues)
		search.Get("/logs", searchH.SearchActivityLogs)
	}
}

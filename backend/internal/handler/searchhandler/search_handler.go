package searchhandler

import (
	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/pkg/opensearch"
	"github.com/monitoring-system/backend/pkg/pagination"
	"github.com/monitoring-system/backend/pkg/response"
)

type SearchHandler struct {
	osClient *opensearch.Client
}

func NewSearchHandler(osClient *opensearch.Client) *SearchHandler {
	return &SearchHandler{osClient: osClient}
}

// SearchInspections provides full-text search across audit inspections
func (h *SearchHandler) SearchInspections(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	q := c.Query("q")
	query := buildQuery(q, p.Offset, p.Limit, "inspection_id", "status", "area_id")

	results, total, err := h.osClient.Search(c.UserContext(), "audit-inspections", query)
	if err != nil {
		return response.InternalServerError(c, "search failed", err.Error())
	}
	return response.Paginated(c, "success", results, total, p.Page, p.Limit)
}

// SearchIssues provides full-text search across audit issues
func (h *SearchHandler) SearchIssues(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	q := c.Query("q")
	query := buildQuery(q, p.Offset, p.Limit, "issue_id", "status", "issue_pic_user_id")

	results, total, err := h.osClient.Search(c.UserContext(), "audit-issues", query)
	if err != nil {
		return response.InternalServerError(c, "search failed", err.Error())
	}
	return response.Paginated(c, "success", results, total, p.Page, p.Limit)
}

// SearchActivityLogs provides full-text search across activity logs
func (h *SearchHandler) SearchActivityLogs(c *fiber.Ctx) error {
	p := pagination.FromQuery(c)
	q := c.Query("q")
	query := buildQuery(q, p.Offset, p.Limit, "action", "entity", "details", "actor_id")

	results, total, err := h.osClient.Search(c.UserContext(), "audit-activity-logs", query)
	if err != nil {
		return response.InternalServerError(c, "search failed", err.Error())
	}
	return response.Paginated(c, "success", results, total, p.Page, p.Limit)
}

// buildQuery creates a simple multi_match query against the specified fields
func buildQuery(queryString string, offset, limit int, fields ...string) map[string]interface{} {
	if queryString == "" {
		return map[string]interface{}{
			"from": offset,
			"size": limit,
			"query": map[string]interface{}{
				"match_all": map[string]interface{}{},
			},
			"sort": []map[string]interface{}{
				{"timestamp": map[string]interface{}{"order": "desc"}},
			},
		}
	}

	return map[string]interface{}{
		"from": offset,
		"size": limit,
		"query": map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  queryString,
				"fields": fields,
				"type":   "best_fields",
			},
		},
		"sort": []map[string]interface{}{
			{"_score": map[string]interface{}{"order": "desc"}},
			{"timestamp": map[string]interface{}{"order": "desc"}},
		},
	}
}

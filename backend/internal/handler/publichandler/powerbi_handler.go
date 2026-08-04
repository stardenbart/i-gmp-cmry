package publichandler

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/monitoring-system/backend/internal/domain/apikey"
	"github.com/monitoring-system/backend/internal/middleware"
	"github.com/monitoring-system/backend/pkg/crypto"
	"github.com/monitoring-system/backend/pkg/logger"
	"github.com/monitoring-system/backend/pkg/response"
	"gorm.io/gorm"
)

type PowerBIHandler struct {
	db        *gorm.DB
	cryptoSvc *crypto.Service
	log       *logger.Logger
}

func NewPowerBIHandler(db *gorm.DB, cryptoSvc *crypto.Service, log *logger.Logger) *PowerBIHandler {
	return &PowerBIHandler{
		db:        db,
		cryptoSvc: cryptoSvc,
		log:       log,
	}
}

// ParseSinceParam parses optional since parameter (ISO8601 or YYYY-MM-DD)
func parseSinceParam(c *fiber.Ctx) *time.Time {
	sinceStr := c.Query("since")
	if sinceStr == "" {
		sinceStr = c.Query("updated_at")
	}
	if sinceStr == "" {
		return nil
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z07:00",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, fmtStr := range formats {
		if t, err := time.Parse(fmtStr, sinceStr); err == nil {
			return &t
		}
	}
	return nil
}

// GetAllData returns all 5 requested datasets: headers, results, issues, follow_ups, issue_photos
// GET /api/v1/public/powerbi/data?since=2026-07-01T00:00:00Z
func (h *PowerBIHandler) GetAllData(c *fiber.Ctx) error {
	since := parseSinceParam(c)

	var apiKeyPlantID string
	if keyEntity, ok := c.Locals(middleware.ContextKeyAPIKey).(*apikey.APIKey); ok && keyEntity != nil {
		if keyEntity.PlantID != nil && *keyEntity.PlantID != "" {
			apiKeyPlantID = *keyEntity.PlantID
		}
	}

	// 1. Inspection Headers
	type InspectionHeaderDTO struct {
		InspectionID      string    `json:"inspection_id"`
		AreaID            string    `json:"area_id"`
		AreaName          string    `json:"area_name"`
		KawasanID         string    `json:"kawasan_id"`
		KawasanName       string    `json:"kawasan_name"`
		DetailKawasanID   string    `json:"detail_kawasan_id"`
		DetailKawasanName string    `json:"detail_kawasan_name"`
		InspectorID       string    `json:"inspector_id"`
		InspectorName     string    `json:"inspector_name"`
		Status            string    `json:"status"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
	}

	var headers []InspectionHeaderDTO
	headerQuery := h.db.Table(`"Inspection_Header" ih`).
		Select(`ih."InspectionID" as inspection_id, ih."AreaID" as area_id, COALESCE(am."AreaName", ih."AreaID") as area_name, ih."KawasanID" as kawasan_id, COALESCE(km."KawasanName", ih."KawasanID") as kawasan_name, ih."DetailKawasanID" as detail_kawasan_id, COALESCE(dkm."DetailKawasanName", ih."DetailKawasanID") as detail_kawasan_name, ih."InspectorID" as inspector_id, COALESCE(u."FullName", ih."InspectorID") as inspector_name, ih."InspectionHeaderStatus" as status, ih."InspectionHeaderCreatedAt" as created_at, ih."InspectionheaderUpdatedAt" as updated_at`).
		Joins(`LEFT JOIN "Area_Master" am ON am."AreaID" = ih."AreaID"`).
		Joins(`LEFT JOIN "Kawasan_Master" km ON km."KawasanID" = ih."KawasanID"`).
		Joins(`LEFT JOIN "DetailKawasan_Master" dkm ON dkm."DetailKawasanID" = ih."DetailKawasanID"`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = ih."InspectorID"`)

	if apiKeyPlantID != "" {
		headerQuery = headerQuery.Where(`am."PlantID" = ?`, apiKeyPlantID)
	}

	if since != nil {
		headerQuery = headerQuery.Where(`ih."InspectionheaderUpdatedAt" >= ? OR ih."InspectionHeaderCreatedAt" >= ?`, *since, *since)
	}
	_ = headerQuery.Order(`ih."InspectionHeaderCreatedAt" DESC`).Scan(&headers).Error

	// 2. Inspection Results
	type InspectionResultDTO struct {
		ResultID     string    `json:"result_id"`
		InspectionID string    `json:"inspection_id"`
		UraianID     string    `json:"uraian_id"`
		UraianText   string    `json:"uraian_text"`
		DetailName   string    `json:"detail_name"`
		AspekName    string    `json:"aspek_name"`
		Nilai        int       `json:"nilai"`
		Keterangan   string    `json:"keterangan"`
		CreatedAt    time.Time `json:"created_at"`
	}

	var results []InspectionResultDTO
	resultQuery := h.db.Table(`"Inspection_Result" ir`).
		Select(`ir."ResultID" as result_id, ir."InspectionID" as inspection_id, ir."UraianID" as uraian_id, um."UraianText" as uraian_text, dm."DetailName" as detail_name, am."AspekName" as aspek_name, ir."Nilai" as nilai, ir."Keterangan" as keterangan, ir."ResultCreatedAt" as created_at`).
		Joins(`LEFT JOIN "Uraian_Master" um ON um."UraianID" = ir."UraianID"`).
		Joins(`LEFT JOIN "Detail_Master" dm ON dm."DetailID" = um."DetailID"`).
		Joins(`LEFT JOIN "Aspek_Master" am ON am."AspekID" = dm."AspekID"`)

	if apiKeyPlantID != "" {
		resultQuery = resultQuery.Joins(`JOIN "Inspection_Header" ih_res ON ih_res."InspectionID" = ir."InspectionID"`).
			Joins(`JOIN "Area_Master" am_res ON am_res."AreaID" = ih_res."AreaID"`).
			Where(`am_res."PlantID" = ?`, apiKeyPlantID)
	}

	if since != nil {
		resultQuery = resultQuery.Where(`ir."ResultCreatedAt" >= ?`, *since)
	}
	_ = resultQuery.Order(`ir."ResultCreatedAt" DESC`).Scan(&results).Error

	// 3. Issues
	type IssueDTO struct {
		IssueID        string     `json:"issue_id"`
		ResultID       string     `json:"result_id"`
		InspectionID   string     `json:"inspection_id"`
		IssuePICUserID string     `json:"issue_pic_user_id"`
		PICName        string     `json:"pic_name"`
		DueDate        *time.Time `json:"due_date"`
		IssueStatus    string     `json:"issue_status"`
		Label          string     `json:"label"`
		NeedsWOWR      bool       `json:"needs_wo_wr"`
		WO_ID          string     `json:"wo_id"`
		WR_ID          string     `json:"wr_id"`
		WOWRStatus     string     `json:"wowr_status"`
		Keterangan     string     `json:"keterangan"`
		CreatedAt      time.Time  `json:"created_at"`
		UpdatedAt      time.Time  `json:"updated_at"`
	}

	var issues []IssueDTO
	issueQuery := h.db.Table(`"Issue" i`).
		Select(`i."IssueID" as issue_id, i."ResultID" as result_id, ir."InspectionID" as inspection_id, i."IssuePICUserID" as issue_pic_user_id, COALESCE(u."FullName", i."IssuePICUserID") as pic_name, i."DueDate" as due_date, i."IssueStatus" as issue_status, i."Label" as label, i."NeedsWOWR" as needs_wo_wr, i."WO_ID" as wo_id, i."WR_ID" as wr_id, i."WOWRStatus" as wowr_status, i."Keterangan" as keterangan, i."IssueCreatedAt" as created_at, i."IssueUpdatedAt" as updated_at`).
		Joins(`LEFT JOIN "Inspection_Result" ir ON ir."ResultID" = i."ResultID"`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = i."IssuePICUserID"`)

	if apiKeyPlantID != "" {
		issueQuery = issueQuery.Joins(`JOIN "Inspection_Header" ih_iss ON ih_iss."InspectionID" = ir."InspectionID"`).
			Joins(`JOIN "Area_Master" am_iss ON am_iss."AreaID" = ih_iss."AreaID"`).
			Where(`am_iss."PlantID" = ?`, apiKeyPlantID)
	}

	if since != nil {
		issueQuery = issueQuery.Where(`i."IssueUpdatedAt" >= ? OR i."IssueCreatedAt" >= ?`, *since, *since)
	}
	_ = issueQuery.Order(`i."IssueCreatedAt" DESC`).Scan(&issues).Error

	// Decrypt issue keterangan if encrypted
	for i := range issues {
		if h.cryptoSvc != nil && issues[i].Keterangan != "" {
			issues[i].Keterangan = h.cryptoSvc.DecryptWithFallback(issues[i].Keterangan)
		}
	}

	// 4. Follow Ups & 5. Issue Photos
	type IssuePhotoDTO struct {
		IssuePhotoID   string     `json:"issue_photo_id"`
		IssueID        string     `json:"issue_id"`
		PICUserID      string     `json:"pic_user_id"`
		PICName        string     `json:"pic_name"`
		PhotoType      string     `json:"photo_type"`
		ImageUrl       string     `json:"image_url"`
		FileName       string     `json:"file_name"`
		FollowUpDate   *time.Time `json:"follow_up_date,omitempty"`
		JumlahFollowUp *int       `json:"jumlah_follow_up,omitempty"`
		CreatedAt      time.Time  `json:"created_at"`
		UpdatedAt      time.Time  `json:"updated_at"`
	}

	var allPhotos []IssuePhotoDTO
	photoQuery := h.db.Table(`"Issue_Photo" ip`).
		Select(`ip."IssuePhotoID" as issue_photo_id, ip."IssueID" as issue_id, ip."PICUserID" as pic_user_id, COALESCE(u."FullName", ip."PICUserID") as pic_name, ip."PhotoType" as photo_type, ip."ImageUrl" as image_url, ip."FileName" as file_name, ip."FollowUpDate" as follow_up_date, ip."JumlahFollowUp" as jumlah_follow_up, ip."PhotoCreatedAt" as created_at, ip."PhotoUpdatedAt" as updated_at`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = ip."PICUserID"`)

	if apiKeyPlantID != "" {
		photoQuery = photoQuery.Joins(`JOIN "Issue" i_pho ON i_pho."IssueID" = ip."IssueID"`).
			Joins(`JOIN "Inspection_Result" ir_pho ON ir_pho."ResultID" = i_pho."ResultID"`).
			Joins(`JOIN "Inspection_Header" ih_pho ON ih_pho."InspectionID" = ir_pho."InspectionID"`).
			Joins(`JOIN "Area_Master" am_pho ON am_pho."AreaID" = ih_pho."AreaID"`).
			Where(`am_pho."PlantID" = ?`, apiKeyPlantID)
	}

	if since != nil {
		photoQuery = photoQuery.Where(`ip."PhotoUpdatedAt" >= ? OR ip."PhotoCreatedAt" >= ?`, *since, *since)
	}
	_ = photoQuery.Order(`ip."PhotoCreatedAt" DESC`).Scan(&allPhotos).Error

	var followUps []IssuePhotoDTO
	var photos []IssuePhotoDTO

	for i := range allPhotos {
		if h.cryptoSvc != nil && allPhotos[i].ImageUrl != "" {
			allPhotos[i].ImageUrl = h.cryptoSvc.DecryptWithFallback(allPhotos[i].ImageUrl)
		}
		if allPhotos[i].PhotoType == "FollowUp" {
			followUps = append(followUps, allPhotos[i])
		} else {
			photos = append(photos, allPhotos[i])
		}
	}

	sinceVal := ""
	if since != nil {
		sinceVal = since.Format(time.RFC3339)
	}

	return response.OK(c, "success", fiber.Map{
		"meta": fiber.Map{
			"extracted_at": time.Now().Format(time.RFC3339),
			"since_filter": sinceVal,
		},
		"inspection_headers": headers,
		"inspection_results": results,
		"issues":             issues,
		"follow_ups":         followUps,
		"issue_photos":       photos,
	})
}

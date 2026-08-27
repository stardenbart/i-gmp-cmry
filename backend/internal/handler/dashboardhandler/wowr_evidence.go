package dashboardhandler

import (
	"time"
)

// WOWRReportEvidence is a report-safe representation of an issue photo.
type WOWRReportEvidence struct {
	PhotoID     string    `json:"photo_id"`
	RefPhotoID  string    `json:"ref_photo_id,omitempty"`
	PhotoType   string    `json:"photo_type"`
	ImageURL    string    `json:"image_url"`
	Description string    `json:"description,omitempty"`
	Uploader    string    `json:"uploader,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type wowrEvidenceRow struct {
	IssueID     string    `gorm:"column:issue_id"`
	PhotoID     string    `gorm:"column:photo_id"`
	RefPhotoID  *string   `gorm:"column:ref_photo_id"`
	PhotoType   string    `gorm:"column:photo_type"`
	ImageURL    string    `gorm:"column:image_url"`
	Description string    `gorm:"column:description"`
	Uploader    string    `gorm:"column:uploader"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (h *DashboardHandler) loadWOWREvidence(issueIDs []string) (map[string][]WOWRReportEvidence, error) {
	grouped := make(map[string][]WOWRReportEvidence)
	if len(issueIDs) == 0 {
		return grouped, nil
	}

	var rows []wowrEvidenceRow
	err := h.db.Table(`"Issue_Photo" ip`).
		Select(`ip."IssueID" as issue_id,
			ip."IssuePhotoID" as photo_id,
			ip."RefPhotoID" as ref_photo_id,
			ip."PhotoType" as photo_type,
			ip."ImageUrl" as image_url,
			COALESCE(ip."Keterangan", '') as description,
			COALESCE(u."FullName", ip."PICUserID", '') as uploader,
			ip."PhotoCreatedAt" as created_at`).
		Joins(`LEFT JOIN "Users" u ON u."UserID" = ip."PICUserID"`).
		Where(`ip."IssueID" IN ? AND ip."PhotoType" IN ?`, issueIDs, []string{"Initial", "WOWR"}).
		Order(`ip."IssueID" ASC, ip."PhotoCreatedAt" ASC`).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, row := range rows {
		imageURL := row.ImageURL
		description := row.Description
		if h.cryptoSvc != nil {
			imageURL = h.cryptoSvc.DecryptWithFallback(imageURL)
			description = h.cryptoSvc.DecryptWithFallback(description)
		}
		if imageURL == "" {
			continue
		}
		refPhotoID := ""
		if row.RefPhotoID != nil {
			refPhotoID = *row.RefPhotoID
		}
		grouped[row.IssueID] = append(grouped[row.IssueID], WOWRReportEvidence{
			PhotoID:     row.PhotoID,
			RefPhotoID:  refPhotoID,
			PhotoType:   row.PhotoType,
			ImageURL:    imageURL,
			Description: description,
			Uploader:    row.Uploader,
			CreatedAt:   row.CreatedAt,
		})
	}
	return grouped, nil
}

func evidenceForWOWRItem(all []WOWRReportEvidence, photoID string) (initial, completion []WOWRReportEvidence) {
	initialCount := 0
	for _, photo := range all {
		if photo.PhotoType == "Initial" {
			initialCount++
			if photoID == "" || photo.PhotoID == photoID {
				initial = append(initial, photo)
			}
		}
	}

	for _, photo := range all {
		if photo.PhotoType != "WOWR" {
			continue
		}
		if photoID == "" || photo.RefPhotoID == photoID || (photo.RefPhotoID == "" && initialCount <= 1) {
			completion = append(completion, photo)
		}
	}
	return initial, completion
}

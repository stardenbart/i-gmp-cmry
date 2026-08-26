package inspectionusecase

import (
	"fmt"
	"log"

	"github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
	notificationdomain "github.com/monitoring-system/backend/internal/domain/notification"
	"github.com/monitoring-system/backend/internal/domain/pic"
	"github.com/monitoring-system/backend/pkg/mail"
)

type InspectionEmailNotifier struct {
	mailer         mail.Mailer
	picRepo        pic.PICMappingRepository
	authRepo       auth.UserRepository
	inspectionRepo inspection.InspectionHeaderRepository
	settingRepo    master.SettingRepository
	notificationUC notificationdomain.NotificationUseCase // in-app bell notifications, alongside email
}

func NewInspectionEmailNotifier(
	mailer mail.Mailer,
	picRepo pic.PICMappingRepository,
	authRepo auth.UserRepository,
	inspectionRepo inspection.InspectionHeaderRepository,
	settingRepo master.SettingRepository,
	notificationUC notificationdomain.NotificationUseCase,
) *InspectionEmailNotifier {
	return &InspectionEmailNotifier{
		mailer:         mailer,
		picRepo:        picRepo,
		authRepo:       authRepo,
		inspectionRepo: inspectionRepo,
		settingRepo:    settingRepo,
		notificationUC: notificationUC,
	}
}

// SendInspectionSummary is called when an Area's status becomes Confirmed
func (n *InspectionEmailNotifier) SendInspectionSummary(areaID string, progress inspection.AreaProgress) error {
	// 1. Find PICs for this Area (Manager, Auditor, or Area PIC)
	pics, err := n.picRepo.FindByAreaAndKawasan(areaID, "")
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving PICs for Area %s: %v", areaID, err)
		return err
	}

	var targetEmails []string
	var targetUserIDs []string
	emailSet := make(map[string]bool)
	userSet := make(map[string]bool)
	plantID := ""
	for _, p := range pics {
		user, err := n.authRepo.FindByID(p.UserID)
		if err == nil && user != nil {
			if plantID == "" && user.PlantID != nil {
				plantID = *user.PlantID
			}
			if user.Email != "" && !emailSet[user.Email] {
				emailSet[user.Email] = true
				targetEmails = append(targetEmails, user.Email)
			}
			if !userSet[user.UserID] {
				userSet[user.UserID] = true
				targetUserIDs = append(targetUserIDs, user.UserID)
			}
		}
	}

	// In-app bell notification, independent of whether the PIC has an email
	// on file (email delivery is best-effort further down).
	if n.notificationUC != nil {
		message := fmt.Sprintf("Inspeksi Area %s telah selesai (%d/%d detail kawasan).", areaID, progress.CompletedDetailKawasan, progress.TotalDetailKawasan)
		for _, userID := range targetUserIDs {
			_ = n.notificationUC.CreateSystemNotification(userID, "success", "Inspeksi Area Selesai", message, "/inspections")
		}
	}

	if len(targetEmails) == 0 {
		log.Printf("[EmailNotifier] No PIC found or missing emails for Area %s", areaID)
		return nil
	}

	// 2. Load the matching plant template (with global fallback).
	setting, err := n.settingRepo.FindByKey(master.SettingKeyEmailTemplateInspectionConfirmed, plantID)
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving template: %v", err)
		return err
	}
	templateStr := setting.SettingValue

	// 3. Prepare data for the template
	data := map[string]interface{}{
		"AreaID":                 areaID,
		"TotalDetailKawasan":     progress.TotalDetailKawasan,
		"CompletedDetailKawasan": progress.CompletedDetailKawasan,
		"Status":                 string(progress.Status),
	}

	// 4. Send Email via Mailer
	subject := "✅ Inspeksi Area Selesai: " + areaID
	err = n.mailer.SendTemplateForPlant(plantID, targetEmails, subject, templateStr, data)
	if err != nil {
		log.Printf("[EmailNotifier] Failed to send email to %v: %v", targetEmails, err)
		return err
	}

	log.Printf("[EmailNotifier] Successfully sent Area Confirmed email to %v", targetEmails)
	return nil
}

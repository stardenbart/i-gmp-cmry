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

// SendKawasanInspectionSummary is called when a Kawasan's status becomes
// Confirmed (every DetailKawasan under it completed this month). Only
// notifies users mapped as "Manager" for that Kawasan — narrower audience
// than SendInspectionSummary's Area-wide PIC list, matching the recipient
// this specific event is meant for.
func (n *InspectionEmailNotifier) SendKawasanInspectionSummary(kawasanID string, progress inspection.KawasanProgress) error {
	managers, err := n.picRepo.FindResponsibleUsers(kawasanID, "Manager")
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving Managers for Kawasan %s: %v", kawasanID, err)
		return err
	}

	var targetEmails []string
	var targetUserIDs []string
	emailSet := make(map[string]bool)
	plantID := ""
	for _, m := range managers {
		if plantID == "" && m.PlantID != nil {
			plantID = *m.PlantID
		}
		if m.Email != "" && !emailSet[m.Email] {
			emailSet[m.Email] = true
			targetEmails = append(targetEmails, m.Email)
		}
		targetUserIDs = append(targetUserIDs, m.UserID)
	}

	if n.notificationUC != nil {
		message := fmt.Sprintf("Kawasan %s telah selesai diinspeksi bulan ini (%d/%d detail kawasan).", kawasanID, progress.CompletedDetailKawasan, progress.TotalDetailKawasan)
		for _, userID := range targetUserIDs {
			_ = n.notificationUC.CreateSystemNotification(userID, "success", "Kawasan Selesai Diinspeksi", message, "/inspections")
		}
	}

	if len(targetEmails) == 0 {
		log.Printf("[EmailNotifier] No Manager found or missing emails for Kawasan %s", kawasanID)
		return nil
	}

	setting, err := n.settingRepo.FindByKey(master.SettingKeyEmailTemplateKawasanConfirmed, plantID)
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving Kawasan-confirmed template: %v", err)
		return err
	}

	data := map[string]interface{}{
		"KawasanID":              kawasanID,
		"TotalDetailKawasan":     progress.TotalDetailKawasan,
		"CompletedDetailKawasan": progress.CompletedDetailKawasan,
		"Status":                 string(progress.Status),
	}

	subject := "✅ Kawasan Selesai Diinspeksi: " + kawasanID
	if err := n.mailer.SendTemplateForPlant(plantID, targetEmails, subject, setting.SettingValue, data); err != nil {
		log.Printf("[EmailNotifier] Failed to send Kawasan-confirmed email to %v: %v", targetEmails, err)
		return err
	}

	log.Printf("[EmailNotifier] Successfully sent Kawasan Confirmed email to %v", targetEmails)
	return nil
}

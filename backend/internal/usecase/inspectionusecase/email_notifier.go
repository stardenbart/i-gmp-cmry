package inspectionusecase

import (
	"log"

	"github.com/monitoring-system/backend/internal/domain/auth"
	"github.com/monitoring-system/backend/internal/domain/inspection"
	"github.com/monitoring-system/backend/internal/domain/master"
	"github.com/monitoring-system/backend/internal/domain/pic"
	"github.com/monitoring-system/backend/pkg/mail"
)

type InspectionEmailNotifier struct {
	mailer         mail.Mailer
	picRepo        pic.PICMappingRepository
	authRepo       auth.UserRepository
	inspectionRepo inspection.InspectionHeaderRepository
	settingRepo    master.SettingRepository
}

func NewInspectionEmailNotifier(
	mailer mail.Mailer,
	picRepo pic.PICMappingRepository,
	authRepo auth.UserRepository,
	inspectionRepo inspection.InspectionHeaderRepository,
	settingRepo master.SettingRepository,
) *InspectionEmailNotifier {
	return &InspectionEmailNotifier{
		mailer:         mailer,
		picRepo:        picRepo,
		authRepo:       authRepo,
		inspectionRepo: inspectionRepo,
		settingRepo:    settingRepo,
	}
}

// SendInspectionSummary is called when an Area's status becomes Confirmed
func (n *InspectionEmailNotifier) SendInspectionSummary(areaID string, progress inspection.AreaProgress) error {
	// 1. Get the dynamic email template from Settings
	setting, err := n.settingRepo.FindByKey(master.SettingKeyEmailTemplateInspectionConfirmed)
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving template: %v", err)
		return err
	}
	templateStr := setting.SettingValue

	// 2. Find PICs for this Area (Manager, Auditor, or Area PIC)
	pics, err := n.picRepo.FindByAreaAndKawasan(areaID, "")
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving PICs for Area %s: %v", areaID, err)
		return err
	}

	var targetEmails []string
	emailSet := make(map[string]bool)
	for _, p := range pics {
		user, err := n.authRepo.FindByID(p.UserID)
		if err == nil && user != nil && user.Email != "" {
			if !emailSet[user.Email] {
				emailSet[user.Email] = true
				targetEmails = append(targetEmails, user.Email)
			}
		}
	}

	if len(targetEmails) == 0 {
		log.Printf("[EmailNotifier] No PIC found or missing emails for Area %s", areaID)
		return nil
	}

	// 3. Prepare data for the template
	data := map[string]interface{}{
		"AreaID":                 areaID,
		"TotalDetailKawasan":     progress.TotalDetailKawasan,
		"CompletedDetailKawasan": progress.CompletedDetailKawasan,
		"Status":                 string(progress.Status),
	}

	// 4. Send Email via Mailer
	subject := "✅ Inspeksi Area Selesai: " + areaID
	err = n.mailer.SendTemplate(targetEmails, subject, templateStr, data)
	if err != nil {
		log.Printf("[EmailNotifier] Failed to send email to %v: %v", targetEmails, err)
		return err
	}

	log.Printf("[EmailNotifier] Successfully sent Area Confirmed email to %v", targetEmails)
	return nil
}

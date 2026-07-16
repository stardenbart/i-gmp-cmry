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

	// 2. Find Manager PIC for this Area. 
	// The repository method FindByAreaAndKawasan handles an empty kawasanID to just get Area PICs.
	pics, err := n.picRepo.FindByAreaAndKawasan(areaID, "")
	if err != nil {
		log.Printf("[EmailNotifier] Error retrieving PICs for Area %s: %v", areaID, err)
		return err
	}

	var managerEmails []string
	for _, p := range pics {
		// Only target "Manager" PIC
		if p.KategoriPIC == "Manager" {
			user, err := n.authRepo.FindByID(p.UserID)
			if err == nil && user != nil {
				managerEmails = append(managerEmails, user.Email)
			}
		}
	}

	if len(managerEmails) == 0 {
		log.Printf("[EmailNotifier] No Manager PIC found or missing emails for Area %s", areaID)
		return nil
	}

	// 3. Prepare data for the template
	// In a real scenario, you could query Area Name and detailed inspection results to inject here.
	data := map[string]interface{}{
		"AreaID":                 areaID,
		"TotalDetailKawasan":     progress.TotalDetailKawasan,
		"CompletedDetailKawasan": progress.CompletedDetailKawasan,
	}

	// 4. Send Email via Mailer
	subject := "✅ Inspeksi Area Selesai: " + areaID
	err = n.mailer.SendTemplate(managerEmails, subject, templateStr, data)
	if err != nil {
		log.Printf("[EmailNotifier] Failed to send email to %v: %v", managerEmails, err)
		return err
	}

	log.Printf("[EmailNotifier] Successfully sent Area Confirmed email to %v", managerEmails)
	return nil
}

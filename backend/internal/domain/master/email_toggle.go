package master

import "strings"

// EmailTemplateEnabledKey is the on/off setting for emails built from an
// email template, e.g. EMAIL_TEMPLATE_ISSUE_ASSIGNMENT_ENABLED. It only
// switches the email; in-app notifications are sent regardless.
func EmailTemplateEnabledKey(templateKey string) string {
	return templateKey + "_ENABLED"
}

// ToggleableEmailTemplates are the templates admins can switch off in
// Pengaturan > Template Email. The password-reset OTP is not one of them.
var ToggleableEmailTemplates = []string{
	SettingKeyEmailTemplateIssue,
	SettingKeyEmailTemplateInspectionConfirmed,
	SettingKeyEmailTemplateDeadlineReminder,
}

// IsEmailTemplateEnabled reports whether emails for templateKey may be sent
// for plantID (plant override first, then global). A missing setting means
// enabled; only an explicit "false" switches the email off.
func IsEmailTemplateEnabled(repo SettingRepository, templateKey, plantID string) bool {
	if repo == nil {
		return true
	}
	setting, err := repo.FindByKey(EmailTemplateEnabledKey(templateKey), plantID)
	if err != nil || setting == nil {
		return true
	}
	return !strings.EqualFold(strings.TrimSpace(setting.SettingValue), "false")
}

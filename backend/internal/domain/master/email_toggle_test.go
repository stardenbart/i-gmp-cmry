package master

import (
	"errors"
	"testing"
)

type toggleRepo map[string]string // "KEY|PLANT" -> value; "KEY|" is global

func (r toggleRepo) FindAll(string) ([]Setting, error) { return nil, nil }
func (r toggleRepo) Update(string, string, string, string) error {
	return nil
}
func (r toggleRepo) FindByKey(key, plantID string) (*Setting, error) {
	if v, ok := r[key+"|"+plantID]; ok {
		return &Setting{SettingKey: key, SettingValue: v, PlantID: plantID}, nil
	}
	if v, ok := r[key+"|"]; ok {
		return &Setting{SettingKey: key, SettingValue: v}, nil
	}
	return nil, errors.New("not found")
}

func TestEmailTemplateEnabledKey(t *testing.T) {
	if got := EmailTemplateEnabledKey(SettingKeyEmailTemplateIssue); got != "EMAIL_TEMPLATE_ISSUE_ASSIGNMENT_ENABLED" {
		t.Fatalf("key = %q", got)
	}
}

func TestIsEmailTemplateEnabled(t *testing.T) {
	repo := toggleRepo{
		"EMAIL_TEMPLATE_ISSUE_ASSIGNMENT_ENABLED|":           "false",
		"EMAIL_TEMPLATE_DEADLINE_REMINDER_ENABLED|":          "true",
		"EMAIL_TEMPLATE_DEADLINE_REMINDER_ENABLED|PLT-SENTUL": "FALSE",
	}
	cases := []struct {
		key, plant string
		want       bool
	}{
		{SettingKeyEmailTemplateIssue, "", false},
		{SettingKeyEmailTemplateIssue, "PLT-SENTUL", false},           // global applies to the plant
		{SettingKeyEmailTemplateDeadlineReminder, "PLT-CICURUG", true}, // global true
		{SettingKeyEmailTemplateDeadlineReminder, "PLT-SENTUL", false}, // plant override, case-insensitive
		{SettingKeyEmailTemplateInspectionConfirmed, "PLT-SENTUL", true}, // missing → enabled
	}
	for _, tc := range cases {
		if got := IsEmailTemplateEnabled(repo, tc.key, tc.plant); got != tc.want {
			t.Errorf("IsEmailTemplateEnabled(%s, %q) = %v, want %v", tc.key, tc.plant, got, tc.want)
		}
	}
	if !IsEmailTemplateEnabled(nil, SettingKeyEmailTemplateIssue, "") {
		t.Error("no repository means enabled")
	}
}

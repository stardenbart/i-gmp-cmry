package masterusecase

import (
	"strings"
	"testing"

	"github.com/monitoring-system/backend/internal/domain/master"
)

func TestValidateDynamicSettingRequiresPasswordResetOTPVariables(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{
			name:  "valid raw HTML",
			value: `<p>Kode {{.OTP}} berlaku {{.OTPExpiryMinutes}} menit.</p>`,
		},
		{
			name:  "valid visual editor envelope",
			value: `{"version":1,"html":"<p>{{.OTP}} - {{.OTPExpiryMinutes}}</p>"}`,
		},
		{
			name:    "missing OTP",
			value:   `<p>Berlaku {{.OTPExpiryMinutes}} menit.</p>`,
			wantErr: true,
		},
		{
			name:    "missing expiry",
			value:   `<p>Kode {{.OTP}}</p>`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDynamicSetting(master.SettingKeyEmailTemplateForgotPass, tt.value)
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
			if tt.wantErr && !strings.Contains(err.Error(), "template OTP wajib") {
				t.Fatalf("unexpected validation message: %v", err)
			}
		})
	}
}

func TestValidateDynamicSettingEmailToggles(t *testing.T) {
	for _, key := range []string{
		"EMAIL_TEMPLATE_ISSUE_ASSIGNMENT_ENABLED",
		"EMAIL_TEMPLATE_INSPECTION_CONFIRMED_ENABLED",
		"EMAIL_TEMPLATE_DEADLINE_REMINDER_ENABLED",
	} {
		if err := validateDynamicSetting(key, "false"); err != nil {
			t.Errorf("%s=false rejected: %v", key, err)
		}
		if err := validateDynamicSetting(key, "maybe"); err == nil {
			t.Errorf("%s=maybe should be rejected", key)
		}
	}
}

func TestValidateDynamicSettingSenderName(t *testing.T) {
	for _, ok := range []string{"", "I-GMP Notification", "I-GMP — Plant Sentul"} {
		if err := validateDynamicSetting("SMTP_SENDER_NAME", ok); err != nil {
			t.Errorf("%q rejected: %v", ok, err)
		}
	}
	for _, bad := range []string{"I-GMP\r\nBcc: attacker@example.com", "line\nbreak", strings.Repeat("x", 101)} {
		if err := validateDynamicSetting("SMTP_SENDER_NAME", bad); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

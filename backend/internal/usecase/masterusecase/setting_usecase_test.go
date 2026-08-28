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

package mail

import "testing"

func TestDynamicSMTPMailerPassesPlantToProvider(t *testing.T) {
	requestedPlant := ""
	dynamic := NewDynamicSMTPMailer(SMTPConfig{}, func(plantID string) (SMTPConfig, error) {
		requestedPlant = plantID
		return SMTPConfig{Enabled: true, Host: "smtp.example.com", Port: 587, SenderEmail: "audit@example.com"}, nil
	})

	mailer, ok := dynamic.(*smtpMailer)
	if !ok {
		t.Fatal("expected smtpMailer implementation")
	}
	config, err := mailer.currentConfig("PLANT-001")
	if err != nil {
		t.Fatalf("resolve plant SMTP config: %v", err)
	}
	if requestedPlant != "PLANT-001" {
		t.Fatalf("expected provider plant PLANT-001, got %q", requestedPlant)
	}
	if config.Host != "smtp.example.com" {
		t.Fatalf("unexpected SMTP host %q", config.Host)
	}
}

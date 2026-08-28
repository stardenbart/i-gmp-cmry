package mail

import (
	"net"
	"strings"
	"testing"
	"time"
)

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

func TestResolveTemplateHTML(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "legacy raw HTML",
			value: `<p>Halo {{.FullName}}</p>`,
			want:  `<p>Halo {{.FullName}}</p>`,
		},
		{
			name:  "visual editor envelope",
			value: `{"version":1,"html":"<p>Halo {{.FullName}}</p>","design":{"body":{}}}`,
			want:  `<p>Halo {{.FullName}}</p>`,
		},
		{
			name:  "unrelated JSON remains unchanged",
			value: `{"message":"not an email template"}`,
			want:  `{"message":"not an email template"}`,
		},
		{
			name:  "malformed JSON remains unchanged",
			value: `{"html":`,
			want:  `{"html":`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveTemplateHTML(tt.value); got != tt.want {
				t.Fatalf("resolveTemplateHTML() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSendSMTPMailTimesOutWhenServerDoesNotGreet(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	release := make(chan struct{})
	defer close(release)
	go func() {
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			return
		}
		defer conn.Close()
		<-release
	}()

	started := time.Now()
	err = sendSMTPMail(
		listener.Addr().String(),
		"localhost",
		nil,
		"sender@example.com",
		[]string{"recipient@example.com"},
		[]byte("Subject: test\r\n\r\nbody"),
		75*time.Millisecond,
	)
	if err == nil {
		t.Fatal("expected an SMTP timeout error")
	}
	if !strings.Contains(err.Error(), "failed to initialize SMTP client") {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("SMTP timeout took too long: %v", elapsed)
	}
}

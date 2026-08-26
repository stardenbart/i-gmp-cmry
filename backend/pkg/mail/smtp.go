package mail

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/smtp"
	"strings"
	"text/template"
)

// Mailer defines the interface for sending emails.
type Mailer interface {
	Send(to []string, subject, body string) error
	SendTemplate(to []string, subject string, tmpl string, data interface{}) error
	SendForPlant(plantID string, to []string, subject, body string) error
	SendTemplateForPlant(plantID string, to []string, subject string, tmpl string, data interface{}) error
}

// SMTPConfig is the effective SMTP configuration used for one send attempt.
// A provider can reload it from System_Setting, allowing changes without an
// application restart. Environment values remain available as fallbacks.
type SMTPConfig struct {
	Enabled     bool
	Host        string
	Port        int
	User        string
	Password    string
	SenderEmail string
}

type SMTPConfigProvider func(plantID string) (SMTPConfig, error)

type smtpMailer struct {
	fallback SMTPConfig
	provider SMTPConfigProvider
}

// NewSMTPMailer creates a new instance of SMTP Mailer.
func NewSMTPMailer(host string, port int, user, password, senderEmail string) Mailer {
	return &smtpMailer{
		fallback: SMTPConfig{Enabled: true, Host: host, Port: port, User: user, Password: password, SenderEmail: senderEmail},
	}
}

// NewDynamicSMTPMailer creates a mailer that reloads SMTP settings for every
// email and plant. This makes changes from the Settings page effective immediately.
func NewDynamicSMTPMailer(fallback SMTPConfig, provider SMTPConfigProvider) Mailer {
	return &smtpMailer{fallback: fallback, provider: provider}
}

func (m *smtpMailer) currentConfig(plantID string) (SMTPConfig, error) {
	if m.provider == nil {
		return m.fallback, nil
	}
	return m.provider(plantID)
}

// Send sends a plain text email.
func (m *smtpMailer) Send(to []string, subject, body string) error {
	return m.SendForPlant("", to, subject, body)
}

func (m *smtpMailer) SendForPlant(plantID string, to []string, subject, body string) error {
	cfg, err := m.currentConfig(plantID)
	if err != nil {
		return fmt.Errorf("failed to load SMTP settings: %w", err)
	}
	if err := validateSMTPConfig(cfg); err != nil {
		return err
	}
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", cfg.SenderEmail, join(to), subject, body)

	var auth smtp.Auth
	if cfg.User != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return smtp.SendMail(addr, auth, cfg.SenderEmail, to, []byte(msg))
}

// SendTemplate sends an HTML email parsed from a template string.
func (m *smtpMailer) SendTemplate(to []string, subject string, tmpl string, data interface{}) error {
	return m.SendTemplateForPlant("", to, subject, tmpl, data)
}

func (m *smtpMailer) SendTemplateForPlant(plantID string, to []string, subject string, tmpl string, data interface{}) error {
	cfg, err := m.currentConfig(plantID)
	if err != nil {
		return fmt.Errorf("failed to load SMTP settings: %w", err)
	}
	if err := validateSMTPConfig(cfg); err != nil {
		return err
	}

	// Settings stores both the exported HTML and the editor design in a JSON
	// envelope. Only the HTML belongs in the outgoing email. Raw HTML remains
	// supported for seeded and legacy templates.
	templateHTML := resolveTemplateHTML(tmpl)
	t, err := template.New("email").Parse(templateHTML)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s\r\n%s", cfg.SenderEmail, join(to), subject, mimeHeaders, body.String())

	var auth smtp.Auth
	if cfg.User != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return smtp.SendMail(addr, auth, cfg.SenderEmail, to, []byte(msg))
}

type storedEmailTemplate struct {
	HTML string `json:"html"`
}

func resolveTemplateHTML(value string) string {
	trimmed := strings.TrimSpace(value)
	if !strings.HasPrefix(trimmed, "{") {
		return value
	}

	var stored storedEmailTemplate
	if err := json.Unmarshal([]byte(trimmed), &stored); err != nil || strings.TrimSpace(stored.HTML) == "" {
		return value
	}
	return stored.HTML
}

func validateSMTPConfig(cfg SMTPConfig) error {
	if !cfg.Enabled {
		return fmt.Errorf("SMTP is disabled in system settings")
	}
	if cfg.Host == "" || cfg.Port < 1 || cfg.Port > 65535 {
		return fmt.Errorf("SMTP host or port is invalid")
	}
	if cfg.SenderEmail == "" {
		return fmt.Errorf("SMTP sender email is required")
	}
	return nil
}

func join(s []string) string {
	res := ""
	for i, v := range s {
		if i > 0 {
			res += ", "
		}
		res += v
	}
	return res
}

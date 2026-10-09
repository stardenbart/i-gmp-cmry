package mail

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"text/template"
	"time"
)

const smtpOperationTimeout = 15 * time.Second

// Mailer defines the interface for sending emails.
type Mailer interface {
	Send(to []string, subject, body string) error
	SendTemplate(to []string, subject string, tmpl string, data interface{}) error
	SendForPlant(plantID string, to []string, subject, body string) error
	SendTemplateForPlant(plantID string, to []string, subject string, tmpl string, data interface{}) error
	SendHTMLForPlant(plantID string, msg HTMLMessage) error
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
	// SenderName is the display name in the From header ("Name" <email>).
	// Empty sends the bare address.
	SenderName string
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
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", fromHeader(cfg), join(to), subject, body)

	var auth smtp.Auth
	if cfg.User != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return sendSMTPMail(addr, cfg.Host, auth, cfg.SenderEmail, to, []byte(msg), smtpOperationTimeout)
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
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s\r\n%s", fromHeader(cfg), join(to), subject, mimeHeaders, body.String())

	var auth smtp.Auth
	if cfg.User != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)

	return sendSMTPMail(addr, cfg.Host, auth, cfg.SenderEmail, to, []byte(msg), smtpOperationTimeout)
}

// sendSMTPMail mirrors net/smtp.SendMail while applying one deadline to the
// connection, greeting, STARTTLS handshake, authentication, and message write.
// net/smtp.SendMail has no timeout and can otherwise leave an HTTP request
// hanging indefinitely when an SMTP hostname or firewall rule is incorrect.
func sendSMTPMail(addr, host string, auth smtp.Auth, from string, to []string, msg []byte, timeout time.Duration) error {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return fmt.Errorf("failed to set SMTP deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("failed to initialize SMTP client: %w", err)
	}
	defer client.Close()

	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			return fmt.Errorf("failed to start SMTP TLS: %w", err)
		}
	}
	if auth != nil {
		if ok, _ := client.Extension("AUTH"); !ok {
			return fmt.Errorf("SMTP server does not support authentication")
		}
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP authentication failed: %w", err)
		}
	}
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("SMTP sender was rejected: %w", err)
	}
	for _, recipient := range to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("SMTP recipient %q was rejected: %w", recipient, err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to start SMTP message: %w", err)
	}
	if _, err := writer.Write(msg); err != nil {
		_ = writer.Close()
		return fmt.Errorf("failed to write SMTP message: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to finish SMTP message: %w", err)
	}
	if err := client.Quit(); err != nil {
		return fmt.Errorf("failed to close SMTP session: %w", err)
	}
	return nil
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

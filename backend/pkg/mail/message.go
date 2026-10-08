package mail

import (
	"fmt"
	"mime"
	"net/smtp"
	"strings"
)

// HTMLMessage is a ready-rendered HTML email with optional CC recipients.
type HTMLMessage struct {
	To             []string
	Cc             []string
	Subject        string
	HTML           string
	HighImportance bool
}

// recipients is every SMTP RCPT address (To then Cc), case-insensitively
// de-duplicated, blanks dropped.
func (m HTMLMessage) recipients() []string {
	seen := make(map[string]bool)
	var out []string
	for _, list := range [][]string{m.To, m.Cc} {
		for _, addr := range list {
			addr = strings.TrimSpace(addr)
			key := strings.ToLower(addr)
			if addr == "" || seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, addr)
		}
	}
	return out
}

func buildHTMLMessage(from string, m HTMLMessage) []byte {
	var b strings.Builder
	fmt.Fprintf(&b, "From: %s\r\n", from)
	fmt.Fprintf(&b, "To: %s\r\n", strings.Join(m.To, ", "))
	if len(m.Cc) > 0 {
		fmt.Fprintf(&b, "Cc: %s\r\n", strings.Join(m.Cc, ", "))
	}
	fmt.Fprintf(&b, "Subject: %s\r\n", mime.QEncoding.Encode("UTF-8", m.Subject))
	if m.HighImportance {
		b.WriteString("Importance: High\r\nX-Priority: 1\r\n")
	}
	b.WriteString("MIME-Version: 1.0\r\nContent-Type: text/html; charset=\"UTF-8\"\r\n\r\n")
	b.WriteString(m.HTML)
	return []byte(b.String())
}

// SendHTMLForPlant sends a pre-rendered HTML email (To + Cc) using the
// plant's SMTP settings.
func (m *smtpMailer) SendHTMLForPlant(plantID string, msg HTMLMessage) error {
	cfg, err := m.currentConfig(plantID)
	if err != nil {
		return fmt.Errorf("failed to load SMTP settings: %w", err)
	}
	if err := validateSMTPConfig(cfg); err != nil {
		return err
	}

	var auth smtp.Auth
	if cfg.User != "" {
		auth = smtp.PlainAuth("", cfg.User, cfg.Password, cfg.Host)
	}
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	return sendSMTPMail(addr, cfg.Host, auth, cfg.SenderEmail, msg.recipients(), buildHTMLMessage(cfg.SenderEmail, msg), smtpOperationTimeout)
}

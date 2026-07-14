package mail

import (
	"bytes"
	"fmt"
	"net/smtp"
	"text/template"
)

// Mailer defines the interface for sending emails.
type Mailer interface {
	Send(to []string, subject, body string) error
	SendTemplate(to []string, subject string, tmpl string, data interface{}) error
}

type smtpMailer struct {
	host        string
	port        int
	user        string
	password    string
	senderEmail string
}

// NewSMTPMailer creates a new instance of SMTP Mailer.
func NewSMTPMailer(host string, port int, user, password, senderEmail string) Mailer {
	return &smtpMailer{
		host:        host,
		port:        port,
		user:        user,
		password:    password,
		senderEmail: senderEmail,
	}
}

// Send sends a plain text email.
func (m *smtpMailer) Send(to []string, subject, body string) error {
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n\r\n%s", m.senderEmail, join(to), subject, body)

	auth := smtp.PlainAuth("", m.user, m.password, m.host)
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	return smtp.SendMail(addr, auth, m.senderEmail, to, []byte(msg))
}

// SendTemplate sends an HTML email parsed from a template string.
func (m *smtpMailer) SendTemplate(to []string, subject string, tmpl string, data interface{}) error {
	t, err := template.New("email").Parse(tmpl)
	if err != nil {
		return fmt.Errorf("failed to parse email template: %w", err)
	}

	var body bytes.Buffer
	if err := t.Execute(&body, data); err != nil {
		return fmt.Errorf("failed to execute email template: %w", err)
	}

	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\n%s\r\n%s", m.senderEmail, join(to), subject, mimeHeaders, body.String())

	auth := smtp.PlainAuth("", m.user, m.password, m.host)
	addr := fmt.Sprintf("%s:%d", m.host, m.port)

	return smtp.SendMail(addr, auth, m.senderEmail, to, []byte(msg))
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

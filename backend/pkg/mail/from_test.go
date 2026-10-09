package mail

import (
	"net/mail"
	"strings"
	"testing"
)

func TestFromHeaderUsesSenderName(t *testing.T) {
	got := fromHeader(SMTPConfig{SenderName: "I-GMP Notification", SenderEmail: "noreply@cimory.com"})
	if got != `"I-GMP Notification" <noreply@cimory.com>` {
		t.Fatalf("From = %q", got)
	}
}

func TestFromHeaderWithoutNameIsBareAddress(t *testing.T) {
	if got := fromHeader(SMTPConfig{SenderEmail: "noreply@cimory.com"}); got != "noreply@cimory.com" {
		t.Fatalf("From = %q", got)
	}
}

func TestFromHeaderEncodesNonASCIIName(t *testing.T) {
	got := fromHeader(SMTPConfig{SenderName: "I-GMP — Plant Sentul", SenderEmail: "noreply@cimory.com"})
	if strings.Contains(got, "—") {
		t.Fatalf("non-ASCII name must be RFC 2047 encoded, got %q", got)
	}
	addr, err := mail.ParseAddress(got)
	if err != nil || addr.Name != "I-GMP — Plant Sentul" || addr.Address != "noreply@cimory.com" {
		t.Fatalf("parsed %+v (%v)", addr, err)
	}
}

func TestHTMLMessageCarriesSenderName(t *testing.T) {
	raw := string(buildHTMLMessage(fromHeader(SMTPConfig{SenderName: "I-GMP", SenderEmail: "noreply@cimory.com"}), HTMLMessage{To: []string{"a@cimory.com"}, Subject: "x", HTML: "y"}))
	if !strings.Contains(raw, "From: \"I-GMP\" <noreply@cimory.com>\r\n") {
		t.Fatalf("headers = %q", raw)
	}
}

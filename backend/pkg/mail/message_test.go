package mail

import (
	"mime"
	"strings"
	"testing"
)

func TestBuildHTMLMessageHeaders(t *testing.T) {
	raw := string(buildHTMLMessage("noreply@cimory.com", HTMLMessage{
		To:             []string{"pic1@cimory.com", "pic2@cimory.com"},
		Cc:             []string{"hera.narulita@cimory.com"},
		Subject:        "[I-GMP] Laporan Inspeksi GMP - Produksi — CMD 1 - 08 Okt 2026",
		HTML:           "<p>Halo</p>",
		HighImportance: true,
	}))

	head, body, ok := strings.Cut(raw, "\r\n\r\n")
	if !ok {
		t.Fatal("missing blank line between headers and body")
	}
	for _, want := range []string{
		"From: noreply@cimory.com",
		"To: pic1@cimory.com, pic2@cimory.com",
		"Cc: hera.narulita@cimory.com",
		"MIME-Version: 1.0",
		`Content-Type: text/html; charset="UTF-8"`,
		"Importance: High",
		"X-Priority: 1",
	} {
		if !strings.Contains(head, want) {
			t.Errorf("headers missing %q:\n%s", want, head)
		}
	}
	if body != "<p>Halo</p>" {
		t.Errorf("body = %q", body)
	}

	// Non-ASCII subjects (the em dash) must be RFC 2047 encoded.
	var subject string
	for _, line := range strings.Split(head, "\r\n") {
		if strings.HasPrefix(line, "Subject: ") {
			subject = strings.TrimPrefix(line, "Subject: ")
		}
	}
	decoded, err := new(mime.WordDecoder).DecodeHeader(subject)
	if err != nil || decoded != "[I-GMP] Laporan Inspeksi GMP - Produksi — CMD 1 - 08 Okt 2026" {
		t.Fatalf("subject %q decodes to %q (%v)", subject, decoded, err)
	}
}

func TestBuildHTMLMessageOmitsEmptyCcAndPriority(t *testing.T) {
	raw := string(buildHTMLMessage("noreply@cimory.com", HTMLMessage{To: []string{"a@cimory.com"}, Subject: "x", HTML: "y"}))
	for _, unwanted := range []string{"Cc:", "Importance:", "X-Priority:"} {
		if strings.Contains(raw, unwanted) {
			t.Errorf("unexpected %q in %q", unwanted, raw)
		}
	}
}

func TestHTMLMessageRecipientsIncludeCcOnce(t *testing.T) {
	got := HTMLMessage{
		To: []string{"pic1@cimory.com", "Shared@cimory.com"},
		Cc: []string{"shared@cimory.com", "hera.narulita@cimory.com", " "},
	}.recipients()
	want := []string{"pic1@cimory.com", "Shared@cimory.com", "hera.narulita@cimory.com"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Fatalf("recipients = %v, want %v", got, want)
	}
}

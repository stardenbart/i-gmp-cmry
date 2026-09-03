package dashboardhandler

import (
	"testing"
	"time"
)

func TestCalculateFollowUpGapDaysUsesJakartaCalendarDays(t *testing.T) {
	due := time.Date(2026, 9, 1, 17, 0, 0, 0, time.UTC)       // 2 Sep 00:00 WIB
	followUp := time.Date(2026, 9, 2, 18, 30, 0, 0, time.UTC) // 3 Sep 01:30 WIB

	gap := calculateFollowUpGapDays(&due, &followUp)
	if gap == nil || *gap != 1 {
		t.Fatalf("expected 1 calendar day late, got %v", gap)
	}
}

func TestCalculateFollowUpGapDaysPreservesEarlyAndOnTimeMeaning(t *testing.T) {
	due := time.Date(2026, 9, 10, 8, 0, 0, 0, time.UTC)
	early := time.Date(2026, 9, 8, 1, 0, 0, 0, time.UTC)
	onTime := time.Date(2026, 9, 10, 14, 0, 0, 0, time.UTC)

	if gap := calculateFollowUpGapDays(&due, &early); gap == nil || *gap != -2 {
		t.Fatalf("expected 2 days early, got %v", gap)
	}
	if gap := calculateFollowUpGapDays(&due, &onTime); gap == nil || *gap != 0 {
		t.Fatalf("expected same calendar day, got %v", gap)
	}
}

func TestCalculateFollowUpGapDaysRequiresBothDates(t *testing.T) {
	now := time.Now()
	if gap := calculateFollowUpGapDays(nil, &now); gap != nil {
		t.Fatalf("expected nil without due date, got %v", *gap)
	}
	if gap := calculateFollowUpGapDays(&now, nil); gap != nil {
		t.Fatalf("expected nil without follow-up date, got %v", *gap)
	}
}

func TestFollowUpImageURLsKeepsEveryEvidenceImageInOrder(t *testing.T) {
	evidence := []gmpFollowUpEvidence{
		{PhotoID: "PHOTO-1", ImageURL: "/uploads/follow-up-1.jpg", Keterangan: "Perbaikan pertama"},
		{PhotoID: "PHOTO-2", ImageURL: "", Keterangan: "Tanpa gambar"},
		{PhotoID: "PHOTO-3", ImageURL: "/uploads/follow-up-2.jpg", Keterangan: "Perbaikan kedua"},
	}

	images := followUpImageURLs(evidence)
	if len(images) != 2 || images[0] != "/uploads/follow-up-1.jpg" || images[1] != "/uploads/follow-up-2.jpg" {
		t.Fatalf("unexpected follow-up images: %#v", images)
	}
}

func TestInitialEvidenceKeepsOneDescriptionPerVisual(t *testing.T) {
	evidence := []gmpInitialEvidence{
		{PhotoID: "PHOTO-1", ImageURL: "/uploads/initial-1.jpg", Keterangan: "Lantai kotor"},
		{PhotoID: "PHOTO-2", ImageURL: "/uploads/initial-2.jpg", Keterangan: ""},
	}

	images := initialImageURLs(evidence)
	if len(images) != 2 || images[0] != "/uploads/initial-1.jpg" || images[1] != "/uploads/initial-2.jpg" {
		t.Fatalf("unexpected initial images: %#v", images)
	}
	if descriptions := initialEvidenceDescriptions(evidence, "fallback"); descriptions != "Lantai kotor\nTanpa keterangan" {
		t.Fatalf("unexpected initial descriptions: %q", descriptions)
	}
}

func TestInitialEvidenceDescriptionFallsBackWithoutVisual(t *testing.T) {
	if description := initialEvidenceDescriptions(nil, "Keterangan temuan lama"); description != "Keterangan temuan lama" {
		t.Fatalf("expected legacy fallback, got %q", description)
	}
}

func TestFormatGMPTemplateDescriptionsUsesSemicolonSeparator(t *testing.T) {
	got := formatGMPTemplateDescriptions("Lantai kotor\nSaluran kotor\nDinding bernoda")
	if got != "Lantai kotor; Saluran kotor; Dinding bernoda" {
		t.Fatalf("unexpected template description list: %q", got)
	}
}

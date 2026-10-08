package inspectionusecase

import (
	"strings"
	"testing"
)

func TestParseEmailListAcceptsCommonSeparators(t *testing.T) {
	got := parseEmailList("hera.narulita@cimory.com; risma.dwi@cimory.com,\nagus.mulyono@cimory.com ;; HERA.narulita@cimory.com  bukan-email ")
	want := "hera.narulita@cimory.com,risma.dwi@cimory.com,agus.mulyono@cimory.com"
	if strings.Join(got, ",") != want {
		t.Fatalf("got %v, want %s", got, want)
	}
}

func TestParseEmailListEmpty(t *testing.T) {
	if got := parseEmailList("   "); len(got) != 0 {
		t.Fatalf("got %v, want none", got)
	}
}

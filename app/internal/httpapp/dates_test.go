package httpapp

import (
	"testing"
	"time"
)

func TestParseBusinessDateUsesUTCMidnight(t *testing.T) {
	got, err := parseBusinessDate("2026-10-25")
	if err != nil {
		t.Fatal(err)
	}
	if got.Location() != time.UTC || got.Format("2006-01-02 15:04:05Z07:00") != "2026-10-25 00:00:00Z" {
		t.Fatalf("tanggal bisnis bergeser: %s", got.Format(time.RFC3339))
	}
}

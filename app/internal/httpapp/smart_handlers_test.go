package httpapp

import "testing"

func TestClockMinutes(t *testing.T) {
	tests := []struct {
		value string
		want  int
		valid bool
	}{
		{"00:00", 0, true},
		{"08:30", 510, true},
		{"23:59", 1439, true},
		{"24:00", 0, false},
		{"8:30", 0, false},
	}
	for _, tt := range tests {
		got, err := clockMinutes(tt.value)
		if tt.valid && (err != nil || got != tt.want) {
			t.Fatalf("clockMinutes(%q) = %d, %v; ingin %d", tt.value, got, err, tt.want)
		}
		if !tt.valid && err == nil {
			t.Fatalf("clockMinutes(%q) seharusnya ditolak", tt.value)
		}
	}
}

func TestSmartWorkflowTransitions(t *testing.T) {
	valid := []struct {
		name, from, to string
		check          func(string, string) bool
	}{
		{"reservasi disetujui", "pending", "approved", reservationTransition},
		{"reservasi selesai", "approved", "completed", reservationTransition},
		{"tamu masuk", "approved", "checked_in", visitorTransition},
		{"tamu keluar", "checked_in", "checked_out", visitorTransition},
		{"permintaan ditinjau", "submitted", "reviewing", residentRequestTransition},
		{"permintaan selesai", "approved", "completed", residentRequestTransition},
		{"darurat ditanggapi", "open", "acknowledged", emergencyTransition},
		{"darurat selesai", "acknowledged", "resolved", emergencyTransition},
	}
	for _, tt := range valid {
		if !tt.check(tt.from, tt.to) {
			t.Fatalf("transisi valid %s ditolak: %s -> %s", tt.name, tt.from, tt.to)
		}
	}

	invalid := []struct {
		from, to string
		check    func(string, string) bool
	}{
		{"completed", "approved", reservationTransition},
		{"pending", "checked_out", visitorTransition},
		{"rejected", "approved", residentRequestTransition},
		{"resolved", "open", emergencyTransition},
	}
	for _, tt := range invalid {
		if tt.check(tt.from, tt.to) {
			t.Fatalf("transisi tidak valid diterima: %s -> %s", tt.from, tt.to)
		}
	}
}

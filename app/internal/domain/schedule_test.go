package domain

import (
	"reflect"
	"testing"
	"time"
)

func TestAddMonthsAnchoredDoesNotDriftFrom31(t *testing.T) {
	location := time.FixedZone("WITA", 8*60*60)
	start := time.Date(2025, time.January, 31, 0, 0, 0, 0, location)

	february, err := AddMonthsAnchored(start, 1, 31)
	if err != nil {
		t.Fatal(err)
	}
	march, err := AddMonthsAnchored(start, 2, 31)
	if err != nil {
		t.Fatal(err)
	}

	if got, want := february.Format("2006-01-02"), "2025-02-28"; got != want {
		t.Fatalf("Februari: got %s want %s", got, want)
	}
	if got, want := march.Format("2006-01-02"), "2025-03-31"; got != want {
		t.Fatalf("Maret: got %s want %s", got, want)
	}
}

func TestAddMonthsAnchoredLeapYear(t *testing.T) {
	location := time.UTC
	start := time.Date(2024, time.January, 29, 0, 0, 0, 0, location)
	february, err := AddMonthsAnchored(start, 1, 29)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := february.Format("2006-01-02"), "2024-02-29"; got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestSplitInstallmentsPutsRemainderInLast(t *testing.T) {
	got, err := SplitInstallments(10_000_001, 4)
	if err != nil {
		t.Fatal(err)
	}
	want := []int64{2_500_000, 2_500_000, 2_500_000, 2_500_001}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %v want %v", got, want)
	}
}

func TestBuildMonthlyLeasePlanKeepsAnchorAndDueFallback(t *testing.T) {
	start := time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.April, 30, 0, 0, 0, 0, time.UTC)
	charges, revenue, err := BuildLeasePlan(start, end, 31, 31, "monthly", 1_200_000, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(charges) != 3 || len(revenue) != 3 {
		t.Fatalf("got charges=%d revenue=%d", len(charges), len(revenue))
	}
	if got := charges[1].DueDate.Format("2006-01-02"); got != "2025-02-28" {
		t.Fatalf("jatuh tempo Februari got %s", got)
	}
	if got := charges[2].DueDate.Format("2006-01-02"); got != "2025-03-31" {
		t.Fatalf("jatuh tempo Maret drift: %s", got)
	}
}

func TestBuildAnnualPlanBalancesInvoicesAndRevenue(t *testing.T) {
	start := time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	charges, revenue, err := BuildLeasePlan(start, end, 15, 10, "annual", 12_000_001, 4)
	if err != nil {
		t.Fatal(err)
	}
	var billed, recognized int64
	for _, item := range charges {
		billed += item.Amount
	}
	for _, item := range revenue {
		recognized += item.Amount
	}
	if billed != 12_000_001 || recognized != billed || len(charges) != 4 || len(revenue) != 12 {
		t.Fatalf("billed=%d recognized=%d charges=%d revenue=%d", billed, recognized, len(charges), len(revenue))
	}
}

func TestAnnualPlanRejectsNonTwelveMonthPeriod(t *testing.T) {
	start := time.Date(2025, time.January, 31, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.July, 31, 0, 0, 0, 0, time.UTC)
	if _, _, err := BuildLeasePlan(start, end, 31, 10, "annual", 12_000_000, 2); err == nil {
		t.Fatal("periode tahunan selain 12 bulan harus ditolak")
	}
}

func TestMonthlyPlanRejectsAnnualInstallmentSelection(t *testing.T) {
	start := time.Date(2025, time.January, 15, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, time.April, 15, 0, 0, 0, 0, time.UTC)
	if _, _, err := BuildLeasePlan(start, end, 15, 10, "monthly", 1_000_000, 2); err == nil {
		t.Fatal("kontrak bulanan dengan cicilan tahunan harus ditolak")
	}
}

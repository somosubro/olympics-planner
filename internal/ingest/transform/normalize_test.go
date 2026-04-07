package transform_test

import (
	"testing"

	"olympics-planner/internal/ingest/transform"
)

func TestNormalizeSport_PrefixSALWhenSportColumnEmpty(t *testing.T) {
	// When pdftotext drops the first column, we still have session code SAL01.
	if got := transform.NormalizeSport("", "SAL01"); got != "Sailing" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeSport_UsesColumnWhenPresent(t *testing.T) {
	if got := transform.NormalizeSport("Sailing (Windsurfing & Kite)", "SAL01"); got != "Sailing (Windsurfing & Kite)" {
		t.Fatalf("got %q", got)
	}
}

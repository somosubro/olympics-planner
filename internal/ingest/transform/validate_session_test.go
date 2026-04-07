package transform_test

import (
	"testing"

	"olympics-planner/internal/domain"
	"olympics-planner/internal/ingest/transform"
)

func TestSessionValidationIssues_MinimalValid(t *testing.T) {
	s := domain.Session{
		ID:          "TEN12",
		Sport:       "Tennis",
		SessionCode: "TEN12",
		Date:        "2028-07-15",
		DayOfWeek:   "Saturday",
		StartTime:   "14:00",
		Venue:       "LA Tennis Center",
	}
	if issues := transform.SessionValidationIssues(s); len(issues) != 0 {
		t.Fatalf("expected valid, got %v", issues)
	}
}

func TestSessionValidationIssues_MissingVenue(t *testing.T) {
	s := domain.Session{
		ID:          "TEN12",
		Sport:       "Tennis",
		SessionCode: "TEN12",
		Date:        "2028-07-15",
		DayOfWeek:   "Saturday",
		StartTime:   "14:00",
	}
	if issues := transform.SessionValidationIssues(s); len(issues) == 0 {
		t.Fatal("expected issues")
	}
}

func TestSessionValidationIssues_UnknownSport(t *testing.T) {
	s := domain.Session{
		ID:          "XXX01",
		Sport:       "Unknown",
		SessionCode: "XXX01",
		Date:        "2028-07-15",
		DayOfWeek:   "Saturday",
		StartTime:   "14:00",
		Venue:       "V",
	}
	if issues := transform.SessionValidationIssues(s); len(issues) == 0 {
		t.Fatal("expected unknown sport rejected")
	}
}

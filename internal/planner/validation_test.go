package planner

import (
	"testing"

	"olympics-planner/internal/domain"
)

func TestValidatePlan_RejectsRepeatedSportAcrossDays(t *testing.T) {
	sessions := []domain.Session{
		{ID: "a", Sport: "Tennis", SessionCode: "T1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "10:00", Venue: "V1"},
		{ID: "b", Sport: "Tennis", SessionCode: "T2", Date: "2028-07-16", DayOfWeek: "Sunday", StartTime: "10:00", Venue: "V1"},
	}

	prefs := domain.Preferences{
		AllowedSports: []string{"Tennis"},
		AllowedDays:   []string{"Saturday", "Sunday"},
		Rules:         domain.Rules{}, // omitted noSameSportAcrossDays → default on
	}

	plan := domain.Plan{
		PlanType: domain.PlanTypeTwoDay,
		Days: []domain.PlanDay{
			{Date: "2028-07-15", DayOfWeek: "Saturday", PrimarySessionID: "a", AlternateSessionIDs: nil},
			{Date: "2028-07-16", DayOfWeek: "Sunday", PrimarySessionID: "b", AlternateSessionIDs: nil},
		},
	}

	result := ValidatePlan(plan, sessions, prefs)
	if result.Valid {
		t.Fatalf("expected plan to be invalid")
	}
	if !hasCode(result.Errors, "REPEATED_SPORT_ACROSS_DAYS") {
		t.Fatalf("expected REPEATED_SPORT_ACROSS_DAYS, got %#v", result.Errors)
	}
}

func TestValidatePlan_SessionNotFound(t *testing.T) {
	sessions := []domain.Session{
		{ID: "a", Sport: "Tennis", SessionCode: "T1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "10:00", Venue: "V1"},
	}
	prefs := domain.Preferences{
		AllowedSports: []string{"Tennis"},
		AllowedDays:   []string{"Saturday"},
		Rules:         domain.Rules{NoSameSportAcrossDays: boolPtr(false)},
	}
	plan := domain.Plan{
		PlanType: domain.PlanTypeOneDay,
		Days: []domain.PlanDay{
			{Date: "2028-07-15", DayOfWeek: "Saturday", PrimarySessionID: "missing", AlternateSessionIDs: nil},
		},
	}
	result := ValidatePlan(plan, sessions, prefs)
	if result.Valid {
		t.Fatal("expected invalid")
	}
	if !hasCode(result.Errors, "SESSION_NOT_FOUND") {
		t.Fatalf("expected SESSION_NOT_FOUND, got %#v", result.Errors)
	}
}

func TestValidatePlan_StructuredErrorsNotStrings(t *testing.T) {
	result := ValidatePlan(domain.Plan{}, nil, domain.Preferences{})
	if result.Valid {
		t.Fatal("expected invalid")
	}
	if len(result.Errors) == 0 {
		t.Fatal("expected errors")
	}
	e := result.Errors[0]
	if e.Code == "" || e.Message == "" {
		t.Fatalf("expected code and message, got %#v", e)
	}
}

func hasCode(errs []domain.ValidationError, code string) bool {
	for _, e := range errs {
		if e.Code == code {
			return true
		}
	}
	return false
}

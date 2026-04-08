package planner

import (
	"testing"

	"olympics-planner/internal/domain"
)

func TestValidatePlan_RejectsRepeatedSportAcrossDays(t *testing.T) {
	sessions := []domain.Session{
		{ID: "a", Sport: "Tennis", SessionCode: "T1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "10:00", EndTime: "12:00", Venue: "V1"},
		{ID: "b", Sport: "Tennis", SessionCode: "T2", Date: "2028-07-16", DayOfWeek: "Sunday", StartTime: "10:00", EndTime: "12:00", Venue: "V1"},
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

func TestValidatePlan_RejectsSameSportPrimaryOneDayAlternateOtherDay(t *testing.T) {
	// Cricket on day 1 primary; cricket again on day 2 as an alternate — must fail when rule is on.
	sessions := []domain.Session{
		{ID: "ckt-d1", Sport: "Cricket", SessionCode: "CK1", Date: "2028-07-22", DayOfWeek: "Saturday", StartTime: "09:00", EndTime: "12:00", Venue: "Pomona"},
		{ID: "ten-d2", Sport: "Tennis", SessionCode: "TN1", Date: "2028-07-23", DayOfWeek: "Sunday", StartTime: "11:00", EndTime: "13:00", Venue: "Carson"},
		{ID: "ckt-d2", Sport: "Cricket", SessionCode: "CK2", Date: "2028-07-23", DayOfWeek: "Sunday", StartTime: "18:00", EndTime: "21:00", Venue: "Pomona"},
	}
	prefs := domain.Preferences{
		AllowedSports: []string{"Cricket", "Tennis"},
		AllowedDays:   []string{"Saturday", "Sunday"},
		Rules:         domain.Rules{},
	}
	plan := domain.Plan{
		PlanType: domain.PlanTypeTwoDay,
		Days: []domain.PlanDay{
			{Date: "2028-07-22", DayOfWeek: "Saturday", PrimarySessionID: "ckt-d1", AlternateSessionIDs: nil},
			{Date: "2028-07-23", DayOfWeek: "Sunday", PrimarySessionID: "ten-d2", AlternateSessionIDs: []string{"ckt-d2"}},
		},
	}
	result := ValidatePlan(plan, sessions, prefs)
	if result.Valid {
		t.Fatalf("expected plan invalid: cricket repeats across days via alternate")
	}
	if !hasCode(result.Errors, "REPEATED_SPORT_ACROSS_DAYS") {
		t.Fatalf("expected REPEATED_SPORT_ACROSS_DAYS, got %#v", result.Errors)
	}
}

func TestValidatePlan_SessionNotFound(t *testing.T) {
	sessions := []domain.Session{
		{ID: "a", Sport: "Tennis", SessionCode: "T1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "10:00", EndTime: "12:00", Venue: "V1"},
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

func TestValidatePlan_SessionIDsMode_TwoSportsSameDay(t *testing.T) {
	sessions := []domain.Session{
		{ID: "equ-a", Sport: "Equestrian", SessionCode: "E1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "09:00", EndTime: "11:45", Venue: "V1"},
		{ID: "ckt-b", Sport: "Cricket", SessionCode: "C1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "18:00", EndTime: "21:30", Venue: "V2"},
		{ID: "ten-c", Sport: "Tennis", SessionCode: "T1", Date: "2028-07-16", DayOfWeek: "Sunday", StartTime: "11:00", EndTime: "13:00", Venue: "V3"},
	}
	prefs := domain.Preferences{
		AllowedSports: []string{"Equestrian", "Cricket", "Tennis"},
		AllowedDays:   []string{"Saturday", "Sunday"},
		Rules:         domain.Rules{},
	}
	plan := domain.Plan{
		PlanType: domain.PlanTypeTwoDay,
		Days: []domain.PlanDay{
			{
				Date:       "2028-07-15",
				DayOfWeek:  "Saturday",
				SessionIDs: []string{"equ-a", "ckt-b"},
			},
			{
				Date:             "2028-07-16",
				DayOfWeek:        "Sunday",
				PrimarySessionID: "ten-c",
			},
		},
	}
	result := ValidatePlan(plan, sessions, prefs)
	if !result.Valid {
		t.Fatalf("expected valid plan with sessionIds, got %#v", result.Errors)
	}
}

func TestValidatePlan_ConflictingDaySpec_SessionIdsWithPrimary(t *testing.T) {
	sessions := []domain.Session{
		{ID: "a", Sport: "Tennis", SessionCode: "T1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "10:00", EndTime: "12:00", Venue: "V1"},
	}
	prefs := domain.Preferences{
		AllowedSports: []string{"Tennis"},
		AllowedDays:   []string{"Saturday"},
		Rules:         domain.Rules{NoSameSportAcrossDays: boolPtr(false)},
	}
	plan := domain.Plan{
		PlanType: domain.PlanTypeOneDay,
		Days: []domain.PlanDay{
			{
				Date:             "2028-07-15",
				DayOfWeek:        "Saturday",
				PrimarySessionID: "a",
				SessionIDs:       []string{"a"},
			},
		},
	}
	result := ValidatePlan(plan, sessions, prefs)
	if result.Valid {
		t.Fatal("expected invalid")
	}
	if !hasCode(result.Errors, "CONFLICTING_DAY_SPEC") {
		t.Fatalf("expected CONFLICTING_DAY_SPEC, got %#v", result.Errors)
	}
}

func TestValidatePlan_InsufficientSameDayGap_DefaultFourHours(t *testing.T) {
	sessions := []domain.Session{
		{ID: "x", Sport: "Tennis", SessionCode: "X1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "09:00", EndTime: "11:00", Venue: "V1"},
		{ID: "y", Sport: "Tennis", SessionCode: "X2", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "12:30", EndTime: "14:00", Venue: "V1"},
	}
	prefs := domain.Preferences{
		AllowedSports: []string{"Tennis"},
		AllowedDays:   []string{"Saturday"},
		Rules:         domain.Rules{NoSameSportAcrossDays: boolPtr(false)},
	}
	plan := domain.Plan{
		PlanType: domain.PlanTypeOneDay,
		Days: []domain.PlanDay{
			{Date: "2028-07-15", DayOfWeek: "Saturday", SessionIDs: []string{"x", "y"}},
		},
	}
	result := ValidatePlan(plan, sessions, prefs)
	if result.Valid {
		t.Fatal("expected invalid: gap 1.5h < 4h")
	}
	if !hasCode(result.Errors, "INSUFFICIENT_SAME_DAY_GAP") {
		t.Fatalf("expected INSUFFICIENT_SAME_DAY_GAP, got %#v", result.Errors)
	}
}

func TestValidatePlan_SameDaySpacingDisabledWithZero(t *testing.T) {
	sessions := []domain.Session{
		{ID: "x", Sport: "Tennis", SessionCode: "X1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "09:00", EndTime: "11:00", Venue: "V1"},
		{ID: "y", Sport: "Tennis", SessionCode: "X2", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "12:30", EndTime: "14:00", Venue: "V1"},
	}
	prefs := domain.Preferences{
		AllowedSports: []string{"Tennis"},
		AllowedDays:   []string{"Saturday"},
		Rules: domain.Rules{
			NoSameSportAcrossDays:          boolPtr(false),
			MinHoursBetweenSameDaySessions: float64Ptr(0),
		},
	}
	plan := domain.Plan{
		PlanType: domain.PlanTypeOneDay,
		Days: []domain.PlanDay{
			{Date: "2028-07-15", DayOfWeek: "Saturday", SessionIDs: []string{"x", "y"}},
		},
	}
	result := ValidatePlan(plan, sessions, prefs)
	if !result.Valid {
		t.Fatalf("expected valid when spacing disabled, got %#v", result.Errors)
	}
}

func TestValidatePlan_TooManySessionsPerDay(t *testing.T) {
	sessions := []domain.Session{
		{ID: "x", Sport: "Tennis", SessionCode: "X1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "09:00", EndTime: "11:00", Venue: "V1"},
		{ID: "y", Sport: "Tennis", SessionCode: "X2", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "18:00", EndTime: "20:00", Venue: "V1"},
	}
	prefs := domain.Preferences{
		AllowedSports: []string{"Tennis"},
		AllowedDays:   []string{"Saturday"},
		Rules: domain.Rules{
			NoSameSportAcrossDays: boolPtr(false),
			MaxSessionsPerDay:     intPtr(1),
		},
	}
	plan := domain.Plan{
		PlanType: domain.PlanTypeOneDay,
		Days: []domain.PlanDay{
			{Date: "2028-07-15", DayOfWeek: "Saturday", SessionIDs: []string{"x", "y"}},
		},
	}
	result := ValidatePlan(plan, sessions, prefs)
	if result.Valid {
		t.Fatal("expected invalid")
	}
	if !hasCode(result.Errors, "TOO_MANY_SESSIONS_PER_DAY") {
		t.Fatalf("expected TOO_MANY_SESSIONS_PER_DAY, got %#v", result.Errors)
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

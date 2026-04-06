package planner

import (
    "testing"

    "olympics-planner/internal/domain"
)

func TestValidatePlanRejectsRepeatedSportAcrossDays(t *testing.T) {
    sessions := []domain.Session{
        {ID: "a", Sport: "Tennis"},
        {ID: "b", Sport: "Tennis"},
    }

    prefs := domain.Preferences{
        AllowedSports: []string{"Tennis"},
        AllowedDays:   []string{"Saturday", "Sunday"},
        Rules: domain.Rules{
            NoSameSportAcrossDays: true,
        },
    }

    plan := domain.Plan{
        Sessions: []domain.PlannedSession{
            {Day: "Saturday", SessionID: "a"},
            {Day: "Sunday", SessionID: "b"},
        },
    }

    result := ValidatePlan(plan, sessions, prefs)
    if result.Valid {
        t.Fatalf("expected plan to be invalid")
    }
}

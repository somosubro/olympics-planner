package planner

import (
    "testing"

    "olympics-planner/internal/domain"
)

func TestRankSessionsPrefersHigherPrioritySport(t *testing.T) {
    sessions := []domain.Session{
        {ID: "1", Sport: "Swimming", DayOfWeek: "Saturday"},
        {ID: "2", Sport: "Tennis", DayOfWeek: "Saturday"},
    }

    prefs := domain.Preferences{
        SportPriority: []string{"Tennis", "Swimming"},
    }

    scored := RankSessions(sessions, prefs)

    if scored[1].Score <= scored[0].Score {
        t.Fatalf("expected tennis to score higher than swimming")
    }
}

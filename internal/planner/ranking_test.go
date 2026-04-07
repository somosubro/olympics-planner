package planner

import (
	"testing"

	"olympics-planner/internal/domain"
)

func TestRankSessionsPrefersHigherPrioritySport(t *testing.T) {
	sessions := []domain.Session{
		{ID: "1", Sport: "Swimming", SessionCode: "S1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "10:00", Venue: "V"},
		{ID: "2", Sport: "Tennis", SessionCode: "T2", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "10:00", Venue: "V"},
	}

	prefs := domain.Preferences{
		AllowedSports: []string{"Tennis", "Swimming"},
		AllowedDays:   []string{"Saturday"},
		SportPriority: []string{"Tennis", "Swimming"},
		Rules:         domain.Rules{NoSameSportAcrossDays: boolPtr(false)},
	}

	scored := RankSessions(sessions, prefs, false)

	if len(scored) != 2 {
		t.Fatalf("got %d ranked", len(scored))
	}
	if scored[0].Session.Sport != "Tennis" {
		t.Fatalf("expected tennis first, got %#v", scored[0].Session.Sport)
	}
	if scored[0].Score <= scored[1].Score {
		t.Fatalf("expected tennis to score higher than swimming")
	}
}

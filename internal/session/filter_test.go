package session

import (
	"testing"

	"olympics-planner/internal/domain"
)

func TestApplyFilter_DateAndSport(t *testing.T) {
	sessions := []domain.Session{
		{ID: "a", Sport: "Tennis", Date: "2028-07-15", DayOfWeek: "Saturday"},
		{ID: "b", Sport: "Tennis", Date: "2028-07-16", DayOfWeek: "Sunday"},
	}
	f := Filter{
		Dates:  []string{"2028-07-15"},
		Sports: []string{"Tennis"},
	}
	out := ApplyFilter(sessions, f)
	if len(out) != 1 || out[0].ID != "a" {
		t.Fatalf("got %#v", out)
	}
}

func TestApplyFilter_ExcludedSports(t *testing.T) {
	sessions := []domain.Session{
		{ID: "a", Sport: "Cricket"},
		{ID: "b", Sport: "Tennis"},
	}
	f := Filter{ExcludedSports: []string{"Cricket"}}
	out := ApplyFilter(sessions, f)
	if len(out) != 1 || out[0].ID != "b" {
		t.Fatalf("got %#v", out)
	}
}

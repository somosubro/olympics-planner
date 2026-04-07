package session

import (
	"olympics-planner/internal/domain"
)

// Filter holds GET /api/v1/sessions query dimensions (docs/api-spec.md §10).
// Empty slices mean "no constraint" for that dimension.
type Filter struct {
	Dates          []string
	DaysOfWeek     []string
	Sports         []string
	AllowedSports  []string
	ExcludedSports []string
}

// Matches returns whether s passes all non-empty filter dimensions (conjunctive; OR within each).
func Matches(s domain.Session, f Filter) bool {
	if len(f.Dates) > 0 && !contains(f.Dates, s.Date) {
		return false
	}
	if len(f.DaysOfWeek) > 0 && !contains(f.DaysOfWeek, s.DayOfWeek) {
		return false
	}
	if len(f.ExcludedSports) > 0 && contains(f.ExcludedSports, s.Sport) {
		return false
	}
	if len(f.Sports) > 0 && !contains(f.Sports, s.Sport) {
		return false
	}
	if len(f.AllowedSports) > 0 && !contains(f.AllowedSports, s.Sport) {
		return false
	}
	return true
}

// ApplyFilter applies f to all sessions and returns matching rows (stable order).
func ApplyFilter(sessions []domain.Session, f Filter) []domain.Session {
	out := make([]domain.Session, 0)
	for _, s := range sessions {
		if Matches(s, f) {
			out = append(out, s)
		}
	}
	return out
}

func contains(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

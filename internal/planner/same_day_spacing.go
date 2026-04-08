package planner

import (
	"fmt"
	"sort"
	"strings"

	"olympics-planner/internal/domain"
)

// sameDaySpacingErrors checks consecutive sessions (by start time) on the same day:
// no overlap, and gap from previous end to next start >= minHours.
func sameDaySpacingErrors(ids []string, byID map[string]domain.Session, minHours float64, dayField string) []domain.ValidationError {
	if len(ids) < 2 || minHours <= 0 {
		return nil
	}
	sess := make([]domain.Session, 0, len(ids))
	for _, id := range ids {
		id = strings.TrimSpace(id)
		s, ok := byID[id]
		if !ok {
			return nil
		}
		sess = append(sess, s)
	}
	sort.Slice(sess, func(i, j int) bool {
		mi, oki := clockMinutes(sess[i].StartTime)
		mj, okj := clockMinutes(sess[j].StartTime)
		if !oki || !okj {
			return false
		}
		return mi < mj
	})
	var errs []domain.ValidationError
	for i := 0; i < len(sess)-1; i++ {
		prev, next := sess[i], sess[i+1]
		endPrev, ok1 := clockMinutes(prev.EndTime)
		startNext, ok2 := clockMinutes(next.StartTime)
		if !ok1 || !ok2 {
			return []domain.ValidationError{{
				Code:    "INCOMPLETE_SESSION_TIME",
				Message: "session start and end time required for same-day spacing checks",
				Field:   dayField,
			}}
		}
		if startNext < endPrev {
			errs = append(errs, domain.ValidationError{
				Code:    "SAME_DAY_SESSION_OVERLAP",
				Message: fmt.Sprintf("sessions overlap in time (%s and %s)", prev.ID, next.ID),
				Field:   dayField,
			})
			return errs
		}
		gapHours := float64(startNext-endPrev) / 60.0
		if gapHours+1e-9 < minHours {
			errs = append(errs, domain.ValidationError{
				Code:    "INSUFFICIENT_SAME_DAY_GAP",
				Message: fmt.Sprintf("need at least %.2f h between end of one session and start of the next (got %.2f h)", minHours, gapHours),
				Field:   dayField,
			})
			return errs
		}
	}
	return errs
}

func clockMinutes(t string) (int, bool) {
	h, m, ok := parseHHMM(strings.TrimSpace(t))
	if !ok {
		return 0, false
	}
	return h*60 + m, true
}

package domain

import "strings"

// Plan matches docs/data-contract.md §11 (canonical runtime/API plan).
type Plan struct {
	PlanType PlanType  `json:"planType"`
	Days     []PlanDay `json:"days"`
}

// PlanType is documented in data-contract §11.5–11.6.
type PlanType string

const (
	PlanTypeOneDay   PlanType = "one_day"
	PlanTypeTwoDay   PlanType = "two_day"
	PlanTypeMultiDay PlanType = "multi_day"
)

// PlanDay matches docs/data-contract.md §11.7.
// Use either legacy shape (primarySessionId + optional alternateSessionIds) or
// sessionIds (non-empty list of co-equal same-day sessions). Do not combine both on one day.
type PlanDay struct {
	Date                string   `json:"date"`
	DayOfWeek           string   `json:"dayOfWeek"`
	PrimarySessionID    string   `json:"primarySessionId,omitempty"`
	AlternateSessionIDs []string `json:"alternateSessionIds,omitempty"`
	SessionIDs          []string `json:"sessionIds,omitempty"`
}

// EffectiveSessionIDs lists every session id scheduled for this calendar day.
// When sessionIds is non-empty it is authoritative; otherwise primary followed by alternates.
func (d PlanDay) EffectiveSessionIDs() []string {
	if len(d.SessionIDs) > 0 {
		out := make([]string, 0, len(d.SessionIDs))
		for _, id := range d.SessionIDs {
			if t := strings.TrimSpace(id); t != "" {
				out = append(out, t)
			}
		}
		return out
	}
	var out []string
	if t := strings.TrimSpace(d.PrimarySessionID); t != "" {
		out = append(out, t)
	}
	for _, id := range d.AlternateSessionIDs {
		if t := strings.TrimSpace(id); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// FirstSessionIDForScoring returns the first session id used for variety and tie-break ordering.
func (d PlanDay) FirstSessionIDForScoring() string {
	if len(d.SessionIDs) > 0 {
		for _, id := range d.SessionIDs {
			if t := strings.TrimSpace(id); t != "" {
				return t
			}
		}
		return ""
	}
	return strings.TrimSpace(d.PrimarySessionID)
}

// UsesSessionIDs reports whether this day uses the sessionIds list (co-equal same-day sessions).
func (d PlanDay) UsesSessionIDs() bool {
	return len(d.SessionIDs) > 0
}

// TotalSessionCount returns the number of session slots across all days (primary + alternates or sessionIds).
func (p Plan) TotalSessionCount() int {
	n := 0
	for _, d := range p.Days {
		n += len(d.EffectiveSessionIDs())
	}
	return n
}

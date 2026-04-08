package domain

type Preferences struct {
	AllowedSports []string `json:"allowedSports"`
	SportPriority []string `json:"sportPriority"`
	AllowedDays   []string `json:"allowedDays"`
	Rules         Rules    `json:"rules"`
}

type Rules struct {
	// NoSameSportAcrossDays is nil or true by default (one sport per calendar day across the plan).
	// Only an explicit JSON false opts out (e.g. user wants the same sport on multiple days).
	NoSameSportAcrossDays *bool                  `json:"noSameSportAcrossDays,omitempty"`
	PreferDayPairs        [][]string             `json:"preferDayPairs,omitempty"`
	SportSpecific         map[string]interface{} `json:"sportSpecific,omitempty"`

	// MinHoursBetweenSameDaySessions is the minimum hours from one session's end to the next session's start
	// on the same calendar day (sessions ordered by start time). Omitted → default 4 (LA-area travel cushion).
	// Explicit 0 disables this check.
	MinHoursBetweenSameDaySessions *float64 `json:"minHoursBetweenSameDaySessions,omitempty"`

	// MaxSessionsPerDay caps how many sessions may appear on a single calendar day (primary + alternates, or sessionIds length).
	// Omitted or 0 → no cap. Use 1 when the user wants at most one event per day.
	MaxSessionsPerDay *int `json:"maxSessionsPerDay,omitempty"`
}

// EffectiveNoSameSportAcrossDays is true when the field is omitted or set true; false only when explicitly false.
func (r Rules) EffectiveNoSameSportAcrossDays() bool {
	if r.NoSameSportAcrossDays == nil {
		return true
	}
	return *r.NoSameSportAcrossDays
}

// EffectiveMinHoursBetweenSameDaySessions returns the minimum gap (hours) from one session's end to the next's start
// when multiple sessions fall on the same calendar day. Omitted → 4 and enforce; explicit 0 → do not enforce.
func (r Rules) EffectiveMinHoursBetweenSameDaySessions() (hours float64, enforce bool) {
	if r.MinHoursBetweenSameDaySessions == nil {
		return 4, true
	}
	v := *r.MinHoursBetweenSameDaySessions
	if v <= 0 {
		return 0, false
	}
	return v, true
}

// EffectiveMaxSessionsPerDay returns a cap on sessions per day when set to a positive integer.
func (r Rules) EffectiveMaxSessionsPerDay() (max int, enforce bool) {
	if r.MaxSessionsPerDay == nil || *r.MaxSessionsPerDay <= 0 {
		return 0, false
	}
	return *r.MaxSessionsPerDay, true
}

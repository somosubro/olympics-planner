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
}

// EffectiveNoSameSportAcrossDays is true when the field is omitted or set true; false only when explicitly false.
func (r Rules) EffectiveNoSameSportAcrossDays() bool {
	if r.NoSameSportAcrossDays == nil {
		return true
	}
	return *r.NoSameSportAcrossDays
}

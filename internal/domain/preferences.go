package domain

type Preferences struct {
	AllowedSports []string `json:"allowedSports"`
	SportPriority []string `json:"sportPriority"`
	AllowedDays   []string `json:"allowedDays"`
	Rules         Rules    `json:"rules"`
}

type Rules struct {
	NoSameSportAcrossDays bool                   `json:"noSameSportAcrossDays"`
	PreferDayPairs        [][]string             `json:"preferDayPairs,omitempty"`
	SportSpecific         map[string]interface{} `json:"sportSpecific,omitempty"`
}

package domain

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
type PlanDay struct {
	Date                string   `json:"date"`
	DayOfWeek           string   `json:"dayOfWeek"`
	PrimarySessionID    string   `json:"primarySessionId"`
	AlternateSessionIDs []string `json:"alternateSessionIds"`
}

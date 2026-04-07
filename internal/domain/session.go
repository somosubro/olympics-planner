package domain

// Session matches docs/data-contract.md §5. Optional fields use omitempty per §5.2 / §6.
type Session struct {
	ID             string   `json:"id"`
	Sport          string   `json:"sport"`
	SessionCode    string   `json:"sessionCode"`
	Title          string   `json:"title,omitempty"`
	Date           string   `json:"date"`
	DayOfWeek      string   `json:"dayOfWeek"`
	StartTime      string   `json:"startTime"`
	EndTime        string   `json:"endTime,omitempty"`
	Venue          string   `json:"venue"`
	IncludedEvents []string `json:"includedEvents,omitempty"`
}

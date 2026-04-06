package domain

type Session struct {
    ID             string   `json:"id"`
    Sport          string   `json:"sport"`
    SessionCode    string   `json:"sessionCode"`
    Title          string   `json:"title"`
    Date           string   `json:"date"`
    DayOfWeek      string   `json:"dayOfWeek"`
    StartTime      string   `json:"startTime"`
    EndTime        string   `json:"endTime,omitempty"`
    Venue          string   `json:"venue"`
    IncludedEvents []string `json:"includedEvents"`
}

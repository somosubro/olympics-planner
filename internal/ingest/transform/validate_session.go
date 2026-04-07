package transform

import (
	"regexp"
	"strings"

	"olympics-planner/internal/domain"
)

var (
	datePattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	timePattern = regexp.MustCompile(`^\d{2}:\d{2}$`)
)

// SessionValidationIssues reports MVP required-field problems per docs/data-contract.md §6.1.
// Empty slice means the session is acceptable for runtime validation.
func SessionValidationIssues(s domain.Session) []string {
	var issues []string
	if strings.TrimSpace(s.ID) == "" {
		issues = append(issues, "missing id")
	}
	if strings.TrimSpace(s.Sport) == "" || s.Sport == "Unknown" {
		issues = append(issues, "missing or unknown sport")
	}
	if strings.TrimSpace(s.SessionCode) == "" {
		issues = append(issues, "missing sessionCode")
	}
	if strings.TrimSpace(s.Date) == "" || !datePattern.MatchString(s.Date) {
		issues = append(issues, "missing or invalid date")
	}
	if strings.TrimSpace(s.DayOfWeek) == "" {
		issues = append(issues, "missing dayOfWeek")
	}
	if strings.TrimSpace(s.StartTime) == "" || !timePattern.MatchString(s.StartTime) {
		issues = append(issues, "missing or invalid startTime")
	}
	if strings.TrimSpace(s.Venue) == "" {
		issues = append(issues, "missing venue")
	}
	return issues
}

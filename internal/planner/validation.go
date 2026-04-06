package planner

import "olympics-planner/internal/domain"

type ValidationResult struct {
    Valid  bool     `json:"valid"`
    Errors []string `json:"errors"`
}

func ValidatePlan(plan domain.Plan, sessions []domain.Session, prefs domain.Preferences) ValidationResult {
    errors := make([]string, 0)

    sessionsByID := make(map[string]domain.Session, len(sessions))
    for _, session := range sessions {
        sessionsByID[session.ID] = session
    }

    allowedSports := toSet(prefs.AllowedSports)
    allowedDays := toSet(prefs.AllowedDays)
    seenSportByDay := make(map[string]string)

    for _, planned := range plan.Sessions {
        session, ok := sessionsByID[planned.SessionID]
        if !ok {
            errors = append(errors, "session not found: "+planned.SessionID)
            continue
        }

        if _, ok := allowedSports[session.Sport]; !ok {
            errors = append(errors, "disallowed sport: "+session.Sport)
        }

        if _, ok := allowedDays[planned.Day]; !ok {
            errors = append(errors, "disallowed day: "+planned.Day)
        }

        if prefs.Rules.NoSameSportAcrossDays {
            if previousDay, seen := seenSportByDay[session.Sport]; seen && previousDay != planned.Day {
                errors = append(errors, "sport repeated across days: "+session.Sport)
            } else {
                seenSportByDay[session.Sport] = planned.Day
            }
        }
    }

    return ValidationResult{
        Valid:  len(errors) == 0,
        Errors: errors,
    }
}

func toSet(values []string) map[string]struct{} {
    result := make(map[string]struct{}, len(values))
    for _, value := range values {
        result[value] = struct{}{}
    }
    return result
}

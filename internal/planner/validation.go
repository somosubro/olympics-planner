package planner

import (
	"fmt"
	"strings"

	"olympics-planner/internal/domain"
)

// ValidatePlan validates a canonical plan against the session dataset and preferences.
// Errors use stable codes from docs/data-contract.md §13.4.
func ValidatePlan(plan domain.Plan, sessions []domain.Session, prefs domain.Preferences) domain.ValidationResult {
	var errs []domain.ValidationError

	sessionsByID := make(map[string]domain.Session, len(sessions))
	for _, s := range sessions {
		sessionsByID[s.ID] = s
	}

	if len(plan.Days) == 0 {
		errs = append(errs, domain.ValidationError{
			Code:    "INVALID_PLAN_SHAPE",
			Message: "plan must include at least one day entry",
			Field:   "days",
		})
		return domain.ValidationResult{Valid: false, Errors: errs}
	}

	switch plan.PlanType {
	case domain.PlanTypeOneDay:
		if len(plan.Days) != 1 {
			errs = append(errs, domain.ValidationError{
				Code:    "INVALID_PLAN_TYPE_FOR_DAY_COUNT",
				Message: `planType "one_day" requires exactly one day`,
				Field:   "days",
			})
		}
	case domain.PlanTypeTwoDay:
		if len(plan.Days) != 2 {
			errs = append(errs, domain.ValidationError{
				Code:    "INVALID_PLAN_TYPE_FOR_DAY_COUNT",
				Message: `planType "two_day" requires exactly two days`,
				Field:   "days",
			})
		}
	case domain.PlanTypeMultiDay:
		if len(plan.Days) < 3 {
			errs = append(errs, domain.ValidationError{
				Code:    "INVALID_PLAN_TYPE_FOR_DAY_COUNT",
				Message: `planType "multi_day" requires three or more days`,
				Field:   "days",
			})
		}
	default:
		errs = append(errs, domain.ValidationError{
			Code:    "INVALID_PLAN_SHAPE",
			Message: fmt.Sprintf("unknown planType %q", plan.PlanType),
			Field:   "planType",
		})
	}

	if len(errs) > 0 {
		return domain.ValidationResult{Valid: false, Errors: errs}
	}

	allowedSports := toSet(prefs.AllowedSports)
	allowedDays := toSet(prefs.AllowedDays)
	seenSportByDay := make(map[string]string)

	seenIDs := make(map[string]struct{})

	for dayIdx, day := range plan.Days {
		dayField := fmt.Sprintf("days[%d]", dayIdx)
		if strings.TrimSpace(day.PrimarySessionID) == "" {
			errs = append(errs, domain.ValidationError{
				Code:    "EMPTY_DAY_ENTRY",
				Message: "primarySessionId is required",
				Field:   dayField + ".primarySessionId",
			})
			continue
		}

		primary, ok := sessionsByID[day.PrimarySessionID]
		if !ok {
			errs = append(errs, domain.ValidationError{
				Code:    "SESSION_NOT_FOUND",
				Message: fmt.Sprintf("unknown session id %q", day.PrimarySessionID),
				Field:   dayField + ".primarySessionId",
			})
			continue
		}

		if _, dup := seenIDs[day.PrimarySessionID]; dup {
			errs = append(errs, domain.ValidationError{
				Code:    "DUPLICATE_SESSION",
				Message: fmt.Sprintf("session id %q appears more than once in the plan", day.PrimarySessionID),
				Field:   dayField + ".primarySessionId",
			})
		} else {
			seenIDs[day.PrimarySessionID] = struct{}{}
		}

		if _, ok := allowedSports[primary.Sport]; !ok {
			errs = append(errs, domain.ValidationError{
				Code:    "DISALLOWED_SPORT",
				Message: fmt.Sprintf("sport %q is not allowed for this request", primary.Sport),
				Field:   dayField + ".primarySessionId",
			})
		}

		if _, ok := allowedDays[primary.DayOfWeek]; !ok {
			errs = append(errs, domain.ValidationError{
				Code:    "DISALLOWED_DAY",
				Message: fmt.Sprintf("day %q is not allowed for this request", primary.DayOfWeek),
				Field:   dayField + ".primarySessionId",
			})
		}

		if primary.Date != day.Date || primary.DayOfWeek != day.DayOfWeek {
			errs = append(errs, domain.ValidationError{
				Code:    "DATE_DAY_MISMATCH",
				Message: "primary session date/dayOfWeek does not match plan day",
				Field:   dayField,
			})
		}

		altSeen := make(map[string]struct{})
		for _, altID := range day.AlternateSessionIDs {
			if altID == day.PrimarySessionID {
				errs = append(errs, domain.ValidationError{
					Code:    "INVALID_ALTERNATE",
					Message: "alternate must not duplicate primary session id",
					Field:   dayField + ".alternateSessionIds",
				})
				continue
			}
			if _, dup := altSeen[altID]; dup {
				errs = append(errs, domain.ValidationError{
					Code:    "INVALID_ALTERNATE",
					Message: "duplicate alternate session id within the same day",
					Field:   dayField + ".alternateSessionIds",
				})
				continue
			}
			altSeen[altID] = struct{}{}

			if _, dup := seenIDs[altID]; dup {
				errs = append(errs, domain.ValidationError{
					Code:    "DUPLICATE_SESSION",
					Message: fmt.Sprintf("session id %q appears more than once in the plan", altID),
					Field:   dayField + ".alternateSessionIds",
				})
			} else {
				seenIDs[altID] = struct{}{}
			}

			altSession, ok := sessionsByID[altID]
			if !ok {
				errs = append(errs, domain.ValidationError{
					Code:    "SESSION_NOT_FOUND",
					Message: fmt.Sprintf("unknown session id %q", altID),
					Field:   dayField + ".alternateSessionIds",
				})
				continue
			}

			if _, ok := allowedSports[altSession.Sport]; !ok {
				errs = append(errs, domain.ValidationError{
					Code:    "DISALLOWED_SPORT",
					Message: fmt.Sprintf("sport %q is not allowed for this request", altSession.Sport),
					Field:   dayField + ".alternateSessionIds",
				})
			} else if _, ok := allowedDays[altSession.DayOfWeek]; !ok {
				errs = append(errs, domain.ValidationError{
					Code:    "DISALLOWED_DAY",
					Message: fmt.Sprintf("day %q is not allowed for this request", altSession.DayOfWeek),
					Field:   dayField + ".alternateSessionIds",
				})
			}
		}

		if prefs.Rules.NoSameSportAcrossDays {
			if prevDay, seen := seenSportByDay[primary.Sport]; seen && prevDay != day.Date {
				errs = append(errs, domain.ValidationError{
					Code:    "REPEATED_SPORT_ACROSS_DAYS",
					Message: fmt.Sprintf("sport %q appears on multiple days where repeated sports are forbidden", primary.Sport),
					Field:   "days",
				})
			} else {
				seenSportByDay[primary.Sport] = day.Date
			}
		}
	}

	return domain.ValidationResult{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}

func toSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

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

	var sportFirstDate map[string]string
	if prefs.Rules.EffectiveNoSameSportAcrossDays() {
		sportFirstDate = make(map[string]string)
	}

	seenIDs := make(map[string]struct{})

	for dayIdx, day := range plan.Days {
		dayField := fmt.Sprintf("days[%d]", dayIdx)

		sessionIDsMode := len(day.SessionIDs) > 0
		if sessionIDsMode && (strings.TrimSpace(day.PrimarySessionID) != "" || len(day.AlternateSessionIDs) > 0) {
			errs = append(errs, domain.ValidationError{
				Code:    "CONFLICTING_DAY_SPEC",
				Message: `use either sessionIds or primarySessionId/alternateSessionIds for a plan day, not both`,
				Field:   dayField,
			})
			continue
		}

		var ids []string
		if sessionIDsMode {
			ids = day.SessionIDs
		} else {
			if strings.TrimSpace(day.PrimarySessionID) == "" {
				errs = append(errs, domain.ValidationError{
					Code:    "EMPTY_DAY_ENTRY",
					Message: "primarySessionId is required when sessionIds is omitted",
					Field:   dayField + ".primarySessionId",
				})
				continue
			}
			ids = make([]string, 0, 1+len(day.AlternateSessionIDs))
			ids = append(ids, day.PrimarySessionID)
			ids = append(ids, day.AlternateSessionIDs...)
		}

		if max, ok := prefs.Rules.EffectiveMaxSessionsPerDay(); ok && len(ids) > max {
			errs = append(errs, domain.ValidationError{
				Code:    "TOO_MANY_SESSIONS_PER_DAY",
				Message: fmt.Sprintf("at most %d session(s) per day allowed for this request", max),
				Field:   dayField,
			})
			continue
		}

		errDayStart := len(errs)
		dayLocalSeen := make(map[string]struct{})
		var sportsThisDay []string

		for i, rawID := range ids {
			sid := strings.TrimSpace(rawID)
			field := fieldForPlanDaySession(dayField, sessionIDsMode, i == 0)
			if sid == "" {
				errs = append(errs, domain.ValidationError{
					Code:    "EMPTY_DAY_ENTRY",
					Message: "session id must not be empty",
					Field:   field,
				})
				continue
			}

			if _, dup := dayLocalSeen[sid]; dup {
				errs = append(errs, domain.ValidationError{
					Code:    "DUPLICATE_SESSION",
					Message: fmt.Sprintf("session id %q appears more than once in the plan", sid),
					Field:   field,
				})
				continue
			}
			dayLocalSeen[sid] = struct{}{}

			if _, dup := seenIDs[sid]; dup {
				errs = append(errs, domain.ValidationError{
					Code:    "DUPLICATE_SESSION",
					Message: fmt.Sprintf("session id %q appears more than once in the plan", sid),
					Field:   field,
				})
				continue
			}
			seenIDs[sid] = struct{}{}

			sess, ok := sessionsByID[sid]
			if !ok {
				errs = append(errs, domain.ValidationError{
					Code:    "SESSION_NOT_FOUND",
					Message: fmt.Sprintf("unknown session id %q", sid),
					Field:   field,
				})
				continue
			}

			if _, ok := allowedSports[sess.Sport]; !ok {
				errs = append(errs, domain.ValidationError{
					Code:    "DISALLOWED_SPORT",
					Message: fmt.Sprintf("sport %q is not allowed for this request", sess.Sport),
					Field:   field,
				})
			}

			if _, ok := allowedDays[sess.DayOfWeek]; !ok {
				errs = append(errs, domain.ValidationError{
					Code:    "DISALLOWED_DAY",
					Message: fmt.Sprintf("day %q is not allowed for this request", sess.DayOfWeek),
					Field:   field,
				})
			}

			if sess.Date != day.Date || sess.DayOfWeek != day.DayOfWeek {
				errs = append(errs, domain.ValidationError{
					Code:    "DATE_DAY_MISMATCH",
					Message: "session date/dayOfWeek does not match plan day",
					Field:   dayField,
				})
			}

			sportsThisDay = append(sportsThisDay, sess.Sport)
		}

		if len(errs) == errDayStart {
			if hrs, enforce := prefs.Rules.EffectiveMinHoursBetweenSameDaySessions(); enforce {
				errs = append(errs, sameDaySpacingErrors(ids, sessionsByID, hrs, dayField)...)
			}
		}

		// noSameSportAcrossDays: same sport must not appear on more than one calendar day
		// anywhere in the plan (all listed sessions per day).
		if prefs.Rules.EffectiveNoSameSportAcrossDays() && sportFirstDate != nil {
			for _, sp := range sportsThisDay {
				if first, seen := sportFirstDate[sp]; seen {
					if first != day.Date {
						errs = append(errs, domain.ValidationError{
							Code:    "REPEATED_SPORT_ACROSS_DAYS",
							Message: fmt.Sprintf("sport %q appears on multiple days (including alternates) where repeated sports are forbidden", sp),
							Field:   "days",
						})
					}
				} else {
					sportFirstDate[sp] = day.Date
				}
			}
		}
	}

	return domain.ValidationResult{
		Valid:  len(errs) == 0,
		Errors: errs,
	}
}

func fieldForPlanDaySession(dayField string, sessionIDsMode, isFirst bool) string {
	if sessionIDsMode {
		return dayField + ".sessionIds"
	}
	if isFirst {
		return dayField + ".primarySessionId"
	}
	return dayField + ".alternateSessionIds"
}

func toSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

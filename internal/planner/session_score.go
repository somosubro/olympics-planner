package planner

import (
	"sort"
	"strings"

	"olympics-planner/internal/domain"
	"olympics-planner/internal/ingest/transform"
)

// SessionScoreComponents matches api-spec ranked session response.
type SessionScoreComponents struct {
	SportPriority int `json:"sportPriority"`
	DataQuality   int `json:"dataQuality"`
}

// RankedSession matches POST /api/v1/rank/sessions response entries.
type RankedSession struct {
	Session    domain.Session          `json:"session"`
	Score      int                     `json:"score"`
	Components *SessionScoreComponents `json:"components,omitempty"`
}

// SessionRankable applies scoring-spec session checks + preferences (SV2/SV3) for rank/sessions.
func SessionRankable(s domain.Session, prefs domain.Preferences) bool {
	if len(transform.SessionValidationIssues(s)) > 0 {
		return false
	}
	if len(prefs.AllowedSports) == 0 || len(prefs.AllowedDays) == 0 {
		return false
	}
	if !stringInList(prefs.AllowedSports, s.Sport) {
		return false
	}
	if !stringInList(prefs.AllowedDays, s.DayOfWeek) {
		return false
	}
	return true
}

// ScoreSession returns total 0..40 and components per scoring-and-validation-spec.md §13.
func ScoreSession(s domain.Session, prefs domain.Preferences) (int, SessionScoreComponents) {
	sp := sportPriorityPoints(s.Sport, prefs.SportPriority)
	dq := dataQualityPoints(s)
	return sp + dq, SessionScoreComponents{SportPriority: sp, DataQuality: dq}
}

func sportPriorityPoints(sport string, priority []string) int {
	n := len(priority)
	if n == 0 {
		return 0
	}
	for i, sp := range priority {
		if sp == sport {
			if n == 1 {
				return 30
			}
			return 30 - (i*25)/(n-1)
		}
	}
	return 0
}

func dataQualityPoints(s domain.Session) int {
	q := 0
	if strings.TrimSpace(s.Title) != "" {
		q += 4
	}
	if strings.TrimSpace(s.EndTime) != "" {
		q += 2
	}
	if len(s.IncludedEvents) > 0 {
		q += 4
	}
	if q > 10 {
		return 10
	}
	return q
}

// RankSessions ranks valid sessions descending; omits non-rankable per api-spec §12.
func RankSessions(in []domain.Session, prefs domain.Preferences, includeBreakdown bool) []RankedSession {
	var rows []RankedSession
	for _, s := range in {
		if !SessionRankable(s, prefs) {
			continue
		}
		total, comp := ScoreSession(s, prefs)
		row := RankedSession{Session: s, Score: total}
		if includeBreakdown {
			c := comp
			row.Components = &c
		}
		rows = append(rows, row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return compareRanked(rows[i], rows[j], prefs) < 0
	})
	return rows
}

// compareRanked < 0 means a sorts before b (a ranks higher).
func compareRanked(a, b RankedSession, prefs domain.Preferences) int {
	if a.Score != b.Score {
		return b.Score - a.Score
	}
	_, ca := ScoreSession(a.Session, prefs)
	_, cb := ScoreSession(b.Session, prefs)
	if ca.SportPriority != cb.SportPriority {
		return cb.SportPriority - ca.SportPriority
	}
	if ca.DataQuality != cb.DataQuality {
		return cb.DataQuality - ca.DataQuality
	}
	return strings.Compare(a.Session.ID, b.Session.ID)
}

func stringInList(list []string, v string) bool {
	for _, x := range list {
		if x == v {
			return true
		}
	}
	return false
}

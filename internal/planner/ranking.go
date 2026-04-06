package planner

import "olympics-planner/internal/domain"

type ScoredSession struct {
    Session domain.Session `json:"session"`
    Score   int            `json:"score"`
}

func RankSessions(sessions []domain.Session, prefs domain.Preferences) []ScoredSession {
    priorityIndex := make(map[string]int, len(prefs.SportPriority))
    for i, sport := range prefs.SportPriority {
        priorityIndex[sport] = i
    }

    scored := make([]ScoredSession, 0, len(sessions))
    for _, session := range sessions {
        score := 0

        if idx, ok := priorityIndex[session.Sport]; ok {
            score += 100 - idx
        }

        if session.DayOfWeek == "Saturday" || session.DayOfWeek == "Sunday" {
            score += 25
        }

        scored = append(scored, ScoredSession{
            Session: session,
            Score:   score,
        })
    }

    return scored
}

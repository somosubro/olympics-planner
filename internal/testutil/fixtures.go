package testutil

import "olympics-planner/internal/domain"

func SampleSessions() []domain.Session {
    return []domain.Session{
        {
            ID:        "s1",
            Sport:     "Tennis",
            Title:     "Tennis Session 1",
            DayOfWeek: "Saturday",
        },
        {
            ID:        "s2",
            Sport:     "Swimming",
            Title:     "Swimming Session 1",
            DayOfWeek: "Sunday",
        },
    }
}

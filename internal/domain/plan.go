package domain

type Plan struct {
    Sessions []PlannedSession `json:"sessions"`
}

type PlannedSession struct {
    Day       string `json:"day"`
    SessionID string `json:"sessionId"`
}

package domain_test

import (
	"encoding/json"
	"testing"

	"olympics-planner/internal/domain"
)

func TestSession_JSONOmitsEmptyOptionals(t *testing.T) {
	s := domain.Session{
		ID:          "s1",
		Sport:       "Tennis",
		SessionCode: "T1",
		Date:        "2028-07-15",
		DayOfWeek:   "Saturday",
		StartTime:   "10:00",
		Venue:       "Court",
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if _, ok := raw["title"]; ok {
		t.Fatal("expected title omitted")
	}
	if _, ok := raw["endTime"]; ok {
		t.Fatal("expected endTime omitted")
	}
	if _, ok := raw["includedEvents"]; ok {
		t.Fatal("expected includedEvents omitted")
	}
}

func TestSession_JSONRoundTripWithOptionals(t *testing.T) {
	s := domain.Session{
		ID:             "s1",
		Sport:          "Tennis",
		SessionCode:    "T1",
		Title:          "Final",
		Date:           "2028-07-15",
		DayOfWeek:      "Saturday",
		StartTime:      "10:00",
		EndTime:        "12:00",
		Venue:          "Court",
		IncludedEvents: []string{"Men's Singles"},
	}
	b, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	var out domain.Session
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.Title != s.Title || out.EndTime != s.EndTime || len(out.IncludedEvents) != 1 {
		t.Fatalf("round trip mismatch: %#v", out)
	}
}

package domain_test

import (
	"encoding/json"
	"testing"

	"olympics-planner/internal/domain"
)

func TestPlan_JSONRoundTripCanonicalShape(t *testing.T) {
	p := domain.Plan{
		PlanType: domain.PlanTypeTwoDay,
		Days: []domain.PlanDay{
			{
				Date:                "2028-07-15",
				DayOfWeek:           "Saturday",
				PrimarySessionID:    "session-ten-12",
				AlternateSessionIDs: []string{"session-ath-09"},
			},
			{
				Date:                "2028-07-16",
				DayOfWeek:           "Sunday",
				PrimarySessionID:    "session-ckt-03",
				AlternateSessionIDs: nil,
			},
		},
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatal(err)
	}
	if string(raw["planType"]) != `"two_day"` {
		t.Fatalf("planType: %s", raw["planType"])
	}
	var out domain.Plan
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	if out.PlanType != domain.PlanTypeTwoDay || len(out.Days) != 2 {
		t.Fatalf("round trip: %#v", out)
	}
	if out.Days[0].PrimarySessionID != "session-ten-12" {
		t.Fatal("primary id")
	}
}

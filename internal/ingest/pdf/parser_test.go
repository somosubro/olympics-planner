package pdf

import (
	"strings"
	"testing"
)

func TestParseScheduleText_SingleRow(t *testing.T) {
	input := MinimalScheduleLineForTest() + "\n"
	sessions, stats, err := ParseScheduleText(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if stats.LinesScanned != 1 {
		t.Fatalf("lines: got %d", stats.LinesScanned)
	}
	if stats.RowMatches != 1 {
		t.Fatalf("row matches: got %d", stats.RowMatches)
	}
	if len(sessions) != 1 {
		t.Fatalf("sessions: %d", len(sessions))
	}
	s := sessions[0]
	if s.ID != "TEN12" || s.Sport != "Tennis" || s.Date != "2028-07-15" {
		t.Fatalf("session: %#v", s)
	}
	if s.DayOfWeek != "Saturday" || s.StartTime != "14:00" || s.EndTime != "17:00" {
		t.Fatalf("times/day: %#v", s)
	}
}

func TestParseScheduleText_HeaderLinesSkipped(t *testing.T) {
	var b strings.Builder
	b.WriteString("Sport Venue Zone Session Code Date\n")
	b.WriteString(MinimalScheduleLineForTest())
	b.WriteString("\n")
	sessions, stats, err := ParseScheduleText(strings.NewReader(b.String()))
	if err != nil {
		t.Fatal(err)
	}
	if stats.RowMatches != 1 {
		t.Fatalf("row matches: %d", stats.RowMatches)
	}
	if len(sessions) != 1 {
		t.Fatalf("got %d sessions", len(sessions))
	}
}

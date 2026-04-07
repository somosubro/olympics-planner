package pipeline

import (
	"os"
	"path/filepath"
	"testing"

	"olympics-planner/internal/domain"
	ingestpdf "olympics-planner/internal/ingest/pdf"
)

func TestImportFromText_WritesValidSession(t *testing.T) {
	input := "Sport Venue Zone Session Code Date\n" + ingestpdf.MinimalScheduleLineForTest() + "\n"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "in.txt")
	outPath := filepath.Join(dir, "out.json")
	if err := os.WriteFile(inPath, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}

	res, err := ImportFromText(inPath, outPath)
	if err != nil {
		t.Fatal(err)
	}
	if res.Stats.Emitted != 1 {
		t.Fatalf("emitted %d, stats %#v", res.Stats.Emitted, res.Stats)
	}
	if len(res.Sessions) != 1 {
		t.Fatal(len(res.Sessions))
	}
	if res.Sessions[0].ID != "TEN12" {
		t.Fatalf("id %q", res.Sessions[0].ID)
	}
}

func TestFilterAndDedupe_DropsDuplicateID(t *testing.T) {
	sessions := []domain.Session{
		{ID: "a", Sport: "Tennis", SessionCode: "T1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "10:00", Venue: "V"},
		{ID: "a", Sport: "Tennis", SessionCode: "T1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "11:00", Venue: "V"},
	}
	out, drop, dup, drops := filterAndDedupe(sessions)
	if len(out) != 1 || drop != 0 || dup != 1 || len(drops) != 0 {
		t.Fatalf("out=%d drop=%d dup=%d drops=%d", len(out), drop, dup, len(drops))
	}
}

func TestFilterAndDedupe_DropsInvalid(t *testing.T) {
	sessions := []domain.Session{
		{ID: "", Sport: "Tennis", SessionCode: "T1", Date: "2028-07-15", DayOfWeek: "Saturday", StartTime: "10:00", Venue: "V"},
	}
	out, drop, dup, drops := filterAndDedupe(sessions)
	if len(out) != 0 || drop != 1 || dup != 0 || len(drops) != 1 {
		t.Fatalf("out=%d drop=%d dup=%d drops=%d", len(out), drop, dup, len(drops))
	}
	if len(drops[0].Issues) == 0 {
		t.Fatal("expected issues on drop")
	}
}

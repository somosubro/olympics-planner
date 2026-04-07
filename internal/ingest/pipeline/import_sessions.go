package pipeline

import (
	"os"

	"olympics-planner/internal/domain"
	ingestpdf "olympics-planner/internal/ingest/pdf"
	"olympics-planner/internal/ingest/transform"
)

// ImportStats summarizes a full import run (parse + validation + dedupe + write).
type ImportStats struct {
	Emitted             int
	DroppedInvalid      int
	DuplicateIDsSkipped int
	ParseRowMatches     int
	LinesScanned        int
}

// InvalidDrop records one parsed session that failed import-time validation.
type InvalidDrop struct {
	ID          string
	SessionCode string
	Issues      []string
}

// ImportResult is returned by ImportFromPDF and ImportFromText.
type ImportResult struct {
	Sessions     []domain.Session
	Stats        ImportStats
	InvalidDrops []InvalidDrop
}

// ImportFromPDF extracts text from the source PDF, parses sessions from the
// extracted text, validates and dedupes, then writes the sessions JSON.
func ImportFromPDF(pdfPath, outPath string) (ImportResult, error) {
	textPath, err := ingestpdf.ExtractTextWithPDFToText(pdfPath)
	if err != nil {
		return ImportResult{}, err
	}
	defer func() { _ = os.Remove(textPath) }()

	return importFromParsedText(textPath, outPath)
}

// ImportFromText parses an already-extracted pdftotext output file, validates
// and dedupes, then writes the sessions JSON.
func ImportFromText(textPath, outPath string) (ImportResult, error) {
	return importFromParsedText(textPath, outPath)
}

func importFromParsedText(textPath, outPath string) (ImportResult, error) {
	raw, parseStats, err := ingestpdf.ParseScheduleTextFile(textPath)
	if err != nil {
		return ImportResult{}, err
	}
	res := finalizeSessions(raw, parseStats)
	if err := WriteSessionsJSON(outPath, res.Sessions); err != nil {
		return ImportResult{}, err
	}
	return res, nil
}

func finalizeSessions(sessions []domain.Session, parseStats ingestpdf.ParseStats) ImportResult {
	valid, dropped, dup, invalidDrops := filterAndDedupe(sessions)
	return ImportResult{
		Sessions:     valid,
		InvalidDrops: invalidDrops,
		Stats: ImportStats{
			Emitted:             len(valid),
			DroppedInvalid:      dropped,
			DuplicateIDsSkipped: dup,
			ParseRowMatches:     parseStats.RowMatches,
			LinesScanned:        parseStats.LinesScanned,
		},
	}
}

func filterAndDedupe(sessions []domain.Session) (out []domain.Session, droppedInvalid, dupSkipped int, invalidDrops []InvalidDrop) {
	seen := make(map[string]struct{})
	for _, s := range sessions {
		if issues := transform.SessionValidationIssues(s); len(issues) > 0 {
			droppedInvalid++
			invalidDrops = append(invalidDrops, InvalidDrop{
				ID:          s.ID,
				SessionCode: s.SessionCode,
				Issues:      issues,
			})
			continue
		}
		if _, ok := seen[s.ID]; ok {
			dupSkipped++
			continue
		}
		seen[s.ID] = struct{}{}
		out = append(out, s)
	}
	return out, droppedInvalid, dupSkipped, invalidDrops
}

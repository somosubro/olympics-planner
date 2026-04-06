package pipeline

import (
	"os"

	"olympics-planner/internal/domain"
	ingestpdf "olympics-planner/internal/ingest/pdf"
)

// ImportFromPDF extracts text from the source PDF, parses sessions from the
// extracted text, and writes the generated sessions JSON.
func ImportFromPDF(pdfPath, outPath string) ([]domain.Session, error) {
	textPath, err := ingestpdf.ExtractTextWithPDFToText(pdfPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(textPath) }()

	sessions, err := ingestpdf.ParseScheduleTextFile(textPath)
	if err != nil {
		return nil, err
	}
	if err := WriteSessionsJSON(outPath, sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

// ImportFromText parses an already-extracted pdftotext output file and writes
// the generated sessions JSON.
func ImportFromText(textPath, outPath string) ([]domain.Session, error) {
	sessions, err := ingestpdf.ParseScheduleTextFile(textPath)
	if err != nil {
		return nil, err
	}
	if err := WriteSessionsJSON(outPath, sessions); err != nil {
		return nil, err
	}
	return sessions, nil
}

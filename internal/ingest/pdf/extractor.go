package pdf

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExtractTextWithPDFToText runs pdftotext -layout and returns the path to the
// generated .txt file next to the PDF.
func ExtractTextWithPDFToText(pdfPath string) (string, error) {
	if _, err := exec.LookPath("pdftotext"); err != nil {
		return "", errors.New("pdftotext not found. Install poppler, or run pdftotext manually and use import-text")
	}

	outPath := strings.TrimSuffix(pdfPath, ".pdf") + ".txt"
	cmd := exec.Command("pdftotext", "-layout", pdfPath, outPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("pdftotext failed: %w", err)
	}
	return outPath, nil
}

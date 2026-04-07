package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"olympics-planner/internal/ingest/pipeline"
)

func main() {
	if len(os.Args) < 2 {
		usageAndExit()
	}

	switch os.Args[1] {
	case "import":
		runImportPDF(os.Args[2:])
	case "import-text":
		runImportText(os.Args[2:])
	default:
		usageAndExit()
	}
}

func runImportPDF(args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	pdfPath := fs.String("pdf", "", "Path to the LA28 PDF")
	outPath := fs.String("out", "data/sessions.json", "Output JSON path")
	_ = fs.Parse(args)

	if *pdfPath == "" {
		fatal(errors.New("-pdf is required"))
	}

	res, err := pipeline.ImportFromPDF(*pdfPath, *outPath)
	if err != nil {
		fatal(err)
	}

	printImportResult(res, *outPath)
}

func runImportText(args []string) {
	fs := flag.NewFlagSet("import-text", flag.ExitOnError)
	textPath := fs.String("text", "", "Path to text file extracted from the LA28 PDF")
	outPath := fs.String("out", "data/sessions.json", "Output JSON path")
	_ = fs.Parse(args)

	if *textPath == "" {
		fatal(errors.New("-text is required"))
	}

	res, err := pipeline.ImportFromText(*textPath, *outPath)
	if err != nil {
		fatal(err)
	}

	printImportResult(res, *outPath)
}

func printImportResult(res pipeline.ImportResult, outPath string) {
	fmt.Printf("Imported %d sessions into %s\n", res.Stats.Emitted, outPath)
	fmt.Fprintf(os.Stderr, "parse: %d lines scanned, %d schedule rows matched\n",
		res.Stats.LinesScanned, res.Stats.ParseRowMatches)
	if res.Stats.DroppedInvalid > 0 {
		fmt.Fprintf(os.Stderr, "dropped %d invalid session(s) (failed required-field checks)\n", res.Stats.DroppedInvalid)
		for _, d := range res.InvalidDrops {
			ref := d.SessionCode
			if ref == "" {
				ref = d.ID
			}
			if d.ID != "" && d.SessionCode != "" && d.ID != d.SessionCode {
				ref = fmt.Sprintf("%s (id=%s)", d.SessionCode, d.ID)
			}
			fmt.Fprintf(os.Stderr, "  - %s: %s\n", ref, strings.Join(d.Issues, "; "))
		}
	}
	if res.Stats.DuplicateIDsSkipped > 0 {
		fmt.Fprintf(os.Stderr, "skipped %d duplicate session id(s)\n", res.Stats.DuplicateIDsSkipped)
	}
}

func usageAndExit() {
	fmt.Fprintln(os.Stderr, "usage:")
	fmt.Fprintln(os.Stderr, "  import_sessions import -pdf <schedule.pdf> [-out data/sessions.json]")
	fmt.Fprintln(os.Stderr, "  import_sessions import-text -text <schedule.txt> [-out data/sessions.json]")
	os.Exit(1)
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "error:", err)
	os.Exit(1)
}

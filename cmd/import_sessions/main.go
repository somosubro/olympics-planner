package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

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

	sessions, err := pipeline.ImportFromPDF(*pdfPath, *outPath)
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Imported %d sessions into %s\n", len(sessions), *outPath)
}

func runImportText(args []string) {
	fs := flag.NewFlagSet("import-text", flag.ExitOnError)
	textPath := fs.String("text", "", "Path to text file extracted from the LA28 PDF")
	outPath := fs.String("out", "data/sessions.json", "Output JSON path")
	_ = fs.Parse(args)

	if *textPath == "" {
		fatal(errors.New("-text is required"))
	}

	sessions, err := pipeline.ImportFromText(*textPath, *outPath)
	if err != nil {
		fatal(err)
	}

	fmt.Printf("Imported %d sessions into %s\n", len(sessions), *outPath)
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

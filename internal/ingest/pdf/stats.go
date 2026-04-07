package pdf

// ParseStats summarizes a single parse of schedule text (pdftotext -layout output).
type ParseStats struct {
	LinesScanned int
	RowMatches   int
}

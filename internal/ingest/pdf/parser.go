package pdf

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"olympics-planner/internal/domain"
	"olympics-planner/internal/ingest/transform"
)

var monthMap = map[string]string{
	"January":   "01",
	"February":  "02",
	"March":     "03",
	"April":     "04",
	"May":       "05",
	"June":      "06",
	"July":      "07",
	"August":    "08",
	"September": "09",
	"October":   "10",
	"November":  "11",
	"December":  "12",
}

var rowRe = regexp.MustCompile(`^(.*?)\s{2,}(.*?)\s{2,}(.*?)\s{2,}([A-Z]{3,4}\d{2,3})\s{2,}([A-Za-z]+,\s+[A-Za-z]+\s+\d{1,2})\s{2,}(-?\d+)\s{2,}(N/A|Preliminary|Quarterfinal|Semifinal|Final|Bronze)\s{2,}(.*?)\s{2,}(\d{1,2}:\d{2})\s{2,}(\d{1,2}:\d{2}|0:00)$`)
var dateOnlyRe = regexp.MustCompile(`^([A-Za-z]+),\s+([A-Za-z]+)\s+(\d{1,2})$`)

// ParseScheduleTextFile reads text produced by pdftotext -layout from the LA28
// "Competition Schedule by Event" PDF and returns sessions in domain form.
func ParseScheduleTextFile(path string) ([]domain.Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return ParseScheduleText(f)
}

// ParseScheduleText parses schedule text from any reader containing pdftotext
// -layout output for the LA28 schedule PDF.
func ParseScheduleText(r io.Reader) ([]domain.Session, error) {
	var sessions []domain.Session
	scanner := bufio.NewScanner(r)

	type acc struct {
		code, sport, venue, date, day, start, end string
		descLines                                 []string
	}

	var current *acc

	flush := func() {
		if current == nil {
			return
		}
		desc := transform.DedupeStrings(current.descLines)
		sport := transform.NormalizeSport(current.sport, current.code)
		title := sessionTitle(desc, sport, current.code)
		sessions = append(sessions, domain.Session{
			ID:             current.code,
			Sport:          sport,
			SessionCode:    current.code,
			Title:          title,
			Date:           current.date,
			DayOfWeek:      current.day,
			StartTime:      current.start,
			EndTime:        current.end,
			Venue:          current.venue,
			IncludedEvents: desc,
		})
		current = nil
	}

	for scanner.Scan() {
		line := scanner.Text()
		norm := transform.NormalizeWhitespace(line)
		if norm == "" {
			continue
		}

		if strings.Contains(norm, "Olympic Competition Schedule by Event Version") ||
			strings.HasPrefix(norm, "As of ") ||
			strings.HasPrefix(norm, "This competition schedule is subject to change") ||
			strings.HasPrefix(norm, "2028 Games.") ||
			strings.HasPrefix(norm, "Events listed in the Session Description") ||
			strings.HasPrefix(norm, "order in which they will occur") ||
			strings.HasPrefix(norm, "the Football (Soccer) tournaments") ||
			strings.HasPrefix(norm, "are in Pacific Time") ||
			strings.HasPrefix(norm, "specified. All dates listed") ||
			strings.HasPrefix(norm, "Sport Venue Zone Session Code Date") {
			continue
		}

		m := rowRe.FindStringSubmatch(line)
		if m != nil {
			flush()

			sport := transform.NormalizeWhitespace(m[1])
			venue := transform.NormalizeWhitespace(m[2])
			code := transform.NormalizeWhitespace(m[4])
			dateStr := transform.NormalizeWhitespace(m[5])
			descFirst := transform.NormalizeWhitespace(m[8])
			start := transform.NormalizeTime(m[9])
			end := transform.NormalizeTime(m[10])

			date, day := parseEventPDFDate(dateOnlyRe, dateStr)

			current = &acc{
				code:      code,
				sport:     sport,
				venue:     venue,
				date:      date,
				day:       day,
				start:     start,
				end:       end,
				descLines: make([]string, 0, 8),
			}

			if descFirst != "" && !strings.EqualFold(descFirst, "Not Ticketed") {
				current.descLines = append(current.descLines, descFirst)
			}
			continue
		}

		if current != nil {
			clean := transform.NormalizeWhitespace(line)
			if clean == "" || strings.EqualFold(clean, "Not Ticketed") || strings.Contains(clean, "OKC Local Time") {
				continue
			}
			current.descLines = append(current.descLines, clean)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	flush()
	return sessions, nil
}

func sessionTitle(events []string, sport, code string) string {
	if len(events) > 0 && strings.TrimSpace(events[0]) != "" {
		return events[0]
	}
	return fmt.Sprintf("%s %s", sport, code)
}

func parseEventPDFDate(re *regexp.Regexp, s string) (date string, day string) {
	m := re.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return "", ""
	}
	day = m[1]
	monthNum, ok := monthMap[m[2]]
	if !ok {
		return "", day
	}
	dd, err := strconv.Atoi(m[3])
	if err != nil {
		return "", day
	}
	return fmt.Sprintf("2028-%s-%02d", monthNum, dd), day
}

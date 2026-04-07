package transform

import (
	"fmt"
	"strconv"
	"strings"
)

func NormalizeSport(currentSport, code string) string {
	if currentSport != "" {
		return currentSport
	}
	prefix := lettersPrefix(code)
	switch prefix {
	case "ATH":
		return "Athletics"
	case "CKT":
		return "Cricket"
	case "SWM":
		return "Swimming"
	case "TEN":
		return "Tennis"
	case "BKB":
		return "Basketball"
	case "HOC":
		return "Field Hockey"
	case "VBV":
		return "Beach Volleyball"
	case "SHO":
		return "Shooting"
	case "ASW":
		return "Artistic Swimming"
	case "WPO":
		return "Water Polo"
	case "ROW":
		return "Rowing"
	case "CSP":
		return "Canoe Sprint"
	case "DIV":
		return "Diving"
	case "SAL":
		return "Sailing"
	default:
		return "Unknown"
	}
}

func lettersPrefix(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= 'A' && r <= 'Z' {
			b.WriteRune(r)
		} else {
			break
		}
	}
	return b.String()
}

func DedupeStrings(in []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(in))
	for _, s := range in {
		s = strings.TrimSpace(s)
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}

func NormalizeWhitespace(s string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(s)), " ")
}

func NormalizeTime(s string) string {
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return s
	}
	h, err := strconv.Atoi(parts[0])
	if err != nil {
		return s
	}
	return fmt.Sprintf("%02d:%s", h, parts[1])
}

package http

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"

	sess "olympics-planner/internal/session"
)

var dateQueryPattern = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

// ParseSessionsFilter parses GET /api/v1/sessions query params (api-spec §10).
func ParseSessionsFilter(r *http.Request) (sess.Filter, *ErrorBody) {
	q := r.URL.Query()
	f := sess.Filter{
		Dates:          collect(q, "date"),
		DaysOfWeek:     collect(q, "dayOfWeek"),
		Sports:         collect(q, "sports"),
		AllowedSports:  collect(q, "allowedSports"),
		ExcludedSports: collect(q, "excludedSports"),
	}
	for _, d := range f.Dates {
		if !dateQueryPattern.MatchString(d) {
			return f, &ErrorBody{
				Code:    "INVALID_QUERY_PARAMETER",
				Message: "date must use YYYY-MM-DD format",
				Field:   "date",
			}
		}
	}
	return f, nil
}

func collect(q url.Values, key string) []string {
	var out []string
	for _, v := range q[key] {
		for _, part := range strings.Split(v, ",") {
			part = strings.TrimSpace(part)
			if part != "" {
				out = append(out, part)
			}
		}
	}
	return out
}

package planner

import (
	"sort"
	"strconv"
	"strings"

	"olympics-planner/internal/domain"
)

// PlanScoreComponents matches api-spec plan ranking response.
type PlanScoreComponents struct {
	DayPair            int `json:"dayPair"`
	SummedSessionScore int `json:"summedSessionScore"`
	Variety            int `json:"variety"`
	Convenience        int `json:"convenience"`
}

// RankedPlan is one entry in POST /api/v1/rank/plans response.
type RankedPlan struct {
	Plan       domain.Plan          `json:"plan"`
	Score      int                  `json:"score"`
	Components *PlanScoreComponents `json:"components,omitempty"`
}

// InvalidPlanEntry is used when includeInvalidPlans is true.
type InvalidPlanEntry struct {
	Plan       domain.Plan             `json:"plan"`
	Validation domain.ValidationResult `json:"validation"`
}

// RankPlansResponse is the JSON body for rank/plans.
type RankPlansResponse struct {
	Plans        []RankedPlan       `json:"plans"`
	InvalidPlans []InvalidPlanEntry `json:"invalidPlans,omitempty"`
}

// ScorePlan computes 0..100 plan score from primaries in dataset (scoring spec §14).
func ScorePlan(plan domain.Plan, sessionsByID map[string]domain.Session, prefs domain.Preferences) (int, PlanScoreComponents) {
	dps := dayPairComponent(plan)
	ss := summedSessionComponent(plan, sessionsByID, prefs)
	if ss > 50 {
		ss = 50
	}
	v := varietyComponent(plan, sessionsByID)
	c := convenienceComponent(plan, sessionsByID)
	total := dps + ss + v + c
	if total > 100 {
		total = 100
	}
	return total, PlanScoreComponents{
		DayPair:            dps,
		SummedSessionScore: ss,
		Variety:            v,
		Convenience:        c,
	}
}

func dayPairComponent(plan domain.Plan) int {
	days := plan.Days
	if len(days) < 2 {
		return 0
	}
	if len(days) == 2 {
		return dayPairScoreForTwoDays(days[0].DayOfWeek, days[1].DayOfWeek)
	}
	minS := 100
	for i := 0; i < len(days)-1; i++ {
		s := dayPairScoreForTwoDays(days[i].DayOfWeek, days[i+1].DayOfWeek)
		if s < minS {
			minS = s
		}
	}
	if minS == 100 {
		return 5
	}
	return minS
}

func dayPairScoreForTwoDays(a, b string) int {
	p := unorderedPairKey(a, b)
	switch p {
	case unorderedPairKey("Friday", "Saturday"):
		return 20
	case unorderedPairKey("Saturday", "Sunday"):
		return 30
	case unorderedPairKey("Sunday", "Monday"):
		return 15
	case unorderedPairKey("Saturday", "Monday"):
		return 12
	case unorderedPairKey("Friday", "Sunday"):
		return 8
	default:
		return 5
	}
}

func unorderedPairKey(a, b string) string {
	if a < b {
		return a + "\x00" + b
	}
	return b + "\x00" + a
}

func summedSessionComponent(plan domain.Plan, byID map[string]domain.Session, prefs domain.Preferences) int {
	sum := 0
	for _, d := range plan.Days {
		s, ok := byID[d.PrimarySessionID]
		if !ok {
			return 0
		}
		sc, _ := ScoreSession(s, prefs)
		sum += sc
	}
	if sum > 50 {
		return 50
	}
	return sum
}

func varietyComponent(plan domain.Plan, byID map[string]domain.Session) int {
	if len(plan.Days) < 2 {
		return 0
	}
	sports := make([]string, 0, len(plan.Days))
	for _, d := range plan.Days {
		s, ok := byID[d.PrimarySessionID]
		if !ok {
			return 0
		}
		sports = append(sports, s.Sport)
	}
	if len(plan.Days) == 2 {
		if sports[0] != sports[1] {
			return 10
		}
		return 0
	}
	seen := make(map[string]struct{})
	for _, sp := range sports {
		if _, ok := seen[sp]; ok {
			return 0
		}
		seen[sp] = struct{}{}
	}
	return 10
}

func convenienceComponent(plan domain.Plan, byID map[string]domain.Session) int {
	if len(plan.Days) == 0 {
		return 0
	}
	allOK := true
	for _, d := range plan.Days {
		s, ok := byID[d.PrimarySessionID]
		if !ok {
			return 0
		}
		if !startTimeInWindow(s.StartTime, 10, 0, 18, 0) {
			allOK = false
			break
		}
	}
	if !allOK {
		return 0
	}
	if len(plan.Days) == 1 {
		return 5
	}
	return 10
}

func startTimeInWindow(start string, h1, m1, h2, m2 int) bool {
	hh, mm, ok := parseHHMM(start)
	if !ok {
		return false
	}
	t := hh*60 + mm
	lo := h1*60 + m1
	hi := h2*60 + m2
	return t >= lo && t <= hi
}

func parseHHMM(s string) (h, m int, ok bool) {
	s = strings.TrimSpace(s)
	if len(s) < 4 || len(s) > 5 {
		return 0, 0, false
	}
	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return 0, 0, false
	}
	var err error
	if h, err = parseIntField(parts[0]); err != nil {
		return 0, 0, false
	}
	if m, err = parseIntField(parts[1]); err != nil {
		return 0, 0, false
	}
	if h < 0 || h > 23 || m < 0 || m > 59 {
		return 0, 0, false
	}
	return h, m, true
}

func parseIntField(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

// RankPlans sorts valid plans by score; invalid excluded unless includeInvalid.
func RankPlans(plans []domain.Plan, dataset []domain.Session, prefs domain.Preferences, includeBreakdown, includeInvalid bool) RankPlansResponse {
	byID := make(map[string]domain.Session, len(dataset))
	for _, s := range dataset {
		byID[s.ID] = s
	}

	var ranked []RankedPlan
	var invalid []InvalidPlanEntry

	for _, p := range plans {
		v := ValidatePlan(p, dataset, prefs)
		if !v.Valid {
			if includeInvalid {
				invalid = append(invalid, InvalidPlanEntry{Plan: p, Validation: v})
			}
			continue
		}
		total, comp := ScorePlan(p, byID, prefs)
		entry := RankedPlan{Plan: p, Score: total}
		if includeBreakdown {
			c := comp
			entry.Components = &c
		}
		ranked = append(ranked, entry)
	}

	sort.SliceStable(ranked, func(i, j int) bool {
		return compareRankedPlans(ranked[i], ranked[j]) < 0
	})

	return RankPlansResponse{Plans: ranked, InvalidPlans: invalid}
}

func compareRankedPlans(a, b RankedPlan) int {
	if a.Score != b.Score {
		return b.Score - a.Score
	}
	ca, cb := a.Components, b.Components
	if ca != nil && cb != nil {
		if ca.DayPair != cb.DayPair {
			return cb.DayPair - ca.DayPair
		}
		if ca.SummedSessionScore != cb.SummedSessionScore {
			return cb.SummedSessionScore - ca.SummedSessionScore
		}
		if ca.Variety != cb.Variety {
			return cb.Variety - ca.Variety
		}
		if ca.Convenience != cb.Convenience {
			return cb.Convenience - ca.Convenience
		}
	}
	idsA := primaryIDs(a.Plan)
	idsB := primaryIDs(b.Plan)
	sort.Strings(idsA)
	sort.Strings(idsB)
	return strings.Compare(strings.Join(idsA, ","), strings.Join(idsB, ","))
}

func primaryIDs(p domain.Plan) []string {
	out := make([]string, 0, len(p.Days))
	for _, d := range p.Days {
		out = append(out, d.PrimarySessionID)
	}
	return out
}

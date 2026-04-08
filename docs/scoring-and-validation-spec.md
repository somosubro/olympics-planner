# Scoring and Validation Spec

## 1. Purpose

This document defines how the Olympics Sessions Planner validates and ranks sessions and plans.

It exists to prevent correctness drift between:
- GPT prompt behavior
- backend logic
- importer output
- future UI behavior

This document is the source of truth for:
- what makes a session valid
- what makes a plan valid
- what makes one valid option rank above another
- which rules are hard rejects versus soft preferences
- how overrides are applied
- how scores are calculated for MVP
- how validation and scoring results are represented

---

## 2. Core Principle

Validation and scoring are separate concerns.

### Validation
Validation answers:
- Is this session allowed?
- Is this plan allowed?

Validation is binary.

A session or plan is either:
- valid
- invalid

### Scoring
Scoring answers:
- Among valid options, which one is better?

Scoring is comparative.

A valid plan can still rank lower than another valid plan.

---

## 3. Rule Categories

All rules must be classified into one of these categories.

### 3.1 Hard Rules
Hard rules are mandatory.

If a hard rule fails:
- the session is invalid, or
- the plan is invalid

Hard rules must never be overridden by ranking.

Examples:
- sport is not allowed
- day is not allowed
- session does not exist in dataset
- same sport appears across both days when forbidden
- plan shape does not match request shape

### 3.2 Soft Rules
Soft rules are preferences.

If a soft rule is not satisfied:
- the session or plan remains valid
- its score may be lower

Examples:
- prefer Saturday/Sunday over Friday/Monday
- prefer higher-priority sports
- prefer stronger sessions
- prefer more variety
- prefer more convenient timing

---

## 4. Scope

Validation and scoring exist at two levels:
- session level
- plan level

### 4.1 Session-Level
Used when filtering and ranking individual sessions.

### 4.2 Plan-Level
Used when validating and ranking assembled one-day or multi-day plans.

---

## 5. Data and Time Assumptions

### 5.1 Source of Truth
All runtime validation and scoring must use imported session data from the planner dataset.

No runtime logic may invent sessions not present in the dataset.

### 5.2 Timezone
All session times are interpreted in the local LA time represented by the source schedule unless a future data contract explicitly changes this.

### 5.3 MVP Derived Fields
The PRD mentions optional derived fields such as:
- `stage`
- `zone`
- `keywords`
- `interesting`
- `finalsHeavy`
- `marquee`
- `durationMins`

For MVP:
- these fields may exist in importer output
- but scoring must not depend on them unless they are formally included in the runtime data contract

MVP scoring should therefore rely only on fields guaranteed by the runtime contract, unless and until those derived fields are formally promoted into the domain model.

---

## 6. Plan Object Shape (MVP)

This section defines the minimum plan shape used for validation.

### 6.1 Canonical Session Identity
For runtime validation and plan objects:
- `primarySessionId`, `alternateSessionIds`, and `sessionIds` entries must refer to canonical `Session.id`
- `sessionCode` is a display/business field, not the primary identity key in plan objects

If a future API accepts `sessionCode` as input, it must resolve it to canonical `Session.id` before validation.

### 6.2 One-Day Plan
A one-day plan contains:
- `planType = "one_day"`
- exactly one `days` entry
- exactly one `primarySessionId` for that day
- zero or more `alternateSessionIds`

Example:

```json
{
  "planType": "one_day",
  "days": [
    {
      "date": "2028-07-15",
      "dayOfWeek": "Saturday",
      "primarySessionId": "session-ten-12",
      "alternateSessionIds": ["session-ath-09", "session-swm-04"]
    }
  ]
}
```

### 6.3 Two-Day / Weekend Plan
A two-day plan contains:
- `planType = "two_day"`
- exactly two `days` entries
- exactly one `primarySessionId` per day
- zero or more `alternateSessionIds` per day

A weekend plan is a specific two-day plan where:
- one day is Saturday
- one day is Sunday

Example:

```json
{
  "planType": "two_day",
  "days": [
    {
      "date": "2028-07-15",
      "dayOfWeek": "Saturday",
      "primarySessionId": "session-ten-12",
      "alternateSessionIds": ["session-ath-09"]
    },
    {
      "date": "2028-07-16",
      "dayOfWeek": "Sunday",
      "primarySessionId": "session-ckt-03",
      "alternateSessionIds": ["session-div-02"]
    }
  ]
}
```

### 6.3.1 Weekend requests and `planType`
When the user request is for a **weekend** (see §7.3), the assembled plan must:
- use `planType = "two_day"`
- contain **exactly two** `days` entries
- use **Saturday** and **Sunday** as `dayOfWeek` (and `date` values that match those weekdays)

Do not represent a default weekend plan as `multi_day`.

### 6.3.2 Same day, multiple sessions (`sessionIds`)
A plan day may instead list **`sessionIds`**: a non-empty array of session ids for that calendar day (co-equal; not primary vs alternate). Do not combine `sessionIds` with `primarySessionId` / `alternateSessionIds` on the same day.

### 6.4 Multi-Day Plan
A multi-day plan contains:
- `planType = "multi_day"`
- **three or more** `days` entries
- exactly one `primarySessionId` per day
- zero or more `alternateSessionIds` per day

`multi_day` must not be used for exactly two days. For exactly two days, use `two_day` (§6.3).

### 6.4.1 Normative: `two_day` vs `multi_day`
- **`two_day`**: exactly **two** `days` entries (any allowed two-day shape the user asked for, including Sat+Sun weekend).
- **`multi_day`**: **three or more** `days` entries.

A two-day Saturday+Sunday **weekend** plan uses `planType: "two_day"`, not `multi_day`.

Example:

```json
{
  "planType": "multi_day",
  "days": [
    {
      "date": "2028-07-15",
      "dayOfWeek": "Saturday",
      "primarySessionId": "session-ten-12",
      "alternateSessionIds": []
    },
    {
      "date": "2028-07-16",
      "dayOfWeek": "Sunday",
      "primarySessionId": "session-ckt-03",
      "alternateSessionIds": ["session-div-02"]
    },
    {
      "date": "2028-07-22",
      "dayOfWeek": "Saturday",
      "primarySessionId": "session-ath-14",
      "alternateSessionIds": ["session-swm-08"]
    }
  ]
}
```

### 6.5 Alternates
Alternates are part of the same validated plan object.

For MVP:
- alternates are validated
- alternates are not part of the main plan score unless explicitly stated in a future revision
- alternates must never duplicate the primary session for the same day
- alternates should ideally provide meaningful diversity

### 6.6 Empty Day Entries
A plan must not contain empty day objects.

Invalid examples:
- a `days` entry with no `primarySessionId` **and** no non-empty `sessionIds`
- a `days` entry with no usable date/day information
- a `days` entry with an empty object

### 6.7 MVP Daily Session Limit
For MVP:
- **Legacy shape:** at most **one primary** per day; zero or more alternates on that day.
- **`sessionIds` shape:** multiple sessions on one day are represented as a single list (see §6.3.2); the server does not resolve time-overlap conflicts between them.

This rule avoids introducing same-day scheduling conflict logic in MVP beyond listing multiple ids.

---

## 7. Default Edge-Case Rules

These rules must be applied consistently unless the request explicitly overrides them.

### 7.1 Empty `allowedSports`
Empty `allowedSports` means **allow none**.

This is effectively an invalid usable configuration and should produce no valid sessions.

Reason:
Treating empty as allow-all is dangerous and can silently violate user or system expectations.

### 7.2 Empty `allowedDays`
Empty `allowedDays` means **allow none**.

This is effectively an invalid usable configuration and should produce no valid sessions.

### 7.3 Meaning of “Weekend”
For MVP, “weekend” means:
- `Saturday + Sunday`

It does not mean:
- Friday + Saturday
- Friday + Sunday
- Friday through Sunday

Those may still be valid multi-day combinations if explicitly requested, but the default interpretation of “weekend” is Saturday + Sunday.

### 7.3.1 Weekend plans and `planType`
A natural-language **weekend** plan must be serialized as `planType = "two_day"` with **Saturday and Sunday** only, unless the user explicitly asked for a different two-day pair (in which case it is still `two_day`, but not the default weekend meaning).

### 7.4 Missing Title
If `title` is missing, display may fall back to a synthesized title:
- `<sport> - <sessionCode>`

However:
- title is still strongly preferred for UX
- missing title should incur a small data-quality penalty if score penalties are enabled

### 7.5 Missing `endTime`
Missing `endTime` is allowed for MVP validation if all other required fields are present.

However:
- missing `endTime` may incur a small data-quality penalty
- timing-based scoring must not depend on `endTime` when missing

### 7.6 Missing `includedEvents`
Missing or empty `includedEvents` is allowed for MVP validation.

However:
- this should incur a small data-quality penalty
- such sessions may still be displayed, but with degraded detail quality

---

## 8. Session Validation Rules

A session is valid only if all of the following are true.

### SV1. Session Must Exist
The session must exist in the imported dataset.

#### Pass
- session ID resolves to a known session

#### Fail
- unknown session
- invented session
- stale reference not found in dataset

---

### SV2. Sport Must Be Allowed
The session’s sport must be in the allowed sport set.

#### Pass
- sport is included in configured allowed sports
- sport is allowed by current request overrides

#### Fail
- sport is not allowed by base preferences
- sport is explicitly excluded by request overrides

---

### SV3. Day Must Be Allowed
The session’s day of week must be in the allowed day set.

#### Pass
- day is in configured allowed days
- day is allowed by current request overrides

#### Fail
- day is outside configured allowed days
- user request restricts to a different set of days

---

### SV4. Required Fields Must Be Present
A session must contain the minimum required fields needed for runtime validation.

#### Required for validation
- `id`
- `sport`
- `sessionCode`
- `date`
- `dayOfWeek`
- `startTime`
- `venue`

#### Required for high-quality display
- `title`

#### Preferred but not required for MVP validation
- `endTime`
- `includedEvents`

If `title` is missing:
- display may fall back to `<sport> - <sessionCode>`

If `endTime` or `includedEvents` are missing:
- validation may still pass
- a small data-quality penalty may apply during scoring

---

### SV5. Date and Day Must Be Consistent
If both `date` and `dayOfWeek` are present, they must represent the same real day.

#### Fail
- `date = 2028-07-15` but `dayOfWeek = Sunday` when the real day is different

This is primarily an importer/data-quality validation rule, but runtime validation may also enforce it.

---

## 9. Plan Validation Rules

A plan is valid only if all of the following are true.

### PV1. All Sessions Must Be Valid
Every primary and alternate session in the plan must pass session-level validation.

---

### PV2. No Duplicate Session IDs
A plan must not contain the same session more than once anywhere in the validated object.

#### Fail
- same session repeated as primary and alternate on the same day
- same session repeated across multiple days
- same session repeated across alternate lists in the same day
- same session repeated across alternate lists on different days

---

### PV3. Plan Shape Must Match Request Shape
The plan must match the requested planning shape.

Examples:
- one-day request -> exactly one day entry and `planType` must be `one_day`
- `two_day` shape -> `planType` must be `two_day` and exactly two `days` entries
- `multi_day` shape -> `planType` must be `multi_day` and **three or more** `days` entries
- weekend request -> must satisfy §6.3.1 (exactly two `days`: Saturday and Sunday)
- Saturday-only request -> all day entries must be Saturday
- one session per day MVP rule must be respected

This rule belongs partly to application/GPT orchestration, but once a plan object is assembled, it must still be validated here.

---

### PV4. No Same Sport Across Different Days When Forbidden
If `noSameSportAcrossDays = true`, the same sport must not appear across different days in the same two_day or multi_day plan, counting **every session in the plan** (primary **and** alternate session IDs).

#### Pass
- Saturday Tennis + Sunday Athletics (no sport repeats across days)
- Saturday Diving + Sunday Swimming (different sports each day)

#### Fail
- Saturday Tennis + Sunday Tennis (any combination of primary/alternate slots)
- Saturday Athletics primary + Sunday Tennis primary + Sunday Athletics alternate (athletics appears on two calendar days)

This rule applies to **primary and alternate** session slots so add-ons cannot repeat a sport on another day.

---

### PV5. No Disallowed Sports in Final Plan
The final plan must not contain disallowed sports.

This is intentionally redundant with session validation.

---

### PV6. No Disallowed Days in Final Plan
The final plan must not contain disallowed days.

This is intentionally redundant with session validation.

---

### PV7. Alternates Must Also Be Valid
If alternates are present:
- each alternate must pass session validation
- alternates must respect the same hard filters
- alternates must not duplicate the primary session for that day
- alternates must not duplicate each other within the same day

---

### PV8. No Empty Day Entries
Every day entry in a plan must contain:
- a valid date and/or day representation
- a valid `primarySessionId`

A day entry with no primary session is invalid in MVP.

---

## 10. Override Resolution

The system supports temporary conversational overrides for a single request.

Examples:
- “avoid cricket”
- “prioritize tennis”
- “only Saturdays”

Overrides are resolved before validation and scoring.

### 10.1 Override Precedence
Order of precedence, highest to lowest:
1. current request override
2. persisted preference set, if later added
3. default preference configuration

### 10.2 Hard Restriction Overrides
Overrides that narrow the allowed set must be treated as hard validation inputs.

Examples:
- “only Saturdays”
- “avoid cricket”
- “only athletics and tennis”

These affect:
- which sessions are eligible
- which plans are valid

### 10.3 Soft Preference Overrides
Overrides that change ranking but not eligibility must be treated as scoring inputs.

Examples:
- “prioritize tennis”
- “prefer relaxed schedule”
- “prefer marquee sessions”

These affect:
- ranking among valid sessions
- ranking among valid plans

They do not make otherwise invalid sessions valid.

---

## 11. Scoring Overview

Scoring applies only to valid sessions and valid plans.

The ranking model must be:
- deterministic
- explainable
- additive for MVP
- configurable later
- easy to test

Scoring happens in two layers:
- session scoring
- plan scoring

For MVP:
- plan ranking uses a numeric weighted model
- all weights and ranges defined here are normative for MVP unless explicitly changed in a later revision

---

## 12. MVP Numeric Scoring Model

This section defines the concrete scoring model for MVP.

### 12.1 Session Score Range
Each valid session receives a score in the range:
- `0` to `40`

### 12.2 Plan Score Range
Each valid plan receives a score in the range:
- `0` to `100`

### 12.3 Plan Score Formula

For MVP:

```text
plan_score =
  day_pair_score
  + summed_session_score
  + variety_score
  + convenience_score
```

Where:
- `day_pair_score` is in `0..30`
- `summed_session_score` is capped at `0..50`
- `variety_score` is in `0..10`
- `convenience_score` is in `0..10`

Maximum total:
- `100`

### 12.4 Dominance Order
Numeric weights must preserve this practical order:

1. validity first
2. day pair quality
3. session quality
4. variety
5. convenience
6. minor tie-breakers

This means:
- invalid plans are never scored as final candidates
- stronger day pairs should usually beat weaker day pairs when session quality is close
- much stronger sessions may still beat a better day pair if the difference is large enough

---

## 13. Session Scoring

Session scoring ranks individual valid sessions.

### 13.1 Session Score Components

For MVP, a session score is:

```text
session_score =
  sport_priority_score
  + data_quality_score
```

Optional future inputs such as:
- stage
- finals-heavy
- marquee
- interesting
- duration

must not affect MVP session score unless they are formally part of the runtime data contract.

### 13.2 Session Score Range
For MVP:
- `sport_priority_score`: `0..30`
- `data_quality_score`: `0..10`

Total:
- `0..40`

### 13.3 Sport Priority Score
Higher-priority sports should score higher.

This score is based on the configured sport priority ordering among allowed sports.

#### Recommended mapping
Given `N` allowed sports in priority order:
- top priority sport receives `30`
- bottom priority sport receives at least `5`
- middle sports are linearly distributed between them

Example with 4 sports:
1. Tennis -> 30
2. Athletics -> 22
3. Swimming -> 14
4. Diving -> 6

The exact mapping function may be implemented programmatically, but it must remain deterministic.

### 13.4 Data Quality Score
This component rewards sessions with complete usable display/runtime data.

Suggested MVP breakdown:
- `+4` if `title` present
- `+2` if `endTime` present
- `+4` if `includedEvents` is non-empty

Total max:
- `10`

This is a small quality adjustment only. It must not dominate sport priority.

### 13.5 Future Session Quality Inputs
The following are intentionally deferred unless formally promoted into the runtime contract:
- `stage`
- `finalsHeavy`
- `marquee`
- `interesting`
- sport-specific strength heuristics

If and when they are promoted, this document should be revised and the numeric scoring model updated.

---

## 14. Plan Scoring

Plan scoring ranks valid plans against each other.

### 14.1 Plan Score Components
For MVP:

```text
plan_score =
  day_pair_score
  + summed_session_score
  + variety_score
  + convenience_score
```

### 14.2 Day Pair Score
Some day combinations are preferable.

For MVP default multi-day ranking:
- Saturday + Sunday -> `30`
- Friday + Saturday -> `20`
- Sunday + Monday -> `15`
- Saturday + Monday -> `12`
- Friday + Sunday -> `8`
- other valid pairs -> `5`

For `multi_day` plans with **three or more** days, `day_pair_score` is the **minimum** of the §14.2 scores for each **consecutive** pair of days in `days` order. If a pair is not listed in §14.2, use **other valid pairs** (`5`).

For one-day plans:
- `day_pair_score = 0`

If the request explicitly specifies a day shape, only valid matching shapes should be compared.

### 14.3 Summed Session Score
A plan should benefit from containing stronger individual sessions.

For MVP:
- **Legacy plan days** (`primarySessionId` / `alternateSessionIds`): sum the session scores of **primary sessions only** (alternates are validated but do not add to this sum).
- **Session-list days** (`sessionIds` non-empty): sum the session scores of **every** id in `sessionIds` for that day.
- cap the total at `50`

Examples:
- one-day plan with session score `34` -> `34`
- two-day plan with session scores `28 + 26` -> capped at `50` if needed

This cap prevents session totals from completely overwhelming day-pair structure.

### 14.4 Variety Score
Plans with better variety may score higher.

For MVP:
- different sports across days -> `+10`
- one-day plans -> `0`
- repeated sport across days when repetition is allowed -> `0`

If `noSameSportAcrossDays = true`, repeated sport across days is invalid and this score is not reached.

For `multi_day` plans with **three or more** days:
- `variety_score = 10` if every primary session’s **sport** is **unique** across all days
- otherwise `variety_score = 0`

For exactly two days, the existing rule applies (different sports across the two days -> `10`).

### 14.5 Convenience Score
Convenience score is a small soft preference for simpler, more usable plans.

For MVP:
- one-day plan with start time between 10:00 and 18:00 -> `+5`
- two_day or multi_day plan where, **for each day**, every counted session starts between 10:00 and 18:00 -> `+10`
  - legacy days: only the **primary** session is counted here (unchanged)
  - `sessionIds` days: **every** listed session must fall in the window
- otherwise -> `0`

This is intentionally simple and must remain low-weight.

---

## 15. Tie-Breaking Rules

If two valid plans share the same **`plan_score`**, break ties using **component values first**, then IDs.

### 15.1 Plan tie-breakers (in order)
1. Higher **`day_pair_score`** (component)
2. Higher **`summed_session_score`** (component)
3. Higher **`variety_score`** (component)
4. Higher **`convenience_score`** (component)
5. Higher **total session count** (all ids across all days—primary + alternates + `sessionIds`)
6. Lexicographic order of **sorted first-session `Session.id` per day** (ascending; first id in `sessionIds`, else primary)

### 15.2 Session tie-breakers (in order)
1. Higher **`session_score`**
2. Higher **`sport_priority_score`** component
3. Higher **`data_quality_score`** component
4. Lexicographic **`Session.id`**

All tie-breakers must be deterministic.

---

## 16. Alternate Selection Rules

Alternates are valid options shown alongside the main recommendation.

### 16.1 Rules for Alternates
- must pass the same validation rules as primary picks
- must not duplicate the primary session for the same day
- must not duplicate each other within the same day
- should ideally provide meaningful variety
- should still be high quality

### 16.2 Alternate Diversity Preference
Alternates should prefer diversity where possible:
- different sport
- different session profile
- different vibe or intensity

Alternates should not simply be near-duplicates unless the pool is very small.

### 16.3 Alternates and Scoring
For MVP:
- alternates are validated
- alternates are not included in the main `plan_score`
- alternates may be ranked separately using `session_score`

---

## 17. Worked Examples

### 17.1 Example A: Valid One-Day Plan
Request:
- Saturday only
- allowed sports: Tennis, Athletics

Plan:
- one Saturday day entry
- primary session is a valid Tennis session
- alternates are valid Athletics session(s)

Result:
- valid

### 17.2 Example B: Invalid Multi-Day Plan
Request:
- weekend plan
- `noSameSportAcrossDays = true`

Plan:
- Saturday primary: Tennis
- Sunday primary: Tennis

Result:
- invalid

Reason:
- repeated sport across days

### 17.3 Example C: Better Day Pair Wins When Session Quality Is Close
Plan 1:
- Saturday session: Athletics, score `22`
- Sunday session: Tennis, score `18`
- day pair = Saturday + Sunday -> `30`
- summed session score = `40`
- variety = `10`
- convenience = `10`

Total:
- `30 + 40 + 10 + 10 = 90`

Plan 2:
- Friday session: Athletics, score `24`
- Saturday session: Tennis, score `19`
- day pair = Friday + Saturday -> `20`
- summed session score = `43`
- variety = `10`
- convenience = `10`

Total:
- `20 + 43 + 10 + 10 = 83`

Result:
- Plan 1 ranks higher

Reason:
- day pair advantage outweighs the small session-quality gap

### 17.4 Example D: Much Stronger Sessions Can Beat a Better Day Pair
Plan 1:
- Saturday session score `14`
- Sunday session score `12`
- day pair = `30`
- summed session score = `26`
- variety = `10`
- convenience = `10`

Total:
- `76`

Plan 2:
- Friday session score `28`
- Saturday session score `24`
- day pair = `20`
- summed session score capped at `50`
- variety = `10`
- convenience = `10`

Total:
- `90`

Result:
- Plan 2 ranks higher

Reason:
- session-quality advantage is large enough to overcome the weaker day pair

---

## 18. Validation Result Contract

Validation responses must use stable machine-readable codes.

Recommended shape:

```json
{
  "valid": false,
  "errors": [
    {
      "code": "REPEATED_SPORT_ACROSS_DAYS",
      "message": "Tennis appears on multiple days in a plan where repeated sports are forbidden.",
      "field": "days"
    }
  ]
}
```

### 18.1 Session Validation Error Codes
- `SESSION_NOT_FOUND`
- `DISALLOWED_SPORT`
- `DISALLOWED_DAY`
- `MISSING_REQUIRED_FIELD`
- `DATE_DAY_MISMATCH`

### 18.2 Plan Validation Error Codes
- `INVALID_PLAN_SHAPE`
- `INVALID_PLAN_TYPE_FOR_DAY_COUNT`
- `DUPLICATE_SESSION`
- `REPEATED_SPORT_ACROSS_DAYS`
- `EMPTY_DAY_ENTRY`
- `CONFLICTING_DAY_SPEC`
- `TOO_MANY_SESSIONS_PER_DAY`
- `INSUFFICIENT_SAME_DAY_GAP`
- `SAME_DAY_SESSION_OVERLAP`
- `INCOMPLETE_SESSION_TIME`

### 18.3 Recommended Field Values
Examples of `field` values:
- `id`
- `sport`
- `dayOfWeek`
- `date`
- `days`
- `days[0].primarySessionId`
- `days[1].alternateSessionIds`
- `days[0].sessionIds`

Clients and tests must rely on `code`, not on parsing free-text `message`.

---

## 19. Scoring Result Contract

For debugging and internal inspection, scoring should return component breakdowns.

Recommended shape:

```json
{
  "score": 90,
  "components": {
    "dayPair": 30,
    "summedSessionScore": 40,
    "variety": 10,
    "convenience": 10
  }
}
```

For session-level ranking:

```json
{
  "score": 28,
  "components": {
    "sportPriority": 22,
    "dataQuality": 6
  }
}
```

This breakdown may remain internal even if end users do not see it by default.

---

## 20. Implementation Guidance

### 20.1 Validate First
Never score invalid sessions or invalid plans as final candidates.

### 20.2 Keep Validation Pure
Validation must not depend on ranking weights.

### 20.3 Keep Scoring Configurable Later
Although MVP defines fixed numeric ranges, later versions may externalize weights into config.

### 20.4 Keep GPT Out of Hard Correctness
GPT may propose, filter, assemble, or format, but final validity must come from deterministic application logic.

### 20.5 Keep MVP Narrow
MVP explicitly avoids:
- multiple primary sessions per day
- same-day scheduling conflict logic
- travel-time optimization
- runtime dependence on importer-only derived fields

---

## 21. Testing Requirements

At minimum, tests should cover:

### Validation Tests
- session exists / does not exist
- allowed sport / disallowed sport
- allowed day / disallowed day
- duplicate session
- repeated sport across days
- invalid plan shape
- empty day entry
- valid alternates / invalid alternates

### Scoring Tests
- higher-priority sport outranks lower-priority sport
- Saturday + Sunday outranks Friday + Saturday when session quality is close
- large session-quality advantage can overcome weaker day pair
- variety bonus applies correctly
- tie-breakers are stable and deterministic
- missing title/endTime/includedEvents apply correct quality adjustments

### Regression Tests
- known “best weekend plan” scenarios
- known “Saturday only” scenarios
- known override scenarios such as “avoid cricket”

---

## 22. Open Design Decisions

The following still need later decisions:
- whether derived fields like `stage`, `marquee`, and `finalsHeavy` should be promoted into the runtime contract
- whether convenience scoring should become richer than simple time-window heuristics
- whether alternates should influence plan-level ranking in a later version
- whether one-day plan ranking should add explicit preferred-day scoring

---

## 23. Summary

This system must:
- reject invalid sessions and plans deterministically
- score only valid options
- keep hard rules separate from soft preferences
- use an explicit numeric model for MVP
- remain explainable and testable
- preserve the architecture split:
  - GPT for flexibility and orchestration
  - backend for filtering, ranking, and validation

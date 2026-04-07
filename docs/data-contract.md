# Data Contract

## 1. Purpose

This document defines the canonical data shapes used by the Olympics Sessions Planner.

It exists to prevent drift between:
- importer output
- runtime domain models
- scoring and validation logic
- HTTP/API request and response shapes
- GPT/application-layer orchestration

This document is the source of truth for:
- canonical field names
- required versus optional fields
- identity and uniqueness rules
- which fields are MVP runtime fields versus importer-only or future fields
- the core shapes for sessions, preferences, plans, filters, and result objects

---

## 2. Design Principles

### 2.1 One Canonical Runtime Shape
The runtime planner must operate on a single canonical `Session` shape.

Importer code may use intermediate/raw structs, but all runtime logic must use the canonical contract defined here.

### 2.2 Identity Must Be Explicit
A session must have one canonical primary identifier used by:
- plan objects
- validation
- ranking
- API request/response references

### 2.3 Required and Optional Fields Must Be Clear
Fields must be classified as:
- required for runtime validation
- required for high-quality display
- optional for MVP
- future / importer-only

### 2.4 Runtime Must Not Depend on Unpromoted Fields
If a field is not explicitly part of the runtime contract, validation and scoring must not depend on it.

---

## 3. Versioning and Transitional Notes

This document defines the **MVP v1 target data contract**.

Any future contract-breaking change should either:
- update this document with a version note, or
- introduce a new versioned contract section

### 3.1 Current Code vs Target Contract
Some current Go domain types may not yet fully match this target contract.

In particular:
- the Go `domain.Plan` struct is aligned with the canonical `Plan` shape in §11 (`planType`, `days`, `primarySessionId`, …)
- the filter object defined in this document is the intended service/API contract and may not yet exist as a matching Go domain type

When current code and this document differ, this document defines the target runtime/API contract that future refactors should align to.

---

## 4. Canonical Session Identity

### 4.1 Primary Identity
The canonical session identity field is:

- `id: string`

This is the only identifier that should be used in:
- `Plan` objects
- validation
- runtime internal references
- API request/response references where session identity is required

### 4.2 Session Code
Each session also has:

- `sessionCode: string`

`sessionCode` is:
- a business/display identifier
- useful for humans and debugging
- not the canonical identity field in runtime plan objects

### 4.3 Identity Resolution Rule
If an external caller provides `sessionCode`, the application/API layer must resolve it to canonical `id` before validation and scoring.

### 4.4 Session `id` Values and the Importer
The contract requires `id` to be a **unique string** per session in the runtime dataset. Examples in this document use prefixed forms such as `session-ten-12` for clarity.

The **PDF importer may set `id` to the official schedule session code** (for example `TEN12` or `ATH09`) when that code is unique in the imported file. In that case `sessionCode` typically matches `id` or repeats the same business code. That is acceptable for MVP as long as:
- every session row has a distinct `id` in the dataset, and
- plans and APIs always reference sessions by that same `id` string.

A later normalization step may introduce stable prefixed IDs; until then, treat **`id` as whatever unique string the pipeline wrote** to `data/sessions.json`.

---

## 5. Session Contract

### 5.1 Canonical Runtime Session Shape

```json
{
  "id": "session-ten-12",
  "sport": "Tennis",
  "sessionCode": "TEN12",
  "title": "Tennis - Session TEN12",
  "date": "2028-07-15",
  "dayOfWeek": "Saturday",
  "startTime": "14:00",
  "endTime": "17:00",
  "venue": "LA Tennis Center",
  "includedEvents": [
    "Men's Singles Quarterfinal",
    "Women's Singles Quarterfinal"
  ]
}
```

### 5.2 Session Field Definitions

#### `id`
- type: `string`
- required: yes
- meaning: canonical unique runtime identifier

#### `sport`
- type: `string`
- required: yes
- meaning: normalized sport name used in filtering, validation, and ranking

#### `sessionCode`
- type: `string`
- required: yes
- meaning: business/display session code from source data

#### `title`
- type: `string`
- required: required for high-quality display, not strictly required for MVP validation
- fallback: may be synthesized as `<sport> - <sessionCode>` if missing
- serialization note: optional string fields may be represented as omitted or as an empty string at the serialization layer; validation semantics treat both as missing

#### `date`
- type: `string`
- format: `YYYY-MM-DD`
- required: yes

#### `dayOfWeek`
- type: `string`
- required: yes
- allowed values:
  - `Monday`
  - `Tuesday`
  - `Wednesday`
  - `Thursday`
  - `Friday`
  - `Saturday`
  - `Sunday`

#### `startTime`
- type: `string`
- format: `HH:MM`
- required: yes
- meaning: 24-hour local LA time

#### `endTime`
- type: `string`
- format: `HH:MM`
- required: no for MVP validation
- note: preferred for display and convenience scoring

#### `venue`
- type: `string`
- required: yes

#### `includedEvents`
- type: `string[]`
- required: no for MVP validation
- note: preferred for high-quality display

---

## 6. Required vs Optional Session Fields

### 6.1 Required for Runtime Validation
These fields must be present for a session to pass MVP runtime validation:
- `id`
- `sport`
- `sessionCode`
- `date`
- `dayOfWeek`
- `startTime`
- `venue`

### 6.2 Required for High-Quality Display
These fields are not strictly required for MVP validation, but are required for best UX:
- `title`

### 6.3 Preferred but Optional for MVP
These fields improve output quality or future scoring:
- `endTime`
- `includedEvents`

---

## 7. Derived Fields

The following fields may exist in importer output or future runtime shapes:
- `stage`
- `zone`
- `keywords`
- `interesting`
- `finalsHeavy`
- `marquee`
- `durationMins`

### 7.1 MVP Status of Derived Fields
For MVP:
- these fields are **not part of the required runtime contract**
- runtime validation must not depend on them
- MVP scoring must not depend on them

### 7.2 Promotion Rule
If any derived field becomes required for:
- runtime ranking
- runtime validation
- API output guarantees

then it must be promoted into the canonical runtime contract in this document.

---

## 8. Normalization Rules

### 8.1 Sport
`sport` must be normalized to a stable human-readable canonical value, for example:
- `Athletics`
- `Cricket`
- `Swimming`
- `Tennis`
- `Field Hockey`
- `Basketball`
- `Beach Volleyball`

Avoid mixed naming variants in runtime data.

### 8.2 Date
`date` must use:
- `YYYY-MM-DD`

### 8.3 Time
`startTime` and `endTime` must use:
- `HH:MM`
- 24-hour local LA time as represented by the source schedule

### 8.4 Day of Week
`dayOfWeek` must use full English weekday names:
- `Monday`
- `Tuesday`
- `Wednesday`
- `Thursday`
- `Friday`
- `Saturday`
- `Sunday`

---

## 9. Preferences Contract

### 9.1 Canonical Preferences Shape

```json
{
  "allowedSports": ["Athletics", "Tennis", "Swimming"],
  "sportPriority": ["Athletics", "Tennis", "Swimming"],
  "allowedDays": ["Friday", "Saturday", "Sunday", "Monday"],
  "rules": {
    "noSameSportAcrossDays": true,
    "preferDayPairs": [
      ["Saturday", "Sunday"],
      ["Friday", "Saturday"],
      ["Sunday", "Monday"]
    ],
    "sportSpecific": {}
  }
}
```

### 9.2 Preferences Field Definitions

#### `allowedSports`
- type: `string[]`
- required: yes
- meaning: sports eligible for validation and planning

#### `sportPriority`
- type: `string[]`
- required: yes
- meaning: ordered list from highest to lowest soft preference

#### `allowedDays`
- type: `string[]`
- required: yes
- meaning: allowed day-of-week set for validation

#### `rules`
- type: `object`
- required: yes
- meaning: grouped rule configuration for validation and ranking

#### `rules.noSameSportAcrossDays`
- type: `boolean`
- required: yes
- meaning: whether repeated sports across different days invalidate a `two_day` or `multi_day` plan

#### `rules.preferDayPairs`
- type: `string[][]`
- required: no for MVP data loading, but recommended
- meaning: ordered list of preferred day pairs for plan ranking

#### `rules.sportSpecific`
- type: `object` (map of sport key to sport-specific rule payload)
- required: no; may be omitted or `{}` for MVP
- meaning: reserved for future sport-specific validation or scoring hints; must not be required for MVP runtime behavior until promoted in this document

---

## 10. Filter Contract

Filters define how sessions are retrieved before ranking or plan assembly.

This filter object defines the intended service/API shape; a matching Go domain type may be added later.

### 10.1 Canonical Filter Shape (Target / Service/API Contract)

```json
{
  "dates": ["2028-07-15", "2028-07-16"],
  "daysOfWeek": ["Saturday", "Sunday"],
  "sports": ["Athletics", "Tennis"],
  "allowedSports": ["Athletics", "Tennis"],
  "excludedSports": ["Cricket"]
}
```

### 10.2 Filter Field Definitions

#### `dates`
- type: `string[]`
- required: no
- meaning: include only these dates

#### `daysOfWeek`
- type: `string[]`
- required: no
- meaning: include only these day names

#### `sports`
- type: `string[]`
- required: no
- meaning: explicit sport filter

#### `allowedSports`
- type: `string[]`
- required: no
- meaning: hard allowlist restriction

#### `excludedSports`
- type: `string[]`
- required: no
- meaning: hard denylist restriction

### 10.3 Filter Semantics
For MVP:
- all provided filters are conjunctive
- a session must satisfy all specified filter dimensions to be included

Example:
- if both `dates` and `sports` are provided, a session must match one of the dates **and** one of the sports

---

## 11. Plan Contract

### 11.1 Canonical Plan Shape
The canonical `Plan` shape defined here is the target runtime/API contract.

The Go `domain.Plan` type implements this shape.

### 11.2 One-Day Plan

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

### 11.3 Two-Day Plan

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

### 11.4 Multi-Day Plan

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

### 11.5 Plan Type Semantics

#### `one_day`
- exactly one `days` entry

#### `two_day`
- exactly two `days` entries

#### `multi_day`
- three or more `days` entries

A default weekend request is represented as:
- `planType = "two_day"`
- exactly two day entries
- Saturday and Sunday

### 11.6 Plan Field Definitions

#### `planType`
- type: `string`
- required: yes
- allowed MVP values:
  - `one_day`
  - `two_day`
  - `multi_day`

#### `days`
- type: `PlanDay[]`
- required: yes
- meaning: ordered list of day entries in the plan

### 11.7 Plan Day Shape

```json
{
  "date": "2028-07-15",
  "dayOfWeek": "Saturday",
  "primarySessionId": "session-ten-12",
  "alternateSessionIds": ["session-ath-09"]
}
```

#### `date`
- type: `string`
- format: `YYYY-MM-DD`
- required: yes for MVP

#### `dayOfWeek`
- type: `string`
- required: yes for MVP

#### `primarySessionId`
- type: `string`
- required: yes
- must reference canonical `Session.id`

#### `alternateSessionIds`
- type: `string[]`
- required: yes
- may be empty

---

## 12. Plan Invariants

For MVP:
- exactly one primary session per day
- alternates are optional
- no empty day entries
- no duplicate session IDs anywhere in the same validated plan object
- alternates must not duplicate the day’s primary session
- alternates must not duplicate each other within the same day
- `primarySessionId` and `alternateSessionIds` must reference canonical `Session.id`

---

## 13. Validation Result Contract

### 13.1 Shape

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

### 13.2 Field Definitions

#### `valid`
- type: `boolean`
- required: yes

#### `errors`
- type: `ValidationError[]`
- required: yes
- empty when `valid = true`

### 13.3 Validation Error Shape

```json
{
  "code": "DISALLOWED_SPORT",
  "message": "Cricket is not allowed for this request.",
  "field": "sport"
}
```

#### `code`
- type: `string`
- required: yes
- must be machine-readable and stable

#### `message`
- type: `string`
- required: yes
- human-readable diagnostic text

#### `field`
- type: `string`
- required: no but strongly recommended

### 13.4 Standard Error Codes

#### Session validation errors
- `SESSION_NOT_FOUND`
- `DISALLOWED_SPORT`
- `DISALLOWED_DAY`
- `MISSING_REQUIRED_FIELD`
- `DATE_DAY_MISMATCH`

#### Plan validation errors
- `INVALID_PLAN_SHAPE`
- `INVALID_PLAN_TYPE_FOR_DAY_COUNT`
- `DUPLICATE_SESSION`
- `REPEATED_SPORT_ACROSS_DAYS`
- `INVALID_ALTERNATE`
- `EMPTY_DAY_ENTRY`

---

## 14. Scoring Result Contract

### 14.1 Plan Scoring Result

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

### 14.2 Session Scoring Result

```json
{
  "score": 28,
  "components": {
    "sportPriority": 22,
    "dataQuality": 6
  }
}
```

### 14.3 Field Definitions

#### `score`
- type: `number`
- required: yes

#### `components`
- type: `object`
- required: yes
- meaning: score breakdown for debugging and internal explainability

---

## 15. Importer Contract Boundary

### 15.1 Importer May Use Intermediate Shapes
Importer code may use:
- raw extracted text structs
- partially parsed structs
- enriched/import-specific structs

These are not runtime contracts.

### 15.2 Importer Output Must Map to Runtime Session
Before runtime use, importer output must be transformed into the canonical runtime `Session` shape defined in this document.

The primary sessions file on disk is JSON whose **root value is an array** of `Session` objects (see `docs/importer-spec.md` §15.1). Do not use a wrapper object such as `{ "sessions": [...] }` for that file.

See **§4.4** for how `id` may be set from schedule codes during import.

### 15.3 Importer-Only Fields
If importer output includes extra fields not listed in the canonical runtime contract, they must be treated as:
- importer-only metadata, or
- future contract candidates

They must not silently become runtime dependencies.

---

## 16. Example Minimal Session Object

This is the smallest acceptable MVP runtime session object that can still pass validation:

```json
{
  "id": "session-ath-09",
  "sport": "Athletics",
  "sessionCode": "ATH09",
  "date": "2028-07-15",
  "dayOfWeek": "Saturday",
  "startTime": "10:00",
  "venue": "LA Memorial Coliseum"
}
```

Notes:
- `title`, `endTime`, and `includedEvents` are absent
- this may still pass MVP validation
- display and scoring quality may be degraded

---

## 17. Example High-Quality Session Object

```json
{
  "id": "session-ath-09",
  "sport": "Athletics",
  "sessionCode": "ATH09",
  "title": "Athletics - Session ATH09",
  "date": "2028-07-15",
  "dayOfWeek": "Saturday",
  "startTime": "10:00",
  "endTime": "13:00",
  "venue": "LA Memorial Coliseum",
  "includedEvents": [
    "Men's 1500m Semifinal",
    "Women's High Jump Final"
  ]
}
```

---

## 18. Open Decisions

The following remain open and must be resolved before expanding the contract:
- whether derived fields such as `stage`, `marquee`, and `finalsHeavy` should be promoted into runtime contract
- whether convenience/timing data should gain a richer contract
- whether alternates should later influence main plan ranking
- whether API requests should accept `sessionCode` directly or only canonical `id`

---

## 19. Summary

This contract defines:
- the canonical runtime `Session` shape
- the canonical `Preferences` shape
- the canonical target `Filter` shape
- the canonical target `Plan` shape
- validation and scoring result objects
- the boundary between importer data and runtime data

Runtime code must depend only on the contracts defined here, and current implementation should be refactored toward this target contract.

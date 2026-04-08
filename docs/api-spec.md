# API Spec

## 1. Purpose

This document defines the MVP backend API for the Olympics Sessions Planner.

It translates the product, scoring/validation, and data-contract docs into concrete request and response behavior.

This API is a **correctness layer**, not an orchestration layer.

It exists to provide:
- session retrieval
- filtering
- plan validation
- session/plan ranking

It must **not** implement planning workflows such as:
- generate weekend plan
- generate Saturday plan
- generate multi-weekend plan

Those orchestration flows belong in GPT or the application layer.

---

## 2. Design Principles

### 2.1 Backend Is a Correctness Layer
The backend should provide:
- `get_sessions`
- `rank_sessions`
- `validate_plan`

It should not own planning flow orchestration.

### 2.2 Canonical Identity
All runtime identity references must use canonical `Session.id`.

If callers provide `sessionCode`, the application or API adapter layer must resolve it before validation and ranking.

### 2.3 Deterministic Behavior
Given the same inputs:
- validation results must be identical
- ranking results must be identical
- tie-breakers must be deterministic

### 2.4 Invalid Inputs Must Fail Clearly
The API should return stable machine-readable errors, not vague text blobs.

---

## 3. Versioning

This document defines the **MVP v1** API.

Base path examples in this document use:

```text
/api/v1
```

All endpoint paths in this document are under `/api/v1` unless a future revision introduces a different versioning scheme. The shapes and semantics defined here are **MVP v1**.

---

## 4. Content Type

Unless otherwise specified:

### Requests
- `Content-Type: application/json`

### Responses
- `Content-Type: application/json`

---

## 5. Authentication

Authentication is out of scope for MVP.

For MVP:
- API may run without auth in local/dev usage
- auth can be added later without changing the core request/response contracts

---

## 6. Canonical Models Used by the API

The API uses the canonical contracts defined in:
- `docs/data-contract.md`
- `docs/scoring-and-validation-spec.md`

Core objects:
- `Session`
- `Preferences`
- `Plan`
- `ValidationResult`
- `ScoringResult` — matches plan and session scoring result shapes in `docs/data-contract.md` §14 (`score` plus `components`).

---

## 7. Session dataset and ID resolution

The backend maintains a **canonical session dataset** (how it is loaded is an implementation detail). The default on-disk artifact is **`data/sessions.json`** (override with `SESSIONS_FILE`); that file is **valid JSON whose root is an array** of `Session` objects — not an object wrapper such as `{ "sessions": [...] }` (see `docs/data-contract.md` §15.2, `docs/importer-spec.md` §15.1). **HTTP** responses for `GET /api/v1/sessions` use a JSON object with a `sessions` property; that envelope applies only to the API, not to the sessions file format.

Session identity rules, including importer-style `id` values, are defined in `docs/data-contract.md` §4.4.

- **`GET /api/v1/sessions`** returns rows from that dataset after applying query filters.
- **`POST /api/v1/validate`** and **`POST /api/v1/rank/plans`**: every session id in the plan **must** resolve to a session in that dataset—whether listed as `primarySessionId`, in `alternateSessionIds`, or in `sessionIds`. Unknown IDs produce validation errors (for example `SESSION_NOT_FOUND`) in a **200** response body using the validation result shape, not an HTTP transport error. When **`preferences.rules.minHoursBetweenSameDaySessions`** is omitted, validation applies a **default 4-hour** minimum gap (end of earlier session → start of next) for multiple sessions on the same calendar day; set it to **`0`** to disable. See `docs/data-contract.md` §9.2.
- **`POST /api/v1/rank/sessions`**: the request body includes **full canonical `Session` objects**. The server ranks those payloads using `preferences` and does **not** require a dataset lookup by ID for scoring. Session-level validation still applies per `docs/scoring-and-validation-spec.md`; inputs that fail validation are handled as specified in §12.

---

## 8. Endpoints Overview

### Included in MVP
- `GET /api/v1/health`
- `GET /api/v1/sessions`
- `POST /api/v1/validate`
- `POST /api/v1/rank/sessions`
- `POST /api/v1/rank/plans`

### Explicitly Excluded from MVP
Do **not** add orchestration endpoints such as:
- `POST /api/v1/generate-weekend-plan`
- `POST /api/v1/generate-saturday-plan`
- `POST /api/v1/generate-multi-day-plan`

Those are application/GPT responsibilities, not backend responsibilities.

---

## 9. GET /api/v1/health

### Purpose
Basic liveness/health endpoint.

### Request
No request body.

### Response: 200

```json
{
  "status": "ok"
}
```

### Notes
This endpoint is intentionally minimal.

---

## 10. GET /api/v1/sessions

### Purpose
Return sessions filtered by query parameters.

This endpoint is for:
- session browsing
- candidate retrieval before ranking
- debugging/filter inspection

### Query Parameters

#### `date`
- type: repeated string or comma-separated string
- format: `YYYY-MM-DD`
- optional

Examples:
- `?date=2028-07-15`
- `?date=2028-07-15&date=2028-07-16`

#### `dayOfWeek`
- type: repeated string or comma-separated string
- optional

Examples:
- `?dayOfWeek=Saturday`
- `?dayOfWeek=Saturday&dayOfWeek=Sunday`

#### `sports`
- type: repeated string or comma-separated string
- optional
- meaning: same as filter `sports` in `docs/data-contract.md` §10 (explicit sport filter)

Examples:
- `?sports=Tennis`
- `?sports=Athletics&sports=Tennis`

#### `allowedSports`
- type: repeated string or comma-separated string
- optional
- meaning: hard allowlist at request time (same as filter `allowedSports` in the data contract)

#### `excludedSports`
- type: repeated string or comma-separated string
- optional
- meaning: hard denylist at request time (same as filter `excludedSports` in the data contract)

#### `includeDebug`
- type: boolean
- optional
- default: `false`
- MVP behavior: **reserved for future use**; no additional response fields are defined in MVP.

### Query parameter mapping (data contract)

HTTP query names align with `docs/data-contract.md` §10 (Filter Contract):

| Query parameter | Data-contract filter field |
|-----------------|----------------------------|
| `date` (repeated) | `dates` |
| `dayOfWeek` | `daysOfWeek` |
| `sports` | `sports` |
| `allowedSports` | `allowedSports` |
| `excludedSports` | `excludedSports` |

### Query Semantics
For MVP:
- all provided filter dimensions are conjunctive
- within a single dimension, multiple values are ORed

Example:
- `date in {2028-07-15, 2028-07-16}`
- AND `sports in {Athletics, Tennis}`

### Example Request

```text
GET /api/v1/sessions?dayOfWeek=Saturday&sports=Tennis&sports=Athletics
```

### Response: 200

Returns **`200`** with `sessions` (possibly empty). If no rows match the filters, the response is **`200`** with:

```json
{
  "sessions": []
}
```

When matches exist:

```json
{
  "sessions": [
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
  ]
}
```

### Response: 400
For malformed query parameters.

Example:

```json
{
  "error": {
    "code": "INVALID_QUERY_PARAMETER",
    "message": "date must use YYYY-MM-DD format",
    "field": "date"
  }
}
```

---

## 11. POST /api/v1/validate

### Purpose
Validate a plan against:
- the canonical session dataset (see §7)
- effective **preferences** for this request

This endpoint does not rank. It only validates.

Conversational or request-specific overrides (for example “avoid cricket”) are **merged into a single effective `preferences` object by the caller** before calling this endpoint. There is **no** separate `overrides` field in MVP.

### Request Body

```json
{
  "plan": {
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
  },
  "preferences": {
    "allowedSports": ["Tennis", "Cricket", "Athletics", "Diving"],
    "sportPriority": ["Cricket", "Athletics", "Tennis", "Diving"],
    "allowedDays": ["Friday", "Saturday", "Sunday", "Monday"],
    "rules": {
      "noSameSportAcrossDays": true,
      "preferDayPairs": [
        ["Saturday", "Sunday"],
        ["Friday", "Saturday"]
      ]
    }
  }
}
```

### Request Fields

#### `plan`
- required
- must match canonical `Plan` contract

#### `preferences`
- required
- must match canonical `Preferences` contract
- must already reflect any merged overrides for this request
- **`rules.noSameSportAcrossDays`:** if omitted, the server treats it as **`true`** (no same sport on more than one day unless explicitly set `false`).

### Response: 200

```json
{
  "valid": true,
  "errors": []
}
```

### Response: 200 with validation failures

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

### Response: 400
For malformed request body.

Example:

```json
{
  "error": {
    "code": "INVALID_REQUEST_BODY",
    "message": "plan is required",
    "field": "plan"
  }
}
```

---

## 12. POST /api/v1/rank/sessions

### Purpose
Rank candidate sessions supplied in the request body using **`preferences`**.

For MVP there is **no** separate filter object: narrowing is expressed only through **`preferences`** and through which sessions the caller includes in the **`sessions`** array. This endpoint does **not** build plans; it ranks sessions only.

### Request Body

```json
{
  "sessions": [
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
        "Men's Singles Quarterfinal"
      ]
    },
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
        "Women's High Jump Final"
      ]
    }
  ],
  "preferences": {
    "allowedSports": ["Tennis", "Athletics"],
    "sportPriority": ["Athletics", "Tennis"],
    "allowedDays": ["Saturday", "Sunday"],
    "rules": {
      "noSameSportAcrossDays": true,
      "preferDayPairs": [["Saturday", "Sunday"]]
    }
  },
  "includeScoreBreakdown": true
}
```

### Request Fields

#### `sessions`
- required
- array of canonical `Session`
- entries that **fail session-level validation** are **omitted** from **`rankedSessions`** (silent exclusion). The response remains **`200`** unless the request body is structurally invalid (for example malformed JSON or missing required top-level fields).

#### `preferences`
- required

#### `includeScoreBreakdown`
- optional boolean
- default: `false`

### Response: 200

```json
{
  "rankedSessions": [
    {
      "session": {
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
          "Women's High Jump Final"
        ]
      },
      "score": 28,
      "components": {
        "sportPriority": 22,
        "dataQuality": 6
      }
    },
    {
      "session": {
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
          "Men's Singles Quarterfinal"
        ]
      },
      "score": 18,
      "components": {
        "sportPriority": 12,
        "dataQuality": 6
      }
    }
  ]
}
```

The `rankedSessions` array contains **`{ session, score, components }`** entries in descending score order.

### Notes
- ranked output must be sorted highest score first
- tie-breakers must follow `docs/scoring-and-validation-spec.md`

---

## 13. POST /api/v1/rank/plans

### Purpose
Rank candidate plans that are already assembled by the application/GPT layer. Session IDs in each plan are resolved against the canonical dataset (§7).

This endpoint:
- validates plans
- excludes invalid plans from ranked output
- returns ranked valid plans

It does **not** generate plans.

### Request Body

```json
{
  "plans": [
    {
      "planType": "two_day",
      "days": [
        {
          "date": "2028-07-15",
          "dayOfWeek": "Saturday",
          "primarySessionId": "session-ath-09",
          "alternateSessionIds": ["session-ten-12"]
        },
        {
          "date": "2028-07-16",
          "dayOfWeek": "Sunday",
          "primarySessionId": "session-ckt-03",
          "alternateSessionIds": ["session-div-02"]
        }
      ]
    },
    {
      "planType": "two_day",
      "days": [
        {
          "date": "2028-07-14",
          "dayOfWeek": "Friday",
          "primarySessionId": "session-ath-05",
          "alternateSessionIds": []
        },
        {
          "date": "2028-07-15",
          "dayOfWeek": "Saturday",
          "primarySessionId": "session-ten-12",
          "alternateSessionIds": []
        }
      ]
    }
  ],
  "preferences": {
    "allowedSports": ["Athletics", "Tennis", "Cricket", "Diving"],
    "sportPriority": ["Cricket", "Athletics", "Tennis", "Diving"],
    "allowedDays": ["Friday", "Saturday", "Sunday", "Monday"],
    "rules": {
      "noSameSportAcrossDays": true,
      "preferDayPairs": [
        ["Saturday", "Sunday"],
        ["Friday", "Saturday"],
        ["Sunday", "Monday"]
      ]
    }
  },
  "includeScoreBreakdown": true,
  "includeInvalidPlans": false
}
```

### Request Fields

#### `plans`
- required
- array of canonical `Plan`

#### `preferences`
- required

#### `includeScoreBreakdown`
- optional boolean
- default: `false`

#### `includeInvalidPlans`
- optional boolean
- default: `false`

If `includeInvalidPlans = true`, invalid plans may be returned in a separate collection for debugging.

### Response: 200

```json
{
  "plans": [
    {
      "plan": {
        "planType": "two_day",
        "days": [
          {
            "date": "2028-07-15",
            "dayOfWeek": "Saturday",
            "primarySessionId": "session-ath-09",
            "alternateSessionIds": ["session-ten-12"]
          },
          {
            "date": "2028-07-16",
            "dayOfWeek": "Sunday",
            "primarySessionId": "session-ckt-03",
            "alternateSessionIds": ["session-div-02"]
          }
        ]
      },
      "score": 90,
      "components": {
        "dayPair": 30,
        "summedSessionScore": 40,
        "variety": 10,
        "convenience": 10
      }
    },
    {
      "plan": {
        "planType": "two_day",
        "days": [
          {
            "date": "2028-07-14",
            "dayOfWeek": "Friday",
            "primarySessionId": "session-ath-05",
            "alternateSessionIds": []
          },
          {
            "date": "2028-07-15",
            "dayOfWeek": "Saturday",
            "primarySessionId": "session-ten-12",
            "alternateSessionIds": []
          }
        ]
      },
      "score": 83,
      "components": {
        "dayPair": 20,
        "summedSessionScore": 43,
        "variety": 10,
        "convenience": 10
      }
    }
  ]
}
```

### Optional Debug Response With Invalid Plans

```json
{
  "plans": [
    {
      "plan": {
        "planType": "two_day",
        "days": [
          {
            "date": "2028-07-15",
            "dayOfWeek": "Saturday",
            "primarySessionId": "session-ath-09",
            "alternateSessionIds": []
          },
          {
            "date": "2028-07-16",
            "dayOfWeek": "Sunday",
            "primarySessionId": "session-ckt-03",
            "alternateSessionIds": []
          }
        ]
      },
      "score": 90,
      "components": {
        "dayPair": 30,
        "summedSessionScore": 40,
        "variety": 10,
        "convenience": 10
      }
    }
  ],
  "invalidPlans": [
    {
      "plan": {
        "planType": "two_day",
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
            "primarySessionId": "session-ten-44",
            "alternateSessionIds": []
          }
        ]
      },
      "validation": {
        "valid": false,
        "errors": [
          {
            "code": "REPEATED_SPORT_ACROSS_DAYS",
            "message": "Tennis appears on multiple days in a plan where repeated sports are forbidden.",
            "field": "days"
          }
        ]
      }
    }
  ]
}
```

---

## 14. Error Contract

### 14.1 Error Shape

For non-validation API failures, use:

```json
{
  "error": {
    "code": "INVALID_REQUEST_BODY",
    "message": "preferences is required",
    "field": "preferences"
  }
}
```

### 14.2 Error Fields

#### `code`
- machine-readable
- stable
- required

#### `message`
- human-readable
- required

#### `field`
- optional but strongly recommended

### 14.3 Recommended Generic Error Codes
- `INVALID_REQUEST_BODY`
- `INVALID_QUERY_PARAMETER`
- `UNSUPPORTED_MEDIA_TYPE`
- `INTERNAL_ERROR`

### 14.4 Validation Error Codes
Validation-specific codes are defined in:
- `docs/scoring-and-validation-spec.md`
- `docs/data-contract.md`

### 14.5 MVP: `400` responses use a single `error` object
For **`400 Bad Request`**, MVP responses use **one** top-level `error` object as in §14.1. Multiple field-level issues may be summarized in `message` or addressed in a future revision; listing multiple errors in one response is **out of scope** for MVP unless explicitly added later.

---

## 15. Status Codes

### `200 OK`
Used for:
- successful health checks
- session retrieval
- validation responses
- ranking responses

Validation failures still return `200` because they are business results, not transport errors.

### `400 Bad Request`
Used for:
- malformed JSON
- invalid query params
- missing required top-level fields
- structurally invalid request body

### `415 Unsupported Media Type`
Used when request content type is not JSON for JSON endpoints.

### `500 Internal Server Error`
Used for unexpected backend failures.

### `404 Not Found`
MVP does **not** use **`404`** for unknown session IDs embedded in plans or validation payloads. Unknown IDs are reported with stable validation error codes (for example `SESSION_NOT_FOUND`) in a **200** response using the validation result shape (see §7 and `docs/scoring-and-validation-spec.md`). **`404`** remains available for future resource-style routes if added later.

---

## 16. Identifier Rules

### Canonical Rule
All runtime references must use canonical `Session.id`.

### If `sessionCode` Is Accepted Later
If future API ergonomics require accepting `sessionCode`:
- resolution must happen before validation and ranking
- the API should still normalize to canonical `Session.id` internally

The canonical contract remains `id`-based.

---

## 17. Backend / App Layer Boundary

### Backend Responsibilities
- return sessions
- filter sessions
- validate plans
- rank sessions
- rank plans

### Application / GPT Responsibilities
- interpret natural-language requests
- decide plan shape
- build candidate plans
- ask for alternates
- orchestrate multi-step planning flows
- present results conversationally

This boundary is mandatory for MVP.

---

## 18. Examples of Allowed and Disallowed Designs

### Allowed
- `GET /api/v1/sessions?sports=Tennis&dayOfWeek=Saturday`
- `POST /api/v1/validate`
- `POST /api/v1/rank/plans`

### Disallowed
- `POST /api/v1/generate-weekend-plan`
- `POST /api/v1/best-saturday-plan`
- `POST /api/v1/plan-for-three-weekends`

---

## 19. Open Decisions

The following remain open:
- whether to support both repeated and comma-separated query params
- whether `GET /api/v1/sessions` should support pagination in MVP
- whether invalid sessions/plans should be optionally returned in rank endpoints by default or only when debug flags are set
- whether `sessionCode` lookup convenience should be exposed at the API edge

---

## 20. Summary

This API spec defines a thin correctness-layer backend.

It supports:
- health checks
- session retrieval
- plan validation
- session ranking
- plan ranking

It explicitly does **not** support planning orchestration workflows, which remain the responsibility of GPT or the application layer.

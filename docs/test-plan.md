# Test Plan

This document describes how to test the Olympics Sessions Planner MVP: layers, scope, traceability to specs, and alignment with [implementation-plan-deployable-chunks.md](diagrams/implementation-plan-deployable-chunks.md).

**Normative references:** [api-spec.md](api-spec.md), [data-contract.md](data-contract.md), [scoring-and-validation-spec.md](scoring-and-validation-spec.md), [importer-spec.md](importer-spec.md).

---

## 1. Objectives

- Prove **determinism**: same inputs → same validation results, scores, ordering, and tie-breakers.
- Prove **contract compliance**: HTTP shapes, status codes, and error bodies match the API spec.
- Prove **separation of concerns**: no orchestration endpoints; validation gates scoring; invalid candidates excluded per spec.
- Keep **fast feedback**: unit tests run in milliseconds; integration tests use real HTTP + fixtures; e2e stays thin until the API surface is complete.

---

## 2. Chunk-to-test mapping

Deployable chunks from the implementation plan map to test expectations as follows. Use this to know **what to add in each PR** before moving on.

| Chunk | Focus | Primary tests |
|-------|--------|----------------|
| **0** — Contract alignment | `Preferences` nested `rules`; `Session` / JSON tags; `Plan` refactor or adapter; `Session.id`; optional-field serialization | **Unit:** domain serialization round-trips; preferences load; any adapter boundary tests. **Contract conformance** (§4): sessions file root array, preferences shape. |
| **1** — Importer | Canonical `data/sessions.json` emission | **Strongly recommended:** frozen extracted-text **golden** input → expected canonical JSON output (§10). **Unit:** parser rows, rejected headers, normalization §4.4. |
| **2** — Read path | `GET /api/v1/health`, `GET /api/v1/sessions` | **Integration:** health JSON; sessions filtering matrix; `400` malformed query; `200` empty list. **Contract conformance:** GET response envelope `{ "sessions": [...] }`. |
| **3** — Validation | `POST /api/v1/validate` | **Unit:** validation rules + structured error `code`s. **Integration:** full request/response vs api-spec §11. **Regression vectors** (§8): unknown IDs → `200` + errors, not `404`. |
| **4** — Session ranking | `POST /api/v1/rank/sessions` | **Unit:** scoring components, bounds, tie-breakers. **Integration:** `rankedSessions`, omitted invalid rows, `includeScoreBreakdown`. **Regression vectors** (§8). |
| **5** — Plan ranking | `POST /api/v1/rank/plans` | **Unit:** plan score components. **Integration:** ranked output, invalid excluded by default, `includeInvalidPlans`. **Regression vectors** (§8). |
| **6** — Hardening | Regression suite + CI | Expand coverage; stable fixtures; CI gates; **forbidden routes** (§5). |

**Rule of thumb:** add narrow tests **with** each chunk; Chunk 6 expands fixtures and CI, not “tests only at the end.”

---

## 3. Test layers

| Layer | Scope | Typical location | Speed |
|-------|--------|------------------|-------|
| **Unit** | Pure functions: parse helpers, filters, validation rules, scoring math, tie-breakers | `internal/...` next to code | Fast |
| **Integration** | HTTP server + real `ServeMux` routes, JSON request/response, file-backed repos | `tests/integration` | Medium |
| **Importer** | Text → normalized sessions → `WriteSessionsJSON`; **golden** outputs | `internal/ingest/...` or `tests/importer` | Medium |
| **E2E** | Optional: full binary, env, and fixture data as a smoke | `tests/e2e` | Slow |

---

## 4. Contract conformance tests

This project is **doc-driven**; silent drift in shapes is high risk. Name and maintain tests that assert the following at the appropriate boundary (load, HTTP encode/decode, or golden file).

| Invariant | What to assert |
|-----------|----------------|
| **Sessions file on disk** | Root value is a **JSON array** of `Session` objects — **not** `{ "sessions": [...] }` ([data-contract §15.2](data-contract.md), [importer-spec §15.1](importer-spec.md)). |
| **HTTP response envelopes** | `GET /api/v1/sessions` returns an object with a **`sessions`** property; rank endpoints use `rankedSessions` / documented plan wrappers per [api-spec](api-spec.md). |
| **Preferences** | Nested **`rules`** object; aligns with `docs/data-contract.md` §9 (not a flat-only legacy shape). |
| **Plan at API boundary** | Request/response `plan` bodies match canonical **`planType`**, **`days`**, **`primarySessionId`**, **`alternateSessionIds`** ([data-contract §11](data-contract.md)). |
| **Validation / scoring results** | `ValidationResult` and `ScoringResult` shapes (including `components` where specified) match [data-contract §13–14](data-contract.md) and [api-spec](api-spec.md). |

Use these as **regression guards** whenever domain types or handlers change.

---

## 5. Forbidden backend behavior

Architecture forbids orchestration endpoints. Add a **lightweight** guard so new routes cannot slip in unnoticed:

- **Route inventory test:** register the production router (or a shared route list) and assert **no** handler is bound to paths that match forbidden patterns, for example:
  - `/generate-weekend-plan`, `/generate-saturday-plan`, `/generate-multi-day-plan`
  - informal variants like `/best-saturday-plan` or `/generate-weekend-plan` under `/api/v1`
- **Policy:** any new `POST` under `/api/v1` is reviewed against [api-spec §8](api-spec.md) (included vs excluded endpoints) and [architecture.md](architecture.md).

Exact path strings should match the **excluded** list in the API spec (see “Explicitly Excluded from MVP”).

---

## 6. Traceability matrix (MVP behavior)

| Capability | Spec section | What to test |
|------------|--------------|--------------|
| Sessions file load | data-contract §15.2, importer §15.1 | Root JSON array; invalid file → clear error at startup or first read |
| `GET /api/v1/health` | api-spec §9 | `200`, `{"status":"ok"}` |
| `GET /api/v1/sessions` | api-spec §10 | Query params: `date`, `dayOfWeek`, `sports`, `allowedSports`, `excludedSports`; conjunctive dimensions; OR within dimension; empty `200`; malformed query → `400` with `INVALID_QUERY_PARAMETER` shape |
| `POST /api/v1/validate` | api-spec §11, scoring §8–9 | Valid plan → `valid: true`; structured `errors` with `code`/`message`/`field`; unknown session id → business error in `200`, not `404` |
| `POST /api/v1/rank/sessions` | api-spec §12, scoring §11–12 | `rankedSessions` order; score ranges; invalid sessions omitted; `includeScoreBreakdown` |
| `POST /api/v1/rank/plans` | api-spec §13, scoring §13–15 | Plan components; invalid plans excluded by default; `includeInvalidPlans` |
| Importer CLI | importer-spec §8, §15 | `-out` writes canonical array; deterministic on fixed input; diagnostics (counts / failures) |

---

## 7. Unit test catalog (by component)

### 7.1 Session repository

- Load valid `data/sessions.json` (array root).
- Reject or error clearly on: empty file, non-array root, malformed JSON, duplicate `id` (if you enforce uniqueness at load time).

### 7.2 GET sessions filtering

- Matrix: single vs repeated query keys; comma-separated vs repeated params (per api-spec acceptance rules).
- Edge: all filters empty → all sessions (or documented subset); impossible combination → `200` + empty list.

### 7.3 Validation

- Session-level: missing session, wrong sport/day, required fields, date vs `dayOfWeek` mismatch (per scoring spec).
- Plan-level: duplicates, `planType` vs `days` length, alternates, empty day, same-sport-across-days when rule on.
- Output shape: each error has stable **`code`** enum matching api-spec / scoring spec (refactor tests when moving from string errors to structured errors).

### 7.4 Session scoring

- `sportPriority`, `dataQuality` (or spec-named components), numeric bounds, tie-breaker order.

### 7.5 Plan scoring

- Components: `dayPair`, `summedSessionScore`, `variety`, `convenience` with documented caps.
- Multi-day vs one-day behavior; `preferDayPairs` in preferences.

### 7.6 Importer

- Parser: representative rows → expected `Session` fields; header/footer rows rejected.
- Normalization: `id` / `sessionCode` rules per data-contract §4.4.
- Pipeline: idempotent write path.

---

## 8. Ranking and validation regression vectors

These are subtle **contract behaviors** that should not drift. Prefer explicit tests for each.

| Vector | Expected behavior |
|--------|---------------------|
| **Tie on score** | Equal total session (or plan) score → **deterministic** ordering per documented **tie-breaker** order ([scoring-and-validation-spec](scoring-and-validation-spec.md)). |
| **Invalid plan excluded from rank/plans** | `POST /api/v1/rank/plans` omits invalid plans from the main ranked list **by default** when `includeInvalidPlans` is false or omitted. |
| **`includeInvalidPlans=true`** | Invalid plans appear in a **separate** collection for debugging; shape per [api-spec §13](api-spec.md). |
| **Unknown session IDs** | Embedded in plans or validation payloads → **validation errors** in **`200`** response (e.g. `SESSION_NOT_FOUND`), **not** HTTP `404` ([api-spec §7](api-spec.md)). |
| **Weekend semantics** | “Weekend” for **`two_day`** means **Saturday + Sunday** (per scoring spec); **not** conflated with **`multi_day`** plans. Tests should use fixtures that distinguish `planType: "two_day"` vs `planType: "multi_day"` when asserting weekend rules. |

---

## 9. Integration test catalog

Use `httptest` against the real router; load fixtures from `testdata/` or `internal/testutil`.

| Test | Assertion |
|------|-----------|
| Health | `GET /api/v1/health` → 200 + body |
| Sessions happy path | Filter returns expected subset |
| Sessions bad query | `400` + `error.code` + `field` |
| Validate | Request bodies from api-spec examples; assert JSON shape |
| Rank sessions | Assert `rankedSessions` length and order for fixed fixture |
| Rank plans | Assert scores, default exclusion of invalid, and optional invalid bucket |

---

## 10. Importer golden tests (strongly recommended)

Importer drift is a **major** risk; treat **golden** regression as a **core** asset, not an optional technique.

- **Inputs:** a **frozen** extracted-text fixture (or minimal PDF-derived text) checked into the repo — no live LA28 downloads in CI.
- **Outputs:** expected **`data/sessions.json`-style** file: JSON array at root, canonical `Session` fields per [data-contract](data-contract.md).
- **Assertions:** byte-for-byte or normalized JSON equality; on intentional parser changes, update the golden with review.
- **Why:** catches regressions in extract, parse, normalize, and ID assignment in one place.

---

## 11. Fixture policy

Keep fixtures **maintainable** as the suite grows:

- **Small** — minimal rows needed to exercise a rule (often 2–4 sessions, 1–2 plans).
- **Deterministic** — fixed IDs, dates, and times; no clocks or randomness.
- **Human-readable** — prefer JSON or text over opaque blobs unless testing binary edge cases.
- **Close to docs** — mirror [data-contract](data-contract.md) §16–17 and [api-spec](api-spec.md) examples so spec and tests stay aligned.
- **No live LA28 sources in default CI** — use synthetic schedules or truncated extracts; avoid network and licensing issues.

---

## 12. Non-functional / quality gates

- **`go test ./...`** (and `go test -race ./...` in CI for packages with concurrency).
- **Lint** (`make lint` or equivalent) in CI.
- **No** reliance on network or live LA28 PDFs in default CI: use checked-in text fixtures or minimal PDFs if required.

---

## 13. Current repo vs this plan

| Area | Today | Next step |
|------|--------|-----------|
| **Sessions artifact** | `WriteSessionsJSON` emits a **root JSON array**; `JSONSessionRepository` rejects a non-array root (`ErrSessionsFileNotJSONArray`). | Add more fixtures if you want broader parse-failure coverage. |
| **`domain.Plan`** | Canonical `planType` + `days` + `primarySessionId` + `alternateSessionIds` ([data-contract §11](data-contract.md)). | Extend validation tests as scoring/validation spec coverage grows. |
| Planner | `validation_test`, `ranking_test` against legacy `domain.Plan` | Update when canonical `Plan` lands; add structured error assertions |
| HTTP | `integration` health on `/health` only | Add `/api/v1` routes + sessions + POST tests per §6 |
| E2E | Placeholder skip | Enable after core API routes exist |

---

## 14. Definition of done (testing)

For MVP correctness layer “done” from a testing perspective:

- Core behaviors in §6 have **automated** coverage at unit or integration level.
- **Contract conformance** checks (§4) exist for file load, HTTP envelopes, preferences, plan boundary, and result shapes.
- **Importer golden** (or equivalent) covers parse → canonical output.
- **Forbidden** route / orchestration guard (§5) in CI.
- At least one **integration** test per MVP endpoint (health, sessions, validate, rank/sessions, rank/plans).
- **Regression vectors** in §8 covered where applicable.
- CI runs tests + lint on every change.

---

## 15. Out of scope (MVP)

- Load testing and performance SLOs (unless added later).
- Production PDFs that cannot be redistributed (use synthetic or truncated inputs in CI).

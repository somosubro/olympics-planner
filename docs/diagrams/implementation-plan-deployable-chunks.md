# Implementation Plan (Deployable Chunks)

## 1. Purpose

This document defines a deployable, value-forward implementation plan for the Olympics Sessions Planner MVP.

It is based on the current project documents:
- `docs/prd.md`
- `docs/architecture.md`
- `docs/data-contract.md`
- `docs/scoring-and-validation-spec.md`
- `docs/api-spec.md`
- `docs/importer-spec.md`

The goal is to deliver usable value quickly in small, deployable chunks while preserving the agreed architecture:

- GPT/app layer owns orchestration and natural-language planning
- backend owns correctness only
- importer is offline and separate from runtime

---

## 2. Delivery Strategy

We will build in **deployable chunks**, not big-bang phases.

Each chunk should:
- produce visible value
- be runnable or deployable on its own
- reduce risk for the next chunk
- keep code aligned with the specs

The chunk order is optimized for:
1. contract alignment
2. getting real data into the system
3. shipping read-path value early
4. adding correctness endpoints incrementally
5. adding ranking after validation is stable

---

## 3. Guiding Constraints

These remain non-negotiable throughout implementation:

1. **No backend orchestration endpoints**
   - no `generate-weekend-plan`
   - no `best-saturday-plan`
   - no plan-generation workflows in backend

2. **Canonical session dataset**
   - runtime reads canonical sessions from `data/sessions.json`
   - root of the file is a JSON array of `Session`

3. **Preferences separate from sessions**
   - preferences live in a separate file
   - importer does not embed preferences into sessions output

4. **Validation before scoring**
   - invalid sessions/plans do not become final ranked results

5. **Specs drive implementation**
   - when docs and code differ, refactor code toward docs unless the docs are deliberately updated

---

## 4. Chunk Overview

| Chunk | Goal | Deployable Value |
|------|------|------------------|
| 0 | Contract alignment and repo stabilization | Removes drift and makes all later work safer |
| 1 | Importer emits canonical sessions file | Real LA28 data available in runtime-ready form |
| 2 | Health + sessions read API | Immediate visible value: browse/filter sessions |
| 3 | Validation engine + validate endpoint | Deterministic correctness available |
| 4 | Session ranking endpoint | Rank candidate sessions |
| 5 | Plan ranking endpoint | Rank candidate plans assembled by GPT/app |
| 6 | Regression safety and release hardening | Stable MVP with confidence |

---

## 5. Chunk 0 — Contract Alignment and Repo Stabilization

### Goal
Eliminate the biggest spec/code drifts before adding more behavior.

### Why first
The docs already define:
- canonical `Session`
- canonical `Preferences`
- target `Plan`
- canonical sessions file shape
- API shapes

If this is not aligned first, every later chunk will build on drift.

### Tasks
- align `Preferences` shape to nested `rules`
- align `data/preferences.json` to the documented contract
- confirm canonical session file location and shape:
  - `data/sessions.json`
  - root JSON array
- align repository/config loading to those paths
- decide and implement one of:
  - refactor `domain.Plan` now to target shape
  - or add an explicit temporary adapter layer
- make `Session.id` handling consistent with the data contract
- confirm optional fields and JSON serialization behavior for:
  - `title`
  - `endTime`
  - `includedEvents`

### Deployable output
- codebase compiles with aligned core models/config
- no unresolved drift in core runtime contracts

### Definition of done
- `Preferences` in code, docs, and data file match
- sessions file path/shape is consistent everywhere
- `Plan` mismatch is either removed or explicitly bridged in code
- no ambiguity about canonical `Session.id`

---

## 6. Chunk 1 — Importer Produces Canonical Sessions

### Goal
Produce a reliable canonical sessions artifact from LA28 source files.

### Why this chunk matters
Everything else depends on the session dataset.
Without a stable importer, the backend is either blocked or faking data.

### Tasks
- finish `cmd/import_sessions`
- support both:
  - PDF import
  - text import
- keep pipeline structure:
  - extract
  - parse
  - normalize
  - validate
  - emit
- ensure emitted file is:
  - `data/sessions.json`
  - JSON array at root
  - canonical `Session` records only
- keep importer-only enrichment out of runtime dependency path
- add import diagnostics:
  - emitted count
  - rejected row count
  - parse failures
- ensure preferences are not emitted in sessions file

### Deployable output
- runnable importer command
- real runtime-ready LA28 sessions artifact

### Definition of done
- importer runs successfully against a known source PDF or extracted text
- `data/sessions.json` is emitted as a JSON array of canonical sessions
- output shape matches `docs/data-contract.md`
- import is deterministic for the same input
- rejected rows are diagnosable

### User-visible value
Even before ranking or validation, the project now has real structured LA28 session data.

---

## 7. Chunk 2 — Read Path: Health + Sessions API

### Goal
Ship the first backend slice that is immediately useful: session browsing and filtering.

### Why this chunk should come early
It gives fast visible value:
- confirm data loads correctly
- confirm filtering semantics
- allow humans and GPT/app layer to inspect real sessions

### Scope
Implement:
- `GET /api/v1/health`
- `GET /api/v1/sessions`

### Tasks
- expose all MVP HTTP routes under the **`/api/v1`** prefix (see `docs/api-spec.md`; paths such as `GET /health` in code must become `GET /api/v1/health`, etc.)
- implement session repository loading `data/sessions.json`
- implement config/env loading for:
  - `PORT`
  - `SESSIONS_FILE`
  - `PREFERENCES_FILE`
- implement filter mapping from HTTP query params to filter semantics
- support:
  - `date`
  - `dayOfWeek`
  - `sports`
  - `allowedSports`
  - `excludedSports`
- enforce documented semantics:
  - dimensions are conjunctive
  - values within one dimension are ORed
- return:
  - `200` with `sessions: []` on no match
  - `400` on malformed query params

### Deployable output
- backend service can start
- health endpoint works
- sessions endpoint returns real filtered data

### Definition of done
- `GET /api/v1/health` returns `{"status":"ok"}`
- `GET /api/v1/sessions` returns filtered sessions correctly
- no-match returns `200` with empty list
- malformed query params return correct `400`

### User-visible value
You can now:
- browse sessions by date/day/sport
- inspect the imported dataset
- power an early UI/GPT read-only flow

---

## 8. Chunk 3 — Validation Engine + Validate Endpoint

### Goal
Ship the first correctness core: deterministic plan validation.

### Why before ranking
Validation is the hard-rule core.
Ranking without stable validation creates noisy and misleading behavior.

### Scope
Implement:
- validation library/service
- `POST /api/v1/validate`

### Tasks
- implement session-level validation:
  - session exists
  - allowed sport
  - allowed day
  - required fields
  - date/day consistency
- implement plan-level validation:
  - all sessions valid
  - no duplicate session IDs
  - plan shape checks
  - weekend / one-day / two-day / multi-day rules
  - no same sport across days when configured
  - alternates validity
  - no empty day entries
- return documented validation result shape (`valid` plus structured `errors` entries with `code`, `message`, optional `field` per `docs/api-spec.md` §11 — not plain string lists)
- return documented validation error codes
- ensure unknown session IDs are validation errors in `200`, not `404`

### Deployable output
- backend can validate candidate plans assembled by GPT/app
- correctness boundary is now real

### Definition of done
- `POST /api/v1/validate` works against canonical dataset + preferences
- all documented validation codes are supported
- transport errors are separated from business validation failures

### User-visible value
The system can now answer:
- “is this plan valid?”
- “why is this plan invalid?”

That is immediately useful even before ranking.

---

## 9. Chunk 4 — Session Ranking Endpoint

### Goal
Add deterministic ranking for candidate sessions.

### Why now
Once sessions can be loaded and plans can be validated, ranking individual sessions is the next low-risk value slice.

### Scope
Implement:
- session scoring
- `POST /api/v1/rank/sessions`

### Tasks
- implement session scoring model:
  - `sportPriority`
  - `dataQuality`
- implement deterministic tie-breakers
- implement ranked response shape:
  - `rankedSessions`
  - `score`
  - `components`
- omit invalid input sessions from ranked output per API spec
- support optional `includeScoreBreakdown`

### Deployable output
- backend can rank candidate sessions supplied by caller

### Definition of done
- ranking output matches score ranges in the spec
- tie-breakers are deterministic
- invalid sessions are handled per API contract
- output order is correct and stable

### User-visible value
The system can now answer:
- “which candidate sessions are better?”
- “show me the strongest sessions for this day/sport set”

This is enough to support a basic GPT/app flow that asks backend for sessions then ranks them.

---

## 10. Chunk 5 — Plan Ranking Endpoint

### Goal
Add deterministic ranking for already-assembled plans.

### Why this is the first “planner-feeling” backend slice
This is the point where GPT/app can:
- assemble candidate plans
- ask backend to validate and rank them
- present the best plan

Without putting orchestration in the backend.

### Scope
Implement:
- plan scoring
- `POST /api/v1/rank/plans`

### Tasks
- validate plans before scoring
- exclude invalid plans from ranked output by default
- support optional **`includeInvalidPlans`** (per `docs/api-spec.md`) to return invalid plans in a separate collection for debugging
- implement plan score components:
  - `dayPair`
  - `summedSessionScore`
  - `variety`
  - `convenience`
- implement multi-day scoring behavior per spec
- implement plan tie-breakers
- return score breakdowns when requested

### Deployable output
- backend can rank candidate plans deterministically

### Definition of done
- valid plans are ranked correctly
- invalid plans are excluded by default
- debug invalid-plan output works when enabled
- all scoring behavior matches the scoring spec

### User-visible value
Now GPT/app can deliver:
- “best weekend plan”
- “best Saturday-only candidate among these”
- “top 3 candidate plans”

without violating the architecture.

---

## 11. Chunk 6 — Regression Safety and Release Hardening

### Goal
Make the MVP safe to iterate on.

### Why this is its own chunk
Once importer + read path + validation + ranking exist, drift becomes the biggest risk.

You should still add **narrow, chunk-scoped tests** while implementing chunks 1–5 (importer smoke tests, handler tests, validation cases). This chunk is where that coverage is **expanded**, **fixtures are stabilized**, and **CI** is expected to gate releases.

### Tasks
- add importer tests:
  - extraction
  - parsing
  - normalization
  - import validation
- add validation unit tests
- add session ranking unit tests
- add plan ranking unit tests
- add API integration tests
- add stable fixtures for:
  - sessions
  - preferences
  - valid plans
  - invalid plans
- add regression scenarios:
  - best weekend candidate set
  - Saturday only
  - excluded-sport preference override
  - repeated sport invalidation
  - malformed query params
- harden logs and diagnostics
- ensure CI runs:
  - tests
  - lint
  - formatting checks

### Deployable output
- MVP correctness layer is stable and safer to evolve

### Definition of done
- core importer/validation/ranking/API behavior has regression coverage
- CI protects against spec drift
- major scenarios from the docs are test-backed

### User-visible value
This chunk is less flashy, but it is what prevents the planner from becoming unreliable.

---

## 12. Suggested Release Sequence

If you want visible checkpoints quickly, release in this order:

### Release A — Read-Only Explorer
Includes:
- Chunk 0
- Chunk 1
- Chunk 2

Value:
- import real LA28 data
- browse/filter sessions via API
- early GPT/app integration can already answer “what exists?”

### Release B — Validation MVP
Includes:
- Release A
- Chunk 3

Value:
- users can test whether a candidate plan is valid
- GPT/app can assemble plans and ask backend for hard-rule enforcement

### Release C — Ranking MVP
Includes:
- Release B
- Chunk 4
- Chunk 5

Value:
- users can get best-of candidate sessions and candidate plans
- this is the first full correctness-layer MVP

### Release D — Hardened MVP
Includes:
- Release C
- Chunk 6

Value:
- safer iteration
- more confidence in data and behavior
- ready for broader use

---

## 13. Recommended Immediate Next Tasks

Do these next, in order:

1. align `Preferences` shape everywhere (`rules.*`)
2. lock the sessions file format to:
   - `data/sessions.json`
   - root JSON array
3. decide whether to refactor `domain.Plan` now or bridge it temporarily
4. finish importer emission to canonical runtime `Session`
5. implement `GET /api/v1/health`
6. implement `GET /api/v1/sessions`

That gets you to the first deployable, visibly useful chunk fastest.

---

## 14. Risks and Mitigations

### Risk 1 — Importer/source drift
LA28 source layout changes break parsing.

Mitigation:
- text import mode
- regression fixtures
- rejected-row diagnostics

### Risk 2 — Code/doc drift
Implementation diverges from contracts.

Mitigation:
- docs remain normative
- tests mirror spec behavior
- CI gates formatting/lint/tests

### Risk 3 — Backend orchestration creep
Helpers evolve into forbidden planning endpoints.

Mitigation:
- keep endpoints thin
- review API additions against PRD/architecture
- only rank or validate caller-supplied candidates

### Risk 4 — Data-quality ambiguity
Missing title/endTime/includedEvents cause inconsistent behavior.

Mitigation:
- keep required vs optional rules explicit
- encode those rules in validation/scoring tests

---

## 15. Definition of Done for the MVP Correctness Layer

The MVP correctness layer is done when all of the following are true:

- canonical contracts are implemented or consciously bridged
- importer emits canonical runtime session data
- preferences shape is aligned across docs, code, and data files
- `GET /api/v1/health` works
- `GET /api/v1/sessions` works
- `POST /api/v1/validate` works
- `POST /api/v1/rank/sessions` works
- `POST /api/v1/rank/plans` works
- validation and ranking behavior match documented specs
- regression tests cover core scenarios
- backend contains no orchestration endpoints

---

## 16. Summary

The fastest path to visible value is:

1. align contracts
2. finish importer
3. ship read-only session retrieval
4. ship validation
5. ship ranking
6. harden with tests

This preserves the architecture defined in the docs:
- importer as offline ingestion
- backend as correctness layer
- GPT/app as orchestration layer

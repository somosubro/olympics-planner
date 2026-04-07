# PRD: GPT-Fronted Olympics Sessions Planner

## 1. Overview

The Olympics Sessions Planner is a conversational planning tool that helps users build one-day, weekend, and multi-weekend attendance plans from official LA28 Olympic session schedule data.

The product combines:
- a **GPT-based conversational layer** for interpreting user intent and presenting plans
- a **structured data layer** built from official session schedule data
- a future **thin backend correctness layer** for filtering, ranking, and validation

The planner is intended to make a large, dense Olympic schedule usable for normal people through natural language.

---

## 2. Problem Statement

Olympic session schedules are difficult for most users to navigate because they are:
- large
- dense
- PDF-based
- not optimized for conversational exploration
- difficult to compare across days and weekends

A user trying to plan attendance currently has to:
- inspect a raw schedule manually
- understand session structure
- compare many possible combinations
- remember constraints such as allowed sports, preferred days, and variety rules

This product should make Olympic planning easy by allowing users to express preferences in natural language and receive valid, clear, explainable plans.

---

## 3. Vision

Allow a user to say things like:

- “Give me the best weekend plan”
- “Only Saturday options”
- “Prioritize tennis and athletics”
- “Avoid cricket”
- “Show one good plan for each available weekend”

…and receive clean, accurate plans built only from real session data.

The experience should feel flexible and conversational, while remaining deterministic and rule-safe underneath.

---

## 4. Goals

### Primary Goals
- Generate strong plans from real LA28 session data
- Support natural language planning requests
- Enforce hard constraints reliably
- Present session and plan details clearly for non-technical users
- Maintain a clean separation between conversational logic and correctness logic

### Secondary Goals
- Make plan generation explainable
- Support future persistence and saved plans
- Support future collaboration and group planning
- Create a polished project that can grow beyond a prototype

---

## 5. Non-Goals

For the MVP, this product will **not**:
- sell tickets
- integrate with LA28 accounts
- guarantee seat availability or pricing
- optimize hotels, transportation, or traffic
- provide live official schedule sync via external APIs
- support multi-user collaboration
- support voting or shared plan editing
- persist plans across users or chats
- encode rigid planning flows in the backend

---

## 6. Target Users

### Primary User
A casual attendee or small-group planner who:
- does not want to read the raw Olympic schedule PDF
- wants to express preferences conversationally
- values clear recommendations over manual filtering

### Secondary User
A more detail-oriented planner who:
- wants to inspect available sessions by date or sport
- wants alternates and comparisons
- may later want persistence or saved plans

---

## 7. User Needs

Users need to be able to:
- see what sessions are available on a date or set of dates
- ask for plans in natural language
- prioritize some sports over others
- avoid certain sports or days
- receive plans that respect hard constraints
- understand why a plan was chosen
- inspect full session details when needed

---

## 8. Product Principles

### 8.1 Real Data Only
The planner must use only sessions that exist in the source dataset.

### 8.2 Hard Rules Are Non-Negotiable
Invalid sports, invalid days, invalid sessions, or invalid multi-day combinations must never appear in final plans.

### 8.3 GPT Is the Flexibility Layer
GPT is responsible for:
- natural language understanding
- planning shape
- temporary conversational overrides
- output presentation

### 8.4 Backend Is the Correctness Layer
The future backend is responsible for:
- filtering
- ranking
- validation

### 8.5 No Backend Planning Flows
The backend must **not** implement endpoints like:
- `generateWeekendPlan`
- `generateThreeDayPlan`

Planning flow orchestration belongs in GPT or the application layer, not the backend.

### 8.6 Visibility Beats Mystery
Users should be able to inspect the sessions in a plan and understand what they are attending.

---

## 9. MVP Scope

### In Scope

#### Data Ingestion
- Import the official LA28 schedule PDF
- Parse schedule data into structured JSON
- Produce normalized `data/sessions.json` (top-level JSON array of sessions per `docs/data-contract.md` / `docs/importer-spec.md`)

#### Session Browsing
- Show available sessions by date, day, and sport
- Return complete session details

#### Conversational Planning
- Support one-day plans
- Support weekend/two-day plans
- Support multi-weekend planning
- Support temporary preference overrides in conversation

#### Rule Enforcement
- allowed sports
- allowed days
- no invented sessions
- no same-sport-across-days for multi-day plans when configured

#### Ranking
- prefer valid plans first
- prefer strong day pairs
- prefer higher-priority sports
- prefer better-quality sessions
- promote variety where relevant

#### Output Formatting
Each session should include:
- readable title
- session code
- date/day
- start time
- end time
- venue
- full included events list

---

## 10. Out of Scope for MVP

- user accounts
- saved plans
- shared links
- collaboration
- notifications
- live data sync
- ticket pricing
- travel optimization
- commute-based recommendations
- mobile app support

---

## 11. Functional Requirements

### FR1. Import Official Schedule Data
The system must support importing the official LA28 schedule PDF and generating structured session JSON.

### FR2. Store Normalized Session Objects
The system must store normalized session objects with fields such as:
- sport
- session code
- date
- day of week
- start time
- end time
- venue
- included events

### FR3. Support Session Retrieval by Filters
The system must support session retrieval by filters such as:
- date
- day of week
- sport
- allowed/disallowed sport set

### FR4. Support Conversational Planning Requests
The system must support requests such as:
- best weekend plan
- Saturday-only plan
- one plan per available weekend
- prioritize or avoid specific sports
- show alternates

### FR5. Enforce Hard Constraints
The system must reject invalid plans including:
- sessions not in dataset
- sessions on disallowed days
- sessions for disallowed sports
- repeated sport across days where the rule forbids it

### FR6. Rank Candidate Sessions and Plans
The system must rank valid candidate options using configurable preferences.

### FR7. Explain Output Clearly
The system must return readable plans with enough detail for a user to understand the recommendation.

### FR8. Allow Temporary Conversational Overrides
The user must be able to express temporary request-specific preferences such as:
- “avoid cricket”
- “prioritize tennis”
- “only Saturdays”

These overrides apply only to the current request unless persistence is added later.

---

## 12. Non-Functional Requirements

### Accuracy
No invented sessions. No fabricated details.

### Determinism
Given the same inputs and rules, backend filtering, ranking, and validation should behave consistently.

### Maintainability
Importer logic, planner logic, and runtime API logic must stay separate.

### Explainability
Scoring must be understandable and decomposable.

### Extensibility
The system should support future persistence, sharing, and collaboration without requiring a rewrite of core planner logic.

### Usability
Responses must be readable by non-technical users.

---

## 13. Data Model

### Session
Core fields:
- `id`
- `sport`
- `sessionCode`
- `title`
- `date`
- `dayOfWeek`
- `startTime`
- `endTime`
- `venue`
- `includedEvents`

Optional derived fields:
- `stage`
- `zone`
- `keywords`
- `interesting`
- `finalsHeavy`
- `marquee`
- `durationMins`

### Preferences
Core fields:
- allowed sports
- sport priority order
- allowed days
- no-same-sport-across-days rule
- day pair preferences
- sport-specific ranking rules

---

## 14. Architecture

### 14.1 Conversational Layer
GPT or chat application layer:
- interprets user intent
- determines planning shape
- requests filtered/ranked data
- assembles plans
- explains and formats output

### 14.2 Backend Layer
Thin correctness layer with functions such as:
- `get_sessions(filters)`
- `rank_sessions(sessions, preferences)`
- `validate_plan(plan)`

This layer must not own planning flow orchestration.

### 14.3 Ingestion Layer
Separate import pipeline:
- PDF -> extracted text
- extracted text -> parsed sessions
- parsed sessions -> normalized `data/sessions.json`

---

## 15. Success Criteria

The MVP is successful if a user can:
- ask for a plan in natural language
- receive a valid plan based only on real data
- inspect session details clearly
- request variations and alternates
- trust that the planner is not inventing or violating rules

---

## 16. User Stories

### User Story 1
As a casual attendee, I want to ask for the best weekend plan so I do not have to inspect the full schedule manually.

### User Story 2
As a user with preferences, I want to prioritize some sports and avoid others so the plan reflects my interests.

### User Story 3
As a detail-oriented user, I want to see all available sessions for a day so I can inspect the raw options.

### User Story 4
As a planner comparing weekends, I want one recommended plan per available weekend so I can compare options quickly.

### User Story 5
As a user, I want the planner to avoid invalid combinations automatically so I do not have to track all constraints myself.

---

## 17. Risks

### Data Extraction Risk
LA28 PDFs may change layout, which can break parsing heuristics.

### Scoring Ambiguity Risk
Without a clear scoring specification, plan quality may drift or feel inconsistent.

### GPT Drift Risk
If GPT owns too much correctness logic, output may become inconsistent.

### Model Drift Risk
Importer-enriched session fields may drift away from the runtime session model if not explicitly defined.

---

## 18. Open Questions

- Should the sessions file (`data/sessions.json`) contain only normalized base fields, or also derived ranking hints?
- Should preferences live separately from session data, or be bundled in generated output?
- How much explanation should be shown by default versus on request?
- Do we rank sessions only, or entire candidate plans?
- Should alternates be ranked independently or selected for diversity?

---

## 19. Recommended Next Documents

Recommended follow-up docs, in order:
1. Scoring and validation spec
2. Data contract
3. API spec
4. Importer spec
5. Milestones / implementation plan

---

## 20. ### Gap Analysis / Areas to Tighten

1. **Optional derived fields**
   The PRD lists optional derived fields such as `stage`, `zone`, `keywords`, `interesting`, `finalsHeavy`, and `marquee`, but these are not yet part of a formal runtime data contract. This is acceptable for MVP only if ranking does not depend on them yet. Otherwise, they should either be:
   - added to the domain and importer contract explicitly, or
   - deferred clearly as non-MVP / importer-only fields.

2. **FR3 filtering contract**
   The PRD expects retrieval by date, day, sport, and allow/disallow filters. This should be reflected explicitly in the service and HTTP/API contracts, rather than relying on loading the full dataset and filtering informally in GPT or application code.

3. **No backend planning flows**
   The PRD explicitly says the backend must remain a correctness layer (`get_sessions`, `rank_sessions`, `validate_plan`) and must not absorb orchestration logic such as “generate weekend plan.” This should remain an active architecture check as implementation evolves.
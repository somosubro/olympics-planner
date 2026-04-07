# Importer Spec

## 1. Purpose

This document defines the importer pipeline that converts official LA28 schedule source material into canonical runtime session data for the Olympics Sessions Planner.

The importer exists to:
- read the official LA28 source schedule
- extract structured session information
- normalize that information into canonical runtime fields
- validate imported data quality
- emit runtime-ready data artifacts

This document is the source of truth for:
- importer responsibilities
- importer inputs and outputs
- pipeline stages
- normalization rules
- validation rules at import time
- importer-only versus runtime fields
- expected failure behavior

---

## 2. Scope

The importer is an **offline ingestion tool**, not a runtime API component.

It is responsible for transforming source schedule material into structured data used by the runtime planner.

### In Scope
- importing from LA28 schedule PDF
- importing from pre-extracted text
- parsing schedule rows and continuation lines
- normalizing fields
- generating canonical runtime session output
- optionally enriching sessions with importer-only derived metadata
- validating imported output
- writing output files

### Out of Scope
- serving HTTP traffic
- planning or ranking sessions at runtime
- GPT/application orchestration
- persistence of user plans
- ticketing or price integration
- live synchronization with remote LA28 systems

---

## 3. Design Principles

### 3.1 Importer Is Separate from Runtime Planner
Importer code must remain separate from runtime API and planner logic.

Recommended separation:
- `cmd/import_sessions`
- `internal/ingest/...`

### 3.2 Canonical Output, Flexible Input
Source PDFs may be messy and layout-dependent.
Importer internals may use raw/intermediate shapes, but output must map cleanly into the canonical runtime data contract.

### 3.3 Deterministic Output
Given the same input source and importer version, output should be deterministic.

### 3.4 Fail Clearly
Importer failures should be explicit and diagnosable.

### 3.5 Do Not Silently Promote Fields
Derived/enriched fields discovered during parsing must not silently become runtime dependencies unless promoted in `docs/data-contract.md`.

---

## 4. Inputs

The importer supports two input modes.

### 4.1 PDF Input
The primary source input is an official LA28 schedule PDF.

Example:
- `LA28OlympicGamesCompetitionScheduleByEventV3.0.pdf`

### 4.2 Extracted Text Input
The importer may also accept pre-extracted text, typically produced from the PDF using a tool such as `pdftotext`.

This mode is useful for:
- debugging parsing
- testing
- environments where direct PDF extraction is not available

### 4.3 Input Assumptions
The importer assumes:
- the source document is an official LA28 session schedule artifact or equivalent extracted text
- the extracted text is produced in a layout-preserving form where rows and continuation lines are still recoverable with heuristics

---

## 5. Outputs

### 5.1 Primary Output
The primary importer output is canonical runtime session data, written as **a single JSON file containing a top-level array** of session objects (see §15.1). The default output path is:

- `data/sessions.json`

The CLI must accept an output path flag (for example `-out`) so callers can override the destination.

### 5.2 Optional Secondary Outputs
Optional outputs may include:
- extracted text file
- debug parse artifacts
- validation reports
- rejected-row diagnostics

These are optional and may be implementation-specific.

### 5.3 Output Contract
Primary emitted runtime data must conform to `docs/data-contract.md`.

The sessions artifact must be valid JSON whose **root value is an array** of objects, each object one canonical `Session`. Do **not** wrap the array in an object such as `{ "sessions": [...] }`.

At minimum, each emitted session must aim to populate:
- `id`
- `sport`
- `sessionCode`
- `date`
- `dayOfWeek`
- `startTime`
- `venue`

Preferred additional fields:
- `title`
- `endTime`
- `includedEvents`

---

## 6. Pipeline Stages

The importer pipeline has five conceptual stages:

1. extract
2. parse
3. normalize
4. validate
5. emit

### 6.1 Extract
Convert PDF input into raw text.

Examples:
- call `pdftotext -layout`
- read a `.txt` file directly in text mode

### 6.2 Parse
Convert raw text into intermediate session records.

Parsing is expected to be heuristic and layout-aware.

### 6.3 Normalize
Convert intermediate parsed fields into canonical normalized data:
- sport names
- dates
- day of week
- times
- venue names
- canonical session identity

### 6.4 Validate
Run importer-time validation checks:
- required fields present
- date/day consistency
- uniqueness constraints
- basic structural sanity

### 6.5 Emit
Write output artifacts:
- canonical sessions JSON (top-level array per §5.3 and §15.1)
- optional debug artifacts

---

## 7. Recommended Package Structure

Recommended project placement:

```text
cmd/
  import_sessions/
    main.go

internal/
  ingest/
    pdf/
      extractor.go
      parser.go
    transform/
      normalize.go
      validate.go
    pipeline/
      import_sessions.go
```

### Package Responsibilities

#### `cmd/import_sessions`
CLI entrypoint only:
- parse flags
- choose import mode
- invoke pipeline
- print success/failure

#### `internal/ingest/pdf`
Source-extraction and source-format parsing:
- PDF extraction
- text scanning
- row parsing

#### `internal/ingest/transform`
Normalization and import-time validation:
- normalize fields
- derive canonical values
- validate imported data

#### `internal/ingest/pipeline`
Orchestration:
- run extract/parse/normalize/validate
- emit output

---

## 8. Import Modes

### 8.1 Import From PDF
CLI shape (subcommand + flags):

```text
go run ./cmd/import_sessions import -pdf data/raw/schedule.pdf -out data/sessions.json
```

Expected behavior:
- verify `pdftotext` or equivalent extraction tool exists
- extract text from PDF
- parse extracted text
- normalize and validate sessions
- write output JSON

### 8.2 Import From Text
CLI shape:

```text
go run ./cmd/import_sessions import-text -text data/raw/schedule.txt -out data/sessions.json
```

Expected behavior:
- read text directly
- parse text
- normalize and validate sessions
- write output JSON

---

## 9. Intermediate Models

Importer internals may use intermediate models that do not match the runtime contract.

### 9.1 Raw Extracted Record
Example conceptual shape:

```json
{
  "rawSport": "Tennis",
  "rawVenue": "LA Tennis Center",
  "rawZone": "West",
  "rawSessionCode": "TEN12",
  "rawDate": "Saturday, July 15",
  "rawSessionType": "Quarterfinal",
  "rawDescriptionLines": [
    "Men's Singles Quarterfinal",
    "Women's Singles Quarterfinal"
  ],
  "rawStartTime": "14:00",
  "rawEndTime": "17:00"
}
```

### 9.2 Intermediate Parsed Session
After parsing but before full runtime normalization, importer may use a richer internal struct containing:
- raw and normalized values side by side
- parse diagnostics
- importer-only enrichment fields

These are allowed internally but are not runtime contracts.

---

## 10. Runtime Output Model

The importer must emit canonical runtime sessions shaped according to `docs/data-contract.md`.

### 10.1 Canonical Runtime Session Example

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

### 10.2 ID Generation
The importer must assign a **unique string `id`** to every emitted session. Rules for MVP:

- **Deterministic** for the same source input and importer version.
- **Unique** across all sessions in the emitted array (see IV2).

**Aligned with `docs/data-contract.md` §4.4**, either strategy is valid:

1. **Session code as `id`** — use the normalized official session code (for example `TEN12`, `ATH09`) when it is unique within the import output. Set `sessionCode` to the same value or keep it as the business code field per data contract.
2. **Prefixed stable `id`** — derive `id` from code for readability, for example `session-ten-12`, `session-ath-09`.

If two rows would collide on `id`, the importer must disambiguate (for example append `-<YYYY-MM-DD>` from the session date, or another deterministic suffix) so IV2 holds.

The runtime treats **`id` as whatever unique string the pipeline wrote** to the sessions file; prefixed forms are recommended for new work but are not mandatory when the schedule code alone is unique.

---

## 11. Normalization Rules

### 11.1 Sport Normalization
The importer must normalize source sport names into canonical runtime values.

Examples:
- `Field Hockey` remains `Field Hockey`
- `Beach Volleyball` remains `Beach Volleyball`

If the source row omits sport, importer may infer sport from session-code prefix when that mapping is trusted.

### 11.2 Session Code Normalization
`sessionCode` must be:
- uppercase as represented by source or normalized equivalent
- stripped of extra whitespace
- stable

### 11.3 Date Normalization
Importer must normalize dates to:
- `YYYY-MM-DD`

### 11.4 Day-of-Week Normalization
Importer must normalize day names to:
- `Monday`
- `Tuesday`
- `Wednesday`
- `Thursday`
- `Friday`
- `Saturday`
- `Sunday`

### 11.5 Time Normalization
Importer must normalize times to:
- `HH:MM`
- 24-hour local LA time as represented by source schedule

If the source uses single-digit hour formatting such as `9:00`, importer should normalize to `09:00`.

### 11.6 Venue Normalization
`venue` should be trimmed and normalized for whitespace.
Avoid duplicate runtime variants that differ only by spacing or trivial text noise.

### 11.7 Included Events Normalization
Continuation lines or event-description lines should be normalized into:
- `includedEvents: string[]`

Importer should:
- trim whitespace
- drop empty lines
- deduplicate identical repeated event strings where safe

### 11.8 Title Normalization
If a trustworthy human-readable title exists, importer may emit it directly.

If not, importer may synthesize a runtime title using:
- `<sport> - Session <sessionCode>`
or equivalent stable formatting

---

## 12. Importer-Only Enrichment Fields

The importer may compute fields such as:
- `stage`
- `zone`
- `keywords`
- `interesting`
- `finalsHeavy`
- `marquee`
- `durationMins`

### 12.1 MVP Rule
These fields may exist in importer internals or optional emitted debug artifacts, but they are not part of the required runtime contract unless promoted in `docs/data-contract.md`.

### 12.2 Runtime Dependency Rule
Runtime validation and MVP scoring must not depend on importer-only fields that have not been promoted.

---

## 13. Import-Time Validation Rules

Importer-time validation is distinct from runtime validation.

Its goal is to ensure the emitted dataset is structurally sane.

### IV1. Required Runtime Fields Must Be Present
Each emitted session should contain, at minimum:
- `id`
- `sport`
- `sessionCode`
- `date`
- `dayOfWeek`
- `startTime`
- `venue`

If one of these is missing, the importer should either:
- reject that session from primary output, or
- emit it only if the chosen failure policy explicitly allows degraded rows

### IV2. Session IDs Must Be Unique
No two emitted sessions may share the same canonical `id`.

### IV3. Session Codes Should Be Stable
`sessionCode` should be non-empty and stable for any row treated as a real session.

### IV4. Date and Day Must Be Consistent
If both are present, `date` and `dayOfWeek` must agree.

### IV5. Times Must Be Parseable
`startTime` must be parseable for any emitted session.
`endTime` is preferred but optional for MVP.

### IV6. Empty Rows Must Not Emit Sessions
Header rows, disclaimer rows, and blank rows must never become emitted sessions.

---

## 14. Handling Partial or Invalid Records

The importer must define explicit behavior for bad or incomplete rows.

### 14.1 Header and Boilerplate Rows
These must be skipped.

Examples:
- PDF title lines
- version lines
- disclaimer text
- column headers

### 14.2 Partial Sessions
If a row appears to represent a session but cannot produce required runtime fields, importer should treat it as invalid for primary output.

### 14.3 Degraded but Usable Sessions
A session with missing:
- `title`
- `endTime`
- `includedEvents`

may still be emitted if required runtime validation fields are present.

### 14.4 Recommended Failure Policy
For MVP:
- fail closed on missing required runtime fields
- allow degraded optional fields
- log or report rejected rows for debugging

---

## 15. Emitted File Shape

### 15.1 Normative Sessions File
The primary sessions file is **valid JSON whose root is an array** of canonical session objects:

```json
[
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
```

An empty import is represented as `[]`.

### 15.2 Preferences Separation
Planner **preferences** are not part of the sessions artifact. They live in a separate file (for example `data/preferences.json`) and are maintained outside the importer.

The importer **must not** embed preferences inside the sessions JSON file. It **must not** be the source of truth for preferences.

---

## 16. Error Handling

### 16.1 Extract Errors
Examples:
- missing PDF file
- `pdftotext` unavailable
- extraction command failed

These should fail the import run with a clear message.

### 16.2 Parse Errors
Examples:
- source text unreadable
- no recoverable session rows found
- malformed date/time fragments

These should fail clearly if they prevent meaningful output.

### 16.3 Validation Errors
Examples:
- duplicate IDs
- missing required fields after normalization
- inconsistent date/day

These should either:
- fail the run, or
- surface in a structured validation report, depending on configured importer mode

### 16.4 Exit Behavior
Recommended CLI behavior:
- exit non-zero on fatal import failure
- print count of emitted sessions on success
- optionally print count of rejected rows when debug mode is enabled

---

## 17. Logging and Diagnostics

The importer should support diagnostics useful for parser drift and source changes.

Recommended diagnostics:
- source file path
- extracted text path, if applicable
- emitted session count
- rejected/invalid row count
- warnings for suspicious parse patterns

Optional future debug artifacts:
- `rejected_rows.json`
- `validation_report.json`

---

## 18. Testing Requirements

At minimum, importer tests should cover:

### 18.1 Extraction Tests
- PDF extraction command wiring
- extracted text file handling
- fallback text-mode import

### 18.2 Parser Tests
- row parsing from realistic extracted text snippets
- continuation-line capture into `includedEvents`
- header/boilerplate skipping
- sport inference from code prefixes when needed

### 18.3 Normalization Tests
- date normalization
- day normalization
- time normalization
- venue normalization
- title synthesis
- deterministic ID generation

### 18.4 Validation Tests
- missing required fields
- duplicate ID detection
- date/day mismatch detection
- degraded optional fields still allowed

### 18.5 Regression Tests
- known sample text -> stable canonical JSON
- known tricky source patterns remain parseable
- source-layout changes are detected early by failing tests

---

## 19. Implementation Conformance

Implementations of this pipeline must:

- keep importer packages separate from runtime API and planner logic (see §7)
- emit only the canonical session array format in the primary output file (§5.3, §15.1)
- treat importer-only enrichment as non-runtime unless promoted in `docs/data-contract.md` (§12)
- never mix preferences into the sessions artifact (§15.2)

---

## 20. Open Decisions

The following remain open:
- whether to standardize on prefixed `id` values versus session-code-only `id` when both satisfy uniqueness (see §10.2)
- whether rejected-row artifacts should be emitted by default
- whether importer should support multiple LA28 source layout versions with explicit version adapters
- whether importer-only enrichment fields should be emitted in a separate debug file
- whether import validation failures should hard-fail by default or support a tolerant mode

---

## 21. Summary

The importer is an offline ingestion pipeline.

It:
- reads LA28 schedule source material
- extracts and parses session rows
- normalizes them into canonical runtime data
- validates imported structure
- emits runtime-ready session artifacts

It must remain separate from runtime planning logic, and it must not silently expand the runtime contract without corresponding updates to the documentation.

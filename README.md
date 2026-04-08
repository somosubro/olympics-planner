# Olympics Schedule Planner

Starter Go backend for the GPT-fronted Olympic sessions planner.

## Current scope

This backend is intentionally thin. It is meant to own:
- session retrieval
- ranking
- validation

It is **not** meant to encode end-user planning flows like weekend planning, day-trip planning, or family itinerary generation.

## Commands

```bash
make run
make test
make test-race
make lint
make build
```

## API (local)

With `data/sessions.json` present, the server exposes `http://localhost:8080/api/v1/` (see `docs/api-spec.md`).

```bash
make run
# or: go run ./cmd/api
curl -s http://localhost:8080/api/v1/health
curl -s 'http://localhost:8080/api/v1/sessions?sports=Tennis&dayOfWeek=Saturday'
```

**Production (Cloud Run):** `https://olympics-schedule-planner-api-530886147910.us-central1.run.app` — e.g. `curl -s "https://olympics-schedule-planner-api-530886147910.us-central1.run.app/api/v1/health"`. Update [`docs/gpt/openapi.yaml`](docs/gpt/openapi.yaml) if `gcloud run deploy` prints a new base URL.

Cross-origin browser calls are allowed via CORS (`Access-Control-Allow-Origin: *`) for local testing.

Optional: open [`test-api.html`](test-api.html) in your browser (file URL) and click the buttons while the server runs.

## Hosting

You do **not** have to use Lambda. A **container** (Dockerfile in repo) on Cloud Run, Fly.io, Railway, Render, ECS, etc. is usually the simplest path. See [`docs/hosting.md`](docs/hosting.md). **Cloud Run:** [`docs/hosting-cloud-run.md`](docs/hosting-cloud-run.md). **GitHub → Cloud Run (CI):** [`docs/hosting-github-deploy.md`](docs/hosting-github-deploy.md).

## Custom GPT (ChatGPT)

**Wire the API:** paste the **entire** [`docs/gpt/instructions.md`](docs/gpt/instructions.md) into the GPT **Instructions** field (copy-paste from `main`; it changes often); import [`docs/gpt/openapi.yaml`](docs/gpt/openapi.yaml) as **Actions** so `servers.url` points at your deployed base URL. One-time setup, Knowledge, privacy URL, and starters: [`docs/gpt/configure-gpt.md`](docs/gpt/configure-gpt.md). **GitHub Pages landing:** [`docs/index.html`](docs/index.html) — [`docs/GITHUB_PAGES.md`](docs/GITHUB_PAGES.md). **Future persistence:** [`docs/gpt/roadmap-user-persistence.md`](docs/gpt/roadmap-user-persistence.md). The API must use **HTTPS** (e.g. ngrok while developing). Preferences in JSON: [`docs/preferences-guide.md`](docs/preferences-guide.md). **Regression checklist:** [`docs/gpt-test-plan.md`](docs/gpt-test-plan.md).

### Optional web browsing (trip context)

The **instructions** use **layered** rules: **schedule is always API-only** (`listSessions`, validation, real `session.id` values)—never the web for times, venues, or which sessions exist—and **`listSessions` must run before any web search** when the user asks for plans, ticket choices, or group trips (so the model does not “hunt the Olympics API” or browse before Actions). Optionally turn **Web** **on** for **non-schedule** context **after** that (rough pricing/hospitality, hotels, fan chatter)—clearly **separate** from plan text, with citations and “verify official sellers.” See **§E2** in [`docs/gpt/configure-gpt.md`](docs/gpt/configure-gpt.md) (**Option A** vs **B**).

## Import LA28 schedule (CLI)

Requires [Poppler](https://poppler.freedesktop.org/) `pdftotext` on your `PATH`.

```bash
go run ./cmd/import_sessions import -pdf ./LA28OlympicGamesCompetitionScheduleByEventV3.0.pdf -out data/sessions.json
```

Or from pre-extracted text (`pdftotext -layout …`):

```bash
go run ./cmd/import_sessions import-text -text ./schedule.txt -out data/sessions.json
```

The importer prints counts to stdout/stderr: sessions written, lines scanned, schedule rows matched, and any dropped invalid or duplicate IDs.

## Environment variables

- `PORT` default `8080`
- `SESSIONS_FILE` default `data/sessions.json`
- `PREFERENCES_FILE` default `data/preferences.json`

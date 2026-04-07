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

Cross-origin browser calls are allowed via CORS (`Access-Control-Allow-Origin: *`) for local testing.

Optional: open [`test-api.html`](test-api.html) in your browser (file URL) and click the buttons while the server runs.

## Hosting

You do **not** have to use Lambda. A **container** (Dockerfile in repo) on Cloud Run, Fly.io, Railway, Render, ECS, etc. is usually the simplest path. See [`docs/hosting.md`](docs/hosting.md). **Cloud Run:** [`docs/hosting-cloud-run.md`](docs/hosting-cloud-run.md). **GitHub → Cloud Run (CI):** [`docs/hosting-github-deploy.md`](docs/hosting-github-deploy.md).

## Custom GPT (ChatGPT)

To wire a **Custom GPT** to this API, use [`docs/gpt/instructions.md`](docs/gpt/instructions.md) (copy the Instructions block) and [`docs/gpt/openapi.yaml`](docs/gpt/openapi.yaml) (import as **Actions**). The API must be reachable over **HTTPS** (use a tunnel such as ngrok while developing). Preferences are sent in JSON request bodies; see [`docs/preferences-guide.md`](docs/preferences-guide.md).

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

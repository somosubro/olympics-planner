# Olympics Planner

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

## Environment variables

- `PORT` default `8080`
- `SESSIONS_FILE` default `data/sessions.json`
- `PREFERENCES_FILE` default `data/preferences.json`

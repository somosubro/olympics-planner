# Hosting the API

The server is a normal Go **`http.ListenAndServe`** process. It reads `SESSIONS_FILE` (default `data/sessions.json`) from the local filesystem on each request batch.

## Do you need Lambda?

**No.** Lambda is optional. It fits **spiky, infrequent** traffic and “pay per invoke,” but adds complexity for this shape of app.

| Approach | Fit |
|----------|-----|
| **Container (Docker)** on **Cloud Run**, **Fly.io**, **Railway**, **Render**, **ECS/Fargate**, **App Runner** | **Easiest mental model:** same binary, env vars, optional TLS from the platform. **Recommended** starting point. |
| **VM** (EC2, Droplet) + systemd | Simple if you already run VMs; you manage OS and deploys. |
| **AWS Lambda** | Possible, but **not** a drop-in: today’s code is a **long-lived HTTP server**, not a `lambda.Start` handler. |

### If you still want Lambda

Reasonable patterns:

1. **Lambda Web Adapter (container)** — Run the **same** HTTP server in a container image; API Gateway or Lambda Function URL forwards HTTP to the adapter. Minimal Go changes; you still bundle or fetch `sessions.json` (image layer, EFS, or S3 + download at cold start).
2. **API Gateway + native Lambda handler** — Refactor to handle API Gateway proxy events in Go (`aws-lambda-go`). More invasive; only worth it if you standardize on Lambda.

Cold starts matter: first request may be slow if the session file is large. For steady traffic, **containers on Cloud Run / ECS** are often simpler than Lambda.

## Google Cloud Run

Step-by-step deploy (CLI, HTTPS URL, public API): [`hosting-cloud-run.md`](hosting-cloud-run.md).

## Docker (repo root)

```bash
docker build -t olympics-schedule-planner-api .
docker run --rm -p 8080:8080 olympics-schedule-planner-api
curl -s http://localhost:8080/api/v1/health
```

Override data path or port:

```bash
docker run --rm -p 8080:8080 \
  -v "$(pwd)/data/sessions.json:/app/data/sessions.json:ro" \
  -e SESSIONS_FILE=/app/data/sessions.json \
  olympics-schedule-planner-api
```

## HTTPS for Custom GPT / browsers

ChatGPT Actions need a **public HTTPS** URL. After the API is reachable:

- Managed platforms usually give you HTTPS automatically.
- For your laptop, use a tunnel (e.g. **ngrok**, **Cloudflare Tunnel**) in front of `docker run` or `go run`.

## Environment variables

| Variable | Default | Notes |
|----------|---------|--------|
| `PORT` | `8080` | Must match what the platform expects (often set by the host). |
| `SESSIONS_FILE` | `data/sessions.json` | Path inside the container or on a mounted volume. |
| `PREFERENCES_FILE` | `data/preferences.json` | Loaded by config only; HTTP handlers currently take preferences in JSON bodies. |

## Security

For anything beyond local dev, put **authentication** (API key, JWT, etc.) in front of the API and configure the Custom GPT Action to match. The MVP server does not enforce auth.

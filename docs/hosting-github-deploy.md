# Deploy from GitHub to Cloud Run

This flow uses **Google Cloud Build** connected to your **GitHub** repo: every push (or tag) can build the **Dockerfile**, push to **Artifact Registry**, and deploy to **Cloud Run**—no deploy from your laptop required.

The repo includes a root [`cloudbuild.yaml`](../cloudbuild.yaml) for that pipeline.

## What you need

- GCP project **`olympics-schedule-planner`** (or yours) with **billing** linked.
- **Cloud Run Admin API**, **Cloud Build API**, **Artifact Registry API** enabled (same as local deploy).
- A **GitHub** repository (can be private) with this code.

## 1. Create an Artifact Registry repository (once per region)

Pick the same **`_REGION`** you use for Cloud Run (e.g. `us-central1`). Docker repo name below matches **`_REPO`** in `cloudbuild.yaml` (`olympics-schedule-planner`).

```bash
gcloud config set project olympics-schedule-planner

gcloud artifacts repositories create olympics-schedule-planner \
  --repository-format=docker \
  --location=us-central1 \
  --description="Olympics Schedule Planner API images"
```

If you change region, create another repository in that region or change substitutions in the trigger.

## 2. Grant Cloud Build permission to deploy

Cloud Build runs as **`PROJECT_NUMBER@cloudbuild.gserviceaccount.com`**. It must push images and deploy Cloud Run.

Get the project number:

```bash
gcloud projects describe olympics-schedule-planner --format='value(projectNumber)'
```

Grant roles (replace `PROJECT_ID` and `PROJECT_NUMBER`):

```bash
PROJECT_ID=olympics-schedule-planner
PROJECT_NUMBER=$(gcloud projects describe "$PROJECT_ID" --format='value(projectNumber)')
CB_SA="${PROJECT_NUMBER}@cloudbuild.gserviceaccount.com"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${CB_SA}" \
  --role="roles/run.admin"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${CB_SA}" \
  --role="roles/iam.serviceAccountUser"

gcloud projects add-iam-policy-binding "$PROJECT_ID" \
  --member="serviceAccount:${CB_SA}" \
  --role="roles/artifactregistry.writer"
```

`roles/iam.serviceAccountUser` lets Cloud Build act as the runtime service account Cloud Run uses (needed for deploy).

## 3. Connect GitHub to Cloud Build

1. Open [Cloud Build → Repositories](https://console.cloud.google.com/cloud-build/repositories) (or **Connections** / **2nd gen** depending on UI).
2. **Connect repository** → choose **GitHub (Cloud Build GitHub App)**.
3. Authenticate GitHub and **install the Google Cloud Build app** on your user or org.
4. Select the **repository** that contains this project and **connect** it.

If the console shows **“Host connection”** or **Region** for the connection, pick a region (often same as your Cloud Run region or `us-central1` for the connection resource).

## 4. Create a build trigger

1. Go to [Cloud Build → Triggers](https://console.cloud.google.com/cloud-build/triggers).
2. **Create trigger**.
3. **Event:** e.g. **Push to a branch** → branch `^main$` (or `^master$`).
4. **Source:** your connected repo.
5. **Configuration:** **Cloud Build configuration file (yaml or json)** → location **Repository root** → `/cloudbuild.yaml`.
6. **Substitution variables** (optional): only if you need to override defaults in `cloudbuild.yaml`:
   - `_REGION` = `us-central1` (must match Artifact Registry repo location)
   - `_SERVICE` = `olympics-schedule-planner-api`
   - `_REPO` = `olympics-schedule-planner`
7. Save.

`SHORT_SHA` is provided automatically for GitHub pushes so the image tag is unique per commit.

## 5. Run the pipeline

- **Push** to `main` (or run the trigger **Run** manually in the console).

Watch [Cloud Build → History](https://console.cloud.google.com/cloud-build/builds). On success, Cloud Run updates and the service URL stays the same unless you change region/service name.

## 6. First-time checklist if the build fails

| Issue | What to check |
|--------|----------------|
| **Permission denied** pushing to Artifact Registry | `roles/artifactregistry.writer` on the project (or repo) for Cloud Build SA. |
| **Permission denied** on `gcloud run deploy` | `roles/run.admin` + `roles/iam.serviceAccountUser` for Cloud Build SA. |
| **Repository not found** | Artifact Registry repo exists in **`_REGION`** with id **`_REPO`**. |
| **SHORT_SHA empty** | Trigger must be from a **GitHub** push (not manual with no commit context); or switch tag to `$COMMIT_SHA` in `cloudbuild.yaml`. |

## Changing region or names

Edit **substitutions** in the trigger (or defaults in [`cloudbuild.yaml`](../cloudbuild.yaml)). Create an Artifact Registry repo in the new region if needed.

## Alternative: GitHub Actions

If you prefer workflows entirely in GitHub, use **GitHub Actions** with [google-github-actions/auth](https://github.com/google-github-actions/auth) (Workload Identity Federation recommended) and `gcloud run deploy` or the same Docker build/push/deploy steps. Cloud Build + trigger keeps secrets and IAM in GCP only; Actions is fine if your team standardizes on it.

## Security

`cloudbuild.yaml` uses **`--allow-unauthenticated`** for a public API (e.g. ChatGPT). For a private API, remove that flag and require authentication; update the trigger substitutions or add a separate production trigger.

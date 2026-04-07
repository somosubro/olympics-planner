# Deploy to Google Cloud Run

These steps deploy this repo’s **Dockerfile** as a Cloud Run service. The container listens on **`PORT`** (Cloud Run sets this automatically; the app defaults to `8080`). Schedule data is **baked into the image** from `data/` at build time.

## Prerequisites

1. **Google account** and a **GCP project** ([console.cloud.google.com](https://console.cloud.google.com)).
2. **Billing enabled** on the project (Cloud Run’s free tier still requires a billing account; you typically pay nothing while within [free tier limits](https://cloud.google.com/run/pricing)).
3. **[Google Cloud SDK](https://cloud.google.com/sdk/docs/install)** (`gcloud`) installed on your machine.

### Install `gcloud` (macOS)

If you see **`gcloud: command not found`**:

1. **Install** with [Homebrew](https://brew.sh/):

   ```bash
   brew install --cask gcloud-cli
   ```

   (Older docs sometimes used the cask name `google-cloud-sdk`; either installs the CLI.)

2. **Put Homebrew on your `PATH`** (Apple Silicon default is `/opt/homebrew/bin`). For **bash**, add to `~/.bash_profile` (create the file if needed):

   ```bash
   echo 'export PATH="/opt/homebrew/bin:$PATH"' >> ~/.bash_profile
   source ~/.bash_profile
   ```

   For **zsh** (default on recent macOS), use `~/.zshrc` instead.

3. **Open a new terminal tab/window**, then:

   ```bash
   gcloud --version
   ```

   If it still fails, call it by full path once: `/opt/homebrew/bin/gcloud --version`.

---

## Google Cloud Console setup (project, billing, organization)

Use this if you are **in the console** and need to choose names, billing, and where the project lives.

### Terms that matter

| Term | What it is |
|------|------------|
| **Organization** | Optional top-level node for **companies and schools** (Google Workspace / Cloud Identity). If you sign in with a **personal Gmail** account, you often have **no organization**—that is normal. |
| **Folder** | Optional grouping of projects **under** an organization (IT uses this for teams). You can ignore it unless your admin created folders. |
| **Project** | The **container** for APIs, Cloud Run, billing link, and IAM. Everything you deploy for this app should live in **one project** to start. |
| **Project name** | Friendly label (e.g. “Olympics Schedule Planner”). You can change it later. |
| **Project ID** | **Globally unique**, permanent identifier used in URLs and `gcloud`. This guide uses **`olympics-schedule-planner`** as the example. If that ID is already taken globally, add a suffix (e.g. `olympics-schedule-planner-48291`). |
| **Project number** | A numeric ID Google assigns. You rarely need it for this guide. |
| **Billing account** | The **payment profile** (card, company contract). **Linking** a billing account to a **project** unlocks paid-capable services. One billing account can pay for **many** projects. |

### Personal Gmail vs work / school

- **Personal (`@gmail.com`)**: You’ll usually create a project with **no organization**. The “Organization” dropdown may be empty or unavailable—use **No organization** if shown.
- **Work or school (Google Workspace)**: Your company may have an **Organization**. You might be **required** to create the project **inside** that org, or only in certain folders. If **Create project** is greyed out, ask your **Google Cloud admin** which folder to use or to grant **Project Creator**.

### Step 1 — Open the right place

1. Go to [console.cloud.google.com](https://console.cloud.google.com).
2. Sign in with the Google account you want to own this workload.
3. If the **project picker** at the top shows many projects, note which one is selected; you can create a **new** one next.

### Step 2 — Create a project (if you don’t have one yet)

1. Click the **project dropdown** at the top (next to “Google Cloud”).
2. Click **New project** (or open [Resource Manager](https://console.cloud.google.com/cloud-resource-manager) and click **Create project**).
3. **Project name**: e.g. `Olympics Schedule Planner` (display name only).
4. **Project ID**: Click **Edit** and set **`olympics-schedule-planner`** (or add a suffix if that ID is unavailable).
5. **Organization / Location**:  
   - No company org: choose **No organization** if available.  
   - Company org: pick the **organization** or **folder** your admin told you to use.
6. Click **Create**. Wait until the project appears in the picker, then **select it** so the top bar shows this project as **active**.

### Step 3 — Billing account (first time or new card)

Cloud Run needs a project with **billing linked** (even when usage stays in the free tier).

1. Open [**Billing**](https://console.cloud.google.com/billing) from the left menu (or search the top bar for “Billing”).
2. If you have **no billing account yet**:
   - Click **Add billing account** (wording may vary).
   - Enter **account name** (e.g. “Personal” or your company name), **country**, **currency**, and **payment method** (card or invoicing if eligible).
   - Finish the wizard; Google may run a **small verification** charge that is reversed.
3. **Link billing to your project**:
   - In Billing, open your **billing account** → **My projects** (or **Account management** → linked projects).
   - **Link** or **Change billing** for project **`olympics-schedule-planner`** (or whatever ID you chose) and select this billing account.  
   - Alternatively: from [**Billing**](https://console.cloud.google.com/billing/linkedaccount) with the **project selected**, choose **Link a billing account**.

If billing is managed by your **organization**, you might only see a **company billing account**—use that only if policy allows personal experiments.

### Step 4 — Confirm you’re in the right project

- Top bar: project name matches **Olympics Schedule Planner** (or whatever you created).
- Optional: [**IAM & Admin → Settings**](https://console.cloud.google.com/iam-admin/settings) shows **Project ID** and **Project number**—copy the **Project ID** for `gcloud`.

### Step 5 — Enable APIs in the console (optional)

You can enable APIs **without** the CLI:

1. Open [**APIs & Services → Library**](https://console.cloud.google.com/apis/library).
2. Search and **Enable** each (console names can differ slightly from shorthand):
   - **Cloud Run Admin API** — this is the Cloud Run API (`run.googleapis.com`). There is no separate listing called “Cloud Run API” in the library.
   - **Cloud Build API**
   - **Artifact Registry API**

Or skip this and use `gcloud services enable` in the next section (it enables them for you).

---

## One-time setup (CLI)

### 1. Log in and select the project

```bash
gcloud auth login
gcloud config set project olympics-schedule-planner
```

Use the **Project ID** from **IAM & Admin → Settings** if yours differs from **`olympics-schedule-planner`**.

### 2. Enable required APIs

```bash
gcloud services enable run.googleapis.com cloudbuild.googleapis.com artifactregistry.googleapis.com
```

### 3. Pick a region

Choose a [Cloud Run region](https://cloud.google.com/run/docs/locations) close to you or your users, e.g. `us-central1`, `europe-west1`, `asia-northeast1`. Use the same value as `REGION` below.

**Tier 1 vs Tier 2 pricing (and “free tier”):** The console may label regions as **Tier 1** or **Tier 2**. That describes **how much you pay per request/CPU/memory when usage is billable**—Tier 2 regions are often **more expensive** per unit than Tier 1. It is **not** a choice between “free” and “paid” products.

The **monthly free allowance** (requests, CPU, memory) applies to Cloud Run on your **billing-enabled project**; you do **not** pick “free tier” on the region screen. While your usage stays **within** those monthly free limits, you typically **pay $0** regardless of Tier 1 vs Tier 2. If you later **exceed** the free allowance, prefer a **Tier 1** region if you want the lower per-unit rates. For light personal traffic, pick **latency** first (region near you), then Tier 1 if you care about future cost.

## Deploy from the repo root

From the **repository root** (where `Dockerfile` and `go.mod` live):

```bash
cd /path/to/olympics-schedule-planner
pwd   # should end with your repo folder, not ~ or /Users/you
ls Dockerfile go.mod   # both must exist

gcloud run deploy olympics-schedule-planner-api \
  --source . \
  --region us-central1 \
  --allow-unauthenticated
```

**Important:** `--source .` uploads the **current directory**. If you run this from your **home directory** (`~`), `gcloud` tries to upload your entire home folder (including `Library/News/...` on macOS) and can crash or fail with `FileNotFoundError` on odd paths. Always **`cd` into the cloned repo first**.

The repo includes a [`.gcloudignore`](../.gcloudignore) to keep uploads small when deploying from the correct directory.

**Build failed at `COPY go.mod go.sum`:** This module may have **no `go.sum`** file (no external Go modules). The [Dockerfile](../Dockerfile) copies only `go.mod` first, then runs `go mod download`, so Cloud Build does not require `go.sum`.

- **`olympics-schedule-planner-api`** — Cloud Run service name (change if you like).
- **`--source .`** — Builds the image with **Cloud Build** using your **Dockerfile**, then deploys.
- **`--allow-unauthenticated`** — Lets anyone call the HTTPS URL without an Identity Token (needed for a **public HTTP API** and **ChatGPT Actions**). Omit only if you will use authenticated access only.

On first deploy, `gcloud` may ask to enable APIs or link Artifact Registry; accept the prompts.

When it finishes, the command prints the **service URL**, for example:

`https://olympics-schedule-planner-api-xxxxx-uc.a.run.app`

## Verify

```bash
curl -s "https://YOUR_SERVICE_URL/api/v1/health"
```

You should see JSON like `{"status":"ok"}`.

## Custom GPT / OpenAPI

1. Set `servers[0].url` in [`docs/gpt/openapi.yaml`](gpt/openapi.yaml) to your **Cloud Run URL** (no trailing slash), e.g. `https://olympics-schedule-planner-api-xxxxx-uc.a.run.app`.
2. Re-import the schema in the GPT **Actions** editor if it changed.
3. Paste the full [`docs/gpt/instructions.md`](gpt/instructions.md) into **Instructions** when it changes; full editor steps: [`docs/gpt/configure-gpt.md`](gpt/configure-gpt.md).

## Updating schedule data

`data/sessions.json` is **copied into the image** at build time. After you change it:

```bash
gcloud run deploy olympics-schedule-planner-api \
  --source . \
  --region REGION \
  --allow-unauthenticated
```

Redeploy so Cloud Build produces a new image.

## Optional: reduce cold starts

By default Cloud Run scales to zero; the first request after idle can be slower.

```bash
gcloud run services update olympics-schedule-planner-api \
  --region REGION \
  --min-instances 1
```

This can incur **continuous** cost—check [pricing](https://cloud.google.com/run/pricing).

## Optional: environment variables

To override paths or add config later:

```bash
gcloud run services update olympics-schedule-planner-api \
  --region REGION \
  --set-env-vars "SESSIONS_FILE=/app/data/sessions.json"
```

(Defaults already match the paths in the Dockerfile.)

## Security note

`--allow-unauthenticated` is appropriate only if the API is **intended to be public**. For production, add **API keys or another auth layer** and restrict invocations; ChatGPT Actions can send a secret header once you implement verification in the Go service.

## Deploy from GitHub (CI)

To build and deploy on every push via **Cloud Build** + **GitHub**, follow [`hosting-github-deploy.md`](hosting-github-deploy.md) and use the repo root [`cloudbuild.yaml`](../cloudbuild.yaml).

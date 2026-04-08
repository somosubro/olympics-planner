# Custom GPT — editor guide (setup & reference)

**Instructions for the model** live alone in [`instructions.md`](instructions.md): open that file, select all, paste into ChatGPT **Configure → Instructions**. That file changes most often; keep this guide for one-time setup and maintenance.

Pair the GPT with **Actions** using [`openapi.yaml`](openapi.yaml) (`servers.url` must match your deployed API).

Optional **Knowledge** uploads: [`user-readme.md`](user-readme.md), [`preset-plans.md`](preset-plans.md) — see notes in `instructions.md` and below.

---

## A. Prepare the OpenAPI file (once per deploy URL)

1. Open [`openapi.yaml`](openapi.yaml) in this repo.
2. Find the **`servers:`** section near the top. Under it, set **`url:`** to your Cloud Run base URL **with no path and no trailing slash**. The repo is kept in sync with production, for example: `https://olympics-schedule-planner-api-530886147910.us-central1.run.app` (if you redeploy elsewhere, paste the URL `gcloud` prints). (“`servers[0]`” means the first entry under `servers:`.)
3. Save the file. You will paste its **full contents** into ChatGPT in step D (or **Import from URL** if you host the raw YAML publicly and the UI supports it).

---

## B. Open your GPT in the editor

1. ChatGPT → **My GPTs** (profile / sidebar, or **Explore GPTs** → **My GPTs**).
2. Open the GPT → **Edit** (pencil).

---

## C. Knowledge: turn off or replace static schedule files

1. In **Configure**, open **Knowledge**.
2. **Remove** uploaded PDFs/CSVs/JSON that duplicated the **session schedule** (they can contradict the API). Optional: keep only non-schedule docs that do not conflict with live data.
3. If you clear schedule Knowledge, the model relies on **Actions** + **Instructions** — intended for live sessions.

---

## D. Actions: attach the API

1. **Actions** → **Create new action** (or **Add action**).
2. Remove or replace any old schema pointing at a mock URL or localhost.
3. **Schema** → **Paste** the entire edited `openapi.yaml` (or import from URL).
4. **Authentication**: for a public Cloud Run service with no API key, choose **None**. Add API key / Bearer later if you secure the API.
5. **Save**. Fix validation errors (usually `servers.url` or YAML indentation).

---

## E. Instructions

1. **Configure** → **Instructions**.
2. **Replace** the field with the **full contents** of [`instructions.md`](instructions.md) (entire file).
3. Save.

---

## E2. Capabilities — disable web search / browsing (strongly recommended)

If the GPT editor offers **Web**, **Search**, **Browse**, or **Use web sources**, turn it **OFF** for this planner.

When web is **on**, the model often runs long multi-site searches instead of calling your **Actions** API, invents or mismatches session codes, and is **slow**. Schedule data is **API-only**—see [`instructions.md`](instructions.md) “Data source (API only—no web)”.

(Exact labels vary by ChatGPT version; look under **Configure** → capabilities or advanced settings.)

**Optional web for pricing/social context:** use a **separate** branch of this repo (e.g. `feature/gpt-optional-web`) that restores the Pattern A instructions; do not mix with `main` unless you accept that risk.

---

## F. Name, description, and starters (optional)

Adjust **Name** / **Description** as needed. **Conversation starters:** see [Suggested conversation starters](#suggested-conversation-starters) below.

---

## G. Test before sharing

1. Use **Preview**.
2. Ask for live data, e.g. “List tennis sessions on Saturday.” Confirm responses match your API (not stale uploads).
3. On errors, check **Actions** and Cloud Run **Logs** in Google Cloud Console.

---

## H. Share with family

1. **Save** the GPT.
2. Set **visibility** (only you, link-only, or public).
3. Share the **GPT link** — users use ChatGPT with your GPT; they do not need the raw `*.run.app` URL.

---

## I. Privacy policy URL (if ChatGPT requires it)

GPTs that use **Actions** and are **public** (or link-shared in some cases) may need a working **HTTPS** **Privacy policy** URL.

1. Edit [`privacy-policy.html`](privacy-policy.html) and replace **`replace-with-your-email@example.com`** with a real contact address or URL.
2. Host via **GitHub Pages** (or similar). Exact URL pattern: see [`../GITHUB_PAGES.md`](../GITHUB_PAGES.md) — e.g. `https://<user>.github.io/<repo>/gpt/privacy-policy.html`.
3. GPT editor → **Configure** → **Privacy policy** → that HTTPS URL → save.

If you only need family access, try **Anyone with the link** first — requirements can differ from **Public** in the store.

---

## Configure Actions (short version)

1. API must be **HTTPS** (e.g. Cloud Run).
2. Set **`url`** under **`servers:`** in `openapi.yaml` to your base URL (no trailing slash).
3. **Configure** → **Actions** → paste schema → **Authentication: None** if the API is open.

---

## Suggested conversation starters

- “What tennis sessions are on Saturday July 15, 2028?”
- “I want a two-day weekend plan with athletics and tennis.”
- “Rank these sessions for me—I prefer athletics over swimming.”
- “Is this plan valid?” (then assemble a plan from real IDs and validate)

---

## Related files in `docs/gpt/`

| File | Purpose |
|------|--------|
| [`instructions.md`](instructions.md) | **Paste entire file** into GPT Instructions (changes frequently). |
| [`openapi.yaml`](openapi.yaml) | Actions schema; set `servers.url`. |
| [`user-readme.md`](user-readme.md) | Optional Knowledge — user-facing help. |
| [`preset-plans.md`](preset-plans.md) | Optional Knowledge — story examples only, not authoritative IDs. |
| [`privacy-policy.html`](privacy-policy.html) | Template for Privacy policy URL OpenAI may require. |
| [`roadmap-user-persistence.md`](roadmap-user-persistence.md) | Plan: GitHub Pages landing + future saved preferences/plans on the API. |

**Public landing page (GitHub Pages):** [`../index.html`](../index.html) — big button to ChatGPT. Setup: [`../GITHUB_PAGES.md`](../GITHUB_PAGES.md).

**Repo docs outside this folder:**

- [`../api-spec.md`](../api-spec.md) — full HTTP behavior  
- [`../preferences-guide.md`](../preferences-guide.md) — preference semantics  
- [`../data-contract.md`](../data-contract.md) — `Session`, `Plan`, `Preferences` shapes  

**Deploy:** [`../hosting-cloud-run.md`](../hosting-cloud-run.md), [`../hosting-github-deploy.md`](../hosting-github-deploy.md)

**Manual GPT tests:** [`../gpt-test-plan.md`](../gpt-test-plan.md)

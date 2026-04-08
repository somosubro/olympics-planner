# Roadmap: landing page + saved preferences & plans

This document ties together (1) a **public GitHub Pages** entry point, (2) **future persistence** of user data on **your API**, and (3) **GPT behavior** so users understand how saving and recall work—including limits.

---

## 0. Preference gate (today: GPT only)

The Custom GPT is instructed to **collect and confirm preferences before** building or ranking trip plans; preferences live in **conversation memory** only. That is **not** enforced by the HTTP API (there is no server-side “must have profile first” flag). When you add persistence (below), you can optionally mirror the same rules in the backend later.

---

## 1. What you want (summary)

| Now | Later |
|-----|--------|
| **GitHub Pages** with a **large “Open in ChatGPT”** button so sharing feels like a normal link. | Users **persist** preferences and saved session/plan snapshots, and **recall** them **through the same Custom GPT**. |
| No account system required for the static page. | **Backend** stores data; **GPT** explains flows and calls new HTTP methods (POST/PATCH/GET/DELETE as designed). |

**GitHub Pages cannot** run a database or execute server logic. It only serves static files (`index.html`, etc.). **All persistence** must live on **your Go API** (e.g. Cloud Run) or another hosted service you control.

---

## 2. Phase A — Landing page (ship first)

**Goal:** One memorable HTTPS URL (e.g. `https://<you>.github.io/<repo>/`) that:

- Explains in one sentence what the planner is.
- Offers a **primary CTA**: **Open in ChatGPT** → your real `https://chatgpt.com/g/...` link.
- Links to **privacy policy** (`gpt/privacy-policy.html`).

**Repo:** [`docs/index.html`](../index.html) — replace `REPLACE_WITH_YOUR_GPT_ID` in the `href` with your actual GPT URL from ChatGPT (**My GPTs** → your GPT → copy link).

**Docs:** [GITHUB_PAGES.md](../GITHUB_PAGES.md) — ensure Pages source is **`/docs`** so the site root serves `index.html`.

**Optional:** Short custom domain (DNS CNAME to `*.github.io`) so relatives remember one name; still points to static files only.

---

## 3. Phase B — Persistence (expansion)

### 3.1 Product primitives (suggested)

Define clear nouns so the GPT and API stay aligned:

| Concept | Meaning |
|--------|---------|
| **Profile / wallet** | Logical “place” where a user’s saved items live. |
| **Saved preferences** | Named or versioned `preferences` JSON (or diff from defaults) the user wants to reuse. |
| **Saved plan** | A snapshot: `plan` + optional label, timestamps, maybe a human title. |
| **Session shortlist** | Optional list of `session.id` values the user starred for later (if you add it). |

You do **not** have to implement all at once; start with **one** (e.g. named preferences) and add plans next.

### 3.2 Hard requirement: **identity**

The API must know **which user** owns which row. Custom GPT **Actions** do **not** automatically send a stable OpenAI user id to your backend.

**Common patterns (pick one for MVP, evolve later):**

1. **Bearer token per user** (simplest for a small family product)  
   - You issue a random token (or sign-up flow later).  
   - User pastes it **once** in the GPT conversation or you store it in **GPT Instructions** only for *your* account (bad for sharing).  
   - Better: **Actions authentication** — ChatGPT supports **API key** or **OAuth** in the schema; each family member could paste a personal key in the GPT’s **Authentication** settings if the UI allows per-user keys (limitations vary).  

2. **OAuth 2.0** (best UX at scale)  
   - User clicks “Sign in” on a tiny web page **you** host (not Pages-only), redirects to IdP, returns tokens; GPT uses OAuth in Actions.  
   - More engineering; aligns with “real” accounts.

3. **Opaque user id in conversation** (MVP hack)  
   - User says “my save id is `abc-123`” and the GPT sends `X-User-Id: abc-123` on every call.  
   - Weak security unless combined with a secret; okay for experimentation only.

4. **Magic link email** (middle ground)  
   - Backend sends a link with a token; establishes a long-lived session cookie in a **mini web app**; GPT integration still needs a token the model can send—usually converges on (1) or (2).

**Recommendation for a first iteration:** design the API as **`Authorization: Bearer <token>`** and document that the **first** release might distribute tokens manually (family list); later swap in OAuth without changing resource shapes.

### 3.3 API sketch (REST)

Names are illustrative; version under `/api/v1/`.

| Method | Path | Purpose |
|--------|------|---------|
| `POST` | `/api/v1/me/preferences` | Create saved preferences (body + optional name). |
| `GET` | `/api/v1/me/preferences` | List saved preference sets. |
| `GET` | `/api/v1/me/preferences/{id}` | Fetch one. |
| `PATCH` | `/api/v1/me/preferences/{id}` | Update. |
| `DELETE` | `/api/v1/me/preferences/{id}` | Delete. |
| `POST` | `/api/v1/me/plans` | Save a plan snapshot. |
| `GET` | `/api/v1/me/plans` | List. |
| … | … | Same pattern for shortlists. |

**Storage:** Cloud Firestore, PostgreSQL (Cloud SQL), or Dynamo-style store—pick based on ops comfort. **Never** store ChatGPT threads; store **only** what the user explicitly saves via API.

### 3.4 Constraints (product + legal)

- **Data minimization:** Store only preferences/plan JSON needed for recall; no unnecessary PII.
- **Retention:** Document how long you keep rows; support **delete everything** for a user.
- **Regions:** If EU/UK users matter, note GDPR-style rights in the privacy policy before widening use.
- **Rate limits:** Protect Cloud Run from abuse (per token/IP).
- **Cost:** Every persisted object is tiny; watch read/write volume if the GPT loops.

---

## 4. GPT guidance (what to add to `instructions.md` when features ship)

When persistence exists, extend the Custom GPT instructions so it:

1. **Explains the model in plain language**  
   - “Your saved items live on **our server** under **your account** (token). I can list, save, update, or delete them only by calling the tools the API exposes.”

2. **Says what is *not* stored**  
   - e.g. “I don’t automatically save every chat; I only persist when you ask me to save or when we use a save tool.”

3. **Repeats preference constraints**  
   - Same as today: `rules.noSameSportAcrossDays` default, allowed sports/days, validation via `validatePlan`, etc.—**saving** does not bypass validation unless you explicitly allow draft saves (product choice).

4. **Walks through recall**  
   - “Load my weekend preferences” → `GET` → merge into next `validatePlan` / `rankPlans`.

5. **Security hygiene**  
   - Never ask users to paste tokens in public channels; prefer Actions auth configuration where possible.

Keep this file in sync with **OpenAPI** (`openapi.yaml`): every tool the GPT claims must exist as a real operation.

---

## 5. Ordering of work (suggested)

1. **Landing `index.html`** + replace GPT URL; verify GitHub Pages.  
2. **Auth decision** (Bearer MVP vs OAuth).  
3. **DB + minimal API** (e.g. save/list/delete preferences only).  
4. **OpenAPI + GPT Actions** for new routes; **instructions** paragraph on save/recall.  
5. **Plans snapshots** and optional shortlists.  
6. **Privacy policy** update: what is stored, where, retention, contact.

---

## 6. Does this all make sense?

Yes:

- **Pages** = **front door** and **shareable link**; **no** server state.  
- **Cloud Run API** = **source of truth** for saved data; **POST/PATCH/GET/DELETE** as you grow.  
- **GPT** = **interface** + **guidance**; must be documented in **instructions** and **user-readme** so constraints and flows stay understandable.

This roadmap is the contract between static site, backend, and assistant behavior until you implement each slice.

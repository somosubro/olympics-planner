# User guide: planner preferences

This guide explains how **you** define preferences for the Olympics Sessions Planner. Preferences are **your** rules and soft priorities—not something the importer or PDF decides.

The machine-readable shape is defined in [data-contract.md](data-contract.md) §9. This document is for **people** who edit JSON or use the [preferences builder](../preferences-builder.html).

---

## 1. What preferences are for

- **Validation** (`POST /api/v1/validate`) — Is a candidate plan allowed given which sports and days you care about, and whether the same sport may repeat across days?
- **Ranking** (`POST /api/v1/rank/sessions`, `POST /api/v1/rank/plans`) — Among valid options, what do you *prefer* (sport priority, day pairs)?

The conversational app (or you, when calling the API) sends a **single effective `preferences` object** per request. There is no separate “overrides” field in MVP: merge any chat-specific tweaks into that object before calling the API ([api-spec.md](api-spec.md) §11).

---

## 2. What you set

| Field | In plain language |
|-------|-------------------|
| **`allowedSports`** | Sports you are willing to consider. Sessions outside this set fail validation / ranking rules that depend on preferences. |
| **`sportPriority`** | Order from **most** to **least** wanted. Used when **ranking**—higher-listed sports score higher (within the model in [scoring-and-validation-spec.md](scoring-and-validation-spec.md)). |
| **`allowedDays`** | Which weekdays are in play (e.g. only `Saturday` and `Sunday` for a weekend trip). |
| **`rules.noSameSportAcrossDays`** | **Default `true`**: the **same sport** cannot appear on two different calendar days (including **alternates / add-ons**). Set **`false` only** if the user explicitly wants that. |
| **`rules.preferDayPairs`** | Soft ranking hint: which **pairs of weekdays** you like for multi-day plans (e.g. Saturday + Sunday). Optional but useful for plan scores. |
| **`rules.minHoursBetweenSameDaySessions`** | **Default when omitted: `4`** — minimum hours from one session’s **end** to the next’s **start** on the same day (feasible travel in the LA area; not turn-by-turn routing). Set **`0`** to turn off this check. |
| **`rules.maxSessionsPerDay`** | Cap sessions per calendar day (e.g. **`1`** when the user wants at most one event that day). Omit for no cap. |
| **`rules.sportSpecific`** | Reserved; use `{}` for MVP. |

Every sport name should match how it appears in **`data/sessions.json`** (e.g. `Tennis`, `Athletics`, `Sailing (Windsurfing & Kite)` if that is what was imported).

---

## 3. Empty lists mean “none”

Per [scoring-and-validation-spec.md](scoring-and-validation-spec.md) §7.1–7.2:

- **`allowedSports: []`** → no sport is allowed for those checks.
- **`allowedDays: []`** → no day is allowed.

So for real use, **fill both** with at least one value each. A starter file is provided as [`data/preferences.example.json`](../data/preferences.example.json).

---

## 4. How to create your preferences

### Option A — Visual builder (easiest)

1. Open [`preferences-builder.html`](../preferences-builder.html) in your browser (from the repo folder).
2. Fill in sports, priority order, days, and rules.
3. Copy or download the JSON.

### Option B — Edit JSON by hand

1. Copy `data/preferences.example.json` to `data/preferences.json` (or merge carefully).
2. Edit with your editor; validate against [data-contract.md](data-contract.md) §9.

### Option C — App-owned file

Point the API at your file with the environment variable:

```bash
export PREFERENCES_FILE=/path/to/my-preferences.json
```

The backend loads this path when it needs the default preferences file (depending on deployment). For **`POST` endpoints**, you still send **`preferences` in the JSON body**—that is what the API uses for that call.

---

## 5. Quick examples

**Weekend-only, tennis and athletics, weekend day pair preferred:**

```json
{
  "allowedSports": ["Tennis", "Athletics"],
  "sportPriority": ["Tennis", "Athletics"],
  "allowedDays": ["Saturday", "Sunday"],
  "rules": {
    "noSameSportAcrossDays": true,
    "preferDayPairs": [["Saturday", "Sunday"]],
    "sportSpecific": {}
  }
}
```

**Broader trip (Fri–Mon), cricket first among many sports:**

```json
{
  "allowedSports": ["Cricket", "Athletics", "Tennis", "Swimming"],
  "sportPriority": ["Cricket", "Athletics", "Tennis", "Swimming"],
  "allowedDays": ["Friday", "Saturday", "Sunday", "Monday"],
  "rules": {
    "noSameSportAcrossDays": true,
    "preferDayPairs": [
      ["Saturday", "Sunday"],
      ["Friday", "Saturday"],
      ["Sunday", "Monday"]
    ],
    "sportSpecific": {}
  }
}
```

---

## 6. Checking that it works

1. Run the API (`make run`).
2. Call `POST /api/v1/validate` or `POST /api/v1/rank/plans` with a body that includes your `preferences` object.
3. Adjust `allowedSports` / `allowedDays` if plans are rejected more often than you expect; adjust `sportPriority` and `preferDayPairs` to change ranking—not validity—when multiple plans are valid.

For questions about scores and caps, see [scoring-and-validation-spec.md](scoring-and-validation-spec.md) §12–15.

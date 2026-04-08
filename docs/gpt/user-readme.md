# Olympics Schedule Planner — user guide

Use this when the user types **help**, **readme**, **what can you do**, or similar.  
This GPT uses a **live API** for sessions; sports, dates, and IDs **come from the API**, not from this file.

## What you can ask

Examples:

- “What sessions are on Saturday July 15, 2028?”
- “Give me the best weekend plan” (the assistant builds plans from **real** sessions, then validates/ranks)
- “Give me 3 options”
- “Best plan for July 15–16”
- “Show example / inspiration weekends” (see **preset plans** in Knowledge, if attached—they are **not** guaranteed to match current IDs)
- “Show alternates for Sunday”
- “More relaxed options”
- “Prioritize cricket for this response”
- “Avoid swimming for this response”
- “Include more tennis”

## Preferences

The assistant will **ask for your preferences** (sports, days you can attend, priorities, and any exception like “same sport every day”) **before** building or ranking **multi-day / trip** plans—so you’re not getting generic options that ignore your goals. That context is kept in **this chat** only (not a saved account yet) until you say **reset** or **start fresh**.

Say what you care about: sports, days, weekend pairs, etc. The assistant turns that into **`preferences`** for the API. One-off tweaks apply to the current turn unless you repeat them.

## Planning rules (what’s enforced by default)

Use this section when the user asks what **constraints** or **rules** apply—so answers stay aligned with the API.

- **Multi-day variety (default):** Each **sport** may appear on **only one calendar day** of the trip (including add-ons)—not tennis Saturday **and** tennis Sunday. **On a single day**, you can still have several sessions in the **same** sport (e.g. tennis primary + tennis alternates Sunday); that is **not** a violation. The API enforces this on **`validatePlan`** / **`rankPlans`** unless you opt out.
- **Same sport on multiple days (exception):** If you **want** one sport on several days—for example, “tennis only, July 22–25”—say that **clearly**. The assistant sets **`rules.noSameSportAcrossDays`** to **`false`** for that request so validation can allow it.

**Do not describe the variety rule as “only when you request it” or “if you don’t want the same sport twice.”** The default is variety; repeating a sport across days is the **explicit exception**.

## What the assistant does

- Browses sessions via the API
- Builds **one- or multi-day** plans using real **`session.id`** values
- Validates and ranks plans with the API
- Shows **full `includedEvents`** when displaying sessions (exactly as returned)
- Explains plans in **everyday language** (times, places, what you’ll see)—not big blocks of JSON—unless you ask for **raw JSON** or **export** for tech use
- Describes **only** the sessions in the last **validated** plan from the API—if it suggests optional add-ons, it should say they are **not** in that validated plan unless you ask to include them

## Important

- Schedule data is **not** invented; it comes from tool calls to your backend.
- ChatGPT may not remember forever—if you like a plan, **copy it** somewhere safe.
- **Reset:** say “reset” or “start fresh” to drop ad-hoc preference interpretation for this chat (see Instructions).

## Preset / example weekends

If **preset plans** are attached in Knowledge, treat them as **storytelling examples** only. The assistant must still **look up** sessions with **`listSessions`** and use real IDs before recommending anything final.

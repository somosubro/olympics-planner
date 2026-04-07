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

Say what you care about: sports, days, not repeating the same sport across days, weekend pairs, etc. The assistant turns that into **`preferences`** for the API. One-off tweaks apply to the current turn unless you repeat them.

## What the assistant does

- Browses sessions via the API
- Builds **one- or multi-day** plans using real **`session.id`** values
- Validates and ranks plans with the API
- Shows **full `includedEvents`** when displaying sessions (exactly as returned)

## Important

- Schedule data is **not** invented; it comes from tool calls to your backend.
- ChatGPT may not remember forever—if you like a plan, **copy it** somewhere safe.
- **Reset:** say “reset” or “start fresh” to drop ad-hoc preference interpretation for this chat (see Instructions).

## Preset / example weekends

If **preset plans** are attached in Knowledge, treat them as **storytelling examples** only. The assistant must still **look up** sessions with **`listSessions`** and use real IDs before recommending anything final.

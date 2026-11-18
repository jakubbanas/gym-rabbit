# Gym Rabbit — UX Audit & Improvement Plan

*Audited on 2026-11-18 against the codebase (`frontend/src/`) and live deployment (`https://gym.kuuur.de/`) across desktop (`1280×800`) and mobile (`390×844`).*

---

## What Works Well

- **Low-glare dark aesthetic & oversized primary targets:** The `#0a0a0a` canvas, `h-28` (`112px`) `LOG SET` / `LEFT` / `RIGHT` buttons, and color-coded barbell sleeve visualizer fit mid-workout fatigue well (`frontend/src/components/WorkoutView.tsx`, `frontend/src/components/ToolsDrawer.tsx`).
- **Clear block/exercise labeling:** Block-letter prefixes (`R1.`, `A1.`, `C1.`/`C2.`) and per-set progress bars are easy to scan at a glance (`frontend/src/lib/api.ts`).

---

## 1. Critical & High-Impact Issues

### 1.1 No session persistence or exit confirmation
- **Where:** `frontend/src/App.tsx:12-16`, `frontend/src/components/WorkoutView.tsx:69-74`
- **Problem:** Workout progress (`progress`, `currentBlockIdx`) lives only in React `useState` despite `GEMINI.md` specifying `localStorage` offline persistence. Refreshing the page, switching apps on mobile, or tapping the small header back arrow immediately discards all logged sets with no confirmation.
- **Improvement:**
  - Persist active session state (`workoutId`, `progress`, `currentBlockIdx`, `startedAt`) to `localStorage` on every change.
  - Surface a **Resume active workout** banner on the dashboard if an unfinished session exists.
  - Ask for confirmation when tapping Back if at least one set has been logged.

### 1.2 `last_performed` updates on card open instead of actual work
- **Where:** `frontend/src/App.tsx:18-25`
- **Problem:** `startWorkout` calls `POST /api/workouts/{id}/performed` immediately on card click. Tapping a past workout just to review exercises overwrites `Last: ...` to today.
- **Improvement:** Only call `touchLastPerformed` once the user logs their first set (or when finishing the session).

### 1.3 Unclickable coaching video links & no actual reps/weight logging
- **Where:** `frontend/src/components/WorkoutView.tsx:62, 132-140, 163-220` (and see `docs/agent-prompt.md:63, 79`)
- **Problem:**
  - `docs/agent-prompt.md` instructs the planner to attach video links (`Wideo: https://youtu.be/...`) in `cues`. In `WorkoutView`, cues render as plain text inside a `select-none` container, so video links can neither be tapped nor copied.
  - Plans include `5RM` (`Dokładaj ciężar aż dojdziesz do maksa na 5 powtórzeń`) and `1 × AMRAP` followed by `3 × 70% AMRAPa`, yet there is no way to note the actual weight/reps achieved (still keeping one-tap `LOG SET` as the default).
- **Improvement:**
  - Auto-link URLs inside `cues` (rendered as a clean clickable link/video pill with `target="_blank" rel="noopener noreferrer"`), and remove `select-none` from text/cue blocks.
  - Add an optional quick-adjust / result input on the card (e.g., tap the weight or reps pill to log actual `kg` / `reps` for that set, and optionally compute `70% AMRAP` hints when an earlier AMRAP rep count was recorded).

### 1.4 Floating Wrench FAB collides with content on every screen
- **Where:** `frontend/src/components/ToolsDrawer.tsx:41-49`, `frontend/src/App.tsx:97`
- **Problem:** The `fixed bottom-28 right-6` FAB overlaps:
  - Workout cards on `Dashboard` (no footer exists there),
  - The `60s` rest badge and `LOG SET` button in multi-exercise blocks (`Rozgrzewka R`, `Blok C`),
  - And floats isolated at the far edge on desktop (`1280×800`) while content is centered in `max-w-md`.
- **Improvement:** Replace the floating FAB with a dedicated Barbell Math icon button in the header (and/or clickable weight pills on exercise cards that pre-fill `weight_kg` into the calculator).

### 1.5 Rest timer drifts in background, skips on stray taps, and breaks iOS audio
- **Where:** `frontend/src/components/RestTimer.tsx`, `frontend/src/App.tsx:41-58`, `frontend/src/lib/audio.ts:4-21`
- **Problem:**
  - `setInterval` decrementing `seconds - 1` drifts or pauses when the phone locks or switches apps.
  - `playBeep()` constructs a new `AudioContext` inside `useEffect` at `0s`, which iOS Safari blocks without an earlier user-gesture unlock.
  - `onClick={onSkip}` covers the whole `fixed inset-0` screen, causing accidental skips.
  - `120` at `text-[12rem]` (`192px`) lacks `m:ss` formatting (`2:00`), `-15s`/`+15s` controls, and "Up next" context.
  - Logging the **final set** of a block/exercise still forces a full-screen rest timer when `rest_seconds > 0`.
- **Improvement:**
  - Drive countdown from `endTimestamp = Date.now() + rest * 1000` + `visibilitychange` sync.
  - Initialize/resume a shared `AudioContext` inside the `LOG SET` click handler.
  - Replace full-screen click-to-skip with a dedicated `Skip Rest` button, `-15s` / `+15s` adjusters, `m:ss` formatting for `≥60s`, and a subtle "Up next" line.
  - Consider skipping or shortening the rest prompt when the very last set of the final exercise in the workout is logged.

### 1.6 No Undo for accidental set taps
- **Where:** `frontend/src/components/WorkoutView.tsx:163-220`
- **Problem:** Double-tapping or accidentally tapping `LOG SET` / `LEFT` / `RIGHT` cannot be undone.
- **Improvement:** Add a lightweight `Undo` button (or allow tapping the last filled set segment) to step back one set.

---

## 2. Medium-Impact Improvements

- **Dashboard error handling & clarity (`frontend/src/components/Dashboard.tsx`):**
  - `fetchWorkouts()` catches errors only to `console.error` (lines 52-57), falsely showing `"No workouts planned yet"` if the server is unreachable. Add an error banner with a `Retry` button.
  - When all sessions are in `Past`, `opacity-60` (line 67) dims the entire screen. Keep `Most Recent` or `Past` readable when `Upcoming` is empty, and add block/exercise counts (e.g., `7 blocks · 14 exercises`) on cards.
- **Tappable block navigation (`frontend/src/components/WorkoutView.tsx:82-89`):**
  - Replace the tiny non-interactive header dots with tappable block pills (`R`, `A`, `B`, `C`…) showing completion state (`✓` / active / pending) so lifters can reorder blocks when equipment is occupied.
- **Barbell Math ergonomics (`frontend/src/components/ToolsDrawer.tsx`):**
  - Hide native number input spinbuttons (`▲▼`) inside dark fields.
  - Add `-5`, `-2.5`, `+2.5`, `+5` steppers and a bar-weight toggle (`20 kg`, `15 kg`, `25 kg` trap bar).
  - Pre-fill from the active exercise's `weight_kg` when opened from a workout.
  - Warn when a target weight isn't achievable with `1.25 kg` increments (e.g., `111 kg` currently silently renders plates for `110 kg`).
  - Constrain drawer width to `max-w-md mx-auto` on desktop.
- **Richer Summary screen (`frontend/src/components/Summary.tsx`):**
  - Bodyweight/unweighted sessions (`totalVolume === 0`) currently hide the entire stats card. Always show **Sets Completed**, **Exercises Done**, **Session Duration**, plus a `"Back to Workout"` option in case `Finish Session` was tapped early.

---

## 3. Quick Wins & Accessibility Polish

- **Allow pinch-to-zoom:** Remove `maximum-scale=1.0, user-scalable=no` from `frontend/index.html:6`.
- **Add missing `aria-label`s & minimum 44×44px targets:** Header back button (`WorkoutView.tsx:69`), Tools button (`ToolsDrawer.tsx:41`), and Drawer close `X` (`ToolsDrawer.tsx:69`).
- **Fix orphaned `•` on narrow screens:** In `WorkoutView.tsx:121`, `<span>•</span>` can sit stranded at the end of line 1 when the load wraps onto line 2 (observed on `C2. Chin up`). Use clean pill spacing without a literal bullet character.
- **Accessible exercise prefix spacing:** Render `{exerciseLabel(...)}. ` with a trailing space (`WorkoutView.tsx:113`) so screen readers don't merge `C1.Chin up`.
- **Safe-area padding:** Use `pb-[env(safe-area-inset-bottom)]` on fixed footers and bottom drawers.

---

## 4. Suggested Implementation Roadmap

1. **Phase 1 — Trust & Ergonomics**
   - `localStorage` session persistence + Resume banner + Back confirmation
   - Delay `touchLastPerformed` until first logged set / completion
   - Replace floating wrench FAB with header/card triggers; improve Barbell Math controls & desktop width
   - Linkify cue URLs, remove global `select-none`, fix orphaned `•`, and apply a11y quick wins
2. **Phase 2 — Core Gym Flow**
   - Timestamp-based `RestTimer` + iOS `AudioContext` unlock + `m:ss` + `±15s` + dedicated `Skip` + "Up Next"
   - Per-card `Undo` set + tappable block pills (`R`, `A`, `B`, …) + optional actual reps/kg logging
3. **Phase 3 — Dashboard & Summary Polish**
   - Dashboard fetch error/retry state, richer workout card metadata, improved contrast when `Upcoming` is empty
   - Complete `Summary` metrics (sets, exercises, duration, back-to-workout)

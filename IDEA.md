# Gym Rabbit — Simplified Gym Tracker Specification

Gym Rabbit is a high-focus, friction-free workout tracker designed for two people: the **Trainer** planning the routine, and the **Lifter** executing it under the weights.

---

## 🎯 The Core Concept

* **For the Trainer:** Zero forms. Create, share, or edit workouts using a simple, standardized plain-text template. Copy, paste, and save.
* **For the Lifter:** Swipe, tap, lift. A focused, card-based interface with screen-wide haptic touch targets, an automated screen-wide ambient rest timer, and helpful built-in gym tools.

---

## 📋 Standardized Workout Template (The "Magic Paste" Format)

Workouts are parsed instantly and offline from a single, standard pipe-delimited format.

### The Template
```text
Plan Name: [Descriptive Plan Title]

Block [A-Z]
[Exercise ID]. [Exercise Name] | [Sets]x[Reps] | [Weight] | [Rest]
- [Optional technique cue or note line]
```

### Parsing Rules
* **Auto-Unilateral Mode:** Reps containing a `+` symbol (e.g., `8+8`) automatically configure the exercise as unilateral, spawning separate left and right tracking selectors.
* **Simple Separation:** The pipe (`|`) character is used as a strict separator to keep fields unambiguous.
* **Structured Cues:** Lines starting with `-` immediately following an exercise line become form cues visible during the set.
* **No Dates Required:** The parser only looks for the Plan Name and Block definitions, keeping the text as minimal as possible.

---

## 💡 Concrete Template Examples

### Example 1: Standard Strength Workout (Pure Bilateral)
```text
Plan Name: Hypertrophy Phase 1

Block A
A1. Barbell Back Squat | 3x6-8 | 100kg | 120s
- Keep chest up, control descent
A2. Romanian Deadlift | 3x8-10 | 80kg | 90s

Block B
B1. Bench Press | 4x8 | 75kg | 90s
B2. Pull-Ups | 3xAMRAP | Bodyweight | 60s
```

### Example 2: Unilateral & Balance Workout (Triggers Unilateral Mode via `+`)
```text
Plan Name: Strength & Balance

Block A
A1. Bulgarian Split Squat | 3x8+8 | 16kg DBs | 90s
- Push through the front heel
A2. Dumbbell Arnold Press | 3x10 | 12.5kg | 60s

Block B
B1. Single Arm Row | 3x10+10 | 20kg | 60s
- Drive elbows to hip crease
```

---

## ⚡ App Architecture & UX Flow

### 1. The Dashboard (Saved Routines)
* **Start Session:** Tapping a saved routine instantly starts the workout session.
* **Tweak Plan (Pencil Icon):** Opens the "Magic Paste" view, allowing the user to tweak weights, reps, or cues via plain text before starting.
* **Paste New Routine:** Opens an empty text area to drop in a new plan structure.

### 2. Distraction-Free Gym Mode
* **Focused Exercise Cards:** Displays only one exercise block at a time (e.g., superset A1 & A2). Lifters swipe left/right to navigate blocks.
* **Giant Set Bubbles (Sweaty-Hand Friendly):** Completed sets are logged using giant, rounded buttons with an `h-28` target size for easy logging with sweaty hands. Includes haptic feedback. Unilateral exercises split into giant **[ LEFT ]** and **[ RIGHT ]** columns.
* **Screen-Wide Ambient Rest Timer:** Logging a set triggers a pitch-black screen with a massive glowing countdown. Bypassed instantly by tapping anywhere on the screen.
* **Quick Exit:** A back arrow allows the user to abandon the session and return to the dashboard at any time.

### 3. Global Tools Drawer
* A floating "Tools" button is accessible from every screen (Dashboard, Input, Workout, and Rest Timer).
* **Barbell Math:** A visually stable calculator. The user inputs a target weight (e.g., `112.5kg`). Assuming a standard `20kg` bar, the app calculates the exact competition plates needed (`25, 20, 15, 10, 5, 2.5, 1.25`) and renders a gorgeous visual representation of the loaded barbell sleeve.

### 4. Smart Summary Screen
* When the workout is completed, the app parses the logged sets, weights, and reps to calculate the **Total Volume** lifted.
* Displays the total raw tonnage (e.g., `3,250 kg`).
* Includes a fun, dynamic comparison message based on the weight (e.g., *"That's equivalent to lifting a rhinoceros! 🦏"* or *"You just lifted a full pickup truck! 🛻"*).

---

## 🛠️ Technology Stack
* **Backend:** Go (Golang) — A simple, blazing-fast backend API server handling static file hosting and local SQLite data persistence.
* **Frontend:** React (built with Vite or as a single-page app served by the Go backend):
  * **Styling:** TailwindCSS for rapid, utility-first UI styling (e.g., glassmorphic rest timers, full-screen tap targets).
  * **Icons:** Lucide React icons for crisp, lightweight, and modern SVG iconography.
* **Storage:** Offline-first Browser `localStorage` to instantly persist local set completions, backed up to the Go SQLite server when online.
* **Device APIs:** Haptic Vibration API (`navigator.vibrate`) for tactile set confirmations directly in the browser.

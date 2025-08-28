# Copilot Rewrite Prompts (Go + Web)

This file contains two detailed, standalone prompts for two separate Copilot agents to rebuild the project cleanly and consistently:
- Prompt A: Go Copilot (WASM solver, GoCSV)
- Prompt B: Web Copilot (Vite + React + TypeScript + Material UI)

Both prompts share a single, explicit WASM API contract so each agent can implement independently without cross-dependencies beyond this contract.

---

## Prompt A — Go Copilot (WASM solver using GoCSV)

You are the Go owner for a browser-only, client-side exam scheduling solver. Your deliverable is a Go module that compiles to WebAssembly and exposes a stable JS API. No servers. All data stays in the browser.

Build for Go 1.22+.

### Objectives
- Implement a robust, deterministic exam scheduler in Go.
- Use GoCSV (github.com/gocarina/gocsv) for CSV parsing/serialization with proper quoted-field handling (commas inside quotes).
- Expose a WASM API (syscall/js) with explicit functions and JSON payloads defined below.
- Provide comprehensive unit tests, including edge-case CSVs (quoted commas, unbalanced quotes, variable columns).
- Focus on correctness, stability, and performance (~5–10k students, ~200–400 courses) on modern laptops.

### Project layout (you create under `go/`)
- go/go.mod, go.sum (module name: `exam-scheduler`)
- go/pkg/scheduler/
  - model.go        (types)
  - io.go           (CSV parse/serialize via GoCSV)
  - graph.go        (conflict graph, edge weights=shared students)
  - slots.go        (slot generation from date range + slots/day + holidays + times)
  - dsatur.go       (DSATUR list-coloring with allowed-slot sets + seeded tie-breaks)
  - penalty.go      (configurable penalty for temporal proximity/conflicts)
  - halls.go        (hall packing and allocation, avoid double-booking)
  - schedule.go     (orchestrator: N attempts with PRNG seed; best by penalty)
  - verify.go       (checks: student clashes, hall overbook, per-course constraints)
  - seed.go         (PRNG helper)
  - types_test.go, io_test.go, graph_test.go, dsatur_test.go, halls_test.go (unit tests)
  - testdata/ (tiny CSVs with tricky quoting)
- go/cmd/wasm/main.go (syscall/js bindings)

### Data model (minimal contract)
- StudentID = string
- CourseID = string
- HallID = string
- SlotID = string (stable, deterministic identifier, e.g., `2025-01-07T09:00Z#1`)
- Course { ID CourseID; Enrollments []StudentID }
- Hall { ID HallID; Capacity int; Group string(optional) }
- Slot { ID SlotID; Start time.Time; End time.Time; DayIndex int; IndexInDay int }
- Assignment { CourseID; SlotID; SlotDateTime RFC3339; Halls []HallID; EnrolledCount int; Notes string }

### CSV formats (use GoCSV)
- registrations.csv (long form): `student_id,course_id` (can contain additional columns; ignore others). Handle quoted commas and leading/trailing spaces. Skip blank lines and comment lines starting with `#`.
- halls.csv: `hall,capacity,group?` (group optional; default empty). Capacity must parse to int >= 0. Handle quoted fields.
- allowed-slots.csv (optional): `course_id,slot_id` (or accept 2+ columns with course then allowed slot IDs). If provided via params, restrict scheduling accordingly; otherwise treated as empty.

Parsing requirements:
- Use GoCSV with a custom `csv.Reader` configured: `LazyQuotes=true`, `TrimLeadingSpace=true`, `FieldsPerRecord=-1` to tolerate real-world CSV inconsistencies. Do not lose or merge fields; ensure quoted commas are preserved. Validate required columns exist; ignore unknown columns. For records missing required fields, skip with a warning.
- For serialization of the output schedule CSV, use `encoding/csv.Writer` with RFC4180 quoting rules to ensure fields containing commas or semicolons are quoted.

### Slot generation
- Inputs from params:
  - examStartDate (ISO date, inclusive)
  - examEndDate (ISO date, inclusive)
  - slotsPerDay (default 2)
  - slotTimes: array of local times as `HH:MM` (optional; if omitted, generate evenly spaced times based on slotDuration and day start time)
  - slotDuration (minutes)
  - holidays []ISO date strings to skip
  - timezone: optional IANA TZ string; default to browser local/time.Local assumption; if omitted, use UTC in WASM with clear doc
- Generate `Slot` list deterministically; create `SlotID` stable across runs given same inputs.

### Scheduling rules
- No overlapping exams per student.
- Respect optional per-course allowed slots.
- Hall allocation: choose 1+ halls whose total capacity >= enrollment; avoid double-book within a slot; prefer fewer halls; tie-break deterministically.
- Optional `minGap` (minutes) between exams for a student; if violated, treat as hard constraint or add penalty depending on `penaltyMode`.
- Objective: minimize penalty = weighted sum of (student proximity conflicts by days/gaps, spread, late exams, etc.). Provide simple default penalty model with tunable weights from params.
- N attempts (params.tries), seeded RNG for tie-breaks/order; pick best by penalty.

### WASM API — Authoritative Contract
Export these functions from `go/cmd/wasm/main.go` via `syscall/js`. All return a JSON string.

- version() string
  - Returns: `{ "name": "exam-scheduler-go", "version": "1.0.0", "go": runtime.Version() }`

- runSchedule(regCSV string, hallsCSV string, paramsJSON string) string
  - paramsJSON schema:
    {
      "examStartDate": "YYYY-MM-DD",
      "examEndDate": "YYYY-MM-DD",
      "slotsPerDay": number,
      "slotTimes": ["HH:MM", ...] (optional),
      "slotDuration": number,        // minutes
      "holidays": ["YYYY-MM-DD", ...],
      "tries": number,               // default 100
      "seed": number,                // default random, but must be echoed in stats
      "minGap": number,              // minutes (optional)
      "allowedSlotsCSV": string      // optional CSV text (course_id,slot_id)
    }
  - Returns on success:
    {
      "success": true,
      "scheduleCSV": "course_id,slot_id,slot_datetime,halls,enrolled_count,notes\n...",
      "report": {
        "valid": true,
        "conflicts": 0,
        "unassigned": [],
        "capacityWarnings": [],
        "totalExams": number,
        "totalSlots": number,
        "errors": []
      },
      "stats": {
        "seed": number,
        "totalTime": number,    // ms
        "attempts": number,
        "bestPenalty": number,
        "slotsUsed": number
      }
    }
  - Returns on infeasible/error:
    { "success": false, "error": string, "report": {...}, "stats": {...} }

- verify(regCSV string, scheduleCSV string) string
  - Returns report JSON with violations (clashes, overbooking, unknown halls/slots, unassigned courses) + counts, similar shape to runSchedule.report but without scheduleCSV.

Implementation notes:
- JSON marshalling must be stable and small. Use struct types with `json` tags.
- All randomness must be seeded and deterministic given the same inputs. Echo `seed` in `stats`.
- Do not panic; return structured errors.
- Ensure `halls` field in schedule CSV is a semicolon-separated list; quote the cell if needed per RFC.

### Testing
- Unit tests for:
  - CSV parsing (quoted commas, spaces, variable columns, comments). Include testdata with tricky rows (e.g., `"Doe, Jane",CSE101`).
  - Graph building: edge weights equal shared students, symmetry, degrees.
  - DSATUR coloring with allowed slots list; verify infeasible instance returns error.
  - Slot generation across date ranges; skip holidays; slot ID determinism.
  - Hall packing (single/multiple halls); avoid double-booking; tie-breaks deterministic.
  - Verification: detect student clashes and overbooked halls; report shapes.
- An end-to-end tiny dataset (4–8 courses) acceptance test.

### Build (WASM)
- Target output: `web/public/main.wasm` (do not assume server; this is a static site).
- Ensure compatibility with Go’s `wasm_exec.js` (frontend will load it alongside the wasm binary).

---

## Prompt B — Web Copilot (Vite + React + TypeScript + Material UI)

You own the browser UI. No backend. Integrate the WASM solver via the API specified in Prompt A. Do not change the contract; assume the WASM provides exactly these functions:
- `version(): string`
- `runSchedule(regCSV: string, hallsCSV: string, paramsJSON: string): string // JSON`
- `verify(regCSV: string, scheduleCSV: string): string // JSON`

### Stack
- Vite + React 18 + TypeScript
- Material UI (MUI) for components
- Web Worker for WASM execution (to keep UI responsive)
- Use PapaParse for client-side CSV preview only (strict RFC CSV parsing with quotes); pass original file text to WASM untouched

### Goals
- A 4-step guided UI:
  1) Uploads: registrations.csv and halls.csv
  2) Configure parameters
  3) Generate schedule (progress + results)
  4) Review & Download (CSV/JSON), and re-verify
- Persist everything to localStorage automatically (files as base64, column mappings, parameters, last results, current step). Provide Clear / Reset Session and Import/Export Session JSON.
- No network I/O beyond fetching `wasm_exec.js` and `main.wasm` from `public/`. Never upload user data.

### App structure (under `web/`)
- package.json (vite, react, ts, mui, papaparse)
- public/
  - index.html
  - wasm_exec.js (copied from Go toolchain)
  - main.wasm (built artifact from the Go project)
- src/
  - main.tsx, App.tsx
  - lib/
    - wasm.ts (types + thin wrappers if calling from main thread)
    - wasmWorker.ts (Module Worker that loads wasm_exec.js and main.wasm; posts PROGRESS/RESULT/ERROR)
    - storage.ts (localStorage helpers with versioned schema)
    - csv.ts (PapaParse helpers for preview; NO naive `split(',')` anywhere)
  - components/
    - FilePicker.tsx (preview, column mapping UI; keep original file text for WASM)
    - ParamsForm.tsx (MUI forms with validation; all fields below)
    - ProgressBar.tsx (status + percent)
    - ScheduleTable.tsx (sortable, filterable results; shows enrolledCount, halls, slot)
    - DownloadButtons.tsx (CSV/JSON download)
    - ValidationPanel.tsx (conflicts, capacity warnings, unassigned)

### Parameters (complete)
- examStartDate (date)
- examEndDate (date)
- slotsPerDay (number, default 2)
- slotTimes (array of `HH:MM`, optional). If provided, overrides even spacing.
- slotDuration (minutes)
- holidays (array of dates to skip)
- tries (number, default 100)
- seed (number; default random but shown to user)
- minGap (minutes, optional)
- allowedSlotsCSV (optional CSV text for per-course allowed slots)

UI behaviors:
- Validate dates and that end >= start; at least one slot per day; slotTimes valid `HH:MM`.
- Persist all fields to localStorage under a namespaced key with versioning. Autosave on change.
- Allow clearing session; confirm before wiping.
- Accessible form labels, keyboard navigation, focus order, and color contrast.

### Worker + WASM loading
- Use a Module Worker (type: module). In the worker:
  - Fetch `/wasm_exec.js` and evaluate its contents (ESM workers don’t support importScripts). Example: `const code = await fetch(...).then(r => r.text()); (0,eval)(code)`.
  - Instantiate Go WASM runtime: `const go = new (globalThis as any).Go()` and then fetch and instantiate `/main.wasm` with `WebAssembly.instantiateStreaming` if available (fallback to `arrayBuffer`).
  - Call `go.run(instance)`; then ensure `globalThis.runSchedule/verify/version` exist.
  - Accept messages `{ type: 'GENERATE_SCHEDULE', data: { registrationsCSV, hallsCSV, params } }` and post back `{ type: 'PROGRESS' | 'RESULT' | 'ERROR' }`.
  - Always pass original CSV text to WASM; do not reformat.

### CSV handling (frontend)
- For previews and column mapping only, use PapaParse with `header: true`, `skipEmptyLines: true`, and correct `quoteChar`, `escapeChar`.
- Never use naive string splitting.
- For displaying results from `scheduleCSV`, parse with PapaParse (`header: true`).

### Downloads
- CSV: Prefer the exact `scheduleCSV` returned by WASM; if user edited table locally, rebuild CSV with proper quoting using PapaParse unparse.
- JSON: `{ schedule, validationReport, generatedAt }` where `schedule` reflects the table rows currently displayed.

### Error handling
- Show clear toasts/panels for parsing errors, infeasible schedules, or WASM load failures. Include a “Copy Error Details” button.
- Support cancellation: if user clicks Cancel during generation, terminate the worker and start a fresh one.

### Persistence
- storage.ts: versioned schema, `save()` and `load()` helpers; files stored as base64 with filename and lastModified; restore on boot; include a big "Reset" and an "Export/Import session" feature.

### Minimal visual flow (MUI)
- Stepper on top; Back/Next on top bar; disabled states based on validation; Generate runs the worker.
- Show a preview table (first ~20 rows) for both CSVs with detected headers and column mapping selectors.
- After generation, show a results table with sorting and search, plus validation panel and download buttons.

### Acceptance tests (manual or scripted)
- Tiny dataset (4 courses, overlapping enrollments) fits within 3–4 slots.
- Capacity test: single course needs two halls; allocator picks multiple.
- Allowed slots test: course restricted to a single slot; either respected or infeasibility reported.
- CSV corner cases: quoted commas, extra columns, empty rows, comment lines.

### Build & run
- `npm run dev` launches the app.
- `npm run build` emits a static site; ensure `wasm_exec.js` and `main.wasm` are under `public/` and referenced by the worker via absolute paths (`/wasm_exec.js`, `/main.wasm`).

### Non-functional
- Privacy: Never upload user data. No analytics.
- Performance: Keep main thread responsive; all solver calls in the worker.
- Determinism: Display the `seed` used in stats and allow re-running with same seed.

---

## Shared API Contract (for both agents)

The Web and Go agents must both adhere to this API exactly. The web agent will implement against this contract without needing the Go code present.

- version(): string
- runSchedule(regCSV: string, hallsCSV: string, paramsJSON: string): string // returns JSON as described in Prompt A
- verify(regCSV: string, scheduleCSV: string): string // returns JSON report

Result JSON shapes are identical to those specified in Prompt A. The schedule CSV must have header:

```
course_id,slot_id,slot_datetime,halls,enrolled_count,notes
```

Where `halls` is a semicolon-separated list within a single CSV field, quoted if necessary.

---

## Notes for Both Agents
- Treat CSVs as untrusted; sanitize/validate but don’t mutate prior to scheduling. The solver receives original bytes/text.
- Handle commas inside quoted fields correctly. Do not split on commas manually anywhere.
- All operations are client-side. No network calls with user data.
- Favor small, deterministic, testable units.

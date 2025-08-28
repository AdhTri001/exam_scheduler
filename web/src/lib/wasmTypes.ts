/**
 * TypeScript definitions and documentation for the Go WASM Exam Scheduler
 *
 * This file defines the exact API contract that the Go WASM module exposes.
 * Agent B should implement against these types without needing the Go source code.
 */

// ===== INPUT TYPES =====

/**
 * Configuration for customizing CSV column names.
 * If not provided, default column names will be used.
 */
export interface ColumnMapping {
  /** Column name for student ID (default: "student_id") */
  studentIDColumn?: string;
  /** Column name for course ID (default: "course_id") */
  courseIDColumn?: string;
  /** Column name for hall ID (default: "hall") */
  hallIDColumn?: string;
  /** Column name for hall capacity (default: "capacity") */
  capacityColumn?: string;
  /** Column name for hall group (default: "group") */
  groupColumn?: string;
}

export interface RunScheduleParams {
  /** ISO date string (YYYY-MM-DD), inclusive start date for exams */
  examStartDate: string;

  /** ISO date string (YYYY-MM-DD), inclusive end date for exams */
  examEndDate: string;

  /** Number of exam slots per day (default: 2) */
  slotsPerDay: number;

  /** Optional array of time strings in HH:MM format. If provided, overrides even spacing */
  slotTimes?: string[];

  /** Duration of each exam slot in minutes */
  slotDuration: number;

  /** Array of ISO date strings to skip (holidays) */
  holidays: string[];

  /** Number of scheduling attempts to try (default: 100) */
  tries: number;

  /** Random seed for deterministic results. If 0, will be generated */
  seed: number;

  /** Minimum gap between exams for a student in minutes (optional) */
  minGap?: number;

  /** Optional CSV text for per-course allowed slots restriction */
  allowedSlotsCSV?: string;

  /** IANA timezone string (optional, default: "UTC") */
  timezone?: string;

  /** Custom column mapping for CSV parsing */
  columnMapping?: ColumnMapping;
}

// ===== OUTPUT TYPES =====

export interface ValidationReport {
  /** Whether the schedule is valid (no conflicts) */
  valid: boolean;

  /** Number of student conflicts detected */
  conflicts: number;

  /** Array of course IDs that could not be scheduled */
  unassigned: string[];

  /** Array of capacity warning messages */
  capacityWarnings: string[];

  /** Array of error messages */
  errors: string[];

  /** Array of student clash descriptions */
  studentClashes?: string[];
}

export interface ScheduleStats {
  /** The random seed that was used */
  seed: number;

  /** Total computation time in milliseconds */
  totalTime: number;

  /** Number of attempts made */
  attempts: number;

  /** Best penalty score achieved */
  bestPenalty: number;

  /** Number of time slots actually used */
  slotsUsed: number;
}

export interface SuccessResponse {
  success: true;

  /** CSV string with headers: course_id,slot_id,slot_datetime,halls,enrolled_count,notes */
  scheduleCSV: string;

  /** Validation report for the generated schedule */
  report: ValidationReport;

  /** Statistics about the scheduling process */
  stats: ScheduleStats;
}

export interface ErrorResponse {
  success: false;

  /** Error message describing what went wrong */
  error: string;

  /** Partial validation report if available */
  report?: ValidationReport;

  /** Partial statistics if available */
  stats?: ScheduleStats;
}

export type ScheduleResponse = SuccessResponse | ErrorResponse;

export interface VersionInfo {
  /** Name of the scheduler module */
  name: string;

  /** Version string */
  version: string;

  /** Go runtime version used to build */
  go: string;
}

// ===== WASM API FUNCTIONS =====

/**
 * Global functions exposed by the Go WASM module after loading.
 * These will be available on globalThis after the WASM module is initialized.
 */
export interface WasmAPI {
  /**
   * Get version information about the scheduler
   * @returns JSON string containing VersionInfo
   */
  version(): string;

  /**
   * Run the exam scheduling algorithm
   * @param regCSV - CSV string with registrations (student_id,course_id header required)
   * @param hallsCSV - CSV string with halls (hall,capacity,group header required)
   * @param paramsJSON - JSON string of RunScheduleParams
   * @returns JSON string containing ScheduleResponse
   */
  runSchedule(regCSV: string, hallsCSV: string, paramsJSON: string): string;

  /**
   * Verify an existing schedule for correctness
   * @param regCSV - CSV string with registrations
   * @param scheduleCSV - CSV string with schedule to verify
   * @returns JSON string containing ValidationReport wrapped in success response
   */
  verify(regCSV: string, scheduleCSV: string): string;
}

// ===== USAGE DOCUMENTATION =====

/**
 * HOW TO USE THE WASM MODULE:
 *
 * 1. Load the WASM module in a Web Worker:
 *    ```typescript
 *    // In worker thread:
 *    const wasmCode = await fetch('/wasm_exec.js').then(r => r.text());
 *    (0, eval)(wasmCode); // Load Go's WASM runtime
 *
 *    const go = new (globalThis as any).Go();
 *    const wasmModule = await WebAssembly.instantiateStreaming(
 *      fetch('/main.wasm'),
 *      go.importObject
 *    );
 *    go.run(wasmModule.instance);
 *
 *    // Now globalThis.version, globalThis.runSchedule, globalThis.verify are available
 *    ```
 *
 * 2. Call the API functions:
 *    ```typescript
 *    // Get version
 *    const versionJson = (globalThis as any).version();
 *    const version: VersionInfo = JSON.parse(versionJson);
 *
 *    // Run scheduling
 *    const params: RunScheduleParams = {
 *      examStartDate: "2025-01-20",
 *      examEndDate: "2025-01-24",
 *      slotsPerDay: 2,
 *      slotDuration: 180,
 *      holidays: [],
 *      tries: 100,
 *      seed: 0
 *    };
 *
 *    const resultJson = (globalThis as any).runSchedule(
 *      registrationsCSV,
 *      hallsCSV,
 *      JSON.stringify(params)
 *    );
 *    const result: ScheduleResponse = JSON.parse(resultJson);
 *
 *    if (result.success) {
 *      console.log('Schedule generated:', result.scheduleCSV);
 *      console.log('Stats:', result.stats);
 *    } else {
 *      console.error('Scheduling failed:', result.error);
 *    }
 *    ```
 *
 * 3. Custom Column Mapping (optional):
 *    ```typescript
 *    // For CSV files with different column names
 *    const customMapping: ColumnMapping = {
 *      studentIDColumn: 'student_number',
 *      courseIDColumn: 'subject_code',
 *      hallIDColumn: 'room_id',
 *      capacityColumn: 'max_students'
 *    };
 *
 *    const params: RunScheduleParams = {
 *      // ... other parameters
 *      columnMapping: customMapping
 *    };
 *    ```
 *
 * 4. CSV Format Requirements:
 *
 *    Registrations CSV:
 *    - Default headers: student_id, course_id
 *    - Can be customized via columnMapping parameter
 *    - Additional columns are ignored
 *    - Handles quoted fields with commas: "Doe, Jane",CS101
 *    - Comment lines starting with # are ignored
 *    - Empty lines are skipped
 *
 *    Halls CSV:
 *    - Default headers: hall, capacity
 *    - Optional header: group
 *    - Can be customized via columnMapping parameter
 *    - Capacity must be a non-negative integer
 *    - Handles quoted hall names: "Hall A, Wing 1",100,North
 *
 *    Allowed Slots CSV (optional):
 *    - Headers: course_id, slot_id
 *    - Restricts courses to specific time slots
 *    - If empty/omitted, all courses can use all slots
 *
 *    Output Schedule CSV:
 *    - Headers: course_id,slot_id,slot_datetime,halls,enrolled_count,notes
 *    - slot_datetime is in RFC3339 format
 *    - halls is semicolon-separated list of hall IDs
 *    - Fields with semicolons are automatically quoted by CSV writer
 *
 * 5. Error Handling:
 *    - All functions return JSON strings, never throw exceptions
 *    - Check the 'success' field in responses
 *    - Infeasible schedules return success=false with descriptive error
 *    - CSV parsing errors are reported in the error field
 *    - Validation reports contain detailed conflict information
 *
 * 6. Deterministic Results:
 *    - Provide the same 'seed' value to get identical results
 *    - The actual seed used is returned in stats.seed
 *    - Useful for debugging and reproducible testing
 */

// ===== EXAMPLE DATA =====

export const EXAMPLE_REGISTRATIONS_CSV = `student_id,course_id
s001,CS101
s001,MATH201
s002,CS101
s002,PHYS101
s003,MATH201
s003,PHYS101
"Smith, John",CS101
"Doe, Jane",MATH201`;

export const EXAMPLE_HALLS_CSV = `hall,capacity,group
H001,100,North
H002,80,North
H003,120,South`;

export const EXAMPLE_CUSTOM_REGISTRATIONS_CSV = `student_number,subject_code,department
s001,CS101,Engineering
s001,MATH201,Science
s002,CS101,Engineering
s002,PHYS101,Science`;

export const EXAMPLE_CUSTOM_HALLS_CSV = `room_id,max_students,building
R001,100,Building A
R002,80,Building B
R003,120,Building C`;

export const EXAMPLE_COLUMN_MAPPING: ColumnMapping = {
  studentIDColumn: "student_number",
  courseIDColumn: "subject_code",
  hallIDColumn: "room_id",
  capacityColumn: "max_students",
  groupColumn: "building"
};

export const EXAMPLE_PARAMS: RunScheduleParams = {
  examStartDate: "2025-01-20",
  examEndDate: "2025-01-24",
  slotsPerDay: 2,
  slotTimes: ["09:00", "14:00"],
  slotDuration: 180,
  holidays: ["2025-01-22"], // Skip Wednesday
  tries: 50,
  seed: 12345,
  minGap: 60,
  timezone: "UTC"
};

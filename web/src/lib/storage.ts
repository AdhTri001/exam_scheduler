/**
 * localStorage helpers with a versioned schema to prevent stale data issues.
 */

import type { ColumnMapping, RunScheduleParams } from './wasmTypes';

const STORAGE_KEY = 'exam_scheduler_session';
const SCHEMA_VERSION = 'v1';

export interface AppState {
  // Base64 encoded file content
  registrationsFile?: { name: string; content: string };
  hallsFile?: { name:string; content: string };
  allowedSlotsFile?: { name: string; content: string };

  // Column mappings for CSV files using proper WASM API types
  registrationsColumnMapping?: ColumnMapping;
  hallsColumnMapping?: ColumnMapping;

  // Parameters from the form
  params?: Partial<RunScheduleParams>;

  // Last results
  lastResult?: object;

  // Current step in the UI
  currentStep?: number;
}

interface StoredState {
  version: string;
  state: AppState;
}

/**
 * Saves the application state to localStorage.
 * @param state The state to save.
 */
export function saveState(state: AppState): void {
  try {
    const data: StoredState = {
      version: SCHEMA_VERSION,
      state,
    };
    localStorage.setItem(STORAGE_KEY, JSON.stringify(data));
  } catch (error) {
    console.error("Failed to save state to localStorage:", error);
  }
}

/**
 * Loads the application state from localStorage.
 * @returns The loaded state, or an empty object if not found or version mismatch.
 */
export function loadState(): AppState {
  try {
    const rawData = localStorage.getItem(STORAGE_KEY);
    if (!rawData) {
      return {};
    }

    const data: StoredState = JSON.parse(rawData);

    if (data.version !== SCHEMA_VERSION) {
      console.warn(`Storage schema mismatch. Expected ${SCHEMA_VERSION}, found ${data.version}. Discarding old state.`);
      localStorage.removeItem(STORAGE_KEY);
      return {};
    }

    return data.state || {};
  } catch (error) {
    console.error("Failed to load state from localStorage:", error);
    return {};
  }
}

/**
 * Clears all saved state from localStorage.
 */
export function clearState(): void {
  try {
    localStorage.removeItem(STORAGE_KEY);
  } catch (error) {
    console.error("Failed to clear state from localStorage:", error);
  }
}

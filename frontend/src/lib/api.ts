// Mirrors internal/plan and internal/db on the server. Plans are written by
// AI agents through MCP, so the field names match what the agent sends.

export interface Exercise {
  name: string;
  sets: number;
  /** e.g. "8", "8-10", "AMRAP", "45s" */
  reps: string;
  /** Reps are per side: log left and right separately. */
  per_side?: boolean;
  /** Omitted for bodyweight. */
  weight_kg?: number;
  /** e.g. "per dumbbell", "RPE 8". */
  load_note?: string;
  /** 0 = go straight to the next exercise in the block. */
  rest_seconds: number;
  cues?: string[];
}

/** Exercises done together (superset/circuit), or a single exercise. */
export interface Block {
  name: string;
  exercises: Exercise[];
}

export interface SavedWorkout {
  id: string;
  title: string;
  notes?: string;
  /** YYYY-MM-DD the session is planned for. */
  scheduled_date: string;
  blocks: Block[];
  last_performed?: string;
  created_at: string;
  updated_at: string;
}

/**
 * Exercise label shown in the app. Uses the block name's trailing letter, so
 * "Blok C" gives C1, C2 (matching the coach's plan even with gaps or a warm-up
 * block first); otherwise the block's position: A1, A2, B1, …
 */
export function exerciseLabel(block: Block, blockIdx: number, exIdx: number): string {
  const letter = block.name.match(/(?:^|\s)([A-Za-z])$/)?.[1].toUpperCase() ?? String.fromCharCode(65 + blockIdx);
  return `${letter}${exIdx + 1}`;
}

/** Human-readable load: "100 kg · per dumbbell", "Bodyweight". */
export function formatLoad(ex: Exercise): string {
  const parts = [];
  if (ex.weight_kg) parts.push(`${ex.weight_kg} kg`);
  if (ex.load_note) parts.push(ex.load_note);
  return parts.length ? parts.join(' · ') : 'Bodyweight';
}

// The REST API is read-only (plus marking a session performed).
// Plans are created, edited and deleted through the MCP endpoint.
const API_BASE = '/api';

export async function fetchWorkouts(): Promise<SavedWorkout[]> {
  const res = await fetch(`${API_BASE}/workouts`);
  if (!res.ok) throw new Error('Failed to fetch workouts');
  return res.json();
}

export async function touchLastPerformed(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/workouts/${id}/performed`, { method: 'POST' });
  if (!res.ok) throw new Error('Failed to update last performed');
}

export interface SavedWorkout {
  id: string;
  title: string;
  text: string;
  lastPerformed: string;
  /** YYYY-MM-DD the session is planned for, '' if unscheduled. Optional for older cached data. */
  scheduledDate?: string;
  createdAt: string;
  updatedAt: string;
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

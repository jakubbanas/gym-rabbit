export interface SavedWorkout {
  id: string;
  title: string;
  text: string;
  lastPerformed: string;
  createdAt: string;
  updatedAt: string;
}

const API_BASE = '/api';

export async function fetchWorkouts(): Promise<SavedWorkout[]> {
  const res = await fetch(`${API_BASE}/workouts`);
  if (!res.ok) throw new Error('Failed to fetch workouts');
  return res.json();
}

export async function createWorkout(title: string, text: string): Promise<SavedWorkout> {
  const res = await fetch(`${API_BASE}/workouts`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title, text }),
  });
  if (!res.ok) throw new Error('Failed to create workout');
  return res.json();
}

export async function updateWorkout(id: string, title: string, text: string): Promise<SavedWorkout> {
  const res = await fetch(`${API_BASE}/workouts/${id}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ title, text }),
  });
  if (!res.ok) throw new Error('Failed to update workout');
  return res.json();
}

export async function deleteWorkout(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/workouts/${id}`, { method: 'DELETE' });
  if (!res.ok) throw new Error('Failed to delete workout');
}

export async function touchLastPerformed(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/workouts/${id}/performed`, { method: 'POST' });
  if (!res.ok) throw new Error('Failed to update last performed');
}

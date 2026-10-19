import { useEffect, useState } from 'react';
import { Dumbbell, Plus, Pencil, Trash2, CalendarDays } from 'lucide-react';
import { fetchWorkouts, deleteWorkout, type SavedWorkout } from '../lib/api';

interface DashboardProps {
  onStartWorkout: (text: string, workoutId?: string) => void;
  onEditWorkout: (text: string, workoutId?: string) => void;
  onNewWorkout: () => void;
}

export default function Dashboard({ onStartWorkout, onEditWorkout, onNewWorkout }: DashboardProps) {
  const [workouts, setWorkouts] = useState<SavedWorkout[]>([]);
  const [loading, setLoading] = useState(true);

  const loadWorkouts = async () => {
    try {
      const data = await fetchWorkouts();
      setWorkouts(data);
    } catch (err) {
      console.error('Failed to load workouts:', err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadWorkouts();
  }, []);

  const handleDelete = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    if (!confirm('Delete this workout?')) return;
    try {
      await deleteWorkout(id);
      setWorkouts((prev) => prev.filter((w) => w.id !== id));
    } catch (err) {
      console.error('Failed to delete workout:', err);
    }
  };

  const formatDate = (dateStr: string) => {
    if (!dateStr) return 'Never';
    try {
      return new Date(dateStr).toLocaleDateString('en-GB', {
        day: 'numeric',
        month: 'short',
        year: 'numeric',
      });
    } catch {
      return dateStr;
    }
  };

  return (
    <div className="min-h-screen bg-neutral-950 text-neutral-200 p-6 font-sans flex flex-col max-w-md mx-auto">
      <div className="flex items-center gap-3 mb-8 pt-4">
        <div className="p-2 bg-indigo-500 rounded-xl text-white">
          <Dumbbell size={24} />
        </div>
        <h1 className="text-2xl font-black tracking-tight">Gym Rabbit</h1>
      </div>

      <div className="flex justify-between items-end mb-4">
        <h2 className="text-lg font-bold text-white">Saved Routines</h2>
      </div>

      <div className="space-y-3 mb-8 flex-1">
        {loading ? (
          <div className="text-neutral-500 text-center py-12">Loading...</div>
        ) : workouts.length === 0 ? (
          <div className="text-neutral-500 text-center py-12">
            <p className="text-4xl mb-3">🐰</p>
            <p>No workouts saved yet.</p>
            <p className="text-sm">Paste your first routine below!</p>
          </div>
        ) : (
          workouts.map((workout) => (
            <div
              key={workout.id}
              onClick={() => onStartWorkout(workout.text, workout.id)}
              className="w-full bg-neutral-900 border border-neutral-800 rounded-2xl p-4 text-left active:bg-neutral-800 transition-colors flex items-center justify-between group cursor-pointer"
            >
              <div className="min-w-0 flex-1">
                <h3 className="font-bold text-white mb-1 text-lg truncate">{workout.title}</h3>
                <p className="text-sm text-neutral-400 flex items-center gap-2 flex-wrap">
                  {workout.scheduledDate && (
                    <span className="inline-flex items-center gap-1 text-indigo-300 bg-indigo-500/10 border border-indigo-500/20 rounded-full px-2 py-0.5 text-xs font-semibold">
                      <CalendarDays size={12} aria-hidden="true" />
                      <span className="sr-only">Planned for </span>
                      {formatDate(workout.scheduledDate)}
                    </span>
                  )}
                  <span>Last: {formatDate(workout.lastPerformed)}</span>
                </p>
              </div>
              <div className="flex items-center gap-1 ml-3 shrink-0">
                <button
                  onClick={(e) => {
                    e.stopPropagation();
                    onEditWorkout(workout.text, workout.id);
                  }}
                  className="text-neutral-400 bg-neutral-800 active:bg-neutral-700 active:text-white p-3 rounded-full transition-colors"
                >
                  <Pencil size={18} />
                </button>
                <button
                  onClick={(e) => handleDelete(e, workout.id)}
                  className="text-neutral-500 bg-neutral-800 active:bg-red-900 active:text-red-400 p-3 rounded-full transition-colors"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
          ))
        )}
      </div>

      <button
        onClick={onNewWorkout}
        className="w-full bg-neutral-900 border-2 border-neutral-800 border-dashed text-neutral-300 font-bold text-lg py-5 rounded-2xl active:bg-neutral-800 transition-colors flex items-center justify-center gap-2"
      >
        <Plus size={20} /> Paste New Routine
      </button>
    </div>
  );
}

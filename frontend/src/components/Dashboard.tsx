import { useEffect, useState } from 'react';
import { Dumbbell, CalendarDays } from 'lucide-react';
import { fetchWorkouts, type SavedWorkout } from '../lib/api';

interface DashboardProps {
  onStartWorkout: (workout: SavedWorkout) => void;
}

/** Today as YYYY-MM-DD in local time, comparable with scheduled_date. */
function localToday(): string {
  const d = new Date();
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
}

/** "Mon 21 Oct" for a YYYY-MM-DD date, parsed as a local day (not UTC). */
function formatDay(date: string): string {
  const [y, m, d] = date.split('-').map(Number);
  return new Date(y, m - 1, d).toLocaleDateString('en-GB', { weekday: 'short', day: 'numeric', month: 'short' });
}

function formatTimestamp(ts?: string): string {
  if (!ts) return 'Never';
  return new Date(ts).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
}

function WorkoutCard({ workout, today, onStart }: { workout: SavedWorkout; today: string; onStart: () => void }) {
  const isToday = workout.scheduled_date === today;
  return (
    <button
      type="button"
      onClick={onStart}
      aria-label={`Start ${workout.title}`}
      className={`w-full bg-neutral-900 border rounded-2xl p-4 text-left active:bg-neutral-800 transition-colors cursor-pointer ${
        isToday ? 'border-indigo-500/60 shadow-[0_0_20px_rgba(99,102,241,0.15)]' : 'border-neutral-800'
      }`}
    >
      <p className="text-xs font-semibold mb-1 flex items-center gap-1 text-indigo-300">
        <CalendarDays size={12} aria-hidden="true" />
        {isToday ? 'Today' : formatDay(workout.scheduled_date)}
      </p>
      <h3 className="font-bold text-white mb-1 text-lg truncate">{workout.title}</h3>
      {workout.notes && <p className="text-sm text-neutral-300 mb-1 truncate">{workout.notes}</p>}
      <p className="text-sm text-neutral-400">Last: {formatTimestamp(workout.last_performed)}</p>
    </button>
  );
}

export default function Dashboard({ onStartWorkout }: DashboardProps) {
  const [workouts, setWorkouts] = useState<SavedWorkout[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchWorkouts()
      .then(setWorkouts)
      .catch((err) => console.error('Failed to load workouts:', err))
      .finally(() => setLoading(false));
  }, []);

  // The API returns workouts sorted by date: upcoming ones in order, past ones most recent first.
  const today = localToday();
  const upcoming = workouts.filter((w) => w.scheduled_date >= today);
  const past = workouts.filter((w) => w.scheduled_date < today).reverse();

  const section = (title: string, list: SavedWorkout[], dim = false) => (
    <section className="mb-8">
      <h2 className="text-lg font-bold text-white mb-4">{title}</h2>
      <div className={`space-y-3 ${dim ? 'opacity-60' : ''}`}>
        {list.map((w) => (
          <WorkoutCard key={w.id} workout={w} today={today} onStart={() => onStartWorkout(w)} />
        ))}
      </div>
    </section>
  );

  return (
    <div className="min-h-screen bg-neutral-950 text-neutral-200 p-6 font-sans flex flex-col max-w-md mx-auto">
      <div className="flex items-center gap-3 mb-8 pt-4">
        <div className="p-2 bg-indigo-500 rounded-xl text-white">
          <Dumbbell size={24} />
        </div>
        <h1 className="text-2xl font-black tracking-tight">Gym Rabbit</h1>
      </div>

      {loading ? (
        <div className="text-neutral-500 text-center py-12">Loading...</div>
      ) : workouts.length === 0 ? (
        <div className="text-neutral-500 text-center py-12">
          <p className="text-4xl mb-3">🐰</p>
          <p>No workouts planned yet.</p>
          <p className="text-sm">Ask Claude to add your plan.</p>
        </div>
      ) : (
        <>
          {upcoming.length > 0 ? (
            section('Upcoming', upcoming)
          ) : (
            <p className="text-neutral-500 text-center py-8 mb-4">Nothing planned ahead. Ask Claude for the next block.</p>
          )}
          {past.length > 0 && section('Past', past, true)}
        </>
      )}
    </div>
  );
}

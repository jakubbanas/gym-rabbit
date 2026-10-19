import { Check } from 'lucide-react';
import type { SavedWorkout } from '../lib/api';
import type { WorkoutProgress } from './WorkoutView';

interface SummaryProps {
  plan: SavedWorkout;
  progress: WorkoutProgress;
  onFinish: () => void;
}

export default function Summary({ plan, progress, onFinish }: SummaryProps) {
  let totalVolume = 0;

  plan.blocks.forEach((block, bIdx) => {
    block.exercises.forEach((ex, eIdx) => {
      const weightNum = ex.weight_kg ?? 0;
      const repsMatch = ex.reps.match(/(\d+)/);
      const repsNum = repsMatch ? parseInt(repsMatch[1]) : 0;

      if (weightNum > 0 && repsNum > 0) {
        progress[bIdx][eIdx].forEach((set) => {
          if (ex.per_side) {
            if (set.left) totalVolume += weightNum * repsNum;
            if (set.right) totalVolume += weightNum * repsNum;
          } else {
            if (set.done) totalVolume += weightNum * repsNum;
          }
        });
      }
    });
  });

  let funMessage = '';
  if (totalVolume > 0) {
    if (totalVolume < 500) funMessage = "That's equivalent to lifting a grand piano! 🎹";
    else if (totalVolume < 1500) funMessage = "That's equivalent to lifting an adult grizzly bear! 🐻";
    else if (totalVolume < 3000) funMessage = "That's equivalent to lifting a rhinoceros! 🦏";
    else if (totalVolume < 5000) funMessage = "That's equivalent to lifting an entire hippopotamus! 🦛";
    else if (totalVolume < 10000) funMessage = "That's equivalent to lifting an elephant! 🐘";
    else funMessage = "That's equivalent to lifting a commercial airplane! ✈️";
  }

  return (
    <div className="min-h-screen bg-neutral-950 text-white flex flex-col items-center justify-center p-6 text-center max-w-md mx-auto w-full">
      <div className="w-24 h-24 bg-green-500/20 text-green-500 rounded-full flex items-center justify-center mb-6">
        <Check size={48} strokeWidth={3} />
      </div>
      <h2 className="text-4xl font-black mb-2">Workout Complete</h2>
      <p className="text-neutral-400 mb-8 text-lg">{plan.title}</p>

      {totalVolume > 0 && (
        <div className="bg-neutral-900 border border-neutral-800 rounded-3xl p-6 mb-10 w-full max-w-xs shadow-xl">
          <p className="text-sm text-neutral-400 uppercase tracking-wider font-bold mb-1">
            Total Volume
          </p>
          <p className="text-4xl font-black text-indigo-400 mb-3">
            {totalVolume.toLocaleString()} kg
          </p>
          <p className="text-sm text-neutral-300 italic leading-snug">{funMessage}</p>
        </div>
      )}

      <button
        onClick={onFinish}
        className="w-full max-w-xs bg-neutral-800 text-white font-bold py-5 rounded-2xl active:bg-neutral-700 active:scale-95 transition-transform"
      >
        Finish & Return
      </button>
    </div>
  );
}

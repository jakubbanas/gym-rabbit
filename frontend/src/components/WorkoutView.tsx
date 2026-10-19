import { useRef } from 'react';
import { ArrowLeft, Clock, Check } from 'lucide-react';
import { exerciseLabel, formatLoad, type SavedWorkout, type Exercise } from '../lib/api';
import { triggerHaptic } from '../lib/audio';

// Set progress for a single set
export type SetProgress = { left?: boolean; right?: boolean; done?: boolean };

// Full progress structure: [blockIdx][exerciseIdx][setIdx]
export type WorkoutProgress = SetProgress[][][];

export function initProgress(plan: SavedWorkout): WorkoutProgress {
  return plan.blocks.map((b) =>
    b.exercises.map((ex) =>
      Array.from({ length: ex.sets }).map(() =>
        ex.per_side ? { left: false, right: false } : { done: false }
      )
    )
  );
}

interface WorkoutViewProps {
  plan: SavedWorkout;
  progress: WorkoutProgress;
  currentBlockIdx: number;
  onLogSet: (blockIdx: number, exIdx: number, setIdx: number, side: 'left' | 'right' | null) => void;
  onBlockChange: (idx: number) => void;
  onFinish: () => void;
  onBack: () => void;
}

export default function WorkoutView({
  plan,
  progress,
  currentBlockIdx,
  onLogSet,
  onBlockChange,
  onFinish,
  onBack,
}: WorkoutViewProps) {
  const touchStartX = useRef(0);

  const handleTouchStart = (e: React.TouchEvent) => {
    touchStartX.current = e.touches[0].clientX;
  };

  const handleTouchEnd = (e: React.TouchEvent) => {
    const touchEndX = e.changedTouches[0].clientX;
    const diff = touchStartX.current - touchEndX;
    if (diff > 50 && currentBlockIdx < plan.blocks.length - 1) {
      onBlockChange(currentBlockIdx + 1);
    } else if (diff < -50 && currentBlockIdx > 0) {
      onBlockChange(currentBlockIdx - 1);
    }
  };

  const block = plan.blocks[currentBlockIdx];
  const blockProgress = progress[currentBlockIdx];

  return (
    <div
      className="min-h-screen bg-neutral-950 text-neutral-100 flex flex-col font-sans select-none max-w-md mx-auto w-full"
      onTouchStart={handleTouchStart}
      onTouchEnd={handleTouchEnd}
    >
      {/* Header */}
      <header className="px-6 pt-10 pb-4 flex items-center justify-between sticky top-0 bg-neutral-950/90 backdrop-blur z-10 border-b border-neutral-900">
        <div className="flex items-start gap-3">
          <button
            onClick={onBack}
            className="p-1 -ml-1 mt-1 bg-transparent text-neutral-400 active:text-white transition-colors"
          >
            <ArrowLeft size={24} />
          </button>
          <div>
            <h2 className="text-xs font-bold uppercase tracking-widest text-indigo-400 mb-1">
              {plan.title}
            </h2>
            <h1 className="text-3xl font-black">{block.name}</h1>
          </div>
        </div>
        <div className="flex gap-1">
          {plan.blocks.map((_, idx) => (
            <div
              key={idx}
              className={`h-2 w-2 rounded-full ${idx === currentBlockIdx ? 'bg-indigo-500' : 'bg-neutral-800'}`}
            />
          ))}
        </div>
      </header>

      {/* Cards Scroll Area */}
      <main className="flex-1 p-6 pb-32 space-y-8 overflow-y-auto">
        {block.exercises.map((ex: Exercise, exIdx: number) => {
          const setsProgress = blockProgress[exIdx];

          // Find the current active set to log
          const activeSetIdx = setsProgress.findIndex((s) =>
            ex.per_side ? !s.left || !s.right : !s.done
          );
          const activeSet = setsProgress[activeSetIdx];
          const isComplete = activeSetIdx === -1;

          return (
            <div
              key={exIdx}
              className={`bg-neutral-900 border ${isComplete ? 'border-green-500/30' : 'border-neutral-800'} rounded-3xl p-5 shadow-xl transition-colors`}
            >
              {/* Exercise Header */}
              <div className="flex justify-between items-start mb-4">
                <div className="pr-4">
                  <h3 className="text-xl font-bold leading-tight mb-1">
                    <span className="text-neutral-500 mr-2">{exerciseLabel(currentBlockIdx, exIdx)}.</span>
                    {ex.name}
                  </h3>
                  <div className="flex items-center flex-wrap gap-x-3 gap-y-1 text-sm font-medium text-neutral-400">
                    <span className="text-white bg-neutral-800 px-2 py-0.5 rounded-md">
                      {ex.sets} × {ex.reps}
                      {ex.per_side && ' / side'}
                    </span>
                    <span>•</span>
                    <span className="text-indigo-300">{formatLoad(ex)}</span>
                  </div>
                </div>
                <div className="flex items-center text-neutral-500 text-xs bg-neutral-950 px-2 py-1 rounded-lg gap-1">
                  <Clock size={14} />
                  <span>{ex.rest_seconds > 0 ? `${ex.rest_seconds}s` : 'No rest'}</span>
                </div>
              </div>

              {/* Cues */}
              {ex.cues && ex.cues.length > 0 && !isComplete && (
                <div className="bg-indigo-950/30 rounded-xl p-3 mb-6">
                  {ex.cues.map((cue, i) => (
                    <p key={i} className="text-indigo-200/80 text-sm italic mb-1 last:mb-0">
                      • {cue}
                    </p>
                  ))}
                </div>
              )}

              {/* Set Progress Indicators */}
              <div className="flex gap-2 mb-6">
                {setsProgress.map((s, i) => {
                  const isSetDone = ex.per_side ? s.left && s.right : s.done;
                  const isCurrent = i === activeSetIdx;
                  return (
                    <div
                      key={i}
                      className={`h-2 flex-1 rounded-full transition-all ${
                        isSetDone
                          ? 'bg-green-500'
                          : isCurrent
                            ? 'bg-indigo-500 shadow-[0_0_10px_rgba(99,102,241,0.5)]'
                            : 'bg-neutral-800'
                      }`}
                    />
                  );
                })}
              </div>

              {/* Giant Interactive Bubble */}
              {!isComplete ? (
                ex.per_side ? (
                  <div className="grid grid-cols-2 gap-3 mt-4">
                    <button
                      disabled={activeSet?.left}
                      onClick={() => {
                        triggerHaptic();
                        onLogSet(currentBlockIdx, exIdx, activeSetIdx, 'left');
                      }}
                      className={`h-28 rounded-2xl flex flex-col items-center justify-center font-black text-2xl transition-all active:scale-95 ${
                        activeSet?.left
                          ? 'bg-neutral-800 text-neutral-600 opacity-50'
                          : 'bg-indigo-600 text-white shadow-lg'
                      }`}
                    >
                      <span className="text-sm font-medium opacity-70 mb-1 uppercase tracking-widest">
                        Set {activeSetIdx + 1}
                      </span>
                      LEFT
                    </button>
                    <button
                      disabled={activeSet?.right}
                      onClick={() => {
                        triggerHaptic();
                        onLogSet(currentBlockIdx, exIdx, activeSetIdx, 'right');
                      }}
                      className={`h-28 rounded-2xl flex flex-col items-center justify-center font-black text-2xl transition-all active:scale-95 ${
                        activeSet?.right
                          ? 'bg-neutral-800 text-neutral-600 opacity-50'
                          : 'bg-indigo-600 text-white shadow-lg'
                      }`}
                    >
                      <span className="text-sm font-medium opacity-70 mb-1 uppercase tracking-widest">
                        Set {activeSetIdx + 1}
                      </span>
                      RIGHT
                    </button>
                  </div>
                ) : (
                  <button
                    onClick={() => {
                      triggerHaptic();
                      onLogSet(currentBlockIdx, exIdx, activeSetIdx, null);
                    }}
                    className="w-full h-28 bg-indigo-600 hover:bg-indigo-500 rounded-2xl flex flex-col items-center justify-center font-black text-3xl text-white shadow-lg transition-all active:scale-95 mt-4"
                  >
                    <span className="text-sm font-medium opacity-70 mb-1 uppercase tracking-widest">
                      Set {activeSetIdx + 1}
                    </span>
                    LOG SET
                  </button>
                )
              ) : (
                <div className="w-full h-16 bg-green-500/10 border border-green-500/20 text-green-500 rounded-2xl flex items-center justify-center font-bold text-lg mt-4">
                  <Check size={24} />
                  <span className="ml-2">Completed</span>
                </div>
              )}
            </div>
          );
        })}
      </main>

      {/* Navigation Footer */}
      <footer className="fixed bottom-0 left-0 right-0 p-6 bg-gradient-to-t from-neutral-950 via-neutral-950 to-transparent pointer-events-none">
        <div className="flex justify-between items-center max-w-md mx-auto pointer-events-auto">
          <button
            onClick={() => currentBlockIdx > 0 && onBlockChange(currentBlockIdx - 1)}
            className={`px-6 py-4 rounded-xl font-bold bg-neutral-900 ${currentBlockIdx === 0 ? 'opacity-0 pointer-events-none' : 'active:bg-neutral-800'}`}
          >
            ← Prev
          </button>

          {currentBlockIdx === plan.blocks.length - 1 ? (
            <button
              onClick={onFinish}
              className="px-8 py-4 rounded-xl font-bold bg-green-600 text-white active:bg-green-700 shadow-lg"
            >
              Finish Session
            </button>
          ) : (
            <button
              onClick={() => onBlockChange(currentBlockIdx + 1)}
              className="px-6 py-4 rounded-xl font-bold bg-neutral-900 active:bg-neutral-800"
            >
              Next →
            </button>
          )}
        </div>
      </footer>
    </div>
  );
}

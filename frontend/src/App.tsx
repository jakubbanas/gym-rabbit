import { useState, useCallback } from 'react';
import Dashboard from './components/Dashboard';
import InputView from './components/InputView';
import WorkoutView, { initProgress, type WorkoutProgress } from './components/WorkoutView';
import RestTimer from './components/RestTimer';
import Summary from './components/Summary';
import ToolsDrawer from './components/ToolsDrawer';
import { parseWorkout, type WorkoutPlan } from './lib/parser';
import { touchLastPerformed } from './lib/api';

type AppScreen = 'dashboard' | 'input' | 'workout' | 'rest' | 'summary';

export default function App() {
  const [screen, setScreen] = useState<AppScreen>('dashboard');
  const [workoutPlan, setWorkoutPlan] = useState<WorkoutPlan | null>(null);
  const [progress, setProgress] = useState<WorkoutProgress>([]);
  const [currentBlockIdx, setCurrentBlockIdx] = useState(0);
  const [restTime, setRestTime] = useState(0);

  // Start a session. Saved plans are only marked as performed; pasted plans run
  // as one-off sessions, since plans are saved through MCP, not the web UI.
  // Throws on parse errors so InputView can show them.
  const startWorkout = (text: string, workoutId?: string) => {
    const parsed = parseWorkout(text);

    if (workoutId) {
      touchLastPerformed(workoutId).catch(() => {});
    }

    setWorkoutPlan(parsed);
    setProgress(initProgress(parsed));
    setCurrentBlockIdx(0);
    setScreen('workout');
  };

  const handleLogSet = useCallback(
    (blockIdx: number, exIdx: number, setIdx: number, side: 'left' | 'right' | null) => {
      if (!workoutPlan) return;

      setProgress((prev) => {
        const newProg = prev.map((b) => b.map((e) => e.map((s) => ({ ...s }))));
        if (side) {
          newProg[blockIdx][exIdx][setIdx][side] = true;
        } else {
          newProg[blockIdx][exIdx][setIdx].done = true;
        }
        return newProg;
      });

      // Start rest timer
      const exercise = workoutPlan.blocks[blockIdx].exercises[exIdx];
      setRestTime(exercise.rest);
      setScreen('rest');
    },
    [workoutPlan]
  );

  const handleRestTick = useCallback(() => {
    setRestTime((prev) => prev - 1);
  }, []);

  const handleSkipRest = useCallback(() => {
    setScreen('workout');
    setRestTime(0);
  }, []);

  const handleFinishWorkout = useCallback(() => {
    setScreen('summary');
  }, []);

  const handleReturnToDashboard = useCallback(() => {
    setScreen('dashboard');
    setWorkoutPlan(null);
    setProgress([]);
  }, []);

  return (
    <>
      {screen === 'dashboard' && (
        <Dashboard
          onStartWorkout={(text, workoutId) => {
            try {
              startWorkout(text, workoutId);
            } catch (err) {
              console.error('Failed to start workout:', err);
            }
          }}
          onQuickSession={() => setScreen('input')}
        />
      )}

      {screen === 'input' && (
        <InputView initialText="" onBack={handleReturnToDashboard} onStart={(text) => startWorkout(text)} />
      )}

      {screen === 'workout' && workoutPlan && (
        <WorkoutView
          plan={workoutPlan}
          progress={progress}
          currentBlockIdx={currentBlockIdx}
          onLogSet={handleLogSet}
          onBlockChange={setCurrentBlockIdx}
          onFinish={handleFinishWorkout}
          onBack={handleReturnToDashboard}
        />
      )}

      {screen === 'rest' && (
        <RestTimer seconds={restTime} onTick={handleRestTick} onSkip={handleSkipRest} />
      )}

      {screen === 'summary' && workoutPlan && (
        <Summary plan={workoutPlan} progress={progress} onFinish={handleReturnToDashboard} />
      )}

      {/* Tools overlay always visible */}
      <ToolsDrawer />
    </>
  );
}

import { useState, useCallback } from 'react';
import Dashboard from './components/Dashboard';
import WorkoutView, { initProgress, type WorkoutProgress } from './components/WorkoutView';
import RestTimer from './components/RestTimer';
import Summary from './components/Summary';
import ToolsDrawer from './components/ToolsDrawer';
import { touchLastPerformed, type SavedWorkout } from './lib/api';

type AppScreen = 'dashboard' | 'workout' | 'rest' | 'summary';

export default function App() {
  const [screen, setScreen] = useState<AppScreen>('dashboard');
  const [workoutPlan, setWorkoutPlan] = useState<SavedWorkout | null>(null);
  const [progress, setProgress] = useState<WorkoutProgress>([]);
  const [currentBlockIdx, setCurrentBlockIdx] = useState(0);
  const [restTime, setRestTime] = useState(0);

  const startWorkout = (workout: SavedWorkout) => {
    touchLastPerformed(workout.id).catch(() => {});

    setWorkoutPlan(workout);
    setProgress(initProgress(workout));
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

      // Start rest timer; 0 means go straight to the next exercise in a superset
      const rest = workoutPlan.blocks[blockIdx].exercises[exIdx].rest_seconds;
      if (rest > 0) {
        setRestTime(rest);
        setScreen('rest');
      }
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
        <Dashboard onStartWorkout={startWorkout} />
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

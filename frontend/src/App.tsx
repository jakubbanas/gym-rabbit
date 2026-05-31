import { useState, useCallback } from 'react';
import Dashboard from './components/Dashboard';
import InputView from './components/InputView';
import WorkoutView, { initProgress, type WorkoutProgress } from './components/WorkoutView';
import RestTimer from './components/RestTimer';
import Summary from './components/Summary';
import ToolsDrawer from './components/ToolsDrawer';
import { parseWorkout, type WorkoutPlan } from './lib/parser';
import { createWorkout, updateWorkout, touchLastPerformed } from './lib/api';

type AppScreen = 'dashboard' | 'input' | 'workout' | 'rest' | 'summary';

export default function App() {
  const [screen, setScreen] = useState<AppScreen>('dashboard');
  const [planText, setPlanText] = useState('');
  const [editingWorkoutId, setEditingWorkoutId] = useState<string | undefined>();
  const [workoutPlan, setWorkoutPlan] = useState<WorkoutPlan | null>(null);
  const [progress, setProgress] = useState<WorkoutProgress>([]);
  const [currentBlockIdx, setCurrentBlockIdx] = useState(0);
  const [restTime, setRestTime] = useState(0);

  // Extract plan name from text for saving
  const extractTitle = (text: string): string => {
    const match = text.match(/Plan Name:\s*(.+)/);
    return match ? match[1].trim() : 'Untitled Workout';
  };

  // Save or update workout to backend, then start
  const startWorkoutFromText = async (text: string, workoutId?: string) => {
    try {
      const parsed = parseWorkout(text);
      const initialProgress = initProgress(parsed);

      // Save/update the workout
      const title = extractTitle(text);
      try {
        if (workoutId) {
          await updateWorkout(workoutId, title, text);
        } else {
          const saved = await createWorkout(title, text);
          workoutId = saved.id;
        }
      } catch (err) {
        console.error('Failed to save workout, continuing anyway:', err);
      }

      // Mark as performed
      if (workoutId) {
        touchLastPerformed(workoutId).catch(() => {});
      }

      setWorkoutPlan(parsed);
      setProgress(initialProgress);
      setCurrentBlockIdx(0);
      setScreen('workout');
    } catch (err: any) {
      throw err; // Let InputView handle the error display
    }
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
    setPlanText('');
    setEditingWorkoutId(undefined);
  }, []);

  return (
    <>
      {screen === 'dashboard' && (
        <Dashboard
          onStartWorkout={(text, workoutId) => {
            startWorkoutFromText(text, workoutId).catch((err) => {
              console.error('Failed to start workout:', err);
            });
          }}
          onEditWorkout={(text, workoutId) => {
            setPlanText(text);
            setEditingWorkoutId(workoutId);
            setScreen('input');
          }}
          onNewWorkout={() => {
            setPlanText('');
            setEditingWorkoutId(undefined);
            setScreen('input');
          }}
        />
      )}

      {screen === 'input' && (
        <InputView
          initialText={planText}
          onBack={handleReturnToDashboard}
          onStart={(text) => {
            startWorkoutFromText(text, editingWorkoutId).catch(() => {});
          }}
        />
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

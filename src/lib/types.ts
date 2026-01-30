export interface Exercise {
    id: string;
    name: string;
    sets: number;
    reps: string;
    weight?: string;
    rest?: number; // Rest time in seconds
    notes?: string;
    isUnilateral?: boolean; // For exercises that track L/R separately
}

export interface TrainingBlock {
    block: string; // A, B, C, D
    exercises: Exercise[];
}

export interface WorkoutPlan {
    id: string;
    name: string;
    startDate: string; // Format: DD.MM
    blocks: TrainingBlock[];
}

export interface ExerciseProgress {
    [exerciseId: string]: {
        [setIndex: number]: {
            left?: boolean;
            right?: boolean;
            completed?: boolean; // For non-unilateral exercises
        };
    };
}

export interface AppState {
    currentPlanId: string;
    progress: {
        [planId: string]: ExerciseProgress;
    };
}

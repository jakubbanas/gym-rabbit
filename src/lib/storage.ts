import { browser } from '$app/environment';
import type { AppState, ExerciseProgress } from './types';

const STORAGE_KEY = 'gym-rabbit-state';

const defaultState: AppState = {
    currentPlanId: '',
    progress: {}
};

export function loadState(): AppState {
    if (!browser) return defaultState;

    try {
        const stored = localStorage.getItem(STORAGE_KEY);
        if (stored) {
            return JSON.parse(stored);
        }
    } catch (error) {
        console.error('Failed to load state:', error);
    }

    return defaultState;
}

export function saveState(state: AppState): void {
    if (!browser) return;

    try {
        localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    } catch (error) {
        console.error('Failed to save state:', error);
    }
}

export function updateProgress(
    planId: string,
    exerciseId: string,
    setIndex: number,
    type: 'left' | 'right' | 'completed',
    value: boolean
): AppState {
    const state = loadState();

    if (!state.progress[planId]) {
        state.progress[planId] = {};
    }

    if (!state.progress[planId][exerciseId]) {
        state.progress[planId][exerciseId] = {};
    }

    if (!state.progress[planId][exerciseId][setIndex]) {
        state.progress[planId][exerciseId][setIndex] = {};
    }

    state.progress[planId][exerciseId][setIndex][type] = value;

    saveState(state);
    return state;
}

export function setCurrentPlan(planId: string): AppState {
    const state = loadState();
    state.currentPlanId = planId;
    saveState(state);
    return state;
}

export function getExerciseProgress(
    planId: string,
    exerciseId: string
): ExerciseProgress[string] | undefined {
    const state = loadState();
    return state.progress[planId]?.[exerciseId];
}

export function resetProgress(): AppState {
    const state = loadState();
    state.progress = {};
    saveState(state);
    return state;
}

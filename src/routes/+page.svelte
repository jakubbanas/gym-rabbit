<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { workoutPlans } from "$lib/data";
    import { loadState, updateProgress, setCurrentPlan } from "$lib/storage";
    import type { AppState, WorkoutPlan, Exercise } from "$lib/types";

    let state = $state<AppState>({ currentPlanId: "", progress: {} });
    let currentPlan = $derived<WorkoutPlan | undefined>(
        workoutPlans.find((p) => p.id === state.currentPlanId),
    );

    // Rest timer state
    let showTimer = $state(false);
    let timerSeconds = $state(0);
    let initialTimerSeconds = $state(0);
    let timerInterval: number | null = null;

    onMount(() => {
        state = loadState();

        // Set first plan as default if none selected
        if (!state.currentPlanId && workoutPlans.length > 0) {
            state = setCurrentPlan(workoutPlans[0].id);
        }
    });

    onDestroy(() => {
        if (timerInterval) {
            clearInterval(timerInterval);
        }
    });

    function startRestTimer(seconds: number) {
        if (seconds === 0) return;

        timerSeconds = seconds;
        initialTimerSeconds = seconds;
        showTimer = true;

        if (timerInterval) {
            clearInterval(timerInterval);
        }

        timerInterval = setInterval(() => {
            timerSeconds--;

            if (timerSeconds <= 0) {
                stopTimer();
            }
        }, 1000) as unknown as number;
    }

    function stopTimer() {
        showTimer = false;
        timerSeconds = 0;
        initialTimerSeconds = 0;

        if (timerInterval) {
            clearInterval(timerInterval);
            timerInterval = null;
        }
    }

    function addTime() {
        timerSeconds += 30;
    }

    function formatTime(seconds: number): string {
        const mins = Math.floor(seconds / 60);
        const secs = seconds % 60;
        return `${mins}:${secs.toString().padStart(2, "0")}`;
    }

    function handlePlanChange(planId: string) {
        state = setCurrentPlan(planId);
    }

    function toggleSet(
        exerciseId: string,
        setIndex: number,
        type: "left" | "right" | "completed",
        exercise: Exercise,
    ) {
        if (!currentPlan) return;

        const currentValue =
            state.progress[currentPlan.id]?.[exerciseId]?.[setIndex]?.[type] ??
            false;

        const newValue = !currentValue;
        state = updateProgress(
            currentPlan.id,
            exerciseId,
            setIndex,
            type,
            newValue,
        );

        // Start rest timer when checking a set (not unchecking)
        if (newValue) {
            const restSeconds = exercise.rest ?? 0;
            startRestTimer(restSeconds);
        }
    }

    function isSetChecked(
        exerciseId: string,
        setIndex: number,
        type: "left" | "right" | "completed",
    ): boolean {
        if (!currentPlan) return false;
        return (
            state.progress[currentPlan.id]?.[exerciseId]?.[setIndex]?.[type] ??
            false
        );
    }

    function getCompletedSets(exercise: Exercise): number {
        if (!currentPlan) return 0;

        const exerciseProgress =
            state.progress[currentPlan.id]?.[exercise.id] ?? {};
        let completed = 0;

        for (let i = 0; i < exercise.sets; i++) {
            const setProgress = exerciseProgress[i];
            if (!setProgress) continue;

            if (exercise.isUnilateral) {
                if (setProgress.left && setProgress.right) completed++;
            } else {
                if (setProgress.completed) completed++;
            }
        }

        return completed;
    }
</script>

<div class="container">
    {#if workoutPlans.length > 0}
        <div class="plan-selector">
            <label for="plan-select">Workout Plan:</label>
            <select
                id="plan-select"
                value={state.currentPlanId}
                onchange={(e) => handlePlanChange(e.currentTarget.value)}
            >
                {#each workoutPlans as plan}
                    <option value={plan.id}>
                        {plan.name} - Started {plan.startDate}
                    </option>
                {/each}
            </select>
        </div>
    {/if}

    {#if currentPlan}
        <div class="workout">
            {#each currentPlan.blocks as block}
                <div class="block">
                    <h2 class="block-header">Block {block.block}</h2>

                    {#each block.exercises as exercise}
                        {@const completed = getCompletedSets(exercise)}
                        <div class="exercise">
                            <div class="exercise-header">
                                <h3 class="exercise-name">{exercise.name}</h3>
                                <span class="progress-badge">
                                    {completed}/{exercise.sets}
                                </span>
                            </div>

                            <div class="exercise-details">
                                <span
                                    ><strong>Sets:</strong>
                                    {exercise.sets}</span
                                >
                                <span
                                    ><strong>Reps:</strong>
                                    {exercise.reps}</span
                                >
                                {#if exercise.weight}
                                    <span
                                        ><strong>Weight:</strong>
                                        {exercise.weight}</span
                                    >
                                {/if}
                                {#if exercise.rest}
                                    <span
                                        ><strong>Rest:</strong>
                                        {exercise.rest}</span
                                    >
                                {/if}
                            </div>

                            {#if exercise.notes}
                                <p class="notes">💡 {exercise.notes}</p>
                            {/if}

                            <div class="sets-tracker">
                                {#each Array(exercise.sets) as _, setIndex}
                                    {#if exercise.isUnilateral}
                                        <div class="checkbox-group">
                                            <button
                                                class="custom-checkbox"
                                                class:checked={isSetChecked(
                                                    exercise.id,
                                                    setIndex,
                                                    "left",
                                                )}
                                                onclick={() =>
                                                    toggleSet(
                                                        exercise.id,
                                                        setIndex,
                                                        "left",
                                                        exercise,
                                                    )}
                                            >
                                                <span class="check-label"
                                                    >L</span
                                                >
                                            </button>
                                            <button
                                                class="custom-checkbox"
                                                class:checked={isSetChecked(
                                                    exercise.id,
                                                    setIndex,
                                                    "right",
                                                )}
                                                onclick={() =>
                                                    toggleSet(
                                                        exercise.id,
                                                        setIndex,
                                                        "right",
                                                        exercise,
                                                    )}
                                            >
                                                <span class="check-label"
                                                    >R</span
                                                >
                                            </button>
                                        </div>
                                    {:else}
                                        <button
                                            class="custom-checkbox"
                                            class:checked={isSetChecked(
                                                exercise.id,
                                                setIndex,
                                                "completed",
                                            )}
                                            onclick={() =>
                                                toggleSet(
                                                    exercise.id,
                                                    setIndex,
                                                    "completed",
                                                    exercise,
                                                )}
                                        >
                                            <span class="check-icon">✓</span>
                                        </button>
                                    {/if}
                                {/each}
                            </div>
                        </div>
                    {/each}
                </div>
            {/each}
        </div>
    {:else}
        <p class="empty-state">No workout plans available</p>
    {/if}
</div>

{#if showTimer}
    <div class="timer-overlay" onclick={stopTimer}>
        <div class="timer-content" onclick={(e) => e.stopPropagation()}>
            <div class="timer-display">
                <div class="timer-text">{formatTime(timerSeconds)}</div>
                <div class="timer-label">Rest Time</div>
            </div>

            <div class="timer-progress">
                <div
                    class="timer-progress-bar"
                    style="width: {initialTimerSeconds > 0
                        ? (timerSeconds / initialTimerSeconds) * 100
                        : 0}%"
                ></div>
            </div>

            <div class="timer-actions">
                <button class="timer-btn timer-btn-secondary" onclick={addTime}>
                    +30s
                </button>
                <button class="timer-btn timer-btn-primary" onclick={stopTimer}>
                    Skip Rest
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
    :global(:root) {
        /* Light mode colors */
        --bg-gradient-start: #667eea;
        --bg-gradient-end: #764ba2;
        --card-bg: #ffffff;
        --card-shadow: rgba(0, 0, 0, 0.1);
        --text-primary: #333333;
        --text-secondary: #555555;
        --text-tertiary: #666666;
        --text-header: #ffffff;
        --border-color: #e0e0e0;
        --exercise-bg: #f8f9fa;
        --notes-bg: #ffffff;
        --notes-border: #fbbf24;
        --notes-text: #666666;
        --accent-primary: #667eea;
        --accent-secondary: #764ba2;
    }

    @media (prefers-color-scheme: dark) {
        :global(:root) {
            /* Dark mode colors */
            --bg-gradient-start: #1a1a2e;
            --bg-gradient-end: #16213e;
            --card-bg: #2d2d44;
            --card-shadow: rgba(0, 0, 0, 0.3);
            --text-primary: #e4e4e7;
            --text-secondary: #d4d4d8;
            --text-tertiary: #a1a1aa;
            --text-header: #ffffff;
            --border-color: #3f3f46;
            --exercise-bg: #1f1f2e;
            --notes-bg: #27273a;
            --notes-border: #fbbf24;
            --notes-text: #d4d4d8;
            --accent-primary: #818cf8;
            --accent-secondary: #a78bfa;
        }
    }

    :global(body) {
        margin: 0;
        padding: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
            "Helvetica Neue", Arial, sans-serif;
        background: linear-gradient(
            135deg,
            var(--bg-gradient-start) 0%,
            var(--bg-gradient-end) 100%
        );
        min-height: 100vh;
    }

    .container {
        max-width: 800px;
        margin: 0 auto;
        padding: 20px;
    }

    header {
        text-align: center;
        color: var(--text-header);
        margin-bottom: 30px;
        padding: 20px 0;
    }

    h1 {
        font-size: 2.5rem;
        margin: 0 0 10px 0;
        font-weight: 800;
    }

    .subtitle {
        margin: 0;
        font-size: 1.1rem;
        opacity: 0.9;
    }

    .plan-selector {
        background: var(--card-bg);
        padding: 24px;
        border-radius: 16px;
        margin-bottom: 24px;
        box-shadow: 0 4px 6px var(--card-shadow);
        display: flex;
        align-items: center;
        gap: 20px;
    }

    .plan-selector label {
        font-weight: 600;
        font-size: 1.15rem;
        color: var(--text-primary);
    }

    select {
        flex: 1;
        padding: 16px 20px;
        border: 2px solid var(--border-color);
        border-radius: 12px;
        font-size: 1.15rem;
        background: var(--card-bg);
        color: var(--text-primary);
        cursor: pointer;
        transition: border-color 0.2s;
        min-height: 54px;
    }

    select:hover {
        border-color: var(--accent-primary);
    }

    select:focus {
        outline: none;
        border-color: var(--accent-primary);
        box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
    }

    .workout {
        display: flex;
        flex-direction: column;
        gap: 25px;
    }

    .block {
        background: var(--card-bg);
        border-radius: 16px;
        padding: 24px;
        box-shadow: 0 4px 6px var(--card-shadow);
    }

    .block-header {
        margin: 0 0 24px 0;
        padding-bottom: 14px;
        border-bottom: 3px solid var(--accent-primary);
        font-size: 1.6rem;
        color: var(--accent-primary);
        font-weight: 700;
    }

    .exercise {
        padding: 20px;
        margin-bottom: 20px;
        background: var(--exercise-bg);
        border-radius: 12px;
        border-left: 5px solid var(--accent-secondary);
    }

    .exercise:last-child {
        margin-bottom: 0;
    }

    .exercise-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 16px;
    }

    .exercise-name {
        margin: 0;
        font-size: 1.35rem;
        color: var(--text-primary);
        font-weight: 600;
        line-height: 1.3;
    }

    .progress-badge {
        background: var(--accent-primary);
        color: white;
        padding: 8px 16px;
        border-radius: 24px;
        font-size: 1.05rem;
        font-weight: 700;
        min-width: 60px;
        text-align: center;
    }

    .exercise-details {
        display: flex;
        flex-wrap: wrap;
        gap: 16px;
        margin-bottom: 14px;
        font-size: 1.05rem;
        color: var(--text-secondary);
    }

    .notes {
        background: var(--notes-bg);
        padding: 14px 16px;
        border-radius: 10px;
        margin: 14px 0;
        font-size: 1.05rem;
        color: var(--notes-text);
        border-left: 4px solid var(--notes-border);
        line-height: 1.5;
    }

    .sets-tracker {
        display: flex;
        flex-wrap: wrap;
        gap: 16px;
        margin-top: 18px;
        align-items: center;
    }

    .checkbox-group {
        display: flex;
        gap: 6px;
        padding: 8px;
        background: var(--exercise-bg);
        border-radius: 20px;
        border: 2px solid var(--border-color);
    }

    .custom-checkbox {
        display: flex;
        flex-direction: column;
        align-items: center;
        justify-content: center;
        gap: 8px;
        cursor: pointer;
        user-select: none;
        padding: 16px 20px;
        background: var(--card-bg);
        border-radius: 16px;
        border: 3px solid transparent;
        transition: all 0.2s;
        min-width: 85px;
        min-height: 95px;
        position: relative;
        font-family: inherit;
    }

    .custom-checkbox:hover {
        border-color: var(--accent-primary);
        transform: translateY(-3px);
        box-shadow: 0 6px 20px rgba(102, 126, 234, 0.3);
    }

    .custom-checkbox:active {
        transform: scale(0.95);
    }

    .custom-checkbox .check-icon {
        font-size: 2.5rem;
        font-weight: 900;
        color: var(--border-color);
        transition: all 0.2s;
        line-height: 1;
    }

    .custom-checkbox.checked .check-icon {
        color: var(--accent-primary);
        transform: scale(1.1);
    }

    .custom-checkbox.checked {
        background: var(--accent-primary);
        border-color: var(--accent-primary);
        box-shadow: 0 4px 15px rgba(102, 126, 234, 0.4);
    }

    .custom-checkbox.checked .check-icon {
        color: white;
    }

    .custom-checkbox.checked .check-label {
        color: white;
    }

    .custom-checkbox .check-label {
        font-size: 2.5rem;
        font-weight: 900;
        color: var(--border-color);
        text-transform: uppercase;
        transition: all 0.2s;
        line-height: 1;
    }

    .custom-checkbox.checked .check-label {
        color: white;
        transform: scale(1.1);
    }

    .empty-state {
        text-align: center;
        color: var(--text-header);
        font-size: 1.2rem;
        padding: 40px;
    }

    /* Rest Timer Styles */
    .timer-overlay {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.95);
        display: flex;
        align-items: center;
        justify-content: center;
        z-index: 1000;
        backdrop-filter: blur(10px);
        animation: fadeIn 0.3s ease-out;
    }

    @keyframes fadeIn {
        from {
            opacity: 0;
        }
        to {
            opacity: 1;
        }
    }

    .timer-content {
        display: flex;
        flex-direction: column;
        align-items: center;
        gap: 40px;
        padding: 40px;
        max-width: 500px;
        width: 90%;
    }

    .timer-display {
        text-align: center;
    }

    .timer-text {
        font-size: 6rem;
        font-weight: 800;
        color: #ffffff;
        line-height: 1;
        margin-bottom: 10px;
        font-variant-numeric: tabular-nums;
        text-shadow: 0 0 30px rgba(102, 126, 234, 0.5);
    }

    .timer-label {
        font-size: 1.5rem;
        color: rgba(255, 255, 255, 0.7);
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 2px;
    }

    .timer-progress {
        width: 100%;
        height: 8px;
        background: rgba(255, 255, 255, 0.2);
        border-radius: 4px;
        overflow: hidden;
    }

    .timer-progress-bar {
        height: 100%;
        background: linear-gradient(90deg, #667eea, #764ba2);
        transition: width 1s linear;
        border-radius: 4px;
    }

    .timer-actions {
        display: flex;
        gap: 20px;
        width: 100%;
    }

    .timer-btn {
        flex: 1;
        padding: 20px 36px;
        font-size: 1.25rem;
        font-weight: 700;
        border: none;
        border-radius: 16px;
        cursor: pointer;
        transition: all 0.2s;
        text-transform: uppercase;
        letter-spacing: 1px;
        min-height: 68px;
    }

    .timer-btn-primary {
        background: linear-gradient(135deg, #667eea, #764ba2);
        color: white;
    }

    .timer-btn-primary:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 20px rgba(102, 126, 234, 0.4);
    }

    .timer-btn-secondary {
        background: rgba(255, 255, 255, 0.1);
        color: white;
        border: 2px solid rgba(255, 255, 255, 0.3);
    }

    .timer-btn-secondary:hover {
        background: rgba(255, 255, 255, 0.2);
        border-color: rgba(255, 255, 255, 0.5);
    }

    .timer-btn:active {
        transform: scale(0.98);
    }

    @media (max-width: 640px) {
        .container {
            padding: 15px;
        }

        h1 {
            font-size: 2rem;
        }

        .plan-selector {
            flex-direction: column;
            align-items: stretch;
        }

        .exercise-details {
            flex-direction: column;
            gap: 8px;
        }

        .sets-tracker {
            flex-direction: column;
        }

        .set-item {
            justify-content: space-between;
        }
    }
</style>

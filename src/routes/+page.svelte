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

    function parseRestTime(rest: string | undefined): number {
        if (!rest) return 0;

        const match = rest.match(/(\d+)(s|min)/);
        if (!match) return 0;

        const value = parseInt(match[1]);
        const unit = match[2];

        return unit === "min" ? value * 60 : value;
    }

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
            // For unilateral exercises, only start timer when both sides are checked
            if (exercise.isUnilateral) {
                const otherSide = type === "left" ? "right" : "left";
                const otherSideChecked =
                    state.progress[currentPlan.id]?.[exerciseId]?.[setIndex]?.[
                        otherSide
                    ] ?? false;

                if (otherSideChecked) {
                    const restSeconds = parseRestTime(exercise.rest);
                    startRestTimer(restSeconds);
                }
            } else {
                // For regular exercises, start timer immediately
                const restSeconds = parseRestTime(exercise.rest);
                startRestTimer(restSeconds);
            }
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
                                    <div class="set-item">
                                        <span class="set-label"
                                            >Set {setIndex + 1}</span
                                        >
                                        {#if exercise.isUnilateral}
                                            <div class="checkbox-group">
                                                <label class="checkbox-label">
                                                    <input
                                                        type="checkbox"
                                                        checked={isSetChecked(
                                                            exercise.id,
                                                            setIndex,
                                                            "left",
                                                        )}
                                                        onchange={() =>
                                                            toggleSet(
                                                                exercise.id,
                                                                setIndex,
                                                                "left",
                                                                exercise,
                                                            )}
                                                    />
                                                    <span>L</span>
                                                </label>
                                                <label class="checkbox-label">
                                                    <input
                                                        type="checkbox"
                                                        checked={isSetChecked(
                                                            exercise.id,
                                                            setIndex,
                                                            "right",
                                                        )}
                                                        onchange={() =>
                                                            toggleSet(
                                                                exercise.id,
                                                                setIndex,
                                                                "right",
                                                                exercise,
                                                            )}
                                                    />
                                                    <span>R</span>
                                                </label>
                                            </div>
                                        {:else}
                                            <label class="checkbox-label">
                                                <input
                                                    type="checkbox"
                                                    checked={isSetChecked(
                                                        exercise.id,
                                                        setIndex,
                                                        "completed",
                                                    )}
                                                    onchange={() =>
                                                        toggleSet(
                                                            exercise.id,
                                                            setIndex,
                                                            "completed",
                                                            exercise,
                                                        )}
                                                />
                                                <span>✓</span>
                                            </label>
                                        {/if}
                                    </div>
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
        padding: 20px;
        border-radius: 12px;
        margin-bottom: 20px;
        box-shadow: 0 4px 6px var(--card-shadow);
        display: flex;
        align-items: center;
        gap: 15px;
    }

    .plan-selector label {
        font-weight: 600;
        font-size: 1rem;
        color: var(--text-primary);
    }

    select {
        flex: 1;
        padding: 10px 15px;
        border: 2px solid var(--border-color);
        border-radius: 8px;
        font-size: 1rem;
        background: var(--card-bg);
        color: var(--text-primary);
        cursor: pointer;
        transition: border-color 0.2s;
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
        border-radius: 12px;
        padding: 20px;
        box-shadow: 0 4px 6px var(--card-shadow);
    }

    .block-header {
        margin: 0 0 20px 0;
        padding-bottom: 10px;
        border-bottom: 3px solid var(--accent-primary);
        font-size: 1.5rem;
        color: var(--accent-primary);
        font-weight: 700;
    }

    .exercise {
        padding: 15px;
        margin-bottom: 20px;
        background: var(--exercise-bg);
        border-radius: 8px;
        border-left: 4px solid var(--accent-secondary);
    }

    .exercise:last-child {
        margin-bottom: 0;
    }

    .exercise-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 12px;
    }

    .exercise-name {
        margin: 0;
        font-size: 1.2rem;
        color: var(--text-primary);
        font-weight: 600;
    }

    .progress-badge {
        background: var(--accent-primary);
        color: white;
        padding: 4px 12px;
        border-radius: 20px;
        font-size: 0.9rem;
        font-weight: 600;
    }

    .exercise-details {
        display: flex;
        flex-wrap: wrap;
        gap: 15px;
        margin-bottom: 10px;
        font-size: 0.95rem;
        color: var(--text-secondary);
    }

    .notes {
        background: var(--notes-bg);
        padding: 10px;
        border-radius: 6px;
        margin: 10px 0;
        font-size: 0.9rem;
        color: var(--notes-text);
        border-left: 3px solid var(--notes-border);
    }

    .sets-tracker {
        display: flex;
        flex-wrap: wrap;
        gap: 10px;
        margin-top: 15px;
    }

    .set-item {
        display: flex;
        align-items: center;
        gap: 8px;
        background: var(--card-bg);
        padding: 8px 12px;
        border-radius: 6px;
        border: 2px solid var(--border-color);
    }

    .set-label {
        font-size: 0.9rem;
        font-weight: 600;
        color: var(--text-tertiary);
    }

    .checkbox-group {
        display: flex;
        gap: 8px;
    }

    .checkbox-label {
        display: flex;
        align-items: center;
        gap: 4px;
        cursor: pointer;
        user-select: none;
    }

    .checkbox-label input[type="checkbox"] {
        width: 20px;
        height: 20px;
        cursor: pointer;
        accent-color: var(--accent-primary);
    }

    .checkbox-label span {
        font-size: 0.9rem;
        font-weight: 600;
        color: var(--text-primary);
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
        padding: 16px 32px;
        font-size: 1.1rem;
        font-weight: 600;
        border: none;
        border-radius: 12px;
        cursor: pointer;
        transition: all 0.2s;
        text-transform: uppercase;
        letter-spacing: 1px;
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

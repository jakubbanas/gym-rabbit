<script lang="ts">
    import { onMount } from "svelte";
    import { workoutPlans } from "$lib/data";
    import { loadState, updateProgress, setCurrentPlan } from "$lib/storage";
    import type { AppState, WorkoutPlan, Exercise } from "$lib/types";

    let state = $state<AppState>({ currentPlanId: "", progress: {} });
    let currentPlan = $derived<WorkoutPlan | undefined>(
        workoutPlans.find((p) => p.id === state.currentPlanId),
    );

    onMount(() => {
        state = loadState();

        // Set first plan as default if none selected
        if (!state.currentPlanId && workoutPlans.length > 0) {
            state = setCurrentPlan(workoutPlans[0].id);
        }
    });

    function handlePlanChange(planId: string) {
        state = setCurrentPlan(planId);
    }

    function toggleSet(
        exerciseId: string,
        setIndex: number,
        type: "left" | "right" | "completed",
    ) {
        if (!currentPlan) return;

        const currentValue =
            state.progress[currentPlan.id]?.[exerciseId]?.[setIndex]?.[type] ??
            false;
        state = updateProgress(
            currentPlan.id,
            exerciseId,
            setIndex,
            type,
            !currentValue,
        );
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
    <header>
        <h1>🐰 Gym Rabbit</h1>
        <p class="subtitle">Track your workout progress</p>
    </header>

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

<style>
    :global(body) {
        margin: 0;
        padding: 0;
        font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
            "Helvetica Neue", Arial, sans-serif;
        background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
        min-height: 100vh;
    }

    .container {
        max-width: 800px;
        margin: 0 auto;
        padding: 20px;
    }

    header {
        text-align: center;
        color: white;
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
        background: white;
        padding: 20px;
        border-radius: 12px;
        margin-bottom: 20px;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
        display: flex;
        align-items: center;
        gap: 15px;
    }

    .plan-selector label {
        font-weight: 600;
        font-size: 1rem;
        color: #333;
    }

    select {
        flex: 1;
        padding: 10px 15px;
        border: 2px solid #e0e0e0;
        border-radius: 8px;
        font-size: 1rem;
        background: white;
        cursor: pointer;
        transition: border-color 0.2s;
    }

    select:hover {
        border-color: #667eea;
    }

    select:focus {
        outline: none;
        border-color: #667eea;
        box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
    }

    .workout {
        display: flex;
        flex-direction: column;
        gap: 25px;
    }

    .block {
        background: white;
        border-radius: 12px;
        padding: 20px;
        box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    }

    .block-header {
        margin: 0 0 20px 0;
        padding-bottom: 10px;
        border-bottom: 3px solid #667eea;
        font-size: 1.5rem;
        color: #667eea;
        font-weight: 700;
    }

    .exercise {
        padding: 15px;
        margin-bottom: 20px;
        background: #f8f9fa;
        border-radius: 8px;
        border-left: 4px solid #764ba2;
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
        color: #333;
        font-weight: 600;
    }

    .progress-badge {
        background: #667eea;
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
        color: #555;
    }

    .notes {
        background: white;
        padding: 10px;
        border-radius: 6px;
        margin: 10px 0;
        font-size: 0.9rem;
        color: #666;
        border-left: 3px solid #fbbf24;
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
        background: white;
        padding: 8px 12px;
        border-radius: 6px;
        border: 2px solid #e0e0e0;
    }

    .set-label {
        font-size: 0.9rem;
        font-weight: 600;
        color: #666;
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
        accent-color: #667eea;
    }

    .checkbox-label span {
        font-size: 0.9rem;
        font-weight: 600;
        color: #333;
    }

    .empty-state {
        text-align: center;
        color: white;
        font-size: 1.2rem;
        padding: 40px;
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

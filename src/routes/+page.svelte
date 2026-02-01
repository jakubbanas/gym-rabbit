<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { goto } from "$app/navigation";
    import { workoutPlans } from "$lib/data";
    import { loadState, updateProgress } from "$lib/storage";
    import type { AppState, WorkoutPlan, Exercise } from "$lib/types";
    import RestTimer from "$lib/components/RestTimer.svelte";
    import ExerciseCard from "$lib/components/ExerciseCard.svelte";
    import FloatingMenu from "$lib/components/FloatingMenu.svelte";

    // @ts-expect-error - svelte-check has issues with $state rune typing
    let state = $state({
        currentPlanId: "",
        progress: {} as Record<string, any>,
    });
    let currentPlan = $derived(
        workoutPlans.find((p) => p.id === state.currentPlanId),
    );

    // Rest timer state
    let showTimer = $state(false);
    let timerSeconds = $state(0);
    let initialTimerSeconds = $state(0);
    let timerInterval: number | null = null;

    onMount(() => {
        const loaded = loadState();
        state.currentPlanId = loaded.currentPlanId;
        state.progress = loaded.progress;

        // Redirect to plan selector if no plan is selected
        if (!state.currentPlanId) {
            goto("/plans");
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
    {#if currentPlan}
        <div class="header">
            <h1 class="plan-title">{currentPlan.name}</h1>
            <button class="change-plan-btn" onclick={() => goto("/plans")}>
                Change Plan
            </button>
        </div>

        <div class="workout">
            {#each currentPlan.blocks as block}
                <div class="block">
                    <h2 class="block-header">Block {block.block}</h2>

                    {#each block.exercises as exercise}
                        {@const completed = getCompletedSets(exercise)}
                        <ExerciseCard
                            {exercise}
                            {completed}
                            {isSetChecked}
                            onToggleSet={(exerciseId, setIndex, type) =>
                                toggleSet(exerciseId, setIndex, type, exercise)}
                        />
                    {/each}
                </div>
            {/each}
        </div>
    {:else}
        <p class="empty-state">Loading workout...</p>
    {/if}
</div>

<RestTimer
    show={showTimer}
    seconds={timerSeconds}
    initialSeconds={initialTimerSeconds}
    onStop={stopTimer}
    onAddTime={addTime}
/>

<FloatingMenu />

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

    .header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 25px;
        padding: 20px;
        background: var(--card-bg);
        border-radius: 16px;
        box-shadow: 0 4px 6px var(--card-shadow);
    }

    .plan-title {
        margin: 0;
        font-size: 1.8rem;
        font-weight: 800;
        color: var(--accent-primary);
    }

    .change-plan-btn {
        padding: 12px 24px;
        font-size: 1rem;
        font-weight: 600;
        background: transparent;
        color: var(--accent-primary);
        border: 2px solid var(--accent-primary);
        border-radius: 12px;
        cursor: pointer;
        transition: all 0.2s;
        font-family: inherit;
        min-height: 44px;
        min-width: 44px;
    }

    .change-plan-btn:hover {
        background: var(--accent-primary);
        color: white;
        transform: translateY(-2px);
    }

    .change-plan-btn:active {
        transform: scale(0.98);
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

    .empty-state {
        text-align: center;
        color: var(--text-header);
        font-size: 1.2rem;
        padding: 40px;
    }

    @media (max-width: 640px) {
        .container {
            padding: 15px;
        }

        .header {
            flex-direction: column;
            gap: 15px;
            align-items: stretch;
        }

        .change-plan-btn {
            width: 100%;
        }
    }
</style>

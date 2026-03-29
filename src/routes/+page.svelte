<script lang="ts">
    import { onMount, onDestroy } from "svelte";
    import { goto } from "$app/navigation";
    import { loadState, updateProgress } from "$lib/storage";
    import type { AppState, WorkoutPlan, Exercise } from "$lib/types";
    import RestTimer from "$lib/components/RestTimer.svelte";
    import ExerciseCard from "$lib/components/ExerciseCard.svelte";
    import FloatingMenu from "$lib/components/FloatingMenu.svelte";

    let { data } = $props();

    // @ts-expect-error - svelte-check has issues with $state rune typing
    let state = $state({
        currentPlanId: "",
        progress: {} as Record<string, any>,
    });
    let currentPlan = $derived(
        data.workoutPlans.find((p: WorkoutPlan) => p.id === state.currentPlanId),
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
    .container {
        max-width: 800px;
        margin: 0 auto;
        padding: 20px;
    }

    .header {
        margin-bottom: 25px;
        text-align: center;
    }

    .plan-title {
        margin: 0;
        font-size: 1.8rem;
        font-weight: 800;
        color: var(--accent-primary);
    }

    .workout {
        display: flex;
        flex-direction: column;
        gap: 25px;
    }

    .block {
        background: var(--card-bg);
        border-radius: 16px;
        box-shadow: 0 4px 6px var(--card-shadow);
    }

    .block-header {
        font-size: 1.6rem;
        color: var(--accent-primary);
        font-weight: 700;
        text-align: center;
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
    }
</style>

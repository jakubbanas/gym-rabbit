<script lang="ts">
    import type { Exercise } from "$lib/types";
    import CheckboxButton from "./CheckboxButton.svelte";

    interface Props {
        exercise: Exercise;
        completed: number;
        isSetChecked: (
            exerciseId: string,
            setIndex: number,
            type: "left" | "right" | "completed",
        ) => boolean;
        onToggleSet: (
            exerciseId: string,
            setIndex: number,
            type: "left" | "right" | "completed",
        ) => void;
    }

    let { exercise, completed, isSetChecked, onToggleSet }: Props = $props();

    let isFullyCompleted = $derived(completed === exercise.sets);
</script>

<div class="exercise" class:completed={isFullyCompleted}>
    <div class="exercise-header">
        <h3 class="exercise-name">{exercise.name}</h3>
    </div>

    <div class="exercise-details">
        <span><strong>Sets:</strong> {exercise.sets}</span>
        <span><strong>Reps:</strong> {exercise.reps}</span>
        {#if exercise.weight}
            <span><strong>Weight:</strong> {exercise.weight}</span>
        {/if}
        {#if exercise.rest}
            <span><strong>Rest:</strong> {exercise.rest}</span>
        {/if}
    </div>

    {#if exercise.notes}
        <p class="notes">💡 {exercise.notes}</p>
    {/if}

    <div class="sets-tracker">
        {#each Array(exercise.sets) as _, setIndex}
            {#if exercise.isUnilateral}
                <div class="checkbox-group">
                    <CheckboxButton
                        checked={isSetChecked(exercise.id, setIndex, "left")}
                        onclick={() =>
                            onToggleSet(exercise.id, setIndex, "left")}
                        isUnilateral={true}
                        label="L"
                    />
                    <CheckboxButton
                        checked={isSetChecked(exercise.id, setIndex, "right")}
                        onclick={() =>
                            onToggleSet(exercise.id, setIndex, "right")}
                        isUnilateral={true}
                        label="R"
                    />
                </div>
            {:else}
                <CheckboxButton
                    checked={isSetChecked(exercise.id, setIndex, "completed")}
                    onclick={() =>
                        onToggleSet(exercise.id, setIndex, "completed")}
                />
            {/if}
        {/each}
    </div>
</div>

<style>
    .exercise {
        padding: 20px;
        margin-bottom: 20px;
        background: var(--exercise-bg);
        border-radius: 12px;
        border-left: 5px solid var(--accent-secondary);
        transition: border-color 0.3s ease;
    }

    .exercise.completed {
        border-left-color: #22c55e;
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
        flex-shrink: 0;
    }

    @media (max-width: 640px) {
        .sets-tracker {
            justify-content: center;
            gap: 12px;
        }

        .checkbox-group {
            padding: 6px;
        }
    }
</style>

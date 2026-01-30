<script lang="ts">
    import type { WorkoutPlan } from "$lib/types";

    interface Props {
        plans: WorkoutPlan[];
        currentPlanId: string;
        onChange: (planId: string) => void;
    }

    let { plans, currentPlanId, onChange }: Props = $props();
</script>

{#if plans.length > 0}
    <div class="plan-selector">
        <label for="plan-select">Workout Plan:</label>
        <select
            id="plan-select"
            value={currentPlanId}
            onchange={(e) => onChange(e.currentTarget.value)}
        >
            {#each plans as plan}
                <option value={plan.id}>
                    {plan.name} - Started {plan.startDate}
                </option>
            {/each}
        </select>
    </div>
{/if}

<style>
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

    @media (max-width: 640px) {
        .plan-selector {
            flex-direction: column;
            align-items: stretch;
        }
    }
</style>

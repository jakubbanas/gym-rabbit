<script lang="ts">
    import { onMount } from "svelte";
    import { goto } from "$app/navigation";
    import { workoutPlans } from "$lib/data";
    import { loadState, setCurrentPlan } from "$lib/storage";

    let selectedPlanId = $state("");

    onMount(() => {
        const state = loadState();
        selectedPlanId = state.currentPlanId;
    });

    function handleSelectPlan(planId: string) {
        selectedPlanId = planId;
    }

    function handleContinue() {
        if (selectedPlanId) {
            setCurrentPlan(selectedPlanId);
            goto("/");
        }
    }
</script>

<div class="container">
    <div class="header">
        <h1>🐰 Choose Your Workout Plan</h1>
        <p class="subtitle">Select a training program to get started</p>
    </div>

    <div class="plans-grid">
        {#each workoutPlans as plan}
            <button
                class="plan-card"
                class:selected={selectedPlanId === plan.id}
                onclick={() => handleSelectPlan(plan.id)}
            >
                <h2 class="plan-name">{plan.name}</h2>
                <p class="plan-date">Started {plan.startDate}</p>
                <div class="plan-stats">
                    <span>{plan.blocks.length} blocks</span>
                    <span>•</span>
                    <span>
                        {plan.blocks.reduce(
                            (sum, block) => sum + block.exercises.length,
                            0,
                        )} exercises
                    </span>
                </div>
                {#if selectedPlanId === plan.id}
                    <div class="selected-indicator">✓</div>
                {/if}
            </button>
        {/each}
    </div>

    {#if selectedPlanId}
        <button class="continue-btn" onclick={handleContinue}>
            Continue to Workout
        </button>
    {/if}
</div>

<style>
    .container {
        max-width: 900px;
        margin: 0 auto;
        padding: 40px 20px;
        min-height: 100vh;
        display: flex;
        flex-direction: column;
    }

    .header {
        text-align: center;
        margin-bottom: 40px;
    }

    h1 {
        font-size: 2.5rem;
        margin: 0 0 10px 0;
        font-weight: 800;
        color: var(--text-header);
    }

    .subtitle {
        margin: 0;
        font-size: 1.2rem;
        opacity: 0.9;
        color: var(--text-header);
    }

    .plans-grid {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
        gap: 24px;
        margin-bottom: 40px;
    }

    .plan-card {
        background: var(--card-bg);
        border: 3px solid var(--border-color);
        border-radius: 20px;
        padding: 32px;
        cursor: pointer;
        transition: all 0.3s;
        text-align: left;
        position: relative;
        font-family: inherit;
    }

    .plan-card:hover {
        transform: translateY(-4px);
        box-shadow: 0 8px 24px var(--card-shadow);
        border-color: var(--accent-primary);
    }

    .plan-card.selected {
        border-color: var(--accent-primary);
        background: var(--exercise-bg);
        box-shadow: 0 4px 20px rgba(102, 126, 234, 0.3);
    }

    .plan-name {
        margin: 0 0 8px 0;
        font-size: 1.8rem;
        color: var(--text-primary);
        font-weight: 700;
    }

    .plan-date {
        margin: 0 0 16px 0;
        font-size: 1.1rem;
        color: var(--text-secondary);
    }

    .plan-stats {
        display: flex;
        gap: 12px;
        font-size: 1rem;
        color: var(--text-tertiary);
        align-items: center;
    }

    .selected-indicator {
        position: absolute;
        top: 20px;
        right: 20px;
        width: 48px;
        height: 48px;
        background: var(--accent-primary);
        border-radius: 50%;
        display: flex;
        align-items: center;
        justify-content: center;
        font-size: 1.8rem;
        color: white;
        font-weight: 900;
    }

    .continue-btn {
        padding: 20px 40px;
        font-size: 1.3rem;
        font-weight: 700;
        background: linear-gradient(
            135deg,
            var(--accent-primary),
            var(--accent-secondary)
        );
        color: white;
        border: none;
        border-radius: 16px;
        cursor: pointer;
        transition: all 0.2s;
        text-transform: uppercase;
        letter-spacing: 1px;
        margin: 0 auto;
        min-width: 300px;
        font-family: inherit;
    }

    .continue-btn:hover {
        transform: translateY(-2px);
        box-shadow: 0 8px 24px rgba(102, 126, 234, 0.4);
    }

    .continue-btn:active {
        transform: scale(0.98);
    }

    @media (max-width: 640px) {
        .container {
            padding: 20px 15px;
        }

        h1 {
            font-size: 2rem;
        }

        .plans-grid {
            grid-template-columns: 1fr;
        }

        .continue-btn {
            min-width: 100%;
        }
    }
</style>

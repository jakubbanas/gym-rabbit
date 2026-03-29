<script lang="ts">
    import { enhance } from "$app/forms";
    import type { WorkoutPlan, TrainingBlock, Exercise } from "$lib/types";

    let { data } = $props();
    
    let planName = $state(data.plan.name);
    let startDate = $state(data.plan.startDate);
    let blocks = $state<TrainingBlock[]>(data.plan.blocks);

    function addBlock() {
        const nextBlockLabel = String.fromCharCode(65 + blocks.length); // A, B, C...
        blocks = [...blocks, { block: nextBlockLabel, exercises: [] }];
    }

    function removeBlock(index: number) {
        blocks = blocks.filter((_, i) => i !== index);
    }

    function addExercise(blockIndex: number) {
        const block = blocks[blockIndex];
        const exerciseId = `${block.block.toLowerCase()}${block.exercises.length + 1}`;
        
        const newExercise: Exercise = {
            id: exerciseId,
            name: "",
            sets: 3,
            reps: "10",
            weight: "",
            rest: 60,
            notes: "",
            isUnilateral: false
        };

        const updatedBlocks = [...blocks];
        updatedBlocks[blockIndex].exercises = [...updatedBlocks[blockIndex].exercises, newExercise];
        blocks = updatedBlocks;
    }

    function removeExercise(blockIndex: number, exerciseIndex: number) {
        const updatedBlocks = [...blocks];
        updatedBlocks[blockIndex].exercises = updatedBlocks[blockIndex].exercises.filter((_, i) => i !== exerciseIndex);
        blocks = updatedBlocks;
    }

    function generatePlanId(name: string): string {
        return name.toLowerCase()
            .replace(/\s+/g, '-')
            .replace(/[^a-z0-9-]/g, '') + '-' + Date.now().toString().slice(-4);
    }

    let intent = $state('update');

    let planDataJson = $derived(JSON.stringify({
        id: intent === 'saveAsNew' ? generatePlanId(planName) : data.plan.id,
        name: planName,
        startDate: startDate,
        blocks: blocks
    } as WorkoutPlan));
</script>

<div class="container">
    <div class="header">
        <a href="/plans" class="back-link">← Back to Plans</a>
        <h1>Edit Workout Plan</h1>
    </div>

    <form method="POST" use:enhance class="plan-form">
        <input type="hidden" name="planData" value={planDataJson} />
        <input type="hidden" name="intent" value={intent} />

        <div class="section">
            <div class="form-group">
                <label for="planName">Plan Name</label>
                <input 
                    type="text" 
                    id="planName" 
                    bind:value={planName} 
                    placeholder="e.g., Summer Strength 2026" 
                    required 
                />
            </div>
            <div class="form-group">
                <label for="startDate">Start Date (DD.MM)</label>
                <input 
                    type="text" 
                    id="startDate" 
                    bind:value={startDate} 
                    placeholder="29.03" 
                    required 
                />
            </div>
        </div>

        <div class="blocks-container">
            <h3>Training Blocks</h3>
            
            {#each blocks as block, bIdx}
                <div class="block-editor">
                    <div class="block-header">
                        <h4>Block {block.block}</h4>
                        <button type="button" class="btn-remove" onclick={() => removeBlock(bIdx)}>Remove Block</button>
                    </div>

                    <div class="exercises-list">
                        {#each block.exercises as exercise, eIdx}
                            <div class="exercise-editor">
                                <div class="exercise-row">
                                    <div class="form-group flex-2">
                                        <label>Exercise Name</label>
                                        <input type="text" bind:value={exercise.name} placeholder="e.g., Bench Press" required />
                                    </div>
                                    <div class="form-group flex-1">
                                        <label>Sets</label>
                                        <input type="number" bind:value={exercise.sets} min="1" required />
                                    </div>
                                    <div class="form-group flex-1">
                                        <label>Reps</label>
                                        <input type="text" bind:value={exercise.reps} placeholder="8-10" required />
                                    </div>
                                </div>

                                <div class="exercise-row">
                                    <div class="form-group flex-1">
                                        <label>Weight</label>
                                        <input type="text" bind:value={exercise.weight} placeholder="40kg" />
                                    </div>
                                    <div class="form-group flex-1">
                                        <label>Rest (sec)</label>
                                        <input type="number" bind:value={exercise.rest} min="0" step="10" />
                                    </div>
                                    <div class="form-group flex-1 centered">
                                        <label>Unilateral</label>
                                        <input type="checkbox" bind:checked={exercise.isUnilateral} />
                                    </div>
                                    <button type="button" class="btn-remove-sm" onclick={() => removeExercise(bIdx, eIdx)}>×</button>
                                </div>
                                
                                <div class="form-group">
                                    <label>Notes</label>
                                    <textarea bind:value={exercise.notes} rows="2" placeholder="Form cues, machine setting, etc."></textarea>
                                </div>
                            </div>
                        {/each}
                        
                        <button type="button" class="btn-add-exercise" onclick={() => addExercise(bIdx)}>
                            + Add Exercise to Block {block.block}
                        </button>
                    </div>
                </div>
            {/each}

            <button type="button" class="btn-add-block" onclick={addBlock}>
                + Add Training Block
            </button>
        </div>

        <div class="form-actions">
            <button 
                type="submit" 
                class="btn-submit" 
                onclick={() => intent = 'update'}
                disabled={!planName || blocks.length === 0}
            >
                Update Workout Plan
            </button>
            <button 
                type="submit" 
                class="btn-save-as-new" 
                onclick={() => intent = 'saveAsNew'}
                disabled={!planName || blocks.length === 0}
            >
                Save as New Plan
            </button>
        </div>
    </form>
</div>

<style>
    /* Styling copied from new plan page for consistency */
    .container {
        max-width: 800px;
        margin: 0 auto;
        padding: 20px;
        color: var(--text-primary);
    }

    .header {
        margin-bottom: 30px;
    }

    .back-link {
        color: var(--accent-primary);
        text-decoration: none;
        display: block;
        margin-bottom: 10px;
        font-weight: 600;
    }

    h1 {
        margin: 0;
        color: var(--text-header);
    }

    .plan-form {
        background: var(--card-bg);
        padding: 30px;
        border-radius: 16px;
        box-shadow: 0 4px 6px var(--card-shadow);
    }

    .section {
        display: grid;
        grid-template-columns: 2fr 1fr;
        gap: 20px;
        margin-bottom: 30px;
    }

    .form-group {
        display: flex;
        flex-direction: column;
        gap: 8px;
        margin-bottom: 15px;
    }

    .flex-1 { flex: 1; }
    .flex-2 { flex: 2; }

    label {
        font-weight: 600;
        font-size: 0.9rem;
        color: var(--text-secondary);
    }

    input[type="text"],
    input[type="number"],
    textarea {
        padding: 12px;
        border: 1px solid var(--border-color);
        border-radius: 8px;
        background: var(--exercise-bg);
        color: var(--text-primary);
        font-family: inherit;
    }

    .exercise-row {
        display: flex;
        gap: 15px;
        align-items: flex-end;
    }

    .centered {
        align-items: center;
        justify-content: center;
    }

    .blocks-container {
        margin-top: 40px;
        border-top: 2px solid var(--border-color);
        padding-top: 20px;
    }

    .block-editor {
        background: var(--exercise-bg);
        border: 1px solid var(--border-color);
        border-radius: 12px;
        padding: 20px;
        margin-bottom: 25px;
    }

    .block-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-bottom: 20px;
        border-bottom: 1px solid var(--border-color);
        padding-bottom: 10px;
    }

    .block-header h4 {
        margin: 0;
        font-size: 1.2rem;
        color: var(--accent-primary);
    }

    .exercise-editor {
        background: var(--card-bg);
        border: 1px solid var(--border-color);
        border-radius: 8px;
        padding: 15px;
        margin-bottom: 15px;
        position: relative;
    }

    .btn-remove {
        background: #ef4444;
        color: white;
        border: none;
        padding: 6px 12px;
        border-radius: 6px;
        cursor: pointer;
        font-size: 0.8rem;
    }

    .btn-remove-sm {
        background: #ef4444;
        color: white;
        border: none;
        width: 30px;
        height: 30px;
        border-radius: 50%;
        cursor: pointer;
        display: flex;
        align-items: center;
        justify-content: center;
        font-weight: bold;
        margin-bottom: 15px;
    }

    .btn-add-exercise {
        width: 100%;
        padding: 10px;
        background: transparent;
        border: 2px dashed var(--accent-primary);
        color: var(--accent-primary);
        border-radius: 8px;
        cursor: pointer;
        font-weight: 600;
        transition: all 0.2s;
    }

    .btn-add-exercise:hover {
        background: rgba(102, 126, 234, 0.1);
    }

    .btn-add-block {
        width: 100%;
        padding: 15px;
        background: var(--accent-primary);
        color: white;
        border: none;
        border-radius: 12px;
        cursor: pointer;
        font-weight: 700;
        font-size: 1.1rem;
        margin-bottom: 30px;
    }

    .form-actions {
        margin-top: 40px;
        display: flex;
        gap: 20px;
        justify-content: center;
    }

    .btn-submit {
        padding: 15px 30px;
        background: linear-gradient(135deg, #10b981, #059669);
        color: white;
        border: none;
        border-radius: 12px;
        font-size: 1.1rem;
        font-weight: 700;
        cursor: pointer;
        box-shadow: 0 4px 12px rgba(16, 185, 129, 0.3);
        flex: 1;
        max-width: 300px;
    }

    .btn-save-as-new {
        padding: 15px 30px;
        background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
        color: white;
        border: none;
        border-radius: 12px;
        font-size: 1.1rem;
        font-weight: 700;
        cursor: pointer;
        box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
        flex: 1;
        max-width: 300px;
    }

    .btn-submit:disabled, .btn-save-as-new:disabled {
        background: #9ca3af;
        cursor: not-allowed;
        box-shadow: none;
    }

    @media (max-width: 640px) {
        .section {
            grid-template-columns: 1fr;
        }
        .form-actions {
            flex-direction: column;
            align-items: center;
        }
        .btn-submit, .btn-save-as-new {
            width: 100%;
            max-width: none;
        }
        .exercise-row {
            flex-direction: column;
            align-items: stretch;
        }
        .btn-remove-sm {
            width: 100%;
            border-radius: 6px;
            height: 40px;
        }
    }
</style>

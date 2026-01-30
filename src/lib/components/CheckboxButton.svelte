<script lang="ts">
    interface Props {
        checked: boolean;
        onclick: () => void;
        isUnilateral?: boolean;
        label?: string;
    }

    let { checked, onclick, isUnilateral = false, label }: Props = $props();
</script>

<button class="custom-checkbox" class:checked {onclick}>
    {#if isUnilateral && label}
        <span class="check-label">{label}</span>
    {:else}
        <span class="check-icon">✓</span>
    {/if}
</button>

<style>
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
        min-width: 75px;
        max-width: 85px;
        flex: 1;
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

    @media (max-width: 640px) {
        .custom-checkbox {
            min-width: 70px;
            max-width: 80px;
            min-height: 85px;
            padding: 14px 16px;
        }
    }
</style>

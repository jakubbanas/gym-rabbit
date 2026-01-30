<script lang="ts">
    interface Props {
        show: boolean;
        seconds: number;
        initialSeconds: number;
        onStop: () => void;
        onAddTime: () => void;
    }

    let { show, seconds, initialSeconds, onStop, onAddTime }: Props = $props();

    function formatTime(secs: number): string {
        const mins = Math.floor(secs / 60);
        const s = secs % 60;
        return `${mins}:${s.toString().padStart(2, "0")}`;
    }

    function handleKeyDown(event: KeyboardEvent) {
        if (event.key === "Escape") {
            onStop();
        }
    }

    function handleContentKeyDown(event: KeyboardEvent) {
        // Stop propagation for keyboard events on content
        event.stopPropagation();
    }
</script>

{#if show}
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
        class="timer-overlay"
        role="dialog"
        aria-modal="true"
        aria-label="Rest timer"
        onclick={onStop}
        onkeydown={handleKeyDown}
        tabindex="-1"
    >
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
            class="timer-content"
            onclick={(e) => e.stopPropagation()}
            onkeydown={handleContentKeyDown}
        >
            <div class="timer-display">
                <div class="timer-text">{formatTime(seconds)}</div>
                <div class="timer-label">Rest Time</div>
            </div>

            <div class="timer-progress">
                <div
                    class="timer-progress-bar"
                    style="width: {initialSeconds > 0
                        ? (seconds / initialSeconds) * 100
                        : 0}%"
                ></div>
            </div>

            <div class="timer-actions">
                <button
                    class="timer-btn timer-btn-secondary"
                    onclick={onAddTime}
                >
                    +30s
                </button>
                <button class="timer-btn timer-btn-primary" onclick={onStop}>
                    Skip Rest
                </button>
            </div>
        </div>
    </div>
{/if}

<style>
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
</style>

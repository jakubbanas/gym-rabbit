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
</script>

{#if show}
    <div class="timer-widget" role="timer" aria-label="Rest timer">
        <div class="timer-header">
            <span class="timer-label">Rest Time</span>
            <button class="timer-close" onclick={onStop} aria-label="Skip rest">✕</button>
        </div>

        <div class="timer-text">{formatTime(seconds)}</div>

        <div class="timer-progress">
            <div
                class="timer-progress-bar"
                style="width: {initialSeconds > 0 ? (seconds / initialSeconds) * 100 : 0}%"
            ></div>
        </div>

        <div class="timer-actions">
            <button class="timer-btn timer-btn-secondary" onclick={onAddTime}>+30s</button>
            <button class="timer-btn timer-btn-primary" onclick={onStop}>Skip</button>
        </div>
    </div>
{/if}

<style>
    .timer-widget {
        position: fixed;
        bottom: 24px;
        right: 24px;
        width: 220px;
        background: rgba(20, 20, 30, 0.92);
        backdrop-filter: blur(16px);
        border: 1px solid rgba(102, 126, 234, 0.3);
        border-radius: 20px;
        padding: 16px 18px;
        z-index: 1000;
        box-shadow: 0 8px 32px rgba(0, 0, 0, 0.4), 0 0 0 1px rgba(102, 126, 234, 0.1);
        display: flex;
        flex-direction: column;
        gap: 12px;
        animation: slideIn 0.25s ease-out;
    }

    @keyframes slideIn {
        from {
            opacity: 0;
            transform: translateY(16px) scale(0.95);
        }
        to {
            opacity: 1;
            transform: translateY(0) scale(1);
        }
    }

    .timer-header {
        display: flex;
        justify-content: space-between;
        align-items: center;
    }

    .timer-label {
        font-size: 0.7rem;
        color: rgba(255, 255, 255, 0.5);
        font-weight: 600;
        text-transform: uppercase;
        letter-spacing: 1.5px;
    }

    .timer-close {
        background: none;
        border: none;
        color: rgba(255, 255, 255, 0.4);
        cursor: pointer;
        font-size: 0.85rem;
        padding: 2px 4px;
        border-radius: 6px;
        transition: color 0.15s, background 0.15s;
        line-height: 1;
    }

    .timer-close:hover {
        color: rgba(255, 255, 255, 0.9);
        background: rgba(255, 255, 255, 0.1);
    }

    .timer-text {
        font-size: 2.8rem;
        font-weight: 800;
        color: #ffffff;
        line-height: 1;
        text-align: center;
        font-variant-numeric: tabular-nums;
        text-shadow: 0 0 20px rgba(102, 126, 234, 0.6);
    }

    .timer-progress {
        width: 100%;
        height: 4px;
        background: rgba(255, 255, 255, 0.15);
        border-radius: 2px;
        overflow: hidden;
    }

    .timer-progress-bar {
        height: 100%;
        background: linear-gradient(90deg, #667eea, #764ba2);
        transition: width 1s linear;
        border-radius: 2px;
    }

    .timer-actions {
        display: flex;
        gap: 8px;
    }

    .timer-btn {
        flex: 1;
        padding: 8px 10px;
        font-size: 0.8rem;
        font-weight: 700;
        border: none;
        border-radius: 10px;
        cursor: pointer;
        transition: all 0.15s;
        text-transform: uppercase;
        letter-spacing: 0.5px;
    }

    .timer-btn-primary {
        background: linear-gradient(135deg, #667eea, #764ba2);
        color: white;
    }

    .timer-btn-primary:hover {
        box-shadow: 0 4px 12px rgba(102, 126, 234, 0.5);
        filter: brightness(1.1);
    }

    .timer-btn-secondary {
        background: rgba(255, 255, 255, 0.08);
        color: rgba(255, 255, 255, 0.8);
        border: 1px solid rgba(255, 255, 255, 0.2);
    }

    .timer-btn-secondary:hover {
        background: rgba(255, 255, 255, 0.15);
        border-color: rgba(255, 255, 255, 0.4);
    }

    .timer-btn:active {
        transform: scale(0.96);
    }
</style>

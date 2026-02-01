<script lang="ts">
    import { goto } from "$app/navigation";

    let kmhInput = $state("");
    let minKmMinutes = $state("");
    let minKmSeconds = $state("");

    let kmhResult = $derived(() => {
        const mins = parseFloat(minKmMinutes);
        const secs = parseFloat(minKmSeconds) || 0;

        if (isNaN(mins) || mins <= 0) return "";

        const totalMinutes = mins + secs / 60;
        const kmh = 60 / totalMinutes;
        return kmh.toFixed(2);
    });

    let minKmResult = $derived(() => {
        const speed = parseFloat(kmhInput);

        if (isNaN(speed) || speed <= 0) return { minutes: "", seconds: "" };

        const paceMinutes = 60 / speed;
        const mins = Math.floor(paceMinutes);
        const secs = Math.round((paceMinutes - mins) * 60);

        return {
            minutes: mins.toString(),
            seconds: secs.toString().padStart(2, "0"),
        };
    });

    function handleKmhInput() {
        // Clear the other input when this one is being used
        minKmMinutes = "";
        minKmSeconds = "";
    }

    function handleMinKmInput() {
        // Clear the other input when this one is being used
        kmhInput = "";
    }
</script>

<div class="container">
    <div class="header">
        <button class="back-btn" onclick={() => goto("/")}>← Back</button>
        <h1>🏃 Running Pace Calculator</h1>
    </div>

    <div class="calculators">
        <div class="calculator-card">
            <h2 class="card-title">km/h → min/km</h2>
            <div class="input-group">
                <label for="kmh-input">Speed (km/h)</label>
                <input
                    id="kmh-input"
                    type="number"
                    step="0.1"
                    min="0"
                    bind:value={kmhInput}
                    oninput={handleKmhInput}
                    placeholder="Enter speed"
                />
            </div>
            {#if minKmResult().minutes}
                <div class="result">
                    <span class="result-label">Pace:</span>
                    <span class="result-value">
                        {minKmResult().minutes}:{minKmResult().seconds} min/km
                    </span>
                </div>
            {/if}
        </div>

        <div class="divider">⇄</div>

        <div class="calculator-card">
            <h2 class="card-title">min/km → km/h</h2>
            <div class="input-group">
                <label for="min-input">Pace</label>
                <div class="pace-inputs">
                    <div class="pace-input-wrapper">
                        <input
                            id="min-input"
                            type="number"
                            min="0"
                            bind:value={minKmMinutes}
                            oninput={handleMinKmInput}
                            placeholder="Min"
                        />
                        <span class="input-suffix">min</span>
                    </div>
                    <span class="pace-separator">:</span>
                    <div class="pace-input-wrapper">
                        <input
                            id="sec-input"
                            type="number"
                            min="0"
                            max="59"
                            bind:value={minKmSeconds}
                            oninput={handleMinKmInput}
                            placeholder="Sec"
                        />
                        <span class="input-suffix">sec</span>
                    </div>
                </div>
            </div>
            {#if kmhResult()}
                <div class="result">
                    <span class="result-label">Speed:</span>
                    <span class="result-value">{kmhResult()} km/h</span>
                </div>
            {/if}
        </div>
    </div>

    <div class="info-card">
        <h3>Common Running Paces</h3>
        <div class="pace-examples">
            <div class="pace-example">
                <span class="pace">5:00 min/km</span>
                <span class="speed">= 12.0 km/h</span>
            </div>
            <div class="pace-example">
                <span class="pace">5:30 min/km</span>
                <span class="speed">= 10.9 km/h</span>
            </div>
            <div class="pace-example">
                <span class="pace">6:00 min/km</span>
                <span class="speed">= 10.0 km/h</span>
            </div>
            <div class="pace-example">
                <span class="pace">6:30 min/km</span>
                <span class="speed">= 9.2 km/h</span>
            </div>
        </div>
    </div>
</div>

<style>
    .container {
        max-width: 800px;
        margin: 0 auto;
        padding: 20px;
        min-height: 100vh;
    }

    .header {
        display: flex;
        align-items: center;
        gap: 20px;
        margin-bottom: 30px;
    }

    .back-btn {
        padding: 12px 20px;
        font-size: 1rem;
        font-weight: 600;
        background: var(--card-bg);
        color: var(--accent-primary);
        border: 2px solid var(--accent-primary);
        border-radius: 12px;
        cursor: pointer;
        transition: all 0.2s;
        font-family: inherit;
    }

    .back-btn:hover {
        background: var(--accent-primary);
        color: white;
        transform: translateX(-2px);
    }

    h1 {
        margin: 0;
        font-size: 2rem;
        font-weight: 800;
        color: var(--text-header);
    }

    .calculators {
        display: grid;
        grid-template-columns: 1fr auto 1fr;
        gap: 30px;
        margin-bottom: 30px;
        align-items: start;
    }

    .divider {
        font-size: 3rem;
        color: var(--accent-primary);
        display: flex;
        align-items: center;
        padding-top: 60px;
    }

    .calculator-card {
        background: var(--card-bg);
        border-radius: 16px;
        padding: 28px;
        box-shadow: 0 4px 12px var(--card-shadow);
    }

    .card-title {
        margin: 0 0 24px 0;
        font-size: 1.4rem;
        font-weight: 700;
        color: var(--accent-primary);
        text-align: center;
    }

    .input-group {
        margin-bottom: 20px;
    }

    .input-group label {
        display: block;
        margin-bottom: 10px;
        font-weight: 600;
        color: var(--text-secondary);
        font-size: 0.95rem;
    }

    input[type="number"] {
        width: 100%;
        padding: 16px;
        font-size: 1.2rem;
        border: 2px solid var(--border-color);
        border-radius: 12px;
        background: var(--exercise-bg);
        color: var(--text-primary);
        font-family: inherit;
        transition: all 0.2s;
        box-sizing: border-box;
    }

    input[type="number"]:focus {
        outline: none;
        border-color: var(--accent-primary);
        box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
    }

    .pace-inputs {
        display: flex;
        align-items: center;
        gap: 12px;
    }

    .pace-input-wrapper {
        flex: 1;
        position: relative;
    }

    .pace-input-wrapper input {
        padding-right: 50px;
    }

    .input-suffix {
        position: absolute;
        right: 16px;
        top: 50%;
        transform: translateY(-50%);
        color: var(--text-tertiary);
        font-size: 0.9rem;
        font-weight: 600;
        pointer-events: none;
    }

    .pace-separator {
        font-size: 2rem;
        font-weight: 700;
        color: var(--text-tertiary);
    }

    .result {
        background: linear-gradient(
            135deg,
            var(--accent-primary),
            var(--accent-secondary)
        );
        color: white;
        padding: 20px;
        border-radius: 12px;
        display: flex;
        justify-content: space-between;
        align-items: center;
        animation: fadeIn 0.3s;
    }

    @keyframes fadeIn {
        from {
            opacity: 0;
            transform: translateY(-10px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }

    .result-label {
        font-size: 1rem;
        opacity: 0.9;
    }

    .result-value {
        font-size: 1.5rem;
        font-weight: 700;
    }

    .info-card {
        background: var(--card-bg);
        border-radius: 16px;
        padding: 28px;
        box-shadow: 0 4px 12px var(--card-shadow);
    }

    .info-card h3 {
        margin: 0 0 20px 0;
        font-size: 1.3rem;
        font-weight: 700;
        color: var(--text-primary);
        text-align: center;
    }

    .pace-examples {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 16px;
    }

    .pace-example {
        display: flex;
        justify-content: space-between;
        align-items: center;
        padding: 16px;
        background: var(--exercise-bg);
        border-radius: 10px;
        border: 2px solid var(--border-color);
    }

    .pace {
        font-weight: 700;
        color: var(--accent-primary);
        font-size: 1.1rem;
    }

    .speed {
        color: var(--text-secondary);
        font-size: 1rem;
    }

    @media (max-width: 768px) {
        .calculators {
            grid-template-columns: 1fr;
            gap: 20px;
        }

        .divider {
            padding-top: 0;
            justify-content: center;
            transform: rotate(90deg);
        }

        .pace-examples {
            grid-template-columns: 1fr;
        }
    }

    @media (max-width: 640px) {
        .container {
            padding: 15px;
        }

        .header {
            flex-direction: column;
            align-items: flex-start;
            gap: 15px;
        }

        h1 {
            font-size: 1.6rem;
        }
    }
</style>

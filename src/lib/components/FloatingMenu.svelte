<script lang="ts">
    import { goto } from "$app/navigation";

    let showMenu = $state(false);

    function toggleMenu() {
        showMenu = !showMenu;
    }

    function closeMenu() {
        showMenu = false;
    }

    function handleKeyDown(event: KeyboardEvent) {
        if (event.key === "Escape") {
            closeMenu();
        }
    }

    const menuItems = [
        { id: "change-plan", label: "Change Plan", icon: "📋" },
        { id: "reset-progress", label: "Reset Progress", icon: "🔄" },
        { id: "export-data", label: "Export Data", icon: "📤" },
        { id: "import-data", label: "Import Data", icon: "📥" },
        { id: "clear-all", label: "Clear All Data", icon: "🗑️" },
        { id: "about", label: "About", icon: "ℹ️" },
    ];

    function handleMenuItemClick(itemId: string) {
        if (itemId === "change-plan") {
            goto("/plans");
        } else {
            console.log(`Menu item clicked: ${itemId}`);
            // TODO: Implement menu item actions
        }
        closeMenu();
    }
</script>

<!-- Backdrop -->
{#if showMenu}
    <button
        class="backdrop"
        onclick={closeMenu}
        onkeydown={handleKeyDown}
        aria-label="Close menu"
    ></button>
{/if}

<!-- Menu -->
{#if showMenu}
    <div class="menu" role="menu">
        {#each menuItems as item}
            <button
                class="menu-item"
                onclick={() => handleMenuItemClick(item.id)}
                onkeydown={(e) => {
                    if (e.key === "Enter" || e.key === " ") {
                        handleMenuItemClick(item.id);
                    }
                }}
                role="menuitem"
            >
                <span class="menu-icon">{item.icon}</span>
                <span class="menu-label">{item.label}</span>
            </button>
        {/each}
    </div>
{/if}

<!-- Floating Button -->
<button
    class="floating-btn"
    class:active={showMenu}
    onclick={toggleMenu}
    onkeydown={(e) => {
        if (e.key === "Enter" || e.key === " ") {
            toggleMenu();
        }
    }}
    aria-label="Menu"
    aria-expanded={showMenu}
>
    <span class="floating-btn-icon">{showMenu ? "✕" : "⚙️"}</span>
</button>

<style>
    .backdrop {
        position: fixed;
        top: 0;
        left: 0;
        right: 0;
        bottom: 0;
        background: rgba(0, 0, 0, 0.3);
        backdrop-filter: blur(2px);
        z-index: 998;
        border: none;
        cursor: pointer;
    }

    .menu {
        position: fixed;
        bottom: 100px;
        right: 20px;
        background: var(--card-bg);
        border-radius: 16px;
        box-shadow: 0 8px 32px var(--card-shadow);
        padding: 8px;
        z-index: 999;
        min-width: 220px;
        animation: slideUp 0.2s ease-out;
    }

    @keyframes slideUp {
        from {
            opacity: 0;
            transform: translateY(10px);
        }
        to {
            opacity: 1;
            transform: translateY(0);
        }
    }

    .menu-item {
        display: flex;
        align-items: center;
        gap: 12px;
        width: 100%;
        padding: 14px 16px;
        background: transparent;
        border: none;
        border-radius: 12px;
        cursor: pointer;
        transition: all 0.2s;
        font-family: inherit;
        font-size: 1rem;
        text-align: left;
        color: var(--text-primary);
    }

    .menu-item:hover {
        background: var(--exercise-bg);
        transform: translateX(-2px);
    }

    .menu-item:active {
        transform: scale(0.98);
    }

    .menu-icon {
        font-size: 1.5rem;
        width: 32px;
        height: 32px;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .menu-label {
        font-weight: 600;
    }

    .floating-btn {
        position: fixed;
        bottom: 20px;
        right: 20px;
        width: 64px;
        height: 64px;
        border-radius: 50%;
        background: linear-gradient(
            135deg,
            var(--accent-primary),
            var(--accent-secondary)
        );
        border: none;
        box-shadow: 0 4px 20px rgba(102, 126, 234, 0.4);
        cursor: pointer;
        z-index: 999;
        transition: all 0.3s;
        display: flex;
        align-items: center;
        justify-content: center;
    }

    .floating-btn:hover {
        transform: scale(1.1);
        box-shadow: 0 6px 28px rgba(102, 126, 234, 0.5);
    }

    .floating-btn:active {
        transform: scale(0.95);
    }

    .floating-btn.active {
        background: var(--accent-primary);
        transform: rotate(90deg);
    }

    .floating-btn-icon {
        font-size: 1.8rem;
        transition: transform 0.2s;
    }

    @media (max-width: 640px) {
        .floating-btn {
            bottom: 16px;
            right: 16px;
            width: 56px;
            height: 56px;
        }

        .menu {
            right: 16px;
            bottom: 88px;
            left: 16px;
            min-width: unset;
        }
    }
</style>

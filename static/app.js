import {
  esc, localToday, formatDay, formatTimestamp, formatRestClock, formatRestBadge,
  blockLetter, exerciseLabel, initProgress, progressMatchesWorkout,
  isSetComplete, hasSetProgress, hasAnyLoggedSet, isBlockComplete,
  countSessionStats, undoSetInArray, renderCueHtml, calculatePlates,
} from './logic.js';

const STORAGE_KEY = 'gym-rabbit:session:v1';

const ICONS = {
  dumbbell: `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M14.4 14.4 9.6 9.6"/><path d="M18.657 21.485a2 2 0 1 1-2.829-2.828l1.767-1.768a2 2 0 1 1 2.829 2.829z"/><path d="m21.5 21.5-1.4-1.4"/><path d="M3.9 3.9 2.5 2.5"/><path d="M6.404 12.768a2 2 0 1 1-2.829-2.829l1.768-1.767a2 2 0 1 1-2.828-2.829l2.828-2.828a2 2 0 1 1 2.829 2.828l1.767-1.768a2 2 0 1 1 2.829 2.829z"/></svg>`,
  calendar: `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M8 2v4"/><path d="M16 2v4"/><rect width="18" height="18" x="3" y="4" rx="2"/><path d="M3 10h18"/></svg>`,
  back: `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="m12 19-7-7 7-7"/><path d="M19 12H5"/></svg>`,
  clock: `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>`,
  check: `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M20 6 9 17l-5-5"/></svg>`,
  barbell: `<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.1" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M6 7v10"/><path d="M18 7v10"/><path d="M3 9v6"/><path d="M21 9v6"/><path d="M6 12h12"/></svg>`,
  close: `<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M18 6 6 18"/><path d="m6 6 12 12"/></svg>`,
  undo: `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 7v6h6"/><path d="M21 17a9 9 0 0 0-9-9 9 9 0 0 0-6 2.3L3 13"/></svg>`,
  play: `<svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><polygon points="6 3 20 12 6 21 6 3"/></svg>`,
};

let audioCtx = null;
function unlockAudio() {
  try {
    const Ctx = window.AudioContext || window.webkitAudioContext;
    if (!Ctx) return;
    if (!audioCtx) audioCtx = new Ctx();
    if (audioCtx.state === 'suspended') audioCtx.resume().catch(() => {});
  } catch {}
}

function playBeep() {
  try {
    const Ctx = window.AudioContext || window.webkitAudioContext;
    if (!Ctx) return;
    if (!audioCtx) audioCtx = new Ctx();
    const now = audioCtx.currentTime;
    [880, 1046.5].forEach((freq, idx) => {
      const osc = audioCtx.createOscillator();
      const gain = audioCtx.createGain();
      osc.type = 'sine';
      osc.frequency.setValueAtTime(freq, now + idx * 0.14);
      gain.gain.setValueAtTime(0.001, now + idx * 0.14);
      gain.gain.exponentialRampToValueAtTime(0.45, now + idx * 0.14 + 0.02);
      gain.gain.exponentialRampToValueAtTime(0.001, now + idx * 0.14 + 0.35);
      osc.connect(gain);
      gain.connect(audioCtx.destination);
      osc.start(now + idx * 0.14);
      osc.stop(now + idx * 0.14 + 0.36);
    });
  } catch {}
}

function triggerHaptic(pattern = 45) {
  if (navigator.vibrate) try { navigator.vibrate(pattern); } catch {}
}

function loadSavedSession() {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : null;
  } catch {
    return null;
  }
}

function saveSession(session) {
  try {
    if (!session) localStorage.removeItem(STORAGE_KEY);
    else localStorage.setItem(STORAGE_KEY, JSON.stringify(session));
  } catch {}
}

const state = {
  screen: 'dashboard',
  workouts: [],
  loading: true,
  loadError: null,
  session: loadSavedSession(), // { workoutId, workout, startedAt, currentBlockIdx, touchedPerformed, progress }
  expandedActuals: {}, // key -> true
  draftActuals: {},    // key -> { weight: string, reps: string }
  rest: null,          // { endAt, nextText }
  drawerOpen: false,
  barWeight: 20,
  targetWeight: '',
  confirmAction: null, // null | { kind: 'replace', workout } | { kind: 'discard' }
};

let restInterval = null;
const appRoot = document.getElementById('app');
const restRoot = document.getElementById('rest-root');
const drawerRoot = document.getElementById('drawer-root');

function activeWorkout() {
  if (!state.session) return null;
  return state.workouts.find((w) => w.id === state.session.workoutId) ?? state.session.workout ?? null;
}

async function fetchWorkouts() {
  state.loading = true;
  state.loadError = null;
  render();
  try {
    const res = await fetch('/api/workouts');
    if (!res.ok) throw new Error(`Server returned ${res.status}`);
    state.workouts = await res.json();
    if (state.session) {
      const fresh = state.workouts.find((w) => w.id === state.session.workoutId);
      if (fresh) {
        if (progressMatchesWorkout(state.session.progress, fresh)) {
          state.session.workout = fresh;
          saveSession(state.session);
        } else {
          state.session = null;
          saveSession(null);
        }
      }
    }
  } catch (err) {
    state.loadError = err instanceof Error ? err.message : 'Failed to load workouts';
  } finally {
    state.loading = false;
    render();
  }
}

function touchLastPerformed(id) {
  fetch(`/api/workouts/${encodeURIComponent(id)}/performed`, { method: 'POST' })
    .then((res) => {
      if (res.ok) {
        const w = state.workouts.find((item) => item.id === id);
        if (w) w.last_performed = new Date().toISOString();
      }
    })
    .catch(() => {});
}

function startWorkout(workout, forceReplace = false) {
  const existing = state.session;
  if (existing && existing.workoutId === workout.id && progressMatchesWorkout(existing.progress, workout)) {
    state.screen = 'workout';
    render();
    return;
  }
  const existingWorkout = activeWorkout();
  if (existing && !forceReplace && hasAnyLoggedSet(existingWorkout, existing.progress)) {
    state.confirmAction = { kind: 'replace', workout };
    renderDrawer();
    return;
  }

  state.confirmAction = null;
  state.session = {
    workoutId: workout.id,
    workout,
    startedAt: Date.now(),
    currentBlockIdx: 0,
    touchedPerformed: false,
    progress: initProgress(workout),
  };
  state.expandedActuals = {};
  state.draftActuals = {};
  saveSession(state.session);
  state.screen = 'workout';
  render();
}

function getActiveSetIdx(ex, setsProgress) {
  return setsProgress.findIndex((s) => !isSetComplete(ex, s));
}

function computeNextRestLabel(workout, progress, blockIdx, exIdx) {
  const block = workout.blocks[blockIdx];
  const ex = block.exercises[exIdx];
  const nextSetIdx = getActiveSetIdx(ex, progress[blockIdx][exIdx]);
  if (nextSetIdx !== -1) {
    return `Next: ${exerciseLabel(block, blockIdx, exIdx)}. ${ex.name} · Set ${nextSetIdx + 1} of ${ex.sets}`;
  }
  const nextExIdx = block.exercises.findIndex((o, idx) => idx > exIdx && getActiveSetIdx(o, progress[blockIdx][idx]) !== -1);
  if (nextExIdx !== -1) return `Next: ${exerciseLabel(block, blockIdx, nextExIdx)}. ${block.exercises[nextExIdx].name}`;
  if (blockIdx + 1 < workout.blocks.length) return `Block complete · Up next: ${workout.blocks[blockIdx + 1].name}`;
  return 'Final set logged — ready to finish!';
}

function logSet(blockIdx, exIdx, setIdx, side) {
  const workout = activeWorkout();
  if (!workout || !state.session) return;

  unlockAudio();
  triggerHaptic(45);

  const key = `${blockIdx}-${exIdx}`;
  const ex = workout.blocks[blockIdx].exercises[exIdx];
  const setObj = state.session.progress[blockIdx][exIdx][setIdx];
  const draft = state.draftActuals[key];

  if (draft) {
    if (draft.weight.trim() !== '') {
      const w = parseFloat(draft.weight);
      if (!Number.isNaN(w) && w >= 0) setObj.weightKg = w;
    }
    if (draft.reps.trim() !== '') {
      const r = parseInt(draft.reps, 10);
      if (!Number.isNaN(r) && r >= 0) setObj.reps = r;
    }
  }

  if (side) setObj[side] = true;
  else setObj.done = true;

  // Clear draft reps once the full set is complete, while keeping draft weight for the next set.
  if (draft && isSetComplete(ex, setObj)) {
    draft.reps = '';
  }

  if (!state.session.touchedPerformed) {
    state.session.touchedPerformed = true;
    touchLastPerformed(workout.id);
  }
  saveSession(state.session);

  const stats = countSessionStats(workout, state.session.progress);
  if (ex.rest_seconds > 0 && stats.doneSets < stats.totalSets) {
    startRestTimer(ex.rest_seconds, computeNextRestLabel(workout, state.session.progress, blockIdx, exIdx));
  }
  render();
}

function undoLastSet(blockIdx, exIdx) {
  const workout = activeWorkout();
  if (!workout || !state.session) return;
  if (undoSetInArray(workout.blocks[blockIdx].exercises[exIdx], state.session.progress[blockIdx][exIdx])) {
    saveSession(state.session);
    render();
  }
}

function startRestTimer(seconds, nextText) {
  stopRestTimer(false);
  state.rest = { endAt: Date.now() + seconds * 1000, nextText };
  renderRest();
  restInterval = window.setInterval(tickRestTimer, 200);
}

function tickRestTimer() {
  if (!state.rest) return;
  const remainingMs = state.rest.endAt - Date.now();
  if (remainingMs <= 0) {
    playBeep();
    triggerHaptic([80, 50, 80]);
    stopRestTimer(true);
    return;
  }
  const el = document.getElementById('rest-clock');
  if (el) el.textContent = formatRestClock(remainingMs / 1000);
}

function adjustRestTimer(deltaSeconds) {
  if (!state.rest) return;
  state.rest.endAt = Math.max(Date.now() + 1000, state.rest.endAt + deltaSeconds * 1000);
  tickRestTimer();
}

function stopRestTimer(rerender = true) {
  if (restInterval !== null) {
    clearInterval(restInterval);
    restInterval = null;
  }
  state.rest = null;
  if (rerender) renderRest();
}

document.addEventListener('visibilitychange', () => {
  if (document.visibilityState === 'visible' && state.rest) tickRestTimer();
});

function renderSleeveResultHtml() {
  const raw = state.targetWeight.trim();
  const target = parseFloat(raw);
  const { plates, actualTotal } = calculatePlates(raw, state.barWeight);

  if (!raw || Number.isNaN(target) || target <= 0) {
    return `<div class="sleeve-box empty"><span style="font-size:2rem" aria-hidden="true">🏋️</span><span>Enter a weight to calculate plates</span></div>`;
  }
  if (target <= state.barWeight) {
    return `<div class="sleeve-box empty"><span style="font-size:2rem" aria-hidden="true">🏋️</span><span>Just the ${state.barWeight} kg bar!</span></div>`;
  }

  const diff = actualTotal - target;
  const diffWarn = Math.abs(diff) > 0.01
    ? `<p class="warn-text">Closest achievable load: ${actualTotal} kg (${diff > 0 ? '+' : ''}${diff.toFixed(2).replace(/\.?0+$/, '')} kg)</p>`
    : '';

  return `
    <div class="sleeve-box">
      <div>
        <div class="row-between">
          <span style="font-size:0.85rem;font-weight:700;color:var(--text-muted)">Load per side</span>
          <span style="font-size:0.8rem;color:var(--text-dim)">${plates.reduce((a, b) => a + b, 0)} kg / side</span>
        </div>
        ${diffWarn}
      </div>
      <div class="sleeve-visual">
        <div class="bar-collar"></div><div class="bar-neck"></div>
        ${plates.map((p) => `<div class="plate plate-${String(p).replace('.', '_')}"><span>${p}</span></div>`).join('')}
        <div class="bar-tail"></div>
      </div>
      <div class="plate-chips">${plates.map((p) => `<span class="plate-chip">${p} kg</span>`).join('')}</div>
    </div>
  `;
}

function renderDashboard() {
  const today = localToday();
  const upcoming = state.workouts.filter((w) => w.scheduled_date >= today);
  const past = state.workouts.filter((w) => w.scheduled_date < today).reverse();

  let resumeHtml = '';
  const activeW = activeWorkout();
  if (state.session && activeW) {
    const stats = countSessionStats(activeW, state.session.progress);
    resumeHtml = `
      <section class="resume-banner" aria-label="In-progress workout">
        <p class="workout-eyebrow">Session in progress</p>
        <h2 class="resume-title">${esc(activeW.title)}</h2>
        <p class="muted-sm">${stats.doneSets} of ${stats.totalSets} sets logged · ${esc(activeW.blocks[state.session.currentBlockIdx]?.name ?? '')}</p>
        <div class="resume-actions">
          <button type="button" class="btn-primary-sm" data-action="resume-session">Resume Workout</button>
          <button type="button" class="btn-ghost-sm" data-action="discard-session">Discard</button>
        </div>
      </section>
    `;
  }

  const cardHtml = (w) => {
    const isToday = w.scheduled_date === today;
    const exCount = w.blocks.reduce((sum, b) => sum + b.exercises.length, 0);
    return `
      <button type="button" class="workout-card ${isToday ? 'today' : ''}" data-action="start-workout" data-id="${esc(w.id)}" aria-label="Start ${esc(w.title)}">
        <span class="card-date">${ICONS.calendar} ${isToday ? 'Today' : esc(formatDay(w.scheduled_date))}</span>
        <h3 class="card-title">${esc(w.title)}</h3>
        ${w.notes ? `<p class="card-notes">${esc(w.notes)}</p>` : ''}
        <div class="card-meta-row">
          <span>${w.blocks.length} blocks · ${exCount} exercises</span>
          <span>Last: ${esc(formatTimestamp(w.last_performed))}</span>
        </div>
      </button>
    `;
  };

  let bodyHtml = '';
  if (state.loading) {
    bodyHtml = `<div class="empty-box">Loading workouts…</div>`;
  } else if (state.loadError) {
    bodyHtml = `
      <div class="error-box" role="alert">
        <div><strong>Couldn't load workouts</strong><p class="muted-sm">${esc(state.loadError)}</p></div>
        <button type="button" class="btn-ghost-sm" data-action="retry-load">Retry</button>
      </div>
    `;
  } else if (state.workouts.length === 0) {
    bodyHtml = `
      <div class="empty-box">
        <p style="font-size:2rem;margin:0 0 0.4rem">🐰</p>
        <p style="margin:0 0 0.25rem;font-weight:700;color:var(--text)">No workouts planned yet.</p>
        <p style="margin:0">Ask Claude to add your training plan.</p>
      </div>
    `;
  } else {
    bodyHtml = `
      ${upcoming.length > 0
        ? `<section class="section"><h2 class="section-title">Upcoming</h2><div class="card-stack">${upcoming.map(cardHtml).join('')}</div></section>`
        : `<div class="empty-box">Nothing planned ahead. Ask Claude for the next block.</div>`}
      ${past.length > 0
        ? `<section class="section"><h2 class="section-title">${upcoming.length > 0 ? 'Past' : 'Recent Workouts'}</h2><div class="card-stack ${upcoming.length > 0 ? 'dim' : ''}">${past.map(cardHtml).join('')}</div></section>`
        : ''}
    `;
  }

  return `
    <div class="shell">
      <header class="brand-row row-between">
        <div class="row-gap">
          <div class="brand-badge">${ICONS.dumbbell}</div>
          <h1 class="brand-title">Gym Rabbit</h1>
        </div>
        <button type="button" class="icon-btn accent" data-action="open-barbell" aria-label="Open Barbell Math calculator" title="Barbell Math">${ICONS.barbell}</button>
      </header>
      <main class="main-pad">${resumeHtml}${bodyHtml}</main>
    </div>
  `;
}

function renderWorkout() {
  const workout = activeWorkout();
  if (!workout || !state.session) {
    state.screen = 'dashboard';
    return renderDashboard();
  }

  const blockIdx = Math.min(state.session.currentBlockIdx, workout.blocks.length - 1);
  const block = workout.blocks[blockIdx];
  const blockProgress = state.session.progress[blockIdx];

  let blockAmrapReps = null;
  block.exercises.forEach((ex, ei) => {
    if (/^amrap$/i.test(ex.reps.trim())) {
      const logged = blockProgress[ei].find((s) => s.reps != null && s.reps > 0);
      if (logged) blockAmrapReps = logged.reps;
    }
  });

  const blockPillsHtml = workout.blocks
    .map((b, idx) => {
      const done = isBlockComplete(b, state.session.progress[idx]);
      const cls = ['block-pill', idx === blockIdx ? 'active' : '', done ? 'done' : ''].filter(Boolean).join(' ');
      return `<button type="button" class="${cls}" data-action="goto-block" data-idx="${idx}" aria-label="${esc(b.name)}${done ? ' (completed)' : ''}">${esc(blockLetter(b, idx))}${done ? ' ✓' : ''}</button>`;
    })
    .join('');

  const cardsHtml = block.exercises
    .map((ex, exIdx) => {
      const setsProgress = blockProgress[exIdx];
      const activeSetIdx = getActiveSetIdx(ex, setsProgress);
      const activeSet = setsProgress[activeSetIdx];
      const isComplete = activeSetIdx === -1;
      const anyLogged = setsProgress.some((s) => hasSetProgress(ex, s));
      const key = `${blockIdx}-${exIdx}`;
      const showActuals = Boolean(state.expandedActuals[key]);
      const draft = state.draftActuals[key] ?? { weight: ex.weight_kg != null ? String(ex.weight_kg) : '', reps: '' };

      const pctMatch = ex.reps.match(/(\d+)\s*%\s*amrap/i);
      const amrapHint = pctMatch && blockAmrapReps != null
        ? `<span class="pill hint">Target: ~${Math.max(1, Math.round((blockAmrapReps * parseInt(pctMatch[1], 10)) / 100))} reps (${pctMatch[1]}% of ${blockAmrapReps})</span>`
        : '';

      const loadHtml = ex.weight_kg
        ? `<button type="button" class="pill load" data-action="open-barbell" data-weight="${ex.weight_kg}" aria-label="Calculate barbell plates for ${ex.weight_kg} kg">${ICONS.barbell} ${ex.weight_kg} kg</button>`
        : !ex.load_note ? `<span class="pill load">Bodyweight</span>` : '';

      const cuesHtml = ex.cues?.length && !isComplete
        ? `<div class="cues-box">${ex.cues.map((c) => `<p class="cue-line">• ${renderCueHtml(c, ICONS.play)}</p>`).join('')}</div>`
        : '';

      const segmentsHtml = setsProgress.map((s, i) => {
        const done = isSetComplete(ex, s);
        const partial = !done && hasSetProgress(ex, s);
        const cls = done ? 'done' : partial ? 'partial' : i === activeSetIdx ? 'active' : '';
        return `<div class="progress-seg ${cls}" aria-hidden="true"></div>`;
      }).join('');

      const loggedNotes = setsProgress.map((s, i) => {
        if (!isSetComplete(ex, s) || (s.weightKg == null && s.reps == null)) return null;
        const parts = [];
        if (s.weightKg != null) parts.push(`${s.weightKg}kg`);
        if (s.reps != null) parts.push(`${s.reps}r`);
        return `S${i + 1}: ${parts.join(' × ')}`;
      }).filter(Boolean);

      const controlsHtml = !isComplete || anyLogged
        ? `<div class="ex-controls">
            ${!isComplete ? `<button type="button" class="mini-btn" data-action="toggle-actuals" data-block="${blockIdx}" data-ex="${exIdx}">${showActuals ? 'Hide kg / reps' : 'Log kg / reps'}</button>` : '<span></span>'}
            ${anyLogged ? `<button type="button" class="mini-btn" data-action="undo-set" data-block="${blockIdx}" data-ex="${exIdx}" aria-label="Undo last set of ${esc(ex.name)}">${ICONS.undo} Undo</button>` : ''}
          </div>`
        : '';

      const actualsPanelHtml = showActuals && !isComplete
        ? `<div class="actuals-panel">
            <div>
              <label class="field-label" for="actual-weight-${key}">Weight (kg)</label>
              <input id="actual-weight-${key}" data-actual-key="${key}" data-actual-field="weight" type="number" inputmode="decimal" step="0.5" placeholder="${ex.weight_kg ?? '0'}" value="${esc(draft.weight)}" />
            </div>
            <div>
              <label class="field-label" for="actual-reps-${key}">Reps done</label>
              <input id="actual-reps-${key}" data-actual-key="${key}" data-actual-field="reps" type="number" inputmode="numeric" placeholder="${esc(ex.reps)}" value="${esc(draft.reps)}" />
            </div>
          </div>`
        : '';

      const actionHtml = isComplete
        ? `<div class="completed-banner">${ICONS.check}<span>Completed</span></div>`
        : ex.per_side
          ? `<div class="split-grid">
              <button type="button" class="log-btn" ${activeSet.left ? 'disabled' : ''} data-action="log-set" data-block="${blockIdx}" data-ex="${exIdx}" data-set="${activeSetIdx}" data-side="left" aria-label="Log set ${activeSetIdx + 1} left side for ${esc(ex.name)}"><span class="sub">Set ${activeSetIdx + 1}</span>LEFT</button>
              <button type="button" class="log-btn" ${activeSet.right ? 'disabled' : ''} data-action="log-set" data-block="${blockIdx}" data-ex="${exIdx}" data-set="${activeSetIdx}" data-side="right" aria-label="Log set ${activeSetIdx + 1} right side for ${esc(ex.name)}"><span class="sub">Set ${activeSetIdx + 1}</span>RIGHT</button>
            </div>`
          : `<button type="button" class="log-btn" data-action="log-set" data-block="${blockIdx}" data-ex="${exIdx}" data-set="${activeSetIdx}" aria-label="Log set ${activeSetIdx + 1} for ${esc(ex.name)}"><span class="sub">Set ${activeSetIdx + 1} of ${ex.sets}</span>LOG SET</button>`;

      return `
        <article class="exercise-card ${isComplete ? 'complete' : ''}">
          <div class="ex-top">
            <div>
              <h3 class="ex-title"><span class="ex-label">${esc(exerciseLabel(block, blockIdx, exIdx))}.</span>${esc(ex.name)}</h3>
              <div class="ex-badges">
                <span class="pill">${ex.sets} × ${esc(ex.reps)}${ex.per_side ? ' / side' : ''}</span>
                ${loadHtml}
                ${ex.load_note ? `<span class="pill note">${esc(ex.load_note)}</span>` : ''}
                ${amrapHint}
              </div>
            </div>
            <div class="rest-badge" title="Rest after set">${ICONS.clock}<span>${esc(formatRestBadge(ex.rest_seconds))}</span></div>
          </div>
          ${cuesHtml}
          <div class="progress-track" role="img" aria-label="${setsProgress.filter((s) => isSetComplete(ex, s)).length} of ${ex.sets} sets completed">${segmentsHtml}</div>
          ${loggedNotes.length ? `<div class="logged-summary">${loggedNotes.map((n) => `<span class="pill">${esc(n)}</span>`).join('')}</div>` : ''}
          ${controlsHtml}
          ${actualsPanelHtml}
          ${actionHtml}
        </article>
      `;
    })
    .join('');

  const isLastBlock = blockIdx === workout.blocks.length - 1;
  return `
    <div class="shell">
      <header class="topbar">
        <div class="row-between">
          <div class="row-gap">
            <button type="button" class="icon-btn" data-action="back-dashboard" aria-label="Back to workouts">${ICONS.back}</button>
            <div style="min-width:0">
              <p class="workout-eyebrow">${esc(workout.title)}</p>
              <h1 class="block-heading">${esc(block.name)}</h1>
            </div>
          </div>
          <button type="button" class="icon-btn accent" data-action="open-barbell" aria-label="Open Barbell Math calculator" title="Barbell Math">${ICONS.barbell}</button>
        </div>
        <nav class="block-pills" aria-label="Workout blocks">${blockPillsHtml}</nav>
      </header>
      <main class="main-workout">${cardsHtml}</main>
      <footer class="workout-footer">
        <div class="footer-inner">
          <button type="button" class="nav-btn ${blockIdx === 0 ? 'hidden' : ''}" data-action="prev-block">← Prev</button>
          ${isLastBlock
            ? `<button type="button" class="nav-btn finish" data-action="finish-workout">Finish Session</button>`
            : `<button type="button" class="nav-btn" data-action="next-block">${esc(workout.blocks[blockIdx + 1].name)} →</button>`}
        </div>
      </footer>
    </div>
  `;
}

function renderSummary() {
  const workout = activeWorkout();
  if (!workout || !state.session) {
    state.screen = 'dashboard';
    return renderDashboard();
  }
  const stats = countSessionStats(workout, state.session.progress);
  const elapsedMin = Math.max(1, Math.round((Date.now() - state.session.startedAt) / 60000));

  let funMessage = '';
  if (stats.totalVolume > 0) {
    if (stats.totalVolume < 500) funMessage = "That's equivalent to lifting a grand piano! 🎹";
    else if (stats.totalVolume < 1500) funMessage = "That's equivalent to lifting an adult grizzly bear! 🐻";
    else if (stats.totalVolume < 3000) funMessage = "That's equivalent to lifting a rhinoceros! 🦏";
    else if (stats.totalVolume < 5000) funMessage = "That's equivalent to lifting an entire hippopotamus! 🦛";
    else if (stats.totalVolume < 10000) funMessage = "That's equivalent to lifting an elephant! 🐘";
    else funMessage = "That's equivalent to lifting a commercial airplane! ✈️";
  }

  return `
    <div class="shell">
      <main class="summary-screen">
        <div class="summary-check">${ICONS.check}</div>
        <h1 class="summary-title">Workout Complete</h1>
        <p class="summary-subtitle">${esc(workout.title)}</p>
        <div class="summary-grid">
          <div class="stat-card"><div class="stat-val">${stats.doneSets}/${stats.totalSets}</div><div class="stat-lbl">Sets</div></div>
          <div class="stat-card"><div class="stat-val">${stats.doneExercises}/${stats.totalExercises}</div><div class="stat-lbl">Exercises</div></div>
          <div class="stat-card"><div class="stat-val">${elapsedMin}m</div><div class="stat-lbl">Duration</div></div>
        </div>
        ${stats.totalVolume > 0 ? `<div class="volume-card"><div class="stat-lbl">Total Volume</div><p class="volume-value">${stats.totalVolume.toLocaleString()} kg</p><p class="volume-fun">${esc(funMessage)}</p></div>` : ''}
        <div class="summary-actions">
          <button type="button" class="btn-full primary" data-action="complete-and-return">Finish & Return</button>
          <button type="button" class="btn-full secondary" data-action="return-to-workout">Back to Workout</button>
        </div>
      </main>
    </div>
  `;
}

function renderRest() {
  if (!state.rest) {
    restRoot.innerHTML = '';
    return;
  }
  const remaining = Math.max(0, Math.ceil((state.rest.endAt - Date.now()) / 1000));
  restRoot.innerHTML = `
    <div class="rest-screen" role="dialog" aria-modal="true" aria-label="Rest timer">
      <div class="rest-top">
        <span class="rest-label">Rest Timer</span>
        <button type="button" class="icon-btn accent" data-action="open-barbell" aria-label="Open Barbell Math calculator">${ICONS.barbell}</button>
      </div>
      <div class="rest-center">
        <div id="rest-clock" class="rest-clock">${formatRestClock(remaining)}</div>
        <p class="rest-next">${esc(state.rest.nextText)}</p>
      </div>
      <div class="rest-actions">
        <div class="rest-adjust-row">
          <button type="button" class="rest-adjust-btn" data-action="rest-adjust" data-delta="-15">−15s</button>
          <button type="button" class="rest-adjust-btn" data-action="rest-adjust" data-delta="15">+15s</button>
        </div>
        <button type="button" class="rest-skip-btn" data-action="skip-rest">Skip Rest</button>
      </div>
    </div>
  `;
}

function renderDrawer() {
  const layers = [];
  if (state.drawerOpen) {
    layers.push(`
      <div class="drawer-backdrop" data-action="close-barbell-backdrop" role="dialog" aria-modal="true" aria-label="Barbell Math calculator">
        <div class="drawer-sheet">
          <div class="row-between">
            <h2 class="drawer-title">${ICONS.barbell}<span>Barbell Math</span></h2>
            <button type="button" class="icon-btn" data-action="close-barbell" aria-label="Close Barbell Math">${ICONS.close}</button>
          </div>
          <div class="bar-selector" role="group" aria-label="Barbell weight">
            ${[20, 15, 25].map((bw) => `<button type="button" class="bar-chip ${state.barWeight === bw ? 'active' : ''}" data-action="set-bar-weight" data-weight="${bw}">${bw} kg bar</button>`).join('')}
          </div>
          <div>
            <label class="field-label" for="barbell-weight-input">Target Weight (kg)</label>
            <input id="barbell-weight-input" class="weight-input" type="number" inputmode="decimal" step="0.5" placeholder="e.g. 112.5" value="${esc(state.targetWeight)}" />
          </div>
          <div class="steppers-row">
            <button type="button" class="step-btn" data-action="step-weight" data-delta="-5">−5</button>
            <button type="button" class="step-btn" data-action="step-weight" data-delta="-2.5">−2.5</button>
            <button type="button" class="step-btn" data-action="step-weight" data-delta="2.5">+2.5</button>
            <button type="button" class="step-btn" data-action="step-weight" data-delta="5">+5</button>
          </div>
          <div id="sleeve-result">${renderSleeveResultHtml()}</div>
        </div>
      </div>
    `);
  }

  if (state.confirmAction) {
    const isDiscard = state.confirmAction.kind === 'discard';
    const title = isDiscard ? 'Discard active session?' : 'Replace active session?';
    const currentName = esc(state.session?.workout?.title ?? 'current workout');
    const body = isDiscard
      ? `You have logged sets in <strong>${currentName}</strong>. Discarding will clear your saved progress.`
      : `You have logged sets in <strong>${currentName}</strong>. Starting <strong>${esc(state.confirmAction.workout.title)}</strong> will replace your saved progress.`;
    const confirmLabel = isDiscard ? 'Discard Session' : `Start ${esc(state.confirmAction.workout.title)}`;

    layers.push(`
      <div class="drawer-backdrop center" role="dialog" aria-modal="true" aria-label="${esc(title)}">
        <div class="drawer-sheet modal-card">
          <h2 class="drawer-title">${esc(title)}</h2>
          <p class="muted-sm" style="margin:0;font-size:0.92rem">${body}</p>
          <div style="display:flex;flex-direction:column;gap:0.55rem;margin-top:0.25rem">
            <button type="button" class="btn-full primary" data-action="confirm-modal">${confirmLabel}</button>
            <button type="button" class="btn-full secondary" data-action="cancel-modal">Keep Current Session</button>
          </div>
        </div>
      </div>
    `);
  }

  drawerRoot.innerHTML = layers.join('');
}

function render() {
  if (state.screen === 'workout') appRoot.innerHTML = renderWorkout();
  else if (state.screen === 'summary') appRoot.innerHTML = renderSummary();
  else appRoot.innerHTML = renderDashboard();
  renderRest();
  renderDrawer();
}

document.addEventListener('click', (e) => {
  const target = e.target.closest('[data-action]');
  if (!target) return;
  const action = target.dataset.action;

  switch (action) {
    case 'retry-load':
      fetchWorkouts();
      break;
    case 'start-workout': {
      const w = state.workouts.find((item) => item.id === target.dataset.id);
      if (w) startWorkout(w);
      break;
    }
    case 'resume-session':
      if (activeWorkout()) {
        state.screen = 'workout';
        render();
      }
      break;
    case 'discard-session':
      if (hasAnyLoggedSet(activeWorkout(), state.session?.progress)) {
        state.confirmAction = { kind: 'discard' };
        renderDrawer();
      } else {
        state.session = null;
        saveSession(null);
        render();
      }
      break;
    case 'confirm-modal':
      if (state.confirmAction?.kind === 'replace') startWorkout(state.confirmAction.workout, true);
      else if (state.confirmAction?.kind === 'discard') {
        state.confirmAction = null;
        state.session = null;
        saveSession(null);
        render();
      }
      break;
    case 'cancel-modal':
      state.confirmAction = null;
      renderDrawer();
      break;
    case 'back-dashboard':
      state.screen = 'dashboard';
      render();
      break;
    case 'goto-block':
      if (!state.session) break;
      state.session.currentBlockIdx = Number(target.dataset.idx);
      saveSession(state.session);
      render();
      window.scrollTo({ top: 0, behavior: 'instant' });
      break;
    case 'prev-block':
      if (!state.session || state.session.currentBlockIdx <= 0) break;
      state.session.currentBlockIdx -= 1;
      saveSession(state.session);
      render();
      window.scrollTo({ top: 0, behavior: 'instant' });
      break;
    case 'next-block': {
      const w = activeWorkout();
      if (!w || !state.session || state.session.currentBlockIdx >= w.blocks.length - 1) break;
      state.session.currentBlockIdx += 1;
      saveSession(state.session);
      render();
      window.scrollTo({ top: 0, behavior: 'instant' });
      break;
    }
    case 'log-set':
      logSet(Number(target.dataset.block), Number(target.dataset.ex), Number(target.dataset.set), target.dataset.side || null);
      break;
    case 'undo-set':
      undoLastSet(Number(target.dataset.block), Number(target.dataset.ex));
      break;
    case 'toggle-actuals': {
      const key = `${target.dataset.block}-${target.dataset.ex}`;
      state.expandedActuals[key] = !state.expandedActuals[key];
      if (state.expandedActuals[key] && !state.draftActuals[key]) {
        const w = activeWorkout();
        const ex = w?.blocks[Number(target.dataset.block)]?.exercises[Number(target.dataset.ex)];
        state.draftActuals[key] = { weight: ex?.weight_kg != null ? String(ex.weight_kg) : '', reps: '' };
      }
      render();
      break;
    }
    case 'finish-workout':
      state.screen = 'summary';
      render();
      break;
    case 'return-to-workout':
      state.screen = 'workout';
      render();
      break;
    case 'complete-and-return':
      state.session = null;
      saveSession(null);
      state.screen = 'dashboard';
      render();
      break;
    case 'rest-adjust':
      adjustRestTimer(Number(target.dataset.delta));
      break;
    case 'skip-rest':
      stopRestTimer(true);
      break;
    case 'open-barbell': {
      const w = target.dataset.weight ? parseFloat(target.dataset.weight) : null;
      if (w != null && w > 0) state.targetWeight = String(w);
      state.drawerOpen = true;
      renderDrawer();
      document.getElementById('barbell-weight-input')?.focus();
      break;
    }
    case 'close-barbell':
      state.drawerOpen = false;
      renderDrawer();
      break;
    case 'close-barbell-backdrop':
      if (e.target === target) {
        state.drawerOpen = false;
        renderDrawer();
      }
      break;
    case 'set-bar-weight':
      state.barWeight = Number(target.dataset.weight);
      renderDrawer();
      break;
    case 'step-weight': {
      const base = Number.isNaN(parseFloat(state.targetWeight)) ? state.barWeight : parseFloat(state.targetWeight);
      state.targetWeight = String(Math.max(0, Math.round((base + Number(target.dataset.delta)) * 100) / 100));
      const input = document.getElementById('barbell-weight-input');
      if (input) input.value = state.targetWeight;
      const res = document.getElementById('sleeve-result');
      if (res) res.innerHTML = renderSleeveResultHtml();
      break;
    }
  }
});

document.addEventListener('input', (e) => {
  const el = e.target;
  if (!el) return;
  if (el.id === 'barbell-weight-input') {
    state.targetWeight = el.value;
    const res = document.getElementById('sleeve-result');
    if (res) res.innerHTML = renderSleeveResultHtml();
  } else if (el.dataset.actualKey) {
    const key = el.dataset.actualKey;
    state.draftActuals[key] ??= { weight: '', reps: '' };
    state.draftActuals[key][el.dataset.actualField] = el.value;
  }
});

document.addEventListener('keydown', (e) => {
  if (e.key !== 'Escape') return;
  if (state.drawerOpen) {
    state.drawerOpen = false;
    renderDrawer();
  } else if (state.confirmAction) {
    state.confirmAction = null;
    renderDrawer();
  }
});

let touchStartX = null;
let touchStartY = null;
document.addEventListener('touchstart', (e) => {
  if (state.screen !== 'workout' || state.rest || state.drawerOpen || state.confirmAction) return;
  if (e.target.closest('.block-pills, .sleeve-visual, input')) return;
  touchStartX = e.touches[0].clientX;
  touchStartY = e.touches[0].clientY;
}, { passive: true });

document.addEventListener('touchend', (e) => {
  if (touchStartX == null || state.screen !== 'workout' || state.rest || state.drawerOpen || state.confirmAction) return;
  const dx = touchStartX - e.changedTouches[0].clientX;
  const dy = Math.abs(touchStartY - e.changedTouches[0].clientY);
  touchStartX = null;
  touchStartY = null;
  const w = activeWorkout();
  if (!w || !state.session || Math.abs(dx) <= Math.max(60, dy * 1.5)) return;
  if (dx > 0 && state.session.currentBlockIdx < w.blocks.length - 1) {
    state.session.currentBlockIdx += 1;
  } else if (dx < 0 && state.session.currentBlockIdx > 0) {
    state.session.currentBlockIdx -= 1;
  } else {
    return;
  }
  saveSession(state.session);
  render();
  window.scrollTo({ top: 0, behavior: 'instant' });
}, { passive: true });

fetchWorkouts();

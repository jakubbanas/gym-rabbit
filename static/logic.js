export const PLATE_SIZES = [25, 20, 15, 10, 5, 2.5, 1.25];

export function esc(value) {
  return String(value ?? '')
    .replaceAll('&', '&amp;')
    .replaceAll('<', '&lt;')
    .replaceAll('>', '&gt;')
    .replaceAll('"', '&quot;')
    .replaceAll("'", '&#39;');
}

export function localToday(now = new Date()) {
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
}

export function formatDay(dateStr) {
  const [y, m, d] = dateStr.split('-').map(Number);
  return new Date(y, m - 1, d).toLocaleDateString('en-GB', { weekday: 'short', day: 'numeric', month: 'short' });
}

export function formatTimestamp(ts) {
  if (!ts) return 'Never';
  return new Date(ts).toLocaleDateString('en-GB', { day: 'numeric', month: 'short' });
}

export function formatRestClock(totalSeconds) {
  const s = Math.max(0, Math.ceil(totalSeconds));
  if (s >= 60) {
    return `${Math.floor(s / 60)}:${String(s % 60).padStart(2, '0')}`;
  }
  return String(s);
}

export function formatRestBadge(seconds) {
  if (!seconds || seconds <= 0) return 'No rest';
  if (seconds >= 60 && seconds % 60 === 0) return `${seconds / 60}m`;
  if (seconds >= 60) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  return `${seconds}s`;
}

export function blockLetter(block, blockIdx) {
  return block.name.match(/(?:^|\s)([A-Za-z])$/)?.[1].toUpperCase() ?? String.fromCharCode(65 + blockIdx);
}

export function exerciseLabel(block, blockIdx, exIdx) {
  return `${blockLetter(block, blockIdx)}${exIdx + 1}`;
}

export function initProgress(workout) {
  return workout.blocks.map((b) =>
    b.exercises.map((ex) =>
      Array.from({ length: ex.sets }, () =>
        ex.per_side
          ? { left: false, right: false, weightKg: null, reps: null }
          : { done: false, weightKg: null, reps: null }
      )
    )
  );
}

export function progressMatchesWorkout(progress, workout) {
  if (!workout || !Array.isArray(progress) || progress.length !== workout.blocks.length) return false;
  return workout.blocks.every((b, bi) => {
    const bp = progress[bi];
    if (!Array.isArray(bp) || bp.length !== b.exercises.length) return false;
    return b.exercises.every((ex, ei) => Array.isArray(bp[ei]) && bp[ei].length === ex.sets);
  });
}

export function isSetComplete(ex, setState) {
  return ex.per_side ? Boolean(setState.left && setState.right) : Boolean(setState.done);
}

export function hasSetProgress(ex, setState) {
  return ex.per_side ? Boolean(setState.left || setState.right) : Boolean(setState.done);
}

export function hasAnyLoggedSet(workout, progress) {
  if (!workout || !progressMatchesWorkout(progress, workout)) return false;
  return workout.blocks.some((b, bi) =>
    b.exercises.some((ex, ei) => progress[bi][ei].some((s) => hasSetProgress(ex, s)))
  );
}

export function isBlockComplete(block, blockProgress) {
  return block.exercises.every((ex, ei) => blockProgress[ei].every((s) => isSetComplete(ex, s)));
}

/**
 * Parses planned reps for volume estimation. Returns 0 for percentage targets
 * (e.g. "70% AMRAPa") and timed holds (e.g. "30s", "20-30s").
 */
export function parsePlannedReps(repsStr) {
  const s = String(repsStr ?? '').trim();
  if (!s || s.includes('%') || /^\d+(?:-\d+)?s$/i.test(s)) return 0;
  const m = s.match(/(\d+)/);
  return m ? parseInt(m[1], 10) : 0;
}

export function countSessionStats(workout, progress) {
  let doneSets = 0;
  let totalSets = 0;
  let doneExercises = 0;
  let totalExercises = 0;
  let totalVolume = 0;

  workout.blocks.forEach((b, bi) => {
    b.exercises.forEach((ex, ei) => {
      totalExercises += 1;
      const sets = progress[bi][ei];
      let exDoneSets = 0;
      const plannedWeight = ex.weight_kg ?? 0;
      const plannedReps = parsePlannedReps(ex.reps);

      sets.forEach((s) => {
        totalSets += 1;
        if (isSetComplete(ex, s)) {
          doneSets += 1;
          exDoneSets += 1;
        }
        const w = s.weightKg != null ? s.weightKg : plannedWeight;
        const r = s.reps != null ? s.reps : plannedReps;
        if (w > 0 && r > 0) {
          if (ex.per_side) {
            if (s.left) totalVolume += w * r;
            if (s.right) totalVolume += w * r;
          } else if (s.done) {
            totalVolume += w * r;
          }
        }
      });

      if (exDoneSets === sets.length) doneExercises += 1;
    });
  });

  return { doneSets, totalSets, doneExercises, totalExercises, totalVolume };
}

export function undoSetInArray(ex, sets) {
  for (let i = sets.length - 1; i >= 0; i -= 1) {
    const s = sets[i];
    if (ex.per_side) {
      if (s.right) {
        s.right = false;
        if (!s.left) {
          s.weightKg = null;
          s.reps = null;
        }
        return true;
      }
      if (s.left) {
        s.left = false;
        s.weightKg = null;
        s.reps = null;
        return true;
      }
    } else if (s.done) {
      s.done = false;
      s.weightKg = null;
      s.reps = null;
      return true;
    }
  }
  return false;
}

export function renderCueHtml(cue, playIconHtml = '') {
  const urlRegex = /(https?:\/\/[^\s)]+)/g;
  return String(cue)
    .split(urlRegex)
    .map((part) => {
      if (!/^https?:\/\//i.test(part)) return esc(part);
      const clean = part.replace(/[.,;:!?]+$/, '');
      const trail = part.slice(clean.length);
      const isVideo = /youtu\.?be|vimeo/i.test(clean);
      const label = isVideo ? 'Watch video' : clean.replace(/^https?:\/\//i, '');
      return `<a class="cue-link" href="${esc(clean)}" target="_blank" rel="noopener noreferrer">${isVideo ? playIconHtml : ''}<span>${esc(label)}</span></a>${esc(trail)}`;
    })
    .join('');
}

export function calculatePlates(weightStr, barWeight = 20) {
  const weight = parseFloat(weightStr);
  if (Number.isNaN(weight) || weight <= barWeight) {
    return { plates: [], actualTotal: barWeight };
  }
  let perSide = (weight - barWeight) / 2;
  const plates = [];
  for (const p of PLATE_SIZES) {
    while (perSide >= p - 0.001) {
      plates.push(p);
      perSide -= p;
    }
  }
  const loadedPerSide = plates.reduce((acc, p) => acc + p, 0);
  return { plates, actualTotal: barWeight + loadedPerSide * 2 };
}

export function stepNumberValue(raw, delta, fallback = 0) {
  const parsed = parseFloat(String(raw ?? '').trim());
  const base = Number.isNaN(parsed) ? fallback : parsed;
  const next = Math.max(0, Math.round((base + delta) * 100) / 100);
  return String(next);
}


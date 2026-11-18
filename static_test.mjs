import test from 'node:test';
import assert from 'node:assert/strict';
import {
  blockLetter,
  exerciseLabel,
  initProgress,
  progressMatchesWorkout,
  hasAnyLoggedSet,
  parsePlannedReps,
  countSessionStats,
  undoSetInArray,
  renderCueHtml,
  calculatePlates,
  formatRestClock,
  stepNumberValue,
} from './static/logic.js';

test('exerciseLabel uses trailing block letter or position fallback', () => {
  assert.equal(blockLetter({ name: 'Rozgrzewka R' }, 0), 'R');
  assert.equal(exerciseLabel({ name: 'Blok C' }, 3, 1), 'C2');
  assert.equal(exerciseLabel({ name: 'Superset' }, 1, 0), 'B1');
});

test('parsePlannedReps ignores percentages and timed holds', () => {
  assert.equal(parsePlannedReps('8'), 8);
  assert.equal(parsePlannedReps('8-10'), 8);
  assert.equal(parsePlannedReps('AMRAP'), 0);
  assert.equal(parsePlannedReps('70% AMRAPa'), 0);
  assert.equal(parsePlannedReps('30s'), 0);
  assert.equal(parsePlannedReps('20-30s'), 0);
});

test('countSessionStats uses actuals and skips 70% AMRAP without logged reps', () => {
  const workout = {
    blocks: [
      {
        name: 'Blok C',
        exercises: [
          { name: 'Weighted Chin', sets: 2, reps: '70% AMRAPa', weight_kg: 10 },
          { name: 'Split Squat', sets: 1, reps: '8', per_side: true, weight_kg: 16 },
        ],
      },
    ],
  };
  const progress = initProgress(workout);
  progress[0][0][0] = { done: true, weightKg: 12, reps: 7 };
  progress[0][0][1] = { done: true, weightKg: null, reps: null };
  progress[0][1][0] = { left: true, right: false, weightKg: 18, reps: 8 };

  assert.equal(hasAnyLoggedSet(workout, progress), true);
  const stats = countSessionStats(workout, progress);
  assert.equal(stats.doneSets, 2);
  assert.equal(stats.totalSets, 3);
  assert.equal(stats.doneExercises, 1);
  // 12*7 (set 1) + 0 (set 2 has 70% AMRAPa & null reps) + 18*8 (left side only)
  assert.equal(stats.totalVolume, 84 + 144);
});

test('undoSetInArray rolls back per_side and clears metadata only when empty', () => {
  const ex = { name: 'Split Squat', sets: 2, reps: '8', per_side: true };
  const sets = [
    { left: true, right: true, weightKg: 18, reps: 8 },
    { left: false, right: false, weightKg: null, reps: null },
  ];
  assert.equal(undoSetInArray(ex, sets), true);
  assert.deepEqual(sets[0], { left: true, right: false, weightKg: 18, reps: 8 });
  assert.equal(undoSetInArray(ex, sets), true);
  assert.deepEqual(sets[0], { left: false, right: false, weightKg: null, reps: null });
  assert.equal(undoSetInArray(ex, sets), false);
});

test('renderCueHtml strips trailing punctuation and escapes HTML', () => {
  const html = renderCueHtml('Wideo: (https://youtu.be/abc123). <script>');
  assert.match(html, /href="https:\/\/youtu\.be\/abc123"/);
  assert.match(html, /Watch video/);
  assert.match(html, /&lt;script&gt;/);
  assert.doesNotMatch(html, /abc123\)/);
});

test('calculatePlates handles exact and unachievable targets', () => {
  assert.deepEqual(calculatePlates('112.5', 20), { plates: [25, 20, 1.25], actualTotal: 112.5 });
  assert.deepEqual(calculatePlates('111', 20), { plates: [25, 20], actualTotal: 110 });
  assert.deepEqual(calculatePlates('20', 20), { plates: [], actualTotal: 20 });
});

test('progressMatchesWorkout rejects stale saved progress after plan edits', () => {
  const workout = { blocks: [{ name: 'A', exercises: [{ name: 'Squat', sets: 3, reps: '5' }] }] };
  const progress = initProgress(workout);
  assert.equal(progressMatchesWorkout(progress, workout), true);
  workout.blocks[0].exercises[0].sets = 4;
  assert.equal(progressMatchesWorkout(progress, workout), false);
  assert.equal(formatRestClock(120), '2:00');
  assert.equal(formatRestClock(45), '45');
});

test('stepNumberValue uses fallback when blank, clamps at 0, and rounds decimals', () => {
  assert.equal(stepNumberValue('', 1, 7), '8');
  assert.equal(stepNumberValue('', -1, 7), '6');
  assert.equal(stepNumberValue('80', 2.5, 20), '82.5');
  assert.equal(stepNumberValue('1', -2.5, 0), '0');
});


/**
 * Plays a short beep sound using the Web Audio API to signal rest timer completion.
 */
export function playBeep(): void {
  try {
    const AudioCtx = window.AudioContext || (window as any).webkitAudioContext;
    const ctx = new AudioCtx();
    const osc = ctx.createOscillator();
    const gain = ctx.createGain();
    osc.connect(gain);
    gain.connect(ctx.destination);
    osc.type = 'sine';
    osc.frequency.setValueAtTime(800, ctx.currentTime);
    gain.gain.setValueAtTime(1, ctx.currentTime);
    gain.gain.exponentialRampToValueAtTime(0.001, ctx.currentTime + 0.5);
    osc.start();
    osc.stop(ctx.currentTime + 0.5);
  } catch (e) {
    console.error('Audio playback failed', e);
  }
}

/**
 * Triggers a short haptic vibration if the Vibration API is available.
 */
export function triggerHaptic(): void {
  if (navigator.vibrate) {
    navigator.vibrate(50);
  }
}

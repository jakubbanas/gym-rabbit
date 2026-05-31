import { useEffect } from 'react';
import { playBeep, triggerHaptic } from '../lib/audio';

interface RestTimerProps {
  seconds: number;
  onTick: () => void;
  onSkip: () => void;
}

export default function RestTimer({ seconds, onTick, onSkip }: RestTimerProps) {
  useEffect(() => {
    if (seconds <= 0) {
      playBeep();
      triggerHaptic();
      onSkip();
      return;
    }

    const timer = setInterval(() => {
      onTick();
    }, 1000);

    return () => clearInterval(timer);
  }, [seconds, onTick, onSkip]);

  return (
    <div
      className="fixed inset-0 bg-black text-white flex flex-col items-center justify-center z-50 cursor-pointer"
      onClick={onSkip}
    >
      <div className="text-[12rem] font-black tracking-tighter tabular-nums leading-none text-indigo-500 drop-shadow-[0_0_40px_rgba(99,102,241,0.4)]">
        {seconds}
      </div>
      <p className="text-neutral-400 mt-8 text-lg font-medium opacity-70 animate-pulse">
        Tap anywhere to skip
      </p>
    </div>
  );
}

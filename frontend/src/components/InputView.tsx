import { useState } from 'react';
import { ArrowLeft, Dumbbell } from 'lucide-react';

interface InputViewProps {
  initialText: string;
  onBack: () => void;
  onStart: (text: string) => void;
}

export default function InputView({ initialText, onBack, onStart }: InputViewProps) {
  const [planText, setPlanText] = useState(initialText);
  const [errorMsg, setErrorMsg] = useState('');

  const handleStart = () => {
    if (!planText.trim()) return;
    try {
      onStart(planText);
    } catch (err: any) {
      setErrorMsg(err.message || 'Unknown parsing error');
    }
  };

  return (
    <div className="min-h-screen bg-neutral-950 text-neutral-200 p-6 font-sans flex flex-col max-w-md mx-auto">
      <div className="flex items-center gap-3 mb-6 pt-4">
        <button
          onClick={onBack}
          className="p-2 -ml-2 bg-transparent text-neutral-400 active:text-white transition-colors"
        >
          <ArrowLeft size={24} />
        </button>
        <div className="p-2 bg-indigo-500 rounded-xl text-white">
          <Dumbbell size={24} />
        </div>
        <h1 className="text-2xl font-black tracking-tight">Gym Rabbit</h1>
      </div>

      <p className="text-neutral-400 mb-4 text-sm">
        Review or paste your workout plan using the standard magic-paste format.
      </p>

      <textarea
        className="flex-1 w-full bg-neutral-900 border border-neutral-800 rounded-2xl p-4 text-neutral-300 font-mono text-sm focus:outline-none focus:border-indigo-500 focus:ring-1 focus:ring-indigo-500 mb-6 resize-none min-h-[300px]"
        value={planText}
        onChange={(e) => setPlanText(e.target.value)}
        placeholder={`Plan Name: My Routine\n\nBlock A\nA1. Squat | 3x8 | 100kg | 90s`}
      />

      <button
        onClick={handleStart}
        disabled={!planText.trim()}
        className="w-full bg-indigo-600 disabled:bg-neutral-800 disabled:text-neutral-500 active:bg-indigo-700 text-white font-bold text-lg py-5 rounded-2xl shadow-lg transition-transform active:scale-95"
      >
        Start Session
      </button>

      {/* Error Popup */}
      {errorMsg && (
        <div className="fixed inset-0 bg-black/80 flex items-center justify-center p-6 z-50">
          <div className="bg-neutral-900 border border-neutral-700 rounded-3xl p-6 w-full max-w-sm shadow-2xl">
            <h3 className="text-xl font-bold text-red-400 mb-2">Format Error</h3>
            <p className="text-neutral-300 mb-6 whitespace-pre-wrap text-sm">{errorMsg}</p>
            <button
              onClick={() => setErrorMsg('')}
              className="w-full bg-neutral-800 text-white font-bold py-3 rounded-xl active:bg-neutral-700"
            >
              Got it
            </button>
          </div>
        </div>
      )}
    </div>
  );
}

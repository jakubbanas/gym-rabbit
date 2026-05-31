import { useState } from 'react';
import { Wrench, X } from 'lucide-react';

const PLATE_STYLES: Record<number, string> = {
  25: 'bg-red-600 h-24 text-white',
  20: 'bg-blue-600 h-24 text-white',
  15: 'bg-yellow-500 h-20 text-neutral-900',
  10: 'bg-green-600 h-16 text-white',
  5: 'bg-neutral-100 h-12 text-neutral-900',
  2.5: 'bg-neutral-700 h-10 text-white',
  1.25: 'bg-neutral-500 h-8 text-white',
};

function calculatePlates(weightStr: string): number[] {
  const weight = parseFloat(weightStr);
  if (isNaN(weight) || weight <= 20) return [];
  let perSide = (weight - 20) / 2;
  const plates: number[] = [];
  const availablePlates = [25, 20, 15, 10, 5, 2.5, 1.25];

  for (const p of availablePlates) {
    while (perSide >= p - 0.001) {
      plates.push(p);
      perSide -= p;
    }
  }
  return plates;
}

export default function ToolsDrawer() {
  const [isOpen, setIsOpen] = useState(false);
  const [targetWeight, setTargetWeight] = useState('');

  const plates = calculatePlates(targetWeight);
  const weight = parseFloat(targetWeight);
  const showPlates = targetWeight && weight > 20;

  return (
    <>
      {/* Floating Tools Button */}
      <button
        onClick={(e) => {
          e.stopPropagation();
          setIsOpen(true);
        }}
        className="fixed bottom-28 right-6 w-14 h-14 bg-indigo-600 text-white rounded-full flex items-center justify-center shadow-[0_4px_20px_rgba(79,70,229,0.5)] z-[60] active:scale-95 transition-transform"
      >
        <Wrench size={24} />
      </button>

      {/* Drawer */}
      {isOpen && (
        <div
          className="fixed inset-0 z-[70] flex flex-col justify-end"
          onClick={(e) => e.stopPropagation()}
        >
          {/* Backdrop */}
          <div
            className="absolute inset-0 bg-black/70 backdrop-blur-sm"
            onClick={() => setIsOpen(false)}
          />

          {/* Drawer Content */}
          <div className="bg-neutral-900 w-full rounded-t-3xl p-6 relative z-10 shadow-2xl border-t border-neutral-800">
            <div className="flex justify-between items-center mb-6">
              <h2 className="text-xl font-bold text-white flex items-center gap-2">
                <Wrench size={20} /> Barbell Math
              </h2>
              <button
                onClick={() => setIsOpen(false)}
                className="text-neutral-400 p-2 active:bg-neutral-800 rounded-full transition-colors"
              >
                <X size={24} />
              </button>
            </div>

            <div className="space-y-6">
              <div>
                <label className="text-sm font-bold text-neutral-400 uppercase tracking-wider mb-2 block">
                  Target Weight (kg)
                </label>
                <input
                  type="number"
                  inputMode="decimal"
                  value={targetWeight}
                  onChange={(e) => setTargetWeight(e.target.value)}
                  placeholder="e.g. 112.5"
                  className="w-full bg-neutral-950 border border-neutral-800 rounded-xl p-4 text-2xl font-black text-white focus:outline-none focus:border-indigo-500"
                />
                <p className="text-xs text-neutral-500 mt-2">*Assumes standard 20kg barbell.</p>
              </div>

              {/* Fixed height container to prevent drawer jumping */}
              <div className="h-[250px] w-full flex flex-col">
                {showPlates ? (
                  <div className="bg-neutral-950 rounded-xl p-5 border border-neutral-800 h-full flex flex-col">
                    <p className="text-sm font-bold text-neutral-400 mb-4 shrink-0">
                      Load per side:
                    </p>
                    <div className="flex items-center justify-start overflow-x-auto pb-4 flex-1">
                      {/* Bar collar */}
                      <div className="h-16 w-4 bg-neutral-700 rounded-r border-r border-neutral-600 shrink-0" />
                      {/* Sleeve start */}
                      <div className="h-8 w-2 bg-neutral-500 shrink-0" />
                      {/* Plates */}
                      {plates.map((p, i) => (
                        <div
                          key={i}
                          className={`w-6 flex shrink-0 rounded-sm items-center justify-center text-[10px] font-bold border border-black/50 mx-[1px] shadow-sm ${PLATE_STYLES[p]}`}
                        >
                          <span className="-rotate-90">{p}</span>
                        </div>
                      ))}
                      {/* Sleeve end */}
                      <div className="h-8 w-24 bg-neutral-500 rounded-r-sm shrink-0 border-l border-neutral-600" />
                    </div>
                    <div className="flex flex-wrap gap-2 mt-auto shrink-0 overflow-y-auto max-h-[60px]">
                      {plates.map((p, i) => (
                        <span
                          key={i}
                          className="px-3 py-1.5 bg-neutral-800 rounded-lg text-sm font-bold text-neutral-200"
                        >
                          {p} kg
                        </span>
                      ))}
                    </div>
                  </div>
                ) : (
                  <div className="bg-indigo-900/20 text-indigo-400 p-4 rounded-xl text-center font-bold border border-indigo-500/20 h-full flex flex-col items-center justify-center">
                    <span className="text-4xl mb-3">🏋️</span>
                    {targetWeight && weight <= 20
                      ? 'Just the empty bar!'
                      : 'Enter a weight to calculate plates'}
                  </div>
                )}
              </div>
            </div>
            {/* Spacer for bottom safe area */}
            <div className="h-8" />
          </div>
        </div>
      )}
    </>
  );
}

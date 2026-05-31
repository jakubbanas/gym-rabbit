export interface Exercise {
  id: string;
  name: string;
  sets: number;
  reps: string;
  weight: string;
  isUnilateral: boolean;
  rest: number;
  cues: string[];
}

export interface Block {
  name: string;
  exercises: Exercise[];
}

export interface WorkoutPlan {
  name: string;
  blocks: Block[];
}

/**
 * Parses the "Magic Paste" plain-text format into a structured WorkoutPlan.
 *
 * Template:
 *   Plan Name: [title]
 *   Block [A-Z]
 *   [ID]. [Name] | [Sets]x[Reps] | [Weight] | [Rest]
 *   - [cue line]
 */
export function parseWorkout(text: string): WorkoutPlan {
  const lines = text.split('\n').map((l) => l.trim());
  const plan: WorkoutPlan = { name: 'Gym Session', blocks: [] };
  let currentBlock: Block | null = null;

  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    if (!line) continue;

    if (line.startsWith('Plan Name:')) {
      plan.name = line.replace('Plan Name:', '').trim();
    } else if (line.startsWith('Block')) {
      currentBlock = { name: line, exercises: [] };
      plan.blocks.push(currentBlock);
    } else if (line.startsWith('-')) {
      if (currentBlock && currentBlock.exercises.length > 0) {
        const lastEx = currentBlock.exercises[currentBlock.exercises.length - 1];
        lastEx.cues.push(line.substring(1).trim());
      }
    } else {
      // Exercise line parsing
      const parts = line.split('|').map((p) => p.trim());
      if (parts.length === 4) {
        if (!currentBlock) {
          currentBlock = { name: 'Block A', exercises: [] };
          plan.blocks.push(currentBlock);
        }

        const [idName, setsReps, weight, rest] = parts;

        const nameMatch = idName.match(/^([A-Za-z0-9]+)\.\s+(.*)$/);
        const id = nameMatch ? nameMatch[1] : `Ex${currentBlock.exercises.length + 1}`;
        const name = nameMatch ? nameMatch[2] : idName;

        const srMatch = setsReps.match(/(\d+)x(.*)/);
        const sets = srMatch ? parseInt(srMatch[1]) : 3;
        const reps = srMatch ? srMatch[2] : setsReps;
        const isUnilateral = reps.includes('+');

        let restSeconds = 60;
        const rMatch = rest.match(/(\d+)/);
        if (rMatch) restSeconds = parseInt(rMatch[1]);

        currentBlock.exercises.push({
          id,
          name,
          sets,
          reps,
          weight,
          isUnilateral,
          rest: restSeconds,
          cues: [],
        });
      } else {
        throw new Error(
          `Formatting error on line: "${line}"\nMust be exactly 4 parts separated by pipes (|).`
        );
      }
    }
  }

  if (plan.blocks.length === 0) {
    throw new Error("No workout blocks found. Make sure to use 'Block A' to define sections.");
  }

  return plan;
}

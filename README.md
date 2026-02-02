# sv

Everything you need to build a Svelte project, powered by [`sv`](https://github.com/sveltejs/cli).

## Creating a project

If you're seeing this, you've probably already done this step. Congrats!

```sh
# create a new project
npx sv create my-app
```

To recreate this project with the same configuration:

```sh
# recreate this project
npx sv create --template minimal --types ts --add vitest="usages:unit,component" --install npm gym-rabbit
```

## Developing

Once you've created a project and installed dependencies with `npm install` (or `pnpm install` or `yarn`), start a development server:

```sh
npm run dev

# or start the server and open the app in a new browser tab
npm run dev -- --open
```

## Building

To create a production version of your app:

```sh
npm run build
```

You can preview the production build with `npm run preview`.

> To deploy your app, you may need to install an [adapter](https://svelte.dev/docs/kit/adapters) for your target environment.

## AI

prompt:
```
You'll receive a training plan in pdf format from my personal trainer. 
Analyze only "Siłowy" part.

When there is "C1: (90” po c2)" that means there should be 30s break in c2 block, not in c1. So if there is "C2: (30-60” po c1)" there should be 90s in c1.

When there is no weight expressed in kg, use RPE as weight (eg. RPE8, RPE5-7 etc)

3x6, 3x6-9 means <sets>x<reps>
3x6 = 3 sets and 6 reps
3x6-8 = 3 sets and 6 to 8 reps

3x 5+5 means the exercise is unilateral, 5 reps for each leg
```

schema:
```
{
  "type": "object",
  "properties": {
    "id": {
      "type": "string",
      "description": "Unique identifier for the workout plan, typically in the format 'plan-DD-MM' where DD-MM represents the start date. This ID is used to track progress and reference the plan throughout the application. Examples: 'plan-19-01', 'plan-26-01'",
      "pattern": "^plan-[0-9]{2}-[0-9]{2}$"
    },
    "name": {
      "type": "string",
      "description": "Human-readable name describing the workout plan's focus or training goal. This helps users quickly identify the purpose of the training program. Examples: 'Strength Building', 'Hypertrophy Focus', 'Endurance Training'",
      "minLength": 1
    },
    "startDate": {
      "type": "string",
      "description": "The date when this workout plan begins or began, formatted as 'DD.MM' (day.month). This allows users to track which plan corresponds to which training period. Examples: '19.01', '26.01', '05.03'",
      "pattern": "^[0-9]{2}\\.[0-9]{2}$"
    },
    "blocks": {
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "block": {
            "type": "string",
            "description": "Single letter identifier for the training block. Blocks are typically performed in alphabetical order, with 'A' being the highest intensity/priority exercises and later blocks containing lower intensity or accessory work. Examples: 'A', 'B', 'C', 'D'",
            "pattern": "^[A-Z]$"
          },
          "exercises": {
            "type": "array",
            "items": {
              "type": "object",
              "properties": {
                "id": {
                  "type": "string",
                  "description": "Unique identifier for the exercise within the plan, typically formatted as '{block}{number}' (e.g., 'a1', 'b2'). Optionally includes version suffix (e.g., 'a1-v2') when plans are updated. This ID is used to track completion progress for each exercise. Examples: 'a1', 'b2', 'c1-v2', 'd2-v3'",
                  "pattern": "^[a-z][0-9](-v[0-9]+)?$"
                },
                "name": {
                  "type": "string",
                  "description": "The full name of the exercise as it should be displayed to the user. Should be specific enough to clearly identify the movement. Examples: 'Barbell Back Squat', 'Romanian Deadlift', 'Bulgarian Split Squat', 'Walking Lunges'",
                  "minLength": 1
                },
                "sets": {
                  "type": "integer",
                  "description": "The number of sets to be performed for this exercise. Each set consists of the prescribed number of repetitions followed by a rest period. Typical values: 3, 4, or 5 sets",
                  "minimum": 1,
                  "maximum": 10
                },
                "reps": {
                  "type": "string",
                  "description": "The target number of repetitions to perform in each set. Can be a specific number ('12'), a range ('8-10'), or a description for non-standard rep schemes ('20 steps', 'AMRAP'). Ranges indicate progressive overload targets. Examples: '8-10', '12-15', '6-8', '20 steps', '12', 'AMRAP'",
                  "minLength": 1
                },
                "weight": {
                  "type": "string",
                  "description": "The prescribed weight or resistance to be used for this exercise. Can include units (kg, lbs) and equipment type (DBs for dumbbells, KBs for kettlebells). Optional because some exercises use body weight only. Examples: '80kg', '60kg', '15kg DBs', '24kg', 'bodyweight'"
                },
                "rest": {
                  "type": "integer",
                  "description": "Rest period in seconds between sets of this exercise. Longer rest periods (90-120s) are typical for heavy compound movements, while shorter periods (30-60s) are common for isolation or accessory exercises. Optional field, defaults may be assumed if not specified. Examples: 30, 45, 60, 90, 120",
                  "maximum": 600
                },
                "notes": {
                  "type": "string",
                  "description": "Additional coaching cues, form reminders, or execution details for the exercise. Used to provide technique tips, tempo instructions, or other important information to ensure proper exercise execution. Examples: 'Focus on depth and controlled descent', 'Keep back neutral, hinge at hips', 'Maintain upright torso', '3 second eccentric'"
                },
                "isUnilateral": {
                  "type": "boolean",
                  "description": "Indicates whether this exercise is performed one limb at a time (true) or bilaterally (false/undefined). When true, the application tracks left and right side completion separately to ensure balanced training. Examples of unilateral exercises include Bulgarian split squats, single-leg press, or walking lunges. Default: false"
                }
              },
              "propertyOrdering": [
                "id",
                "name",
                "sets",
                "reps",
                "weight",
                "rest",
                "notes",
                "isUnilateral"
              ],
              "required": [
                "id",
                "name",
                "sets",
                "reps"
              ]
            },
            "description": "List of exercises to be performed within this training block. Exercises within a block are typically related by movement pattern, muscle group, or training intensity.",
            "minItems": 1
          }
        },
        "propertyOrdering": [
          "block",
          "exercises"
        ],
        "required": [
          "block",
          "exercises"
        ]
      },
      "description": "Collection of training blocks that organize exercises into logical groups. Blocks typically represent different intensities, movement patterns, or exercise categories (e.g., main lifts vs. accessories). They are usually performed in sequence during a workout session.",
      "minItems": 1
    }
  },
  "propertyOrdering": [
    "id",
    "name",
    "startDate",
    "blocks"
  ],
  "required": [
    "id",
    "name",
    "startDate",
    "blocks"
  ],
  "description": "A structured workout plan containing multiple training blocks with exercises organized by difficulty or movement pattern. Each plan represents a complete training program with a specific focus (e.g., strength building, hypertrophy)."
}

```

import type { WorkoutPlan } from "./types";

export const workoutPlans: WorkoutPlan[] = [
  {
    id: "plan-jakub-4.5-strength",
    name: "Jakub Banaś 4.5 Siłowy",
    startDate: "01.01",
    blocks: [
      {
        block: "A",
        exercises: [
          {
            id: "a1",
            name: "SSB (Ta dziwna sztanga z uchwytami) Bulgarian split squat",
            sets: 3,
            reps: "6+6",
            weight: "RPE8",
            rest: 105,
            notes: 'RPE 8 na słabszej nodze. Przerwa między nogami 90-120"',
            isUnilateral: true,
          },
        ],
      },
      {
        block: "B",
        exercises: [
          {
            id: "b1",
            name: "Pendlay row – wiosłowanie z martwgo punktu",
            sets: 3,
            reps: "6",
            weight: "RPE8",
            rest: 90,
            notes: "Każde powtórzenie z nowym napięciem.",
          },
        ],
      },
      {
        block: "C",
        exercises: [
          {
            id: "c1",
            name: "Leg extension – maszyna izolowana",
            sets: 3,
            reps: "6",
            weight: "RPE10",
            rest: 45,
            notes:
              '(30-60" po c1) Rest after C1 is 45s (average of 30-60s) before C2. Superset with C2.',
          },
          {
            id: "c2",
            name: "SLDL – taki RDL, ale nogi 100% proste",
            sets: 3,
            reps: "6-8",
            weight: "RPE6-7",
            rest: 90,
            notes:
              'Rób hantlami. Dokładny ruch ważniejszy od ciężaru. (90" po c2) Rest after C2 is 90s before C1.',
          },
        ],
      },
      {
        block: "D",
        exercises: [
          {
            id: "d1",
            name: "Weighted dips",
            sets: 3,
            reps: "6-8",
            weight: "RPE8",
            rest: 30,
            notes:
              "Rest after D1 is 30s (transition) before D2. Superset with D2.",
          },
          {
            id: "d2",
            name: "Uginanie ramion ze sztangą łamaną (Barbell Curl)",
            sets: 3,
            reps: "5-6",
            weight: "RPE9",
            rest: 90,
            notes: '(90" po D2) Rest after D2 is 90s before D1.',
          },
        ],
      },
      {
        block: "E",
        exercises: [
          {
            id: "e1",
            name: "Coppenhaggen adduction",
            sets: 3,
            reps: "5+5",
            weight: "RPE8-9",
            rest: 60,
            notes: "Jak za lekko kładź ciężar na biodro.",
            isUnilateral: true,
          },
        ],
      },
    ],
  },
  {
    id: "plan-jakub-5.0-strength",
    name: "Jakub Banaś 5.0 Siłowy",
    startDate: "16.02",
    blocks: [
      {
        block: "A",
        exercises: [
          {
            id: "a1",
            name: "SSB (sztanga z rogami) Bulgarian split squat",
            sets: 3,
            reps: "6+6",
            weight: "45 kg",
            rest: 90,
            notes: 'RPE 8 na słabszej nodze. Przerwa między nogami 90-120"',
            isUnilateral: true,
          },
        ],
      },
      {
        block: "B",
        exercises: [
          {
            id: "b1",
            name: "Pendlay row - wiosłowanie z martwego punktu",
            sets: 3,
            reps: "6",
            weight: "55 kg",
            rest: 90,
            notes: "Każde powtórzenie z nowym napięciem.",
          },
        ],
      },
      {
        block: "C",
        exercises: [
          {
            id: "c1",
            name: "Leg extension - maszyna izolowana",
            sets: 3,
            reps: "6",
            weight: "75 kg",
            rest: 30,
            notes: "",
          },
          {
            id: "c2",
            name: "SLDL - taki RDL, ale nogi 100% proste",
            sets: 3,
            reps: "6",
            weight: "52.5 kg",
            rest: 90,
            notes:
              'Rób hantlami. Dokładny ruch ważniejszy od ciężaru. (90" po c2) Rest after C2 is 90s before C1.',
          },
        ],
      },
      {
        block: "D",
        exercises: [
          {
            id: "d1",
            name: "Weighted dips",
            sets: 3,
            reps: "6",
            weight: "5 kg",
            rest: 30,
            notes:
              "Rest after D1 is 30s (transition) before D2. Superset with D2.",
          },
          {
            id: "d2",
            name: "Uginanie ramion ze sztangą łamaną (Barbell Curl)",
            sets: 3,
            reps: "5-6",
            weight: "32.5 kg",
            rest: 90,
            notes: '(90" po D2) Rest after D2 is 90s before D1.',
          },
        ],
      },
      {
        block: "E",
        exercises: [
          {
            id: "e1",
            name: "Coppenhaggen adduction",
            sets: 3,
            reps: "7+7",
            weight: "bodyweight",
            rest: 60,
            notes: "Jak za lekko kładź ciężar na biodro.",
            isUnilateral: true,
          },
        ],
      },
    ],
  },
];

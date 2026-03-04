import type { WorkoutPlan } from "./types";

export const workoutPlans: WorkoutPlan[] = [
    {
        id: "plan-paulina-strength",
        name: "Paulina Siłowy",
        startDate: "4.03",
        blocks: [
            {
                block: "A",
                exercises: [
                    {
                        id: "a1",
                        name: "Trap bar deadlift",
                        sets: 5,
                        reps: "5",
                        weight: "?",
                        rest: 30,
                        notes: "przerwy do pełnego wypoczynku."
                    },
                    {
                        id: "a2",
                        name: "High jump",
                        sets: 3,
                        reps: "3",
                        weight: "5kg",
                        rest: 120,
                        notes: "przerwy do pełnego wypoczynku."
                    }
                ]
            },
            {
                block: "B",
                exercises: [
                    {
                        id: "b1",
                        name: "Trap bar pendlay row",
                        sets: 3,
                        reps: "6",
                        weight: "25 kg",
                        rest: 120,
                        notes: 'sztangę ustaw na stepach',
                    },
                ],
            },
            {
                block: "C",
                exercises: [
                    {
                        id: "c1",
                        name: "Przysiad bułgarski",
                        sets: 3,
                        reps: "6+6",
                        weight: "8 kg",
                        rest: 90,
                        notes: 'z pauzami na dole 3x 6+6 - z kettlami 8kg',
                        isUnilateral: true,
                    },
                ],
            },
            {
                block: "D",
                exercises: [
                    {
                        id: "d1",
                        name: "Copenhaggen plank",
                        sets: 3,
                        reps: "20\"",
                        weight: "-",
                        rest: 60,
                        notes: "jak za lekko kładź ciężar na biodro."
                    },
                    {
                        id: "d2",
                        name: "Face pulls leżąc",
                        sets: 3,
                        reps: "6-8",
                        weight: "3kg",
                        rest: 120,
                        notes: "na brzuchu hantlami."
                    }
                ]
            },
            {
                block: "E",
                exercises: [
                    {
                        id: "e1",
                        name: "Maszyna izolowana na dwójkę",
                        sets: 3,
                        reps: "6",
                        weight: "RPE9",
                        rest: 90,
                        notes: 'najlepiej siedząca',
                    },
                ],
            },
        ]
    },
    {
        id: "plan-jakub-5.0-strength",
        name: "Kuba 5.0 Siłowy",
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
                        weight: "55 kg",
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
                        weight: "57.5 kg",
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
                        weight: "80 kg",
                        rest: 30,
                        notes: "",
                    },
                    {
                        id: "c2",
                        name: "SLDL - taki RDL, ale nogi 100% proste",
                        sets: 3,
                        reps: "6",
                        weight: "55 kg",
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
                        weight: "35 kg",
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
                        weight: "7.5 kg",
                        rest: 60,
                        notes: "Jak za lekko kładź ciężar na biodro.",
                        isUnilateral: true,
                    },
                ],
            },
        ],
    },
    {
        id: "marcin-1.0-strength",
        name: "Marcin 1.0 Siłowy",
        startDate: "19.02",
        blocks: [
            {
                block: "A",
                exercises: [
                    {
                        id: "a1",
                        name: "Przysiad do skrzyni (Box Squat)",
                        sets: 3,
                        reps: "10 - 12",
                        weight: "8 - 12 kg",
                        rest: 90,
                        notes: 'Zacznij bez obciążenia, skup się na technice. Skrzynia powinna być na wysokości, która pozwala na utrzymanie napięcia w mięśniach przez cały ruch.',
                    },
                ],
            },
            {
                block: "B",
                exercises: [
                    {
                        id: "b1",
                        name: "Wiosłowanie sztangą",
                        sets: 3,
                        reps: "10 - 12",
                        weight: "25 - 35 kg",
                        rest: 60,
                        notes: "",
                    },
                ],
            },
            {
                block: "C",
                exercises: [
                    {
                        id: "c1",
                        name: "Wyciskanie (maszyna)",
                        sets: 3,
                        reps: "10 - 12",
                        weight: "20 - 30 kg",
                        rest: 60,
                        notes: "",
                    },
                ],
            },
            {
                block: "D",
                exercises: [
                    {
                        id: "d1",
                        name: "Martwy ciąg (Kettle)",
                        sets: 3,
                        reps: "6",
                        weight: "5 kg",
                        rest: 90,
                        notes: "",
                    },
                ],
            },
            {
                block: "E",
                exercises: [
                    {
                        id: "e1",
                        name: "Plank",
                        sets: 3,
                        reps: "RPE10",
                        weight: "bodyweight",
                        rest: 60,
                        notes: "Wytrzymaj 15-20 sekund.",
                    },
                ],
            },
        ],
    },
];

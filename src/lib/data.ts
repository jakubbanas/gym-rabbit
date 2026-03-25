import type { WorkoutPlan } from "./types";

export const workoutPlans: WorkoutPlan[] = [
    {
        id: "plan-paulina-2-strength",
        name: "Paulina 2.0 Siłowy",
        startDate: "21.03",
        blocks: [
            {
                block: "A",
                exercises: [
                    {
                        id: "a1",
                        name: "Trap bar deadlift",
                        sets: 4,
                        reps: "3",
                        weight: "40; 50; 55; 60kg",
                        rest: 180,
                        notes: "przerwy do pełnego wypoczynku."
                    },
                    {
                        id: "a2",
                        name: "High jump",
                        sets: 3,
                        reps: "3",
                        weight: "5kg",
                        rest: 120,
                        notes: "30\" po A1 - najlepiej łączone z trzema ostatnimi seriami - hantle 5kg - MAKSYMALNA INTENCJA."
                    }
                ]
            },
            {
                block: "B",
                exercises: [
                    {
                        id: "b1",
                        name: "Trap bar pendlay row",
                        sets: 4,
                        reps: "6",
                        weight: "27.5kg",
                        rest: 120,
                        notes: "sztangę ustaw na stepach - Przerwa między seriami 90-120\""
                    }
                ]
            },
            {
                block: "C",
                exercises: [
                    {
                        id: "c1",
                        name: "Przysiad bułgarski",
                        sets: 3,
                        reps: "7+7",
                        weight: "10kg",
                        rest: 90,
                        notes: "z pauzami na dole - z hantlami 10kg (Przerwa między nogami 90\")",
                        isUnilateral: true
                    }
                ]
            },
            {
                block: "D",
                exercises: [
                    {
                        id: "d1",
                        name: "Copenhaggen plank",
                        sets: 3,
                        reps: "20\"+20\"",
                        weight: "4kg",
                        rest: 60,
                        notes: "+4kg na biodrze"
                    },
                    {
                        id: "d2",
                        name: "Face pulls leżąc",
                        sets: 3,
                        reps: "8",
                        weight: "3kg",
                        rest: 120,
                        notes: "na brzuchu hantlami - hantle 3kg. (Przerwa po całości do 120\")"
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
                        reps: "5-6",
                        weight: "35kg",
                        rest: 90,
                        notes: "najlepiej siedząca."
                    }
                ]
            }
        ]
    },
    {
        id: "plan-jakub-6.0-strength",
        name: "Kuba 6.0 Siłowy",
        startDate: "16.03",
        blocks: [
            {
                block: "A",
                exercises: [
                    {
                        id: "a1",
                        name: "Trap bar deadlift",
                        sets: 1,
                        reps: "do 4RM",
                        weight: "?",
                        rest: 30,
                        notes: "Przerwa między seriami do pełnego wypoczynku.",
                    },
                    {
                        id: "a2",
                        name: "High jump",
                        sets: 3,
                        reps: "3",
                        weight: "10 kg",
                        rest: 90,
                        notes: '30" po A1 (najlepiej łączone z trzema ostatnimi seriami A1).',
                    },
                ],
            },
            {
                block: "B",
                exercises: [
                    {
                        id: "b1",
                        name: "Chin up (podchwyt)",
                        sets: 3,
                        reps: "3",
                        weight: "7,5 kg",
                        rest: 0,
                        notes: "",
                    },
                ],
            },
            {
                block: "C",
                exercises: [
                    {
                        id: "c1",
                        name: "Single leg seated leg curl",
                        sets: 3,
                        reps: "6+6",
                        weight: "RPE9",
                        rest: 30,
                        notes:
                            'RPE9 na słabszej nodze. (30-60" po C1) C2. Przed seriami docelowymi zrób 1 serię wstępną na mniejszym ciężarze. W obu ćwiczeniach potrzebuję nagranie 2/3 serii głównej.',
                        isUnilateral: true,
                    },
                    {
                        id: "c2",
                        name: "Foot dorsiflexion z KB",
                        sets: 3,
                        reps: "do RPE9",
                        weight: "?",
                        rest: 90,
                        notes: '30-60" po C1. RPE9 na obu nogach. (90" po C2) wracasz do C1.',
                    },
                ],
            },
            {
                block: "D",
                exercises: [
                    {
                        id: "d1",
                        name: "Dips",
                        sets: 1,
                        reps: "do 5RM",
                        weight: "?",
                        rest: 90,
                        notes: '',
                    },
                ],
            },
            {
                block: "E",
                exercises: [
                    {
                        id: "e1",
                        name: "Landmine press w pozycji iso yielding split squat",
                        sets: 3,
                        reps: "6+6",
                        weight: "5 kg",
                        rest: 90,
                        notes:
                            'Zaczynasz ze słabszej strony. Noga wykroczna przeciwna do pracującej ręki. Przerwa między stronami 90-120".',
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
                        name: "Przysiad (barbell squat)",
                        sets: 1,
                        reps: "10",
                        weight: "10 RM",
                        rest: 90,
                        notes: 'Zacznij od małego obciążenia, stopniowo dokładaj do max na 10 powtórzeń.',
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
                        name: "Martwy ciąg sumo",
                        sets: 3,
                        reps: "6",
                        weight: "50 kg",
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

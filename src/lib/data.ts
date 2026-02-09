import type { WorkoutPlan } from './types';

export const workoutPlans: WorkoutPlan[] = [
    {
        id: 'plan-19-01',
        name: 'Strength Building',
        startDate: '19.01',
        blocks: [
            {
                block: 'A',
                exercises: [
                    {
                        id: 'a1',
                        name: 'Barbell Back Squat',
                        sets: 3,
                        reps: '6-8',
                        weight: '80kg',
                        rest: 90,
                        notes: 'Focus on depth and controlled descent'
                    },
                    {
                        id: 'a2',
                        name: 'Romanian Deadlift',
                        sets: 3,
                        reps: '8-10',
                        weight: '60kg',
                        rest: 90,
                        notes: 'Keep back neutral, hinge at hips'
                    }
                ]
            },
            {
                block: 'B',
                exercises: [
                    {
                        id: 'b1',
                        name: 'Bulgarian Split Squat',
                        sets: 3,
                        reps: '10-12',
                        weight: '20kg',
                        rest: 60,
                        notes: 'Maintain upright torso',
                        isUnilateral: true
                    },
                    {
                        id: 'b2',
                        name: 'Leg Curl',
                        sets: 3,
                        reps: '12-15',
                        weight: '45kg',
                        rest: 60,
                        notes: 'Slow eccentric phase'
                    }
                ]
            },
            {
                block: 'C',
                exercises: [
                    {
                        id: 'c1',
                        name: 'Walking Lunges',
                        sets: 3,
                        reps: '20 steps',
                        weight: '15kg DBs',
                        rest: 45,
                        notes: 'Long stride, knee to ground',
                        isUnilateral: true
                    },
                    {
                        id: 'c2',
                        name: 'Calf Raises',
                        sets: 3,
                        reps: '15-20',
                        weight: '60kg',
                        rest: 45,
                        notes: 'Full range of motion'
                    }
                ]
            }
        ]
    },
    {
        id: 'plan-26-01',
        name: 'Hypertrophy Focus',
        startDate: '26.01',
        blocks: [
            {
                block: 'A',
                exercises: [
                    {
                        id: 'a1-v2',
                        name: 'Front Squat',
                        sets: 3,
                        reps: '8-10',
                        weight: '65kg',
                        rest: 120,
                        notes: 'Elbows high, core tight'
                    },
                    {
                        id: 'a2-v2',
                        name: 'Leg Press',
                        sets: 3,
                        reps: '12-15',
                        weight: '180kg',
                        rest: 90,
                        notes: 'Feet shoulder width, full depth'
                    }
                ]
            },
            {
                block: 'B',
                exercises: [
                    {
                        id: 'b1-v2',
                        name: 'Single Leg Press',
                        sets: 3,
                        reps: '12-15',
                        weight: '80kg',
                        rest: 60,
                        notes: 'Control the eccentric',
                        isUnilateral: true
                    },
                    {
                        id: 'b2-v2',
                        name: 'Hamstring Curl',
                        sets: 3,
                        reps: '12-15',
                        weight: '50kg',
                        rest: 60,
                        notes: '3 second eccentric'
                    }
                ]
            },
            {
                block: 'C',
                exercises: [
                    {
                        id: 'c1-v2',
                        name: 'Goblet Squat',
                        sets: 3,
                        reps: '15-20',
                        weight: '24kg',
                        rest: 45,
                        notes: 'Pause at bottom for 2s'
                    },
                    {
                        id: 'c2-v2',
                        name: 'Leg Extension',
                        sets: 3,
                        reps: '15-20',
                        weight: '55kg',
                        rest: 45,
                        notes: 'Squeeze at top'
                    }
                ]
            },
            {
                block: 'D',
                exercises: [
                    {
                        id: 'd1-v2',
                        name: 'Seated Calf Raise',
                        sets: 3,
                        reps: '20-25',
                        weight: '40kg',
                        rest: 30,
                        notes: 'Hold peak contraction for 1s'
                    },
                    {
                        id: 'd2-v2',
                        name: 'Standing Calf Raise',
                        sets: 3,
                        reps: '15-20',
                        weight: '70kg',
                        rest: 30,
                        notes: 'Full stretch at bottom'
                    }
                ]
            }
        ]
    },
    {
        "id": "plan-jakub-4.5-strength",
        "name": "Jakub Banaś 4.5 Siłowy",
        "startDate": "01.01",
        "blocks": [
            {
                "block": "A",
                "exercises": [
                    {
                        "id": "a1",
                        "name": "SSB (Ta dziwna sztanga z uchwytami) Bulgarian split squat",
                        "sets": 3,
                        "reps": "6+6",
                        "weight": "RPE8",
                        "rest": 105,
                        "notes": "RPE 8 na słabszej nodze. Przerwa między nogami 90-120\"",
                        "isUnilateral": true
                    }
                ]
            },
            {
                "block": "B",
                "exercises": [
                    {
                        "id": "b1",
                        "name": "Pendlay row – wiosłowanie z martwgo punktu",
                        "sets": 3,
                        "reps": "6",
                        "weight": "RPE8",
                        "rest": 90,
                        "notes": "Każde powtórzenie z nowym napięciem."
                    }
                ]
            },
            {
                "block": "C",
                "exercises": [
                    {
                        "id": "c1",
                        "name": "Leg extension – maszyna izolowana",
                        "sets": 3,
                        "reps": "6",
                        "weight": "RPE10",
                        "rest": 45,
                        "notes": "(30-60\" po c1) Rest after C1 is 45s (average of 30-60s) before C2. Superset with C2."
                    },
                    {
                        "id": "c2",
                        "name": "SLDL – taki RDL, ale nogi 100% proste",
                        "sets": 3,
                        "reps": "6-8",
                        "weight": "RPE6-7",
                        "rest": 90,
                        "notes": "Rób hantlami. Dokładny ruch ważniejszy od ciężaru. (90\" po c2) Rest after C2 is 90s before C1."
                    }
                ]
            },
            {
                "block": "D",
                "exercises": [
                    {
                        "id": "d1",
                        "name": "Weighted dips",
                        "sets": 3,
                        "reps": "6-8",
                        "weight": "RPE8",
                        "rest": 30,
                        "notes": "Rest after D1 is 30s (transition) before D2. Superset with D2."
                    },
                    {
                        "id": "d2",
                        "name": "Uginanie ramion ze sztangą łamaną (Barbell Curl)",
                        "sets": 3,
                        "reps": "5-6",
                        "weight": "RPE9",
                        "rest": 90,
                        "notes": "(90\" po D2) Rest after D2 is 90s before D1."
                    }
                ]
            },
            {
                "block": "E",
                "exercises": [
                    {
                        "id": "e1",
                        "name": "Coppenhaggen adduction",
                        "sets": 3,
                        "reps": "5+5",
                        "weight": "RPE8-9",
                        "rest": 60,
                        "notes": "Jak za lekko kładź ciężar na biodro.",
                        "isUnilateral": true
                    }
                ]
            }
        ]
    },
    {
        "id": "plan-jakub-5.0-strength",
        "name": "Jakub Banaś 5.0 Siłowy",
        "startDate": "09.02",
        "blocks": [
            {
                "block": "A",
                "exercises": [
                    {
                        "id": "a1",
                        "name": "SSB (sztanga z rogami) Bulgarian split squat",
                        "sets": 3,
                        "reps": "6+6",
                        "weight": "RPE8",
                        "rest": 90,
                        "notes": "RPE 8 na słabszej nodze. Przerwa między nogami 90-120\"",
                        "isUnilateral": true
                    }
                ]
            },
            {
                "block": "B",
                "exercises": [
                    {
                        "id": "b1",
                        "name": "Pendlay row - wiosłowanie z martwego punktu",
                        "sets": 3,
                        "reps": "6",
                        "weight": "RPE8",
                        "rest": 90,
                        "notes": "Każde powtórzenie z nowym napięciem."
                    }
                ]
            },
            {
                "block": "C",
                "exercises": [
                    {
                        "id": "c1",
                        "name": "Leg extension - maszyna izolowana",
                        "sets": 3,
                        "reps": "6",
                        "weight": "67.5 - 70 kg",
                        "rest": 30,
                        "notes": "Rest after C1 is 30s (30-60s) before C2. Superset with C2."
                    },
                    {
                        "id": "c2",
                        "name": "SLDL - taki RDL, ale nogi 100% proste",
                        "sets": 3,
                        "reps": "6",
                        "weight": "RPE6-7",
                        "rest": 90,
                        "notes": "Rób hantlami. Dokładny ruch ważniejszy od ciężaru. (90\" po c2) Rest after C2 is 90s before C1."
                    }
                ]
            },
            {
                "block": "D",
                "exercises": [
                    {
                        "id": "d1",
                        "name": "Weighted dips",
                        "sets": 3,
                        "reps": "8",
                        "weight": "RPE8",
                        "rest": 30,
                        "notes": "Rest after D1 is 30s (transition) before D2. Superset with D2."
                    },
                    {
                        "id": "d2",
                        "name": "Uginanie ramion ze sztangą łamaną (Barbell Curl)",
                        "sets": 3,
                        "reps": "5-6",
                        "weight": "32.5 kg",
                        "rest": 90,
                        "notes": "(90\" po D2) Rest after D2 is 90s before D1."
                    }
                ]
            },
            {
                "block": "E",
                "exercises": [
                    {
                        "id": "e1",
                        "name": "Coppenhaggen adduction",
                        "sets": 3,
                        "reps": "6+6",
                        "weight": "RPE8-9",
                        "rest": 60,
                        "notes": "Jak za lekko kładź ciężar na biodro.",
                        "isUnilateral": true
                    }
                ]
            }
        ]
    },
];

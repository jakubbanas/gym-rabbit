import Database from 'better-sqlite3';
import path from 'path';
import type { WorkoutPlan, TrainingBlock, Exercise } from '../types';

const dbPath = path.resolve('database.sqlite');
const db = new Database(dbPath);

// Enable foreign keys
db.pragma('foreign_keys = ON');

// Initialize schema
db.exec(`
    CREATE TABLE IF NOT EXISTS plans (
        id TEXT PRIMARY KEY,
        name TEXT NOT NULL,
        startDate TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS blocks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        plan_id TEXT NOT NULL,
        name TEXT NOT NULL,
        sort_order INTEGER NOT NULL,
        FOREIGN KEY (plan_id) REFERENCES plans(id) ON DELETE CASCADE
    );

    CREATE TABLE IF NOT EXISTS exercises (
        id TEXT NOT NULL,
        block_id INTEGER NOT NULL,
        name TEXT NOT NULL,
        sets INTEGER NOT NULL,
        reps TEXT NOT NULL,
        weight TEXT,
        rest INTEGER,
        notes TEXT,
        is_unilateral INTEGER DEFAULT 0,
        sort_order INTEGER NOT NULL,
        PRIMARY KEY (id, block_id),
        FOREIGN KEY (block_id) REFERENCES blocks(id) ON DELETE CASCADE
    );
`);

export function getWorkoutPlans(): WorkoutPlan[] {
    const plans = db.prepare('SELECT * FROM plans').all() as any[];
    
    return plans.map(plan => {
        const blocks = db.prepare('SELECT * FROM blocks WHERE plan_id = ? ORDER BY sort_order').all(plan.id) as any[];
        
        const mappedBlocks: TrainingBlock[] = blocks.map(block => {
            const exercises = db.prepare('SELECT * FROM exercises WHERE block_id = ? ORDER BY sort_order').all(block.id) as any[];
            
            return {
                block: block.name,
                exercises: exercises.map(ex => ({
                    id: ex.id,
                    name: ex.name,
                    sets: ex.sets,
                    reps: ex.reps,
                    weight: ex.weight || undefined,
                    rest: ex.rest || undefined,
                    notes: ex.notes || undefined,
                    isUnilateral: !!ex.is_unilateral
                }))
            };
        });

        return {
            id: plan.id,
            name: plan.name,
            startDate: plan.startDate,
            blocks: mappedBlocks
        };
    });
}

export function seedDatabase(workoutPlans: WorkoutPlan[]) {
    const insertPlan = db.prepare('INSERT OR REPLACE INTO plans (id, name, startDate) VALUES (?, ?, ?)');
    const insertBlock = db.prepare('INSERT INTO blocks (plan_id, name, sort_order) VALUES (?, ?, ?)');
    const insertExercise = db.prepare(`
        INSERT INTO exercises (id, block_id, name, sets, reps, weight, rest, notes, is_unilateral, sort_order)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);

    const transaction = db.transaction((plans: WorkoutPlan[]) => {
        // Clear existing data to avoid duplicates during seeding
        db.prepare('DELETE FROM plans').run();
        
        for (const plan of plans) {
            insertPlan.run(plan.id, plan.name, plan.startDate);
            
            plan.blocks.forEach((block, blockIdx) => {
                const blockResult = insertBlock.run(plan.id, block.block, blockIdx);
                const blockId = blockResult.lastInsertRowid;
                
                block.exercises.forEach((ex, exIdx) => {
                    insertExercise.run(
                        ex.id,
                        blockId,
                        ex.name,
                        ex.sets,
                        ex.reps,
                        ex.weight || null,
                        ex.rest || null,
                        ex.notes || null,
                        ex.isUnilateral ? 1 : 0,
                        exIdx
                    );
                });
            });
        }
    });

    transaction(workoutPlans);
}

export function createWorkoutPlan(plan: WorkoutPlan) {
    const insertPlan = db.prepare('INSERT INTO plans (id, name, startDate) VALUES (?, ?, ?)');
    const insertBlock = db.prepare('INSERT INTO blocks (plan_id, name, sort_order) VALUES (?, ?, ?)');
    const insertExercise = db.prepare(`
        INSERT INTO exercises (id, block_id, name, sets, reps, weight, rest, notes, is_unilateral, sort_order)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);

    const transaction = db.transaction((plan: WorkoutPlan) => {
        insertPlan.run(plan.id, plan.name, plan.startDate);
        
        plan.blocks.forEach((block, blockIdx) => {
            const blockResult = insertBlock.run(plan.id, block.block, blockIdx);
            const blockId = blockResult.lastInsertRowid;
            
            block.exercises.forEach((ex, exIdx) => {
                insertExercise.run(
                    ex.id,
                    blockId,
                    ex.name,
                    ex.sets,
                    ex.reps,
                    ex.weight || null,
                    ex.rest || null,
                    ex.notes || null,
                    ex.isUnilateral ? 1 : 0,
                    exIdx
                );
            });
        });
    });

    transaction(plan);
}

export function updateWorkoutPlan(plan: WorkoutPlan) {
    const insertPlan = db.prepare('INSERT INTO plans (id, name, startDate) VALUES (?, ?, ?)');
    const insertBlock = db.prepare('INSERT INTO blocks (plan_id, name, sort_order) VALUES (?, ?, ?)');
    const insertExercise = db.prepare(`
        INSERT INTO exercises (id, block_id, name, sets, reps, weight, rest, notes, is_unilateral, sort_order)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `);

    const transaction = db.transaction((plan: WorkoutPlan) => {
        // Delete existing blocks and exercises (via cascade from plans delete if we delete plan, 
        // but we want to keep the same plan ID, so we delete from blocks)
        db.prepare('DELETE FROM blocks WHERE plan_id = ?').run(plan.id);
        
        // Update plan metadata
        db.prepare('UPDATE plans SET name = ?, startDate = ? WHERE id = ?').run(plan.name, plan.startDate, plan.id);
        
        plan.blocks.forEach((block, blockIdx) => {
            const blockResult = insertBlock.run(plan.id, block.block, blockIdx);
            const blockId = blockResult.lastInsertRowid;
            
            block.exercises.forEach((ex, exIdx) => {
                insertExercise.run(
                    ex.id,
                    blockId,
                    ex.name,
                    ex.sets,
                    ex.reps,
                    ex.weight || null,
                    ex.rest || null,
                    ex.notes || null,
                    ex.isUnilateral ? 1 : 0,
                    exIdx
                );
            });
        });
    });

    transaction(plan);
}

export function getWorkoutPlan(id: string): WorkoutPlan | undefined {
    const plan = db.prepare('SELECT * FROM plans WHERE id = ?').get(id) as any;
    if (!plan) return undefined;

    const blocks = db.prepare('SELECT * FROM blocks WHERE plan_id = ? ORDER BY sort_order').all(plan.id) as any[];
    
    const mappedBlocks: TrainingBlock[] = blocks.map(block => {
        const exercises = db.prepare('SELECT * FROM exercises WHERE block_id = ? ORDER BY sort_order').all(block.id) as any[];
        
        return {
            block: block.name,
            exercises: exercises.map(ex => ({
                id: ex.id,
                name: ex.name,
                sets: ex.sets,
                reps: ex.reps,
                weight: ex.weight || undefined,
                rest: ex.rest || undefined,
                notes: ex.notes || undefined,
                isUnilateral: !!ex.is_unilateral
            }))
        };
    });

    return {
        id: plan.id,
        name: plan.name,
        startDate: plan.startDate,
        blocks: mappedBlocks
    };
}

export default db;

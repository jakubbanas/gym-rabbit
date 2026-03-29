import { fail, redirect, error } from '@sveltejs/kit';
import { getWorkoutPlan, updateWorkoutPlan, createWorkoutPlan } from '$lib/server/database';
import type { PageServerLoad, Actions } from './$types';
import type { WorkoutPlan } from '$lib/types';

export const load: PageServerLoad = async ({ params }) => {
    const plan = getWorkoutPlan(params.id);
    
    if (!plan) {
        throw error(404, 'Plan not found');
    }

    return {
        plan
    };
};

export const actions: Actions = {
    default: async ({ request, params }) => {
        const formData = await request.formData();
        const planDataStr = formData.get('planData');

        if (!planDataStr || typeof planDataStr !== 'string') {
            return fail(400, { message: 'Invalid plan data' });
        }

        try {
            const plan: WorkoutPlan = JSON.parse(planDataStr);
            const intent = formData.get('intent');

            if (intent === 'saveAsNew') {
                // For "Save as New", we let createWorkoutPlan handle it
                // but we should probably ensure the ID is unique or modified
                // In this case, the client will provide a new ID if we update the UI
                createWorkoutPlan(plan);
            } else {
                // Ensure ID matches for update
                if (plan.id !== params.id) {
                    return fail(400, { message: 'Plan ID mismatch' });
                }
                updateWorkoutPlan(plan);
            }
        } catch (e) {
            console.error('Failed to process plan:', e);
            return fail(500, { message: 'Failed to save plan to database' });
        }

        throw redirect(303, '/plans');
    }
};

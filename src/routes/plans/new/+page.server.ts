import { fail, redirect } from '@sveltejs/kit';
import { createWorkoutPlan } from '$lib/server/database';
import type { Actions } from './$types';
import type { WorkoutPlan } from '$lib/types';

export const actions: Actions = {
    default: async ({ request }) => {
        const formData = await request.formData();
        const planDataStr = formData.get('planData');

        if (!planDataStr || typeof planDataStr !== 'string') {
            return fail(400, { message: 'Invalid plan data' });
        }

        try {
            const plan: WorkoutPlan = JSON.parse(planDataStr);
            
            // Basic validation
            if (!plan.name || !plan.id) {
                return fail(400, { message: 'Plan name and ID are required' });
            }

            createWorkoutPlan(plan);
        } catch (e) {
            console.error('Failed to create plan:', e);
            return fail(500, { message: 'Failed to save plan to database' });
        }

        throw redirect(303, '/plans');
    }
};

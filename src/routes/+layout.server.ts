import { getWorkoutPlans } from '$lib/server/database';
import type { LayoutServerLoad } from './$types';

export const load: LayoutServerLoad = async () => {
    return {
        workoutPlans: getWorkoutPlans()
    };
};

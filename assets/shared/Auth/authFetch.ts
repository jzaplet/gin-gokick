import { router } from '@/router';
import { loginReturningTo } from '@/shared/Auth/Login/loginRedirect';
import { forgetSession } from '@/shared/Auth/Session/currentSession';
import { type ApiFetch, createApiFetch } from '@/shared/Fetch';

export const authFetch: ApiFetch = createApiFetch({
    afterResponse: async (result) => {
        if (result.status === 401) {
            forgetSession();
            await router.push(loginReturningTo(router.currentRoute.value));
        }
    },
});

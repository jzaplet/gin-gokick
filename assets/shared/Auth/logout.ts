import { router } from '@/router';
import { type ApiGeneralError, type ApiResponse, apiFetch } from '@/shared/Fetch';
import { isNull } from '@/shared/TypeGuards/typeGuards';
import { forgetSession } from '@/shared/Auth/Session/currentSession';

export const logout = async (): Promise<ApiResponse<null, ApiGeneralError>> => {
    const result = await apiFetch<null>('POST', '/api/auth/logout', { validate: isNull });

    if (result.success) {
        forgetSession();
        await router.push({ name: 'login' });
    }

    return result;
};

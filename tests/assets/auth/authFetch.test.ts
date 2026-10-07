import { afterEach, describe, expect, it, vi } from 'vitest';
import { router } from '@/router';
import { authFetch } from '@/shared/Auth/authFetch';
import { forgetSession, rememberSession, signedIn } from '@/shared/Auth/Session/currentSession';
import { isNull } from '@/shared/TypeGuards/typeGuards';
import { json } from '../fetch/fixtures';

const jan = { id: '0192f3a4-5b6c-7d8e-9f01-23456789abcd' };

describe('authFetch', () => {
    afterEach(forgetSession);

    it('sends a signed-out user to the login and remembers the page', async () => {
        vi.spyOn(globalThis, 'fetch').mockResolvedValue(json({ general: { key: 'auth.sign_in_required' } }, 401));
        rememberSession(jan);
        await router.push('/cs/app/settings?tab=mail');

        const result = await authFetch<null>('GET', '/api/user/me', { validate: isNull });

        expect(result).toEqual({
            success: false,
            status: 401,
            data: { general: { key: 'auth.sign_in_required' } },
        });
        expect(router.currentRoute.value.name).toBe('login');
        expect(router.currentRoute.value.query).toEqual({ redirect: '/settings?tab=mail' });
        expect(signedIn()).toBe(false);
    });

    it.each([
        [
            'a success',
            new Response(null, { status: 200 }),
        ],
        [
            'a refusal',
            json({ general: { key: 'auth.forbidden' } }, 403),
        ],
        [
            'a server error',
            json({ general: { key: 'request.internal' } }, 500),
        ],
    ])('stays on the page after %s', async (_, response) => {
        const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(response);

        rememberSession(jan);
        await router.push('/cs/app/settings');

        const result = await authFetch<null>('POST', '/api/user/me', { validate: isNull });

        expect(result.status).toBe(response.status);
        expect(fetchSpy.mock.calls[0]?.[1]?.method).toBe('POST');
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/settings');
        expect(signedIn()).toBe(true);
    });
});

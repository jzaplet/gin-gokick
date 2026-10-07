import { afterEach, describe, expect, it, vi } from 'vitest';
import { router } from '@/router';
import { logout } from '@/shared/Auth/logout';
import { forgetSession, rememberSession, signedIn } from '@/shared/Auth/Session/currentSession';
import { json } from '../fetch/fixtures';

const jan = { id: '0192f3a4-5b6c-7d8e-9f01-23456789abcd' };

describe('logout', () => {
    afterEach(forgetSession);

    it('ends the session and shows the login', async () => {
        const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(new Response(null, { status: 204 }));

        rememberSession(jan);
        await router.push('/cs/app/dashboard');

        const result = await logout();

        expect(result.success).toBe(true);
        expect(fetchSpy.mock.calls[0]?.[0]).toBe('/api/auth/logout');
        expect(fetchSpy.mock.calls[0]?.[1]?.method).toBe('POST');
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/login');
        expect(signedIn()).toBe(false);
    });

    it('stays on the page and returns the error when the session did not end', async () => {
        vi.spyOn(globalThis, 'fetch').mockResolvedValue(json({ general: { key: 'request.internal' } }, 500));
        rememberSession(jan);
        await router.push('/cs/app/dashboard');

        const result = await logout();

        expect(result).toEqual({
            success: false,
            status: 500,
            data: { general: { key: 'request.internal' } },
        });
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/dashboard');
        expect(signedIn()).toBe(true);
    });
});

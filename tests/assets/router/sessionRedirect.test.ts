import { flushPromises } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { router } from '@/router';
import { forgetSession, rememberSession } from '@/shared/Auth/Session/currentSession';
import { offerEnglish } from '../i18n/languages';

const jan = { id: '0192f3a4-5b6c-7d8e-9f01-23456789abcd' };

describe('the router', () => {
    beforeEach(async () => {
        await router.push('/cs/app/nothing');
    });

    afterEach(forgetSession);

    it.each([
        [
            true,
            '/en/app/login',
            '/en/app/dashboard',
        ],
        [
            true,
            '/en/app/register',
            '/en/app/dashboard',
        ],
        [
            false,
            '/en/app/dashboard',
            '/en/app/login',
        ],
    ])('keeps the language of the address when it redirects, signed in: %s', async (signedIn, path, target) => {
        const fetchSpy = vi.spyOn(globalThis, 'fetch');

        offerEnglish();

        if (signedIn) {
            rememberSession(jan);
        }

        await router.push(path);

        expect(router.currentRoute.value.fullPath).toBe(target);
        expect(fetchSpy).not.toHaveBeenCalled();
    });

    it('leaves the dashboard once another tab ended the session', async () => {
        rememberSession(jan);
        await router.push('/cs/app/dashboard');

        localStorage.removeItem('session');
        window.dispatchEvent(
            new StorageEvent('storage', {
                key: 'session',
                storageArea: localStorage,
            }),
        );
        await flushPromises();

        expect(router.currentRoute.value.name).toBe('login');
    });
});

import * as Sentry from '@sentry/vue';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { forgetSession, rememberSession, signedIn } from '@/shared/Auth/Session/currentSession';
import { followOtherTabs } from '@/shared/Auth/Session/followOtherTabs';
import { restoreSession } from '@/shared/Auth/Session/restoreSession';
import { json } from '../fetch/fixtures';

const jan = { id: '0192f3a4-5b6c-7d8e-9f01-23456789abcd' };

const eva = { id: '0192f3a4-5b6c-7d8e-9f01-23456789abce' };

const sentryUser = (): unknown => Sentry.getIsolationScope().getUser();

const store = (user: unknown): void => {
    localStorage.setItem('session', JSON.stringify(user));
};

describe('the session', () => {
    afterEach(forgetSession);

    it('keeps only the ID of the user in localStorage and in Sentry', () => {
        const withEmail = {
            ...jan,
            email: 'jan@example.com',
        };

        rememberSession(withEmail);

        expect(signedIn()).toBe(true);
        expect(localStorage.getItem('session')).toBe(`{"id":"${jan.id}"}`);
        expect(sentryUser()).toEqual({ id: jan.id });
    });

    it('forgets the user in localStorage and in Sentry', () => {
        rememberSession(jan);

        forgetSession();

        expect(signedIn()).toBe(false);
        expect(localStorage.getItem('session')).toBeNull();
        expect(Sentry.getIsolationScope().getUser()?.id).toBeUndefined();
    });

    it('keeps working in memory when localStorage refuses', () => {
        vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new DOMException('refused', 'SecurityError');
        });

        rememberSession(jan);

        expect(signedIn()).toBe(true);
    });
});

describe('restoreSession', () => {
    afterEach(forgetSession);

    it('restores the stored user and checks the session with the server', async () => {
        const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(json(jan, 200));

        store(jan);
        await restoreSession();

        expect(fetchSpy.mock.calls[0]?.[0]).toBe('/api/auth/session');
        expect(signedIn()).toBe(true);
        expect(sentryUser()).toEqual({ id: jan.id });
    });

    it('keeps the stored user when the server fails', async () => {
        vi.spyOn(globalThis, 'fetch').mockResolvedValue(json({ general: { key: 'request.internal' } }, 500));

        store(jan);
        await restoreSession();

        expect(signedIn()).toBe(true);
        expect(localStorage.getItem('session')).toBe(`{"id":"${jan.id}"}`);
        expect(sentryUser()).toEqual({ id: jan.id });
    });

    it('forgets a session the server ended', async () => {
        vi.spyOn(globalThis, 'fetch').mockResolvedValue(json({ general: { key: 'auth.sign_in_required' } }, 401));

        store(jan);
        await restoreSession();

        expect(signedIn()).toBe(false);
        expect(localStorage.getItem('session')).toBeNull();
    });

    it('keeps a logout that happened while it asked the server', async () => {
        const response = Promise.withResolvers<Response>();

        vi.spyOn(globalThis, 'fetch').mockReturnValue(response.promise);
        store(jan);

        const starting = restoreSession();

        forgetSession();
        response.resolve(json(jan, 200));
        await starting;

        expect(signedIn()).toBe(false);
        expect(localStorage.getItem('session')).toBeNull();
    });

    it('keeps a login that happened while it asked the server', async () => {
        const response = Promise.withResolvers<Response>();

        vi.spyOn(globalThis, 'fetch').mockReturnValue(response.promise);
        store(jan);

        const starting = restoreSession();

        rememberSession(eva);
        response.resolve(json({ general: { key: 'auth.sign_in_required' } }, 401));
        await starting;

        expect(signedIn()).toBe(true);
        expect(localStorage.getItem('session')).toBe(`{"id":"${eva.id}"}`);
    });
});

describe('followOtherTabs', () => {
    afterEach(forgetSession);

    it('follows a session that another tab started or ended', () => {
        const onChange = vi.fn();

        followOtherTabs(onChange);
        store(eva);
        window.dispatchEvent(
            new StorageEvent('storage', {
                key: 'session',
                storageArea: localStorage,
            }),
        );

        expect(signedIn()).toBe(true);
        expect(sentryUser()).toEqual({ id: eva.id });

        localStorage.clear();
        window.dispatchEvent(
            new StorageEvent('storage', {
                key: null,
                storageArea: localStorage,
            }),
        );

        expect(signedIn()).toBe(false);

        window.dispatchEvent(
            new StorageEvent('storage', {
                key: 'theme',
                storageArea: localStorage,
            }),
        );

        expect(onChange).toHaveBeenCalledTimes(2);
    });
});

import { afterEach, describe, expect, it } from 'vitest';
import { forgetSession, signedIn } from '@/shared/Auth/Session/currentSession';
import { followOtherTabs } from '@/shared/Auth/Session/followOtherTabs';
import { readStoredSession, writeStoredSession } from '@/shared/Auth/Session/storedSession';
import { readStoredChoice, writeStoredChoice } from '@/shared/Consent/Choice/storedChoice';
import { appStorage, setAppStorage } from '@/shared/Storage/appStorage';
import type { KeyValueStorage } from '@/shared/Storage/types/KeyValueStorage';

type MemoryStorage = KeyValueStorage & {
    values: Map<string, unknown>;
    changeInAnotherTab: (key: string, value: unknown) => void;
};

const jan = { id: '0192f3a4-5b6c-7d8e-9f01-23456789abcd' };

const tools = {
    ga4: 'G-AB12CD34EF',
    google_ads: '',
    meta_pixel: '',
};

const memoryStorage = (): MemoryStorage => {
    const values = new Map<string, unknown>();
    const followers = new Map<string, () => void>();

    const read = (key: string): unknown => values.get(key);

    const write = (key: string, value: unknown): boolean => {
        values.set(key, value);

        return true;
    };

    const remove = (key: string): boolean => {
        values.delete(key);

        return true;
    };

    const follow = (key: string, listener: () => void): (() => void) => {
        followers.set(key, listener);

        return () => {
            followers.delete(key);
        };
    };

    const changeInAnotherTab = (key: string, value: unknown): void => {
        values.set(key, value);
        followers.get(key)?.();
    };

    return {
        values,
        read,
        write,
        remove,
        follow,
        changeInAnotherTab,
    };
};

describe('appStorage', () => {
    const browser = appStorage();

    afterEach(() => {
        setAppStorage(browser);
        forgetSession();
    });

    it('lets the session and the consent keep their values in another storage', () => {
        const storage = memoryStorage();

        setAppStorage(storage);
        writeStoredSession(jan);
        writeStoredChoice(tools, { analytics: true });

        expect(storage.values.has('session')).toBe(true);
        expect(storage.values.has('consent')).toBe(true);
        expect(localStorage.getItem('session')).toBeNull();
        expect(localStorage.getItem('consent')).toBeNull();
        expect(readStoredSession()).toEqual(jan);
        expect(readStoredChoice(tools)).toEqual({ analytics: true });

        writeStoredSession(undefined);

        expect(storage.values.has('session')).toBe(false);

        let changes = 0;

        followOtherTabs(() => {
            changes += 1;
        });
        storage.changeInAnotherTab('session', jan);

        expect(signedIn()).toBe(true);
        expect(changes).toBe(1);
    });
});

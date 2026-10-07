import { afterEach, describe, expect, it, vi } from 'vitest';
import { localStorageAdapter, webStorage } from '@/shared/Storage/Adapters/webStorage';

const refuse = (method: 'getItem' | 'setItem' | 'removeItem'): void => {
    vi.spyOn(Storage.prototype, method).mockImplementation(() => {
        throw new DOMException('refused', 'SecurityError');
    });
};

const changeInAnotherTab = (key: string | null, storageArea: Storage): void => {
    window.dispatchEvent(
        new StorageEvent('storage', {
            key,
            storageArea,
        }),
    );
};

afterEach(() => {
    localStorage.clear();
    sessionStorage.clear();
});

describe('webStorage', () => {
    it('writes JSON, reads it back and removes it', () => {
        expect(localStorageAdapter.write('session', { id: 'a' })).toBe(true);
        expect(localStorage.getItem('session')).toBe('{"id":"a"}');
        expect(localStorageAdapter.read('session')).toEqual({ id: 'a' });
        expect(localStorageAdapter.remove('session')).toBe(true);
        expect(localStorage.getItem('session')).toBeNull();
    });

    it('reads nothing from a missing or malformed value', () => {
        localStorage.setItem('broken', '{id');

        expect(localStorageAdapter.read('missing')).toBeUndefined();
        expect(localStorageAdapter.read('broken')).toBeUndefined();
    });

    it('reports a storage that refuses instead of throwing', () => {
        refuse('getItem');
        refuse('setItem');
        refuse('removeItem');

        expect(localStorageAdapter.read('session')).toBeUndefined();
        expect(localStorageAdapter.write('session', { id: 'a' })).toBe(false);
        expect(localStorageAdapter.remove('session')).toBe(false);
    });

    it('reports a storage it cannot reach', () => {
        const unreachable = webStorage(() => {
            throw new DOMException('denied', 'SecurityError');
        });

        expect(unreachable.read('session')).toBeUndefined();
        expect(unreachable.write('session', { id: 'a' })).toBe(false);
        expect(unreachable.remove('session')).toBe(false);
    });

    it('follows changes of its key in its own area from other tabs until unfollowed', () => {
        const listener = vi.fn();
        const unfollow = localStorageAdapter.follow('session', listener);

        changeInAnotherTab('session', localStorage);
        changeInAnotherTab(null, localStorage);
        changeInAnotherTab('theme', localStorage);
        changeInAnotherTab('session', sessionStorage);

        expect(listener).toHaveBeenCalledTimes(2);

        unfollow();
        changeInAnotherTab('session', localStorage);

        expect(listener).toHaveBeenCalledTimes(2);
    });
});

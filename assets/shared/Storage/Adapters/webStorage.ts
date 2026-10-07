import type { KeyValueStorage } from '@/shared/Storage/types/KeyValueStorage';

export const webStorage = (area: () => Storage): KeyValueStorage => {
    const read = (key: string): unknown => {
        try {
            const stored = area().getItem(key);
            const value: unknown = stored === null ? undefined : JSON.parse(stored);

            return value;
        } catch {
            return undefined;
        }
    };

    const write = (key: string, value: unknown): boolean => {
        try {
            area().setItem(key, JSON.stringify(value));

            return true;
        } catch {
            return false;
        }
    };

    const remove = (key: string): boolean => {
        try {
            area().removeItem(key);

            return true;
        } catch {
            return false;
        }
    };

    const follow = (key: string, listener: () => void): (() => void) => {
        const changed = (event: StorageEvent): void => {
            if ((event.key === key || event.key === null) && event.storageArea === area()) {
                listener();
            }
        };

        window.addEventListener('storage', changed);

        return () => {
            window.removeEventListener('storage', changed);
        };
    };

    return {
        read,
        write,
        remove,
        follow,
    };
};

export const localStorageAdapter = webStorage(() => localStorage);

import { localStorageAdapter } from '@/shared/Storage/Adapters/webStorage';
import type { KeyValueStorage } from '@/shared/Storage/types/KeyValueStorage';

let current: KeyValueStorage = localStorageAdapter;

export const appStorage = (): KeyValueStorage => current;

export const setAppStorage = (storage: KeyValueStorage): void => {
    current = storage;
};

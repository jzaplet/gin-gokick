import { isSessionUser, type SessionUser } from '@/shared/Auth/Session/types/SessionUser';
import { appStorage } from '@/shared/Storage/appStorage';

export const storageKey = 'session';

export const readStoredSession = (): SessionUser | undefined => {
    const value = appStorage().read(storageKey);

    return isSessionUser(value) ? { id: value.id } : undefined;
};

export const writeStoredSession = (user: SessionUser | undefined): boolean =>
    user === undefined ? appStorage().remove(storageKey) : appStorage().write(storageKey, user);

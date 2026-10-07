import { writeStoredSession } from '@/shared/Auth/Session/storedSession';
import type { SessionUser } from '@/shared/Auth/Session/types/SessionUser';
import { forgetUser, identifyUser } from '@/shared/Sentry/sentryUser';

let current: SessionUser | undefined;

let version = 0;

export const signedIn = (): boolean => current !== undefined;

export const sessionVersion = (): number => version;

export const setCurrentSession = (user: SessionUser | undefined): void => {
    version += 1;
    current = user;

    if (user === undefined) {
        forgetUser();
    } else {
        identifyUser(user.id);
    }
};

export const rememberSession = (user: SessionUser): void => {
    const session = { id: user.id };

    setCurrentSession(session);
    writeStoredSession(session);
};

export const forgetSession = (): void => {
    setCurrentSession(undefined);
    writeStoredSession(undefined);
};

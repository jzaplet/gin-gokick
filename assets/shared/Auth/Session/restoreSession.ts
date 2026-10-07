import {
    forgetSession,
    rememberSession,
    sessionVersion,
    setCurrentSession,
} from '@/shared/Auth/Session/currentSession';
import { readStoredSession, writeStoredSession } from '@/shared/Auth/Session/storedSession';
import { isSessionUser, type SessionUser } from '@/shared/Auth/Session/types/SessionUser';
import { apiFetch } from '@/shared/Fetch';

export const restoreSession = async (): Promise<void> => {
    const user = readStoredSession();

    setCurrentSession(user);

    if (user === undefined) {
        writeStoredSession(undefined);

        return;
    }

    const since = sessionVersion();
    const result = await apiFetch<SessionUser>('GET', '/api/auth/session', { validate: isSessionUser });

    if (since !== sessionVersion()) {
        return;
    }

    if (result.success) {
        rememberSession(result.data);
    } else if (result.status === 401) {
        forgetSession();
    }
};

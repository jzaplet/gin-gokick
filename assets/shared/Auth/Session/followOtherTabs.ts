import { setCurrentSession } from '@/shared/Auth/Session/currentSession';
import { readStoredSession, storageKey } from '@/shared/Auth/Session/storedSession';
import { appStorage } from '@/shared/Storage/appStorage';

export const followOtherTabs = (onChange: () => void): void => {
    appStorage().follow(storageKey, () => {
        setCurrentSession(readStoredSession());
        onChange();
    });
};

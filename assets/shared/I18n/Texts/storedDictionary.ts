import { isRecord } from '@/shared/TypeGuards/typeGuards';
import type { Dictionary } from '@/shared/I18n/Dictionary/types/Dictionary';
import { isMessage } from '@/shared/I18n/Dictionary/types/Message';
import { appStorage } from '@/shared/Storage/appStorage';

export const storageKey = (locale: string): string => `dictionary.${locale}`;

const isMessages = (v: unknown): v is Partial<Dictionary> =>
    isRecord(v) && Object.values(v).every((message) => isMessage(message));

export const readStoredDictionary = (locale: string, version: string): Partial<Dictionary> | undefined => {
    const stored = appStorage().read(storageKey(locale));
    const messages = isRecord(stored) && stored['version'] === version ? stored['messages'] : undefined;

    return isMessages(messages) ? messages : undefined;
};

export const writeStoredDictionary = (locale: string, version: string, messages: Partial<Dictionary>): boolean =>
    appStorage().write(storageKey(locale), {
        version,
        messages,
    });

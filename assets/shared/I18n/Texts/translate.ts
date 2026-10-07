import type { ApiMessage } from '@/shared/Fetch';
import type { MessageKey, MessageParams } from '@/shared/I18n/Dictionary/MessageParams';
import { formatMessage } from '@/shared/I18n/Format/formatMessage';
import { languageTag } from '@/shared/I18n/Page/localeMeta';
import { pageDictionary } from '@/shared/I18n/Texts/pageDictionary';
import { reportUnexpected } from '@/shared/Sentry/reportUnexpected';

export type MessageKeyWith<P> = {
    [K in MessageKey]: MessageParams[K] extends P ? ([P] extends [MessageParams[K]] ? K : never) : never;
}[MessageKey];

export type PlainMessageKey = MessageKeyWith<null>;

export type MessageArgs<K extends MessageKey> = MessageParams[K] extends null
    ? []
    : MessageParams[K] extends object ? [params: MessageParams[K]] : never;

const isMessageKey = (key: unknown): key is MessageKey =>
    typeof key === 'string' && Object.hasOwn(pageDictionary().messages, key);

export const isPlainMessageKey = (key: unknown): key is PlainMessageKey =>
    isMessageKey(key) && typeof pageDictionary().messages[key] === 'string';

const render = (key: string, params: Readonly<Record<string, unknown>>): string => {
    const { locale, messages } = pageDictionary();
    const message = isMessageKey(key) ? messages[key] : undefined;

    if (message === undefined) {
        reportUnexpected(`Message key without a text: ${key}`);

        return key;
    }

    try {
        return formatMessage(message, params, languageTag(locale));
    } catch (error) {
        reportUnexpected(`Message ${key} cannot be formatted: ${String(error)}`);

        return key;
    }
};

export const t = <K extends MessageKey>(key: K, ...[params]: MessageArgs<K>): string => render(key, params ?? {});

export const uiMessage = <K extends MessageKey>(key: K, ...[params]: MessageArgs<K>): ApiMessage =>
    params === undefined
        ? { key }
        : {
                key,
                params,
            };

export const tm = (message: ApiMessage | undefined): string | null =>
    message === undefined ? null : render(message.key, message.params ?? {});

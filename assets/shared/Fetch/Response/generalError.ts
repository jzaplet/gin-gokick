import type { ApiError } from '@/shared/Fetch/types/ApiError';
import type { MessageKey } from '@/shared/I18n/Dictionary/MessageParams';
import type { MessageArgs } from '@/shared/I18n/Texts/translate';

export const generalError = <K extends MessageKey>(
    status: number,
    key: K,
    ...[params]: MessageArgs<K>
): ApiError<never> => ({
    success: false,
    status,
    data: {
        general: params === undefined
            ? { key }
            : {
                    key,
                    params,
                },
    },
});

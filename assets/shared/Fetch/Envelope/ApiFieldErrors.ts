import { type ApiMessage, isApiMessage } from '@/shared/Fetch/Envelope/ApiMessage';
import { isRecord } from '@/shared/TypeGuards/typeGuards';

export type ApiFieldErrors = Record<string, ApiMessage>;

export const isApiFieldErrors = (keys: RegExp, v: unknown): v is ApiFieldErrors =>
    isRecord(v)
    && Object.keys(v).length > 0
    && Object.entries(v).every(([key, value]) => keys.test(key) && isApiMessage(value));

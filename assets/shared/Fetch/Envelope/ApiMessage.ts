import { isRecord, isString, optional } from '@/shared/TypeGuards/typeGuards';

export type ApiMessage = {
    key: string;
    params?: Record<string, unknown>;
};

export const isApiMessage = (v: unknown): v is ApiMessage =>
    isRecord(v) && isString(v['key']) && optional(isRecord)(v['params']);

import { type ApiMessage, isApiMessage } from '@/shared/Fetch/Envelope/ApiMessage';
import { isRecord, isString } from '@/shared/TypeGuards/typeGuards';

export type IdBody = {
    id: string;
};

export type EmailErrors = {
    general?: ApiMessage;
    email?: ApiMessage;
};

export const isIdBody = (v: unknown): v is IdBody => isRecord(v) && isString(v['id']);

const optionalMessage = (v: unknown): boolean => v === undefined || isApiMessage(v);

export const isEmailErrors = (v: unknown): v is EmailErrors =>
    isRecord(v) && optionalMessage(v['general']) && optionalMessage(v['email']);

export const json = (body: unknown, status: number): Response => new Response(JSON.stringify(body), { status });

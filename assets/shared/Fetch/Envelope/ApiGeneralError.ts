import { type ApiMessage, isApiMessage } from '@/shared/Fetch/Envelope/ApiMessage';
import { isRecord } from '@/shared/TypeGuards/typeGuards';

export type ApiGeneralError = {
    general: ApiMessage;
};

export const isApiGeneralError = (v: unknown): v is ApiGeneralError => isRecord(v) && isApiMessage(v['general']);

import type { ApiGeneralError } from '@/shared/Fetch/Envelope/ApiGeneralError';

export type ApiError<TError> = {
    success: false;
    status: number;
    data: TError | ApiGeneralError;
};

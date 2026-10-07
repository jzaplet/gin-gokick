import type { ApiGeneralError } from '@/shared/Fetch/Envelope/ApiGeneralError';
import type { Guard } from '@/shared/TypeGuards/typeGuards';

export type ResponseGuards<TData, TError> = {
    validate: Guard<TData>;
} & ([TError] extends [ApiGeneralError]
    ? { validateError?: Guard<TError> }
    : { validateError: Guard<TError> });

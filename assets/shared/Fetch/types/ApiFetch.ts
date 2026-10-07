import type { ApiGeneralError } from '@/shared/Fetch/Envelope/ApiGeneralError';
import type { ApiResponse } from '@/shared/Fetch/types/ApiResponse';
import type { FetchOptions } from '@/shared/Fetch/types/FetchOptions';
import type { ResponseGuards } from '@/shared/Fetch/types/ResponseGuards';

export type ApiFetch = <TData, TError = ApiGeneralError, TBody = never>(
    method: string,
    url: string,
    options: FetchOptions<TBody> & ResponseGuards<TData, TError>,
) => Promise<ApiResponse<TData, TError>>;

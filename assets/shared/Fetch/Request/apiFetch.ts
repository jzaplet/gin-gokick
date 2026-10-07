import { generalError } from '@/shared/Fetch/Response/generalError';
import { parseResponse } from '@/shared/Fetch/Response/parseResponse';
import type { ApiFetch } from '@/shared/Fetch/types/ApiFetch';
import type { ApiResponse } from '@/shared/Fetch/types/ApiResponse';
import type { FetchOptions } from '@/shared/Fetch/types/FetchOptions';
import type { ResponseGuards } from '@/shared/Fetch/types/ResponseGuards';

export const apiFetch: ApiFetch = async <TData, TError, TBody>(
    method: string,
    url: string,
    options: FetchOptions<TBody> & ResponseGuards<TData, TError>,
): Promise<ApiResponse<TData, TError>> => {
    const init: RequestInit = {
        method,
        headers: {
            'Accept': 'application/json',
            'Content-Type': 'application/json',
            ...options.headers,
        },
        credentials: 'same-origin',
    };

    if (options.body !== undefined) {
        init.body = JSON.stringify(options.body);
    }

    const response = await fetch(url, init).catch(() => null);

    if (response === null) {
        return generalError(0, 'fetch.network_error');
    }

    return parseResponse(response, options);
};

import { apiFetch } from '@/shared/Fetch/Request/apiFetch';
import type { ApiFetch } from '@/shared/Fetch/types/ApiFetch';
import type { ApiResponse } from '@/shared/Fetch/types/ApiResponse';
import type { FetchOptions } from '@/shared/Fetch/types/FetchOptions';
import type { ResponseGuards } from '@/shared/Fetch/types/ResponseGuards';

type Steps = {
    afterResponse: (result: ApiResponse<unknown, unknown>) => Promise<void>;
};

export const createApiFetch = ({ afterResponse }: Steps): ApiFetch =>
    async <TData, TError, TBody>(
        method: string,
        url: string,
        options: FetchOptions<TBody> & ResponseGuards<TData, TError>,
    ): Promise<ApiResponse<TData, TError>> => {
        const result = await apiFetch<TData, TError, TBody>(method, url, options);

        await afterResponse(result);

        return result;
    };

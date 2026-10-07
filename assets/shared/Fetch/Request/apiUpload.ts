import type { ApiGeneralError } from '@/shared/Fetch/Envelope/ApiGeneralError';
import { generalError } from '@/shared/Fetch/Response/generalError';
import { parseResponse } from '@/shared/Fetch/Response/parseResponse';
import type { ApiResponse } from '@/shared/Fetch/types/ApiResponse';
import type { ResponseGuards } from '@/shared/Fetch/types/ResponseGuards';
import type { UploadProgress } from '@/shared/Fetch/types/UploadProgress';

type UploadOptions<TData, TError> = ResponseGuards<TData, TError> & {
    onProgress?: (stats: UploadProgress) => void;
};

export const apiUpload = async <TData, TError = ApiGeneralError>(
    url: string,
    formData: FormData,
    options: UploadOptions<TData, TError>,
): Promise<ApiResponse<TData, TError>> => new Promise((resolve) => {
    const xhr = new XMLHttpRequest();
    const { onProgress } = options;

    if (onProgress !== undefined) {
        xhr.upload.onprogress = (event: ProgressEvent): void => {
            if (event.lengthComputable) {
                onProgress({
                    percent: (event.loaded / event.total) * 100,
                    loaded: event.loaded,
                    total: event.total,
                });
            }
        };
    }

    xhr.onload = (): void => {
        const response = new Response(xhr.responseText === '' ? null : xhr.responseText, {
            status: xhr.status,
            statusText: xhr.statusText,
        });

        void parseResponse(response, options).then(resolve);
    };

    xhr.onerror = (): void => {
        resolve(generalError(xhr.status, 'fetch.network_error'));
    };

    xhr.open('POST', url);
    xhr.setRequestHeader('Accept', 'application/json');
    xhr.send(formData);
});

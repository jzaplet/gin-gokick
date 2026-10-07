import { type ApiGeneralError, isApiGeneralError } from '@/shared/Fetch/Envelope/ApiGeneralError';
import { generalError } from '@/shared/Fetch/Response/generalError';
import type { ApiResponse } from '@/shared/Fetch/types/ApiResponse';
import type { ResponseGuards } from '@/shared/Fetch/types/ResponseGuards';
import { type Guard, isRecord } from '@/shared/TypeGuards/typeGuards';
import { reportUnexpected } from '@/shared/Sentry/reportUnexpected';

type Body = {
    parsed: true;
    json: unknown;
} | { parsed: false };

const uuid = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/gi;

const groupableUrl = (url: string): string => {
    if (url === '') {
        return 'unknown url';
    }

    return (url.split('?')[0] ?? url).replace(uuid, ':id');
};

const readBody = async (response: Response): Promise<Body> => {
    const text = await response.text().catch(() => '');

    if (text === '') {
        return {
            parsed: true,
            json: null,
        };
    }

    try {
        return {
            parsed: true,
            json: JSON.parse(text),
        };
    } catch {
        return { parsed: false };
    }
};

const report = (kind: string, response: Response): void => {
    reportUnexpected(
        `${kind} violates its contract: ${groupableUrl(response.url)} (status ${String(response.status)})`,
    );
};

const errorGuard = <TData, TError>(guards: ResponseGuards<TData, TError>): Guard<TError> | undefined =>
    guards.validateError;

export const parseResponse = async <TData, TError = ApiGeneralError>(
    response: Response,
    guards: ResponseGuards<TData, TError>,
): Promise<ApiResponse<TData, TError>> => {
    const body = await readBody(response);
    const validateError = errorGuard(guards);

    if (response.ok) {
        if (body.parsed === false) {
            return generalError(response.status, 'fetch.malformed_body');
        }

        if (guards.validate(body.json)) {
            return {
                success: true,
                status: response.status,
                data: body.json,
            };
        }

        report('Response', response);

        return generalError(response.status, 'fetch.invalid_shape');
    }

    if (body.parsed === true && isRecord(body.json)) {
        if (validateError?.(body.json) === true || isApiGeneralError(body.json)) {
            return {
                success: false,
                status: response.status,
                data: body.json,
            };
        }

        report('Error response', response);
    }

    return generalError(response.status, 'fetch.error_status', { status: String(response.status) });
};

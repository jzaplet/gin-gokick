import { beforeEach, describe, expect, it, vi } from 'vitest';
import { parseResponse } from '@/shared/Fetch/Response/parseResponse';
import { isNull } from '@/shared/TypeGuards/typeGuards';
import { reportUnexpected } from '@/shared/Sentry/reportUnexpected';
import { type EmailErrors, type IdBody, isEmailErrors, isIdBody, json } from './fixtures';

vi.mock('@/shared/Sentry/reportUnexpected', () => ({ reportUnexpected: vi.fn() }));

describe('parseResponse', () => {
    beforeEach(() => {
        vi.mocked(reportUnexpected).mockClear();
    });

    it('returns the data of a 2xx body that passes its guard', async () => {
        const out = await parseResponse<IdBody, EmailErrors>(json({ id: 'u-1' }, 200), {
            validate: isIdBody,
            validateError: isEmailErrors,
        });

        expect(out).toEqual({
            success: true,
            status: 200,
            data: { id: 'u-1' },
        });
    });

    it('accepts an empty 204 body through isNull', async () => {
        const out = await parseResponse(new Response(null, { status: 204 }), { validate: isNull });

        expect(out).toEqual({
            success: true,
            status: 204,
            data: null,
        });
    });

    it('turns a 2xx body that fails its guard into a general failure and reports it', async () => {
        const out = await parseResponse(json({ id: 42 }, 200), { validate: isIdBody });

        expect(out).toEqual({
            success: false,
            status: 200,
            data: { general: { key: 'fetch.invalid_shape' } },
        });
        expect(reportUnexpected).toHaveBeenCalledOnce();
    });

    it('turns a malformed 2xx body into a general failure instead of a fake success', async () => {
        const out = await parseResponse(new Response('{ not json', { status: 200 }), { validate: isIdBody });

        expect(out).toEqual({
            success: false,
            status: 200,
            data: { general: { key: 'fetch.malformed_body' } },
        });
    });

    it('returns field errors that pass the error guard', async () => {
        const errors = { email: { key: 'validation.email' } };
        const out = await parseResponse(json(errors, 422), {
            validate: isIdBody,
            validateError: isEmailErrors,
        });

        expect(out).toEqual({
            success: false,
            status: 422,
            data: errors,
        });
        expect(reportUnexpected).not.toHaveBeenCalled();
    });

    it('returns a general error without an error guard', async () => {
        const errors = { general: { key: 'auth.invalid_credentials' } };
        const out = await parseResponse(json(errors, 401), { validate: isIdBody });

        expect(out).toEqual({
            success: false,
            status: 401,
            data: errors,
        });
    });

    it('reports an error object that fails both guards', async () => {
        const out = await parseResponse(json({ email: 'taken' }, 422), {
            validate: isIdBody,
            validateError: isEmailErrors,
        });

        expect(out).toEqual({
            success: false,
            status: 422,
            data: {
                general: {
                    key: 'fetch.error_status',
                    params: { status: '422' },
                },
            },
        });
        expect(reportUnexpected).toHaveBeenCalledOnce();
    });

    it.each([
        [
            'an empty body',
            new Response(null, { status: 502 }),
        ],
        [
            'a JSON string',
            json('rate limited', 429),
        ],
        [
            'a JSON array',
            json([
                'a',
                'b',
            ], 400),
        ],
        [
            'an HTML page',
            new Response('<h1>Bad gateway</h1>', { status: 502 }),
        ],
    ])('turns %s from a proxy into a general failure without a report', async (_, response) => {
        const out = await parseResponse(response, { validate: isIdBody });

        expect(out).toEqual({
            success: false,
            status: response.status,
            data: {
                general: {
                    key: 'fetch.error_status',
                    params: { status: String(response.status) },
                },
            },
        });
        expect(reportUnexpected).not.toHaveBeenCalled();
    });
});

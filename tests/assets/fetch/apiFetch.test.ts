import { describe, expect, expectTypeOf, it, vi } from 'vitest';
import { authFetch } from '@/shared/Auth/authFetch';
import { type ApiFetch, type ApiGeneralError, apiFetch, type ResponseGuards } from '@/shared/Fetch';
import type { Guard } from '@/shared/TypeGuards/typeGuards';
import { type EmailErrors, type IdBody, isEmailErrors, isIdBody, json } from './fixtures';

type SignUp = {
    email: string;
    password: string;
};

describe('apiFetch', () => {
    it('sends the method, the JSON body and the session cookie and asks for JSON', async () => {
        const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(json({ id: 'u-1' }, 201));
        const body: SignUp = {
            email: 'jan@example.com',
            password: 'secret',
        };

        await apiFetch<IdBody, EmailErrors, SignUp>('POST', '/api/auth/register', {
            body,
            validate: isIdBody,
            validateError: isEmailErrors,
        });

        const init = fetchSpy.mock.calls[0]?.[1];

        expect(init?.method).toBe('POST');
        expect(init?.credentials).toBe('same-origin');
        expect(init?.body).toBe('{"email":"jan@example.com","password":"secret"}');
        expect(new Headers(init?.headers).get('Content-Type')).toBe('application/json');
        expect(new Headers(init?.headers).get('Accept')).toBe('application/json');
    });

    it('asks only once on 401, retrying is not its job', async () => {
        const fetchSpy = vi.spyOn(globalThis, 'fetch').mockResolvedValue(
            json({ general: { key: 'auth.required' } }, 401),
        );

        const result = await apiFetch<IdBody>('GET', '/api/user/me', { validate: isIdBody });

        expect(fetchSpy).toHaveBeenCalledOnce();
        expect(result).toEqual({
            success: false,
            status: 401,
            data: { general: { key: 'auth.required' } },
        });
    });

    it('turns a network failure into status 0 with a general error', async () => {
        vi.spyOn(globalThis, 'fetch').mockRejectedValue(new TypeError('Failed to fetch'));

        const result = await apiFetch<IdBody>('GET', '/api/user/me', { validate: isIdBody });

        expect(result).toEqual({
            success: false,
            status: 0,
            data: { general: { key: 'fetch.network_error' } },
        });
    });

    it('needs the error guard once a caller declares its own errors', () => {
        expectTypeOf(apiFetch).toEqualTypeOf<ApiFetch>();
        expectTypeOf(authFetch).toEqualTypeOf<ApiFetch>();
        expectTypeOf<Parameters<typeof apiFetch<IdBody, EmailErrors>>[2]['validateError']>()
            .toEqualTypeOf<Guard<EmailErrors>>();
        expectTypeOf<Parameters<typeof authFetch<IdBody, EmailErrors>>[2]['validateError']>()
            .toEqualTypeOf<Guard<EmailErrors>>();
        expectTypeOf<ResponseGuards<IdBody, EmailErrors>['validateError']>().toEqualTypeOf<Guard<EmailErrors>>();
        expectTypeOf<ResponseGuards<IdBody, ApiGeneralError>['validateError']>()
            .toEqualTypeOf<Guard<ApiGeneralError> | undefined>();
    });
});

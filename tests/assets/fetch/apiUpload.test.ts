import { beforeEach, describe, expect, it, vi } from 'vitest';
import { apiUpload } from '@/shared/Fetch';
import { type EmailErrors, type IdBody, isEmailErrors, isIdBody } from './fixtures';

class FakeXhr {
    static instances: FakeXhr[] = [];

    onload: (() => void) | null = null;

    onerror: (() => void) | null = null;

    method = '';

    headers: Record<string, string> = {};

    status = 0;

    statusText = '';

    responseText = '';

    constructor() {
        FakeXhr.instances.push(this);
    }

    open(method: string): void {
        this.method = method;
    }

    setRequestHeader(name: string, value: string): void {
        if (this.method === '') {
            throw new Error('setRequestHeader before open');
        }
        this.headers[name] = value;
    }

    send = vi.fn<() => void>();

    fireLoad(status: number, responseText: string): void {
        this.status = status;
        this.responseText = responseText;
        this.onload?.();
    }

    fireError(): void {
        this.onerror?.();
    }
}

const lastXhr = (): FakeXhr => {
    const xhr = FakeXhr.instances.at(-1);

    if (xhr === undefined) {
        throw new Error('no XMLHttpRequest was constructed');
    }

    return xhr;
};

describe('apiUpload', () => {
    beforeEach(() => {
        FakeXhr.instances = [];
        vi.stubGlobal('XMLHttpRequest', FakeXhr);
    });

    it('settles on a bodyless 204 instead of hanging', async () => {
        const promise = apiUpload<IdBody>('/api/files', new FormData(), { validate: isIdBody });

        lastXhr().fireLoad(204, '');

        expect(await promise).toEqual({
            success: false,
            status: 204,
            data: { general: { key: 'fetch.invalid_shape' } },
        });
    });

    it('returns field errors checked by the error guard', async () => {
        const promise = apiUpload<IdBody, EmailErrors>('/api/files', new FormData(), {
            validate: isIdBody,
            validateError: isEmailErrors,
        });

        lastXhr().fireLoad(422, JSON.stringify({ email: { key: 'validation.email' } }));

        expect(lastXhr().headers['Accept']).toBe('application/json');
        expect(await promise).toEqual({
            success: false,
            status: 422,
            data: { email: { key: 'validation.email' } },
        });
    });

    it('turns a network error into a general error', async () => {
        const promise = apiUpload<IdBody>('/api/files', new FormData(), { validate: isIdBody });

        lastXhr().fireError();

        expect(await promise).toEqual({
            success: false,
            status: 0,
            data: { general: { key: 'fetch.network_error' } },
        });
    });
});

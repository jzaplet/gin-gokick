import * as Sentry from '@sentry/vue';
import { afterEach, describe, expect, it } from 'vitest';
import { createApp } from 'vue';
import { startSentry } from '@/shared/Sentry/startSentry';

const meta = document.createElement('meta');

const start = (): void => {
    meta.name = 'sentry';
    meta.setAttribute('data-dsn', 'https://public@example.com/1');
    document.head.append(meta);

    startSentry(createApp({}));
};

describe('startSentry', () => {
    afterEach(async () => {
        meta.remove();
        await Sentry.close();
    });

    it('keeps the IP address of the visitor out of Sentry', () => {
        start();

        expect(Sentry.getClient()?.getSdkMetadata()?.sdk?.settings?.infer_ip).toBe('never');
    });

    it('keeps the query of every address and drops its fragment', async () => {
        start();

        const options = Sentry.getClient()?.getOptions();
        const event = await options?.beforeSend?.({
            type: undefined,
            request: {
                url: 'https://example.com/app/reset?ref=mail#token=secret',
                headers: { Referer: 'https://example.com/?ref=ad#top' },
            },
        }, {});
        const breadcrumb = options?.beforeBreadcrumb?.({
            category: 'navigation',
            data: {
                from: '/app/login?ref=mail#form',
                to: '/app/reset#token=secret',
            },
        });

        expect(event?.request).toEqual({
            url: 'https://example.com/app/reset?ref=mail',
            headers: { Referer: 'https://example.com/?ref=ad' },
        });
        expect(breadcrumb?.data).toEqual({
            from: '/app/login?ref=mail',
            to: '/app/reset',
        });
    });
});

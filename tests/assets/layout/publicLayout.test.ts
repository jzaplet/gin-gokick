import { enableAutoUnmount, flushPromises } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { router } from '@/router';
import { t } from '@/shared/I18n/Texts/translate';
import { openApp } from '../app/openApp';

enableAutoUnmount(afterEach);

afterEach(() => {
    vi.useRealTimers();
});

const robots = (): string | undefined => document.head.querySelector<HTMLMetaElement>('meta[name="robots"]')?.content;

describe.each([
    '/cs/app/login',
    '/cs/app/register',
    '/cs/app/nothing',
])('the page %s', (path) => {
    it('offers the cookie settings in the footer only when a tool waits for consent', async () => {
        expect((await openApp(path)).find('footer [data-consent-settings]').exists()).toBe(false);

        const tracking = document.createElement('meta');

        tracking.name = 'tracking';
        tracking.dataset['ga4'] = 'G-TEST';
        document.head.append(tracking);
        const settings = (await openApp(path)).find('footer [data-consent-settings]');

        tracking.remove();

        expect(settings.exists()).toBe(true);
    });
});

describe('the public layout', () => {
    it('ends with the copyright of the current year', async () => {
        vi.useFakeTimers({ toFake: ['Date'] });
        vi.setSystemTime(new Date(2031, 5, 1));

        const footer = (await openApp('/cs/app/login')).get('footer');

        expect(footer.text()).toContain(t('footer.copyright', { year: '2031' }));
    });

    it('asks robots not to index a page that does not exist, and only while it shows', async () => {
        expect(robots()).toBeUndefined();

        await openApp('/cs/app/nothing');

        expect(robots()).toBe('noindex');
        expect(document.head.querySelectorAll('meta[name="robots"]')).toHaveLength(1);

        await router.push('/cs/app/login');
        await flushPromises();

        expect(robots()).toBeUndefined();
    });
});

import { afterEach, describe, expect, it, vi } from 'vitest';
import { t } from '@/shared/I18n/Texts/translate';
import {
    banner,
    bannerGone,
    clearConsentPage,
    click,
    inSettings,
    openFromLink,
    saved,
    settingsOpen,
    start,
    toggle,
} from './consentPage';

describe('the stored cookie consent', () => {
    afterEach(clearConsentPage);

    it('applies a stored choice without asking', () => {
        const { choices } = start({ ga4: 'G-AB12CD34EF' }, saved('ga4=G-AB12CD34EF', { analytics: true }));

        expect(banner()).toBeNull();
        expect(choices).toEqual([{ analytics: true }]);
    });

    it.each([
        [
            'a tool of a new category',
            {
                ga4: 'G-AB12CD34EF',
                metaPixel: '1234567890123456',
            },
            saved('ga4=G-AB12CD34EF', { analytics: true }),
        ],
        [
            'a tool joining a decided category',
            {
                ga4: 'G-AB12CD34EF',
                googleAds: 'AW-123456789',
                metaPixel: '1234567890123456',
            },
            saved('ga4=G-AB12CD34EF&meta_pixel=1234567890123456', {
                analytics: true,
                marketing: true,
            }),
        ],
        [
            'a changed ID',
            { ga4: 'G-ZZ98YY76XX' },
            saved('ga4=G-AB12CD34EF', { analytics: true }),
        ],
    ])('asks again from scratch after %s', async (_change, ids, stored) => {
        const { choices, reload } = start(ids, stored);

        expect(banner()).not.toBeNull();
        expect(choices).toEqual([{}]);

        await click(t('consent.details'));

        expect(toggle('analytics')?.checked).toBe(false);

        await inSettings(t('consent.accept_all'));

        expect(choices.at(-1)?.analytics).toBe(true);
        expect(reload).not.toHaveBeenCalled();
    });

    it.each([
        [
            'without its tools',
            '{"analytics":true}',
        ],
        [
            'with a malformed choice',
            '{"tools":"ga4=G-AB12CD34EF","choice":{"analytics":"yes"}}',
        ],
        [
            'as no JSON',
            '{analytics',
        ],
    ])('asks when the stored choice comes %s', (_problem, stored) => {
        const { choices } = start({ ga4: 'G-AB12CD34EF' }, stored);

        expect(banner()).not.toBeNull();
        expect(choices).toEqual([{}]);
    });

    it('opens the settings from a link with the current choice and reloads the page after a withdrawal', async () => {
        const tools = 'ga4=G-AB12CD34EF&meta_pixel=1234567890123456';
        const { reload } = start(
            {
                ga4: 'G-AB12CD34EF',
                metaPixel: '1234567890123456',
            },
            saved(tools, {
                analytics: true,
                marketing: false,
            }),
        );

        const event = await openFromLink();

        expect(event.defaultPrevented).toBe(true);
        expect(settingsOpen()).toBe(true);
        expect(toggle('analytics')?.checked).toBe(true);

        await inSettings(t('consent.necessary_only'));

        expect(localStorage.getItem('consent')).toBe(
            saved(tools, {
                analytics: false,
                marketing: false,
            }),
        );
        expect(reload).toHaveBeenCalledOnce();
    });

    it('grants more without a reload and shows the new choice next time', async () => {
        const { choices, reload } = start({ ga4: 'G-AB12CD34EF' }, saved('ga4=G-AB12CD34EF', { analytics: false }));

        await openFromLink();
        await inSettings(t('consent.accept_all'));

        expect(choices.at(-1)).toEqual({ analytics: true });
        expect(reload).not.toHaveBeenCalled();

        await openFromLink();

        expect(toggle('analytics')?.checked).toBe(true);
    });

    it('keeps the choice for the page when localStorage refuses', async () => {
        vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
            throw new DOMException('refused', 'SecurityError');
        });
        const { choices } = start({ ga4: 'G-AB12CD34EF' });

        await click(t('consent.understood'));

        expect(choices.at(-1)).toEqual({ analytics: true });
        await bannerGone();
    });
});

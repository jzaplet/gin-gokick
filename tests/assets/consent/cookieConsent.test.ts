import { flushPromises } from '@vue/test-utils';
import { afterEach, describe, expect, it } from 'vitest';
import { t } from '@/shared/I18n/Texts/translate';
import {
    banner,
    bannerGone,
    button,
    clearConsentPage,
    click,
    inSettings,
    saved,
    settingsOpen,
    start,
    toggle,
} from './consentPage';

describe('the cookie consent', () => {
    afterEach(clearConsentPage);

    it('offers refusing next to accepting and names the button that consents', () => {
        start({ metaPixel: '1234567890123456' });

        expect(banner()?.querySelector('p')?.textContent).toContain(t('consent.understood'));
        expect(button(t('consent.necessary_only')).closest('section')).toBe(banner());
        expect(button(t('consent.understood')).closest('section')).toBe(banner());
        expect(settingsOpen()).toBe(false);
    });

    it('grants every offered category', async () => {
        const { choices, reload } = start({
            ga4: 'G-AB12CD34EF',
            metaPixel: '1234567890123456',
        });

        await click(t('consent.understood'));

        expect(localStorage.getItem('consent')).toBe(
            saved('ga4=G-AB12CD34EF&meta_pixel=1234567890123456', {
                analytics: true,
                marketing: true,
            }),
        );
        expect(choices).toEqual([{}, {
            analytics: true,
            marketing: true,
        }]);
        await bannerGone();
        expect(reload).not.toHaveBeenCalled();
    });

    it('accepts only the necessary cookies', async () => {
        const { choices, reload } = start({
            ga4: 'G-AB12CD34EF',
            googleAds: 'AW-123456789',
        });

        await click(t('consent.necessary_only'));

        expect(localStorage.getItem('consent')).toBe(
            saved('ga4=G-AB12CD34EF&google_ads=AW-123456789', {
                analytics: false,
                marketing: false,
            }),
        );
        expect(choices.at(-1)).toEqual({
            analytics: false,
            marketing: false,
        });
        expect(reload).not.toHaveBeenCalled();
    });

    it('lists the enabled tools in the settings, switched off, with the necessary cookies always on', async () => {
        start({
            googleAds: 'AW-123456789',
            metaPixel: '1234567890123456',
        });

        await click(t('consent.details'));

        expect(settingsOpen()).toBe(true);
        expect(document.querySelector('dialog')?.textContent).toContain('Google\u00a0Ads');
        expect(document.querySelector('dialog')?.textContent).toContain('Meta\u00a0Pixel');
        expect(toggle('marketing')?.checked).toBe(false);
        expect(toggle('analytics')).toBeNull();
        expect(toggle('necessary')?.checked).toBe(true);
        expect(toggle('necessary')?.disabled).toBe(true);
    });

    it('saves only the switched-on categories', async () => {
        const { choices } = start({
            ga4: 'G-AB12CD34EF',
            metaPixel: '1234567890123456',
        });

        await click(t('consent.details'));
        toggle('marketing')?.click();
        await inSettings(t('consent.save'));

        expect(choices.at(-1)).toEqual({
            analytics: false,
            marketing: true,
        });
        expect(settingsOpen()).toBe(false);
        await bannerGone();
    });

    it('hides the banner behind the settings and brings it back when they close without a choice', async () => {
        const { choices } = start({ ga4: 'G-AB12CD34EF' });

        for (const close of [
            (): void => {
                document.querySelector<HTMLButtonElement>(`dialog button[aria-label="${t('consent.close')}"]`)?.click();
            },
            (): void => {
                document.querySelector('dialog')?.dispatchEvent(new Event('cancel', { cancelable: true }));
            },
        ]) {
            await click(t('consent.details'));
            await bannerGone();

            close();
            await flushPromises();

            expect(settingsOpen()).toBe(false);
            expect(banner()).not.toBeNull();
        }
        expect(localStorage.getItem('consent')).toBeNull();
        expect(choices).toEqual([{}]);
    });
});

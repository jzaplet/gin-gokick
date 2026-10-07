import { describe, expect, it, vi } from 'vitest';
import { router } from '@/router';
import { dictionaries } from '@/shared/I18n/Dictionary/dictionaries';
import { dictionary as englishDictionary } from '@/shared/I18n/Dictionary/Locales/en_US';
import type { Dictionary } from '@/shared/I18n/Dictionary/types/Dictionary';
import { appStorage } from '@/shared/Storage/appStorage';
import { t } from '@/shared/I18n/Texts/translate';
import { offerEnglish } from '../i18n/languages';

describe('router', () => {
    it.each([
        '/cs/app/nothing/here',
        '/cs/app',
        '/cs/app/',
    ])('shows the not-found page on %s', (path) => {
        expect(router.resolve(path).name).toBe('not-found');
    });

    it('keeps the texts of a navigation that won over one still loading another language', async () => {
        const source = dictionaries['en_US'];
        const czech = t('register.title');
        const { promise, resolve } = Promise.withResolvers<Dictionary>();

        if (source === undefined) {
            throw new Error('The test needs the English dictionary');
        }

        appStorage().remove('dictionary.en_US');
        offerEnglish();
        await router.push('/cs/app/login');
        const download = vi.spyOn(source, 'load').mockReturnValue(promise);
        const english = router.push('/en/app/login');

        await vi.waitFor(() => {
            expect(download).toHaveBeenCalledOnce();
        });
        await router.push('/cs/app/register');
        resolve(englishDictionary);
        await english;

        expect(router.currentRoute.value.fullPath).toBe('/cs/app/register');
        expect(t('register.title')).toBe(czech);
        expect(document.documentElement.lang).toBe('cs-CZ');
    });
});

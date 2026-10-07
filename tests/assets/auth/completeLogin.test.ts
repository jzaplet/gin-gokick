import { afterEach, describe, expect, it } from 'vitest';
import { router } from '@/router';
import { completeLogin } from '@/shared/Auth/Login/completeLogin';
import { redirectAfterLogin } from '@/shared/Auth/Login/loginRedirect';
import { forgetSession, signedIn } from '@/shared/Auth/Session/currentSession';
import { dictionary as englishDictionary } from '@/shared/I18n/Dictionary/Locales/en_US';
import { t } from '@/shared/I18n/Texts/translate';
import { localeMeta, offerEnglish } from '../i18n/languages';

const id = '0192f3a4-5b6c-7d8e-9f01-23456789abcd';

describe('redirectAfterLogin', () => {
    it.each([
        '/dashboard',
        '/settings?tab=mail#top',
    ])('returns to %s in the language of the account', (redirect) => {
        expect(redirectAfterLogin(redirect, 'en')).toBe(`/en/app${redirect}`);
    });

    it.each([
        undefined,
        null,
        '',
        ['/dashboard'],
        'dashboard',
        'https://evil.example',
        '//evil.example',
        '/\\evil.example',
    ])('goes to the dashboard instead of %j', (redirect) => {
        expect(redirectAfterLogin(redirect, 'cs')).toEqual({
            name: 'dashboard',
            params: { lang: 'cs' },
        });
    });
});

describe('completeLogin', () => {
    afterEach(forgetSession);

    it.each([
        [
            '/dashboard?tab=mail#top',
            '/en/app/dashboard?tab=mail#top',
        ],
        [
            undefined,
            '/en/app/dashboard',
        ],
    ])('opens the target %s in the language of the account without a reload', async (redirect, path) => {
        offerEnglish();
        await router.push('/cs/app/login');

        await completeLogin({
            id,
            locale: 'en_US',
        }, redirect);

        expect(router.currentRoute.value.fullPath).toBe(path);
        expect(t('login.title')).toBe(englishDictionary['login.title']);
        expect(document.documentElement.lang).toBe('en-US');
        expect(localStorage.getItem('session')).toBe(`{"id":"${id}"}`);
    });

    it.each([
        [
            'en_GB',
            'cs_CZ en_GB',
        ],
        [
            'de_DE',
            'cs_CZ',
        ],
    ])(
        'opens the target in the shown language when the language %s of the account does not load or the page lacks it',
        async (locale, languages) => {
            localeMeta()?.setAttribute('data-languages', languages);
            await router.push('/cs/app/login');
            const shown = t('login.title');

            await completeLogin({
                id,
                locale,
            }, '/dashboard');

            expect(router.currentRoute.value.fullPath).toBe('/cs/app/dashboard');
            expect(t('login.title')).toBe(shown);
            expect(signedIn()).toBe(true);
        },
    );
});

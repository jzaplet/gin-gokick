import { describe, expect, it } from 'vitest';
import { homePageOf, offeredLocale, switcherLocales } from '@/shared/I18n/Page/localeMeta';
import { localeMeta, setHomes } from './languages';

describe('a language', () => {
    it('has the home Go named for it and fails without one or on a malformed one', () => {
        expect(() => homePageOf('en')).toThrow('The page names no home of the language "en"');

        setHomes('cs=/ en=//evil.example');

        expect(() => homePageOf('cs')).toThrow(
            'The page names a home "en=//evil.example" that is no language and its path',
        );

        setHomes('cs=/ en=/en');

        expect(homePageOf('cs')).toBe('/');
        expect(homePageOf('en')).toBe('/en');
    });

    it('finds the locale of the page first and then the one of the switcher', () => {
        localeMeta()?.setAttribute('content', 'en_GB');
        localeMeta()?.setAttribute('data-languages', 'cs_CZ en_US');

        expect(offeredLocale('en')).toBe('en_GB');
        expect(offeredLocale('cs')).toBe('cs_CZ');
        expect(offeredLocale('de')).toBeUndefined();
    });
});

describe('switcherLocales', () => {
    it('has none for a page of one language', () => {
        expect(switcherLocales()).toEqual([]);

        localeMeta()?.setAttribute('data-languages', '');

        expect(switcherLocales()).toEqual([]);
    });
});

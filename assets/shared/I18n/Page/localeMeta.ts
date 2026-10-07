const posixLocale = /^[a-z]{2}_[A-Z]{2}$/;

const homeEntry = /^[a-z]{2}=\/(?:[a-z]{2})?$/;

const localeMeta = (): HTMLMetaElement | null => document.querySelector<HTMLMetaElement>('meta[name="locale"]');

export const servedLocale = (): string => {
    const locale = localeMeta()?.content ?? '';

    if (posixLocale.test(locale) === false) {
        throw new Error(`The page has no locale, its meta says "${locale}"`);
    }

    return locale;
};

export const languageOfLocale = (locale: string): string => locale.slice(0, 2);

export const languageTag = (locale: string): string => locale.replace('_', '-');

export const homePageOf = (language: string): string => {
    const entries = (localeMeta()?.dataset['homes'] ?? '').split(' ').filter((entry) => entry !== '');
    const malformed = entries.find((entry) => homeEntry.test(entry) === false);

    if (malformed !== undefined) {
        throw new Error(`The page names a home "${malformed}" that is no language and its path`);
    }

    const home = entries.find((entry) => entry.startsWith(`${language}=`))?.slice(language.length + 1);

    if (home === undefined) {
        throw new Error(`The page names no home of the language "${language}"`);
    }

    return home;
};

export const switcherLocales = (): string[] => {
    const languages = (localeMeta()?.dataset['languages'] ?? '').split(' ').filter((locale) => locale !== '');
    const malformed = languages.find((locale) => posixLocale.test(locale) === false);

    if (malformed !== undefined) {
        throw new Error(`The page names a language "${malformed}" that is no locale`);
    }

    return languages;
};

export const offeredLocale = (language: string): string | undefined =>
    [
        servedLocale(),
        ...switcherLocales(),
    ].find((locale) => languageOfLocale(locale) === language);

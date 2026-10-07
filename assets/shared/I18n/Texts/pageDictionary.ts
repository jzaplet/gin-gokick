import { shallowRef } from 'vue';
import { dictionaries } from '@/shared/I18n/Dictionary/dictionaries';
import type { Dictionary } from '@/shared/I18n/Dictionary/types/Dictionary';
import type { DictionarySource } from '@/shared/I18n/Dictionary/types/DictionarySource';
import { languageOfLocale, servedLocale, languageTag } from '@/shared/I18n/Page/localeMeta';
import { readStoredDictionary, writeStoredDictionary } from '@/shared/I18n/Texts/storedDictionary';
import type { PageDictionary } from '@/shared/I18n/Texts/types/PageDictionary';
import { reportUnexpected } from '@/shared/Sentry/reportUnexpected';

const loads = new Map<string, Promise<PageDictionary>>();

const loaded = new Map<string, PageDictionary>();

const shown = shallowRef<PageDictionary>();

const sourceOf = (locale: string): DictionarySource => {
    const source = Object.hasOwn(dictionaries, locale) ? dictionaries[locale] : undefined;

    if (source === undefined) {
        throw new Error(`The locale ${locale} has no dictionary`);
    }

    return source;
};

const download = async (locale: string, source: DictionarySource): Promise<Dictionary> => {
    const messages = await source.load();

    writeStoredDictionary(locale, source.version, messages);

    return messages;
};

const load = async (locale: string): Promise<PageDictionary> => {
    const source = sourceOf(locale);
    const messages = readStoredDictionary(locale, source.version) ?? (await download(locale, source));
    const dictionary = {
        locale,
        messages,
    };

    loaded.set(locale, dictionary);

    return dictionary;
};

const dictionaryOf = (locale: string): Promise<PageDictionary> => {
    const cached = loads.get(locale);

    if (cached !== undefined) {
        return cached;
    }

    const loading = load(locale).catch((error: unknown) => {
        loads.delete(locale);
        throw error;
    });

    loads.set(locale, loading);

    return loading;
};

export const loadPageDictionary = async (): Promise<void> => {
    const dictionary = await dictionaryOf(servedLocale());

    shown.value ??= dictionary;
};

export const preloadDictionary = (locale: string): void => {
    dictionaryOf(locale).catch(() => undefined);
};

export const loadDictionary = async (locale: string): Promise<boolean> => {
    try {
        await dictionaryOf(locale);

        return true;
    } catch (error) {
        reportUnexpected(`The dictionary of ${locale} did not load: ${String(error)}`);

        return false;
    }
};

export const showDictionary = (locale: string): boolean => {
    const dictionary = loaded.get(locale);

    if (dictionary === undefined) {
        return false;
    }

    shown.value = dictionary;
    document.documentElement.lang = languageTag(locale);

    return true;
};

export const pageDictionary = (): PageDictionary => {
    if (shown.value === undefined) {
        throw new Error('The dictionary of the page is not loaded yet, the entry awaits loadPageDictionary first');
    }

    return shown.value;
};

export const shownLocale = (): string => pageDictionary().locale;

export const shownLanguage = (): string => languageOfLocale(shownLocale());

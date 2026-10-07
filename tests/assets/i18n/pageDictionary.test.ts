import { beforeEach, describe, expect, it, vi } from 'vitest';
import type { Dictionary } from '@/shared/I18n/Dictionary/types/Dictionary';
import type { DictionarySource } from '@/shared/I18n/Dictionary/types/DictionarySource';
import type { t as translate } from '@/shared/I18n/Texts/translate';
import type { loadDictionary, preloadDictionary, showDictionary } from '@/shared/I18n/Texts/pageDictionary';
import type { reportUnexpected } from '@/shared/Sentry/reportUnexpected';
import { appStorage } from '@/shared/Storage/appStorage';

vi.mock('@/shared/Sentry/reportUnexpected', () => ({ reportUnexpected: vi.fn() }));

type Fresh = {
    source: DictionarySource;
    english: DictionarySource;
    load: () => Promise<void>;
    loadOne: typeof loadDictionary;
    preload: typeof preloadDictionary;
    show: typeof showDictionary;
    report: typeof reportUnexpected;
    t: typeof translate;
};

const czech = 'cs_CZ';

const stored = (): unknown => appStorage().read(`dictionary.${czech}`);

const store = (version: string, messages: unknown): void => {
    appStorage().write(`dictionary.${czech}`, {
        version,
        messages,
    });
};

const fresh = async (): Promise<Fresh> => {
    vi.resetModules();

    const { dictionaries } = await import('@/shared/I18n/Dictionary/dictionaries');
    const {
        loadDictionary,
        loadPageDictionary,
        preloadDictionary,
        showDictionary,
    } = await import('@/shared/I18n/Texts/pageDictionary');
    const { t } = await import('@/shared/I18n/Texts/translate');
    const { reportUnexpected } = await import('@/shared/Sentry/reportUnexpected');
    const source = dictionaries[czech];
    const english = dictionaries['en_US'];

    if (source === undefined || english === undefined) {
        throw new Error('The test needs the Czech and the English dictionary');
    }

    return {
        source,
        english,
        load: loadPageDictionary,
        loadOne: loadDictionary,
        preload: preloadDictionary,
        show: showDictionary,
        report: reportUnexpected,
        t,
    };
};

const roundTrip = (dictionary: Dictionary): unknown => JSON.parse(JSON.stringify(dictionary));

describe('loadPageDictionary', () => {
    beforeEach(() => {
        appStorage().remove(`dictionary.${czech}`);
    });

    it('downloads the dictionary again when the stored version differs', async () => {
        store('an older version', { 'login.title': 'Starý titulek' });

        const { source, load, t } = await fresh();

        await load();
        const downloaded = await source.load();

        expect(t('login.title')).toBe(downloaded['login.title']);
        expect(stored()).toEqual({
            version: source.version,
            messages: roundTrip(downloaded),
        });
    });
});

describe('showDictionary', () => {
    it('shows the texts of another locale once it loaded and tags the page with it', async () => {
        const { english, load, loadOne, show, t } = await fresh();

        await load();

        expect(show('en_US')).toBe(false);
        expect(await loadOne('en_US')).toBe(true);
        expect(show('en_US')).toBe(true);
        expect(t('login.title')).toBe((await english.load())['login.title']);
        expect(document.documentElement.lang).toBe('en-US');
    });

    it('keeps the shown texts and reports a dictionary that does not download', async () => {
        const { english, load, loadOne, report, show, source, t } = await fresh();

        appStorage().remove('dictionary.en_US');
        vi.spyOn(english, 'load').mockRejectedValue(new Error('offline'));
        await load();

        expect(await loadOne('en_US')).toBe(false);
        expect(show('en_US')).toBe(false);
        expect(t('login.title')).toBe((await source.load())['login.title']);
        expect(report).toHaveBeenCalledWith('The dictionary of en_US did not load: Error: offline');
    });
});

import { loadDictionary, showDictionary } from '@/shared/I18n/Texts/pageDictionary';

export const localeMeta = (): HTMLMetaElement | null =>
    document.head.querySelector<HTMLMetaElement>('meta[name="locale"]');

export const setHomes = (homes: string): void => {
    localeMeta()?.setAttribute('data-homes', homes);
};

export const offerEnglish = (): void => {
    localeMeta()?.setAttribute('data-languages', 'cs_CZ en_US');
    setHomes('cs=/ en=/en');
};

export const showEnglish = async (): Promise<void> => {
    await loadDictionary('en_US');
    showDictionary('en_US');
};

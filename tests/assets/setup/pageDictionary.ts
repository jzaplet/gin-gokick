import { beforeEach } from 'vitest';
import { loadPageDictionary, showDictionary } from '@/shared/I18n/Texts/pageDictionary';

await loadPageDictionary();

beforeEach(() => {
    showDictionary('cs_CZ');
});

import type { ConsentCategory } from '@/shared/Consent/types/ConsentCategory';
import type { MessageKeyWith, PlainMessageKey } from '@/shared/I18n/Texts/translate';

export type CategoryRow = {
    category: ConsentCategory;
    title: PlainMessageKey;
    text: MessageKeyWith<{ tools: string }>;
    tools: string;
};

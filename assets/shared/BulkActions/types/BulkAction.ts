import type { PlainMessageKey } from '@/shared/I18n/Texts/translate';

export type BulkAction<K extends string> = {
    key: K;
    label: PlainMessageKey;
};

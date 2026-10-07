import type { PlainMessageKey } from '@/shared/I18n/Texts/translate';

export type GridColumn<S extends string> = {
    key: string;
    label: PlainMessageKey;
    sort?: S;
    align?: 'left' | 'right';
    hideLabel?: boolean;
};

import type { Message, PluralPart } from '@/shared/I18n/Dictionary/types/Message';

export const pound = { type: 'pound' } as const;

export const plural = (
    cases: PluralPart['cases'],
    more: Partial<Pick<PluralPart, 'offset' | 'exact'>> = {},
): Message => [{
    type: 'plural',
    name: 'n',
    offset: more.offset ?? 0,
    exact: more.exact ?? {},
    cases,
}];

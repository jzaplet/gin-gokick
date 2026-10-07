import type { Dictionary } from '@/shared/I18n/Dictionary/types/Dictionary';

export type PageDictionary = {
    readonly locale: string;
    readonly messages: Partial<Dictionary>;
};

import type { Dictionary } from '@/shared/I18n/Dictionary/types/Dictionary';

export type DictionarySource = {
    readonly version: string;
    readonly name: string;
    readonly load: () => Promise<Dictionary>;
};

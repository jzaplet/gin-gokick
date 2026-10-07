import { languageTag } from '@/shared/I18n/Page/localeMeta';
import { shownLocale } from '@/shared/I18n/Texts/pageDictionary';

const formats = new Map<string, Intl.DateTimeFormat>();

const formatOf = (tag: string): Intl.DateTimeFormat => {
    const known = formats.get(tag);

    if (known !== undefined) {
        return known;
    }

    const format = new Intl.DateTimeFormat(tag, {
        dateStyle: 'medium',
        timeStyle: 'short',
    });

    formats.set(tag, format);

    return format;
};

export const formatDateTime = (value: string): string => formatOf(languageTag(shownLocale())).format(new Date(value));

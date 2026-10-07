import { computed, type ComputedRef } from 'vue';
import { pageWindow } from '@/shared/Grid/Pagination/pageWindow';
import { languageTag } from '@/shared/I18n/Page/localeMeta';
import { shownLocale } from '@/shared/I18n/Texts/pageDictionary';
import { t } from '@/shared/I18n/Texts/translate';

type Paging = {
    page: number;
    perPage: number;
    total: number;
};

type GridPagination = {
    range: ComputedRef<string>;
    pages: ComputedRef<number>;
    numbers: ComputedRef<number[]>;
};

export const useGridPagination = (paging: Paging): GridPagination => {
    const pages = computed((): number => Math.max(1, Math.ceil(paging.total / paging.perPage)));

    const range = computed((): string => {
        const format = new Intl.NumberFormat(languageTag(shownLocale()));
        const from = (paging.page - 1) * paging.perPage + 1;

        return t('grid.range', {
            from: format.format(from),
            to: format.format(Math.min(paging.page * paging.perPage, paging.total)),
            total: format.format(paging.total),
        });
    });

    const numbers = computed((): number[] => pageWindow(paging.page, pages.value));

    return {
        range,
        pages,
        numbers,
    };
};

import { computed, type ComputedRef } from 'vue';
import type { BulkAction } from '@/shared/BulkActions/types/BulkAction';
import { languageTag } from '@/shared/I18n/Page/localeMeta';
import { shownLocale } from '@/shared/I18n/Texts/pageDictionary';
import { t } from '@/shared/I18n/Texts/translate';

type Selected = {
    count: number;
    total: number;
    all: boolean;
};

type BulkActionBar<K extends string> = {
    count: ComputedRef<string>;
    selected: ComputedRef<string>;
    selectAll: ComputedRef<string>;
    labelOf: (action: BulkAction<K>) => string;
};

export const useBulkActionBar = <K extends string>(selection: Selected): BulkActionBar<K> => {
    const format = computed(() => new Intl.NumberFormat(languageTag(shownLocale())));
    const count = computed(() => format.value.format(selection.count));
    const total = computed(() => format.value.format(selection.total));

    const selected = computed(() => (selection.all
        ? t('bulk.selected_all', { total: total.value })
        : t('bulk.selected', {
                count: count.value,
                total: total.value,
            })));

    const selectAll = computed(() => t('bulk.select_all', { total: total.value }));

    const labelOf = (action: BulkAction<K>): string => t('bulk.action', {
        action: t(action.label),
        count: count.value,
    });

    return {
        count,
        selected,
        selectAll,
        labelOf,
    };
};

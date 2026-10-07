import { computed, onScopeDispose, reactive, type Reactive, type Ref, watch } from 'vue';
import { createDebounce } from '@/shared/Debounce/createDebounce';
import { sameFilters, trimmed } from '@/shared/Grid/Filters/gridFilters';
import type { GridFilters } from '@/shared/Grid/types/GridFilters';

type FilterChanges = {
    apply: (filters: GridFilters) => Promise<void>;
    changing: () => void;
};

type GridFilterFields<F extends GridFilters> = {
    filters: Reactive<F>;
    appliedFilters: Readonly<Ref<F>>;
    filtered: Readonly<Ref<boolean>>;
    filled: Readonly<Ref<boolean>>;
    fill: (values: GridFilters) => void;
    clearFilters: () => Promise<void>;
};

const filterDelay = 400;

export const useGridFilters = <F extends GridFilters>(
    defaults: F,
    applied: Readonly<Ref<GridFilters>>,
    changes: FilterChanges,
): GridFilterFields<F> => {
    const filters = reactive({ ...defaults });
    const typed: GridFilters = filters;
    const debounce = createDebounce(filterDelay);

    const fill = (values: GridFilters): void => {
        Object.assign(typed, values);
    };

    const filterBy = (wanted: GridFilters): void => {
        if (sameFilters(wanted, applied.value)) {
            debounce.cancel();

            return;
        }

        changes.changing();
        debounce.run(() => {
            void changes.apply(wanted);
        });
    };

    const clearFilters = async (): Promise<void> => {
        fill(defaults);
        debounce.cancel();

        if (sameFilters(applied.value, defaults)) {
            return;
        }

        await changes.apply(defaults);
    };

    const appliedFilters = computed(() => ({
        ...defaults,
        ...applied.value,
    }));

    const filtered = computed(() => sameFilters(applied.value, defaults) === false);

    const filled = computed(() => sameFilters(trimmed(typed), defaults) === false);

    fill(applied.value);
    watch(() => trimmed(typed), filterBy);
    onScopeDispose(debounce.cancel);

    return {
        filters,
        appliedFilters,
        filtered,
        filled,
        fill,
        clearFilters,
    };
};

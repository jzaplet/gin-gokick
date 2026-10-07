import { type DeepReadonly, onMounted, readonly, type Reactive, type Ref, shallowRef, watch } from 'vue';
import { useGridAddress } from '@/shared/Grid/Composables/useGridAddress';
import { useGridFilters } from '@/shared/Grid/Composables/useGridFilters';
import { useGridRows } from '@/shared/Grid/Composables/useGridRows';
import { sameFilters } from '@/shared/Grid/Filters/gridFilters';
import { createSelection } from '@/shared/Grid/Selection/gridSelection';
import { nextSort } from '@/shared/Grid/Sort/sortState';
import type { GridFilters } from '@/shared/Grid/types/GridFilters';
import type { GridSelection } from '@/shared/Grid/types/GridSelection';
import type { GridSort } from '@/shared/Grid/types/GridSort';
import type { GridState } from '@/shared/Grid/types/GridState';
import type { Guard } from '@/shared/TypeGuards/typeGuards';

type GridOptions<S extends string, T, F extends GridFilters> = {
    name: string;
    url: string;
    item: Guard<T>;
    isSort: Guard<S>;
    sort: GridSort<S>;
    filters: F;
    isFilter: { [K in keyof F]: (value: string) => boolean };
    perPage: number;
    selectable?: (item: T) => boolean;
};

export type Grid<S extends string, T extends { id: string }, F extends GridFilters> = {
    items: DeepReadonly<Ref<T[]>>;
    total: Readonly<Ref<number>>;
    page: Readonly<Ref<number>>;
    perPage: number;
    sort: DeepReadonly<Ref<GridSort<S>>>;
    loading: Readonly<Ref<boolean>>;
    failed: Readonly<Ref<boolean>>;
    filters: Reactive<F>;
    appliedFilters: Readonly<Ref<F>>;
    filtered: Readonly<Ref<boolean>>;
    filled: Readonly<Ref<boolean>>;
    selection: GridSelection<T>;
    reload: () => Promise<void>;
    sortBy: (column: S) => Promise<void>;
    goTo: (page: number) => Promise<void>;
    clearFilters: () => Promise<void>;
};

const sameState = <S extends string>(one: GridState<S>, other: GridState<S>): boolean =>
    one.page === other.page
    && one.sort.sortBy === other.sort.sortBy
    && one.sort.sortDir === other.sort.sortDir
    && sameFilters(one.filters, other.filters);

export const useGrid = <S extends string, T extends { id: string }, F extends GridFilters>(
    options: GridOptions<S, T, F>,
): Grid<S, T, F> => {
    const defaults: GridState<S> = {
        page: 1,
        sort: options.sort,
        filters: { ...options.filters },
    };
    const address = useGridAddress(options, defaults);
    const initial = address.read();
    const page = shallowRef(initial.page);
    const sort = shallowRef<GridSort<S>>(initial.sort);
    const applied = shallowRef<GridFilters>(initial.filters);
    const rows = useGridRows(options);
    const selection = createSelection(rows.items, rows.selectable, options.selectable ?? (() => true));
    let shown = initial;

    const current = (): GridState<S> => ({
        page: page.value,
        sort: sort.value,
        filters: applied.value,
    });

    const set = (state: GridState<S>): void => {
        page.value = state.page;
        sort.value = state.sort;
        applied.value = state.filters;
    };

    const load = async (): Promise<void> => {
        const outcome = await rows.load(current());

        if (outcome.kind === 'failed') {
            set(shown);
            address.write(shown);

            return;
        }

        if (outcome.kind === 'past') {
            page.value = outcome.lastPage;
            address.write(current());
            await load();

            return;
        }

        if (outcome.kind === 'shown') {
            shown = current();
        }
    };

    const show = async (state: GridState<S>): Promise<void> => {
        set(state);
        address.write(state);
        await load();
    };

    const applyFilters = async (filters: GridFilters): Promise<void> => show({
        page: 1,
        sort: sort.value,
        filters,
    });

    const filterFields = useGridFilters(options.filters, applied, {
        apply: applyFilters,
        changing: selection.clear,
    });

    const follow = async (wanted: GridState<S>): Promise<void> => {
        if (sameState(wanted, current())) {
            return;
        }

        set(wanted);
        filterFields.fill(wanted.filters);
        await load();
    };

    const clearOnFilters = (next: GridFilters, previous: GridFilters): void => {
        if (sameFilters(next, previous)) {
            return;
        }

        selection.clear();
    };

    const sortBy = async (column: S): Promise<void> => show({
        page: 1,
        sort: nextSort(sort.value, column),
        filters: applied.value,
    });

    const goTo = async (target: number): Promise<void> => {
        if (target === page.value) {
            return;
        }

        await show({
            page: target,
            sort: sort.value,
            filters: applied.value,
        });
    };

    onMounted(load);
    watch(applied, clearOnFilters);
    address.follow(follow);

    return {
        items: readonly(rows.items),
        total: readonly(rows.total),
        page: readonly(page),
        perPage: options.perPage,
        sort: readonly(sort),
        loading: readonly(rows.loading),
        failed: readonly(rows.failed),
        filters: filterFields.filters,
        appliedFilters: filterFields.appliedFilters,
        filtered: filterFields.filtered,
        filled: filterFields.filled,
        selection,
        reload: load,
        sortBy,
        goTo,
        clearFilters: filterFields.clearFilters,
    };
};

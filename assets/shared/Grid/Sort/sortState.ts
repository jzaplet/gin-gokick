import type { DeepReadonly } from 'vue';
import type { GridColumn } from '@/shared/Grid/types/GridColumn';
import type { GridSort } from '@/shared/Grid/types/GridSort';
import { SortDirection } from '@/shared/Grid/types/SortDirection';

type AriaSort = 'ascending' | 'descending';

const isSortedBy = <S extends string>(column: GridColumn<S>, sort: DeepReadonly<GridSort<S>>): boolean =>
    column.sort !== undefined && column.sort === sort.sortBy;

export const ariaSort = <S extends string>(
    column: GridColumn<S>,
    sort: DeepReadonly<GridSort<S>>,
): AriaSort | undefined => {
    if (isSortedBy(column, sort) === false) {
        return undefined;
    }

    return sort.sortDir === SortDirection.Ascending ? 'ascending' : 'descending';
};

export const sortMark = <S extends string>(column: GridColumn<S>, sort: DeepReadonly<GridSort<S>>): string => {
    if (isSortedBy(column, sort) === false) {
        return 'text-slate-400';
    }

    return sort.sortDir === SortDirection.Ascending ? 'rotate-180 text-ink-900' : 'text-ink-900';
};

export const nextSort = <S extends string>(current: GridSort<S>, column: S): GridSort<S> => ({
    sortBy: column,
    sortDir: current.sortBy === column && current.sortDir === SortDirection.Ascending
        ? SortDirection.Descending
        : SortDirection.Ascending,
});

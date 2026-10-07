import type { GridFilters } from '@/shared/Grid/types/GridFilters';
import type { GridSort } from '@/shared/Grid/types/GridSort';

export type GridState<S extends string> = {
    page: number;
    sort: GridSort<S>;
    filters: GridFilters;
};

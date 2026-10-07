import type { SortDirection } from '@/shared/Grid/types/SortDirection';

export type GridSort<S extends string> = {
    sortBy: S;
    sortDir: SortDirection;
};

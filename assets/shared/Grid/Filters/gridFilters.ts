import type { GridFilters } from '@/shared/Grid/types/GridFilters';

export const trimmed = (filters: GridFilters): GridFilters =>
    Object.fromEntries(
        Object.entries(filters).map(([key, value]) => [
            key,
            value.trim(),
        ]),
    );

export const sameFilters = (one: GridFilters, other: GridFilters): boolean =>
    Object.keys(one).every((key) => one[key] === other[key]);

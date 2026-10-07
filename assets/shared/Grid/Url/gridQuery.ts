import type { LocationQuery } from 'vue-router';

export type GridValues = Partial<Record<string, string>>;

type QueryEntry = [string, LocationQuery[string]];

const gridKeys = new Set<string>();

export const queryKey = (grid: string, key: string): string => grid + key.charAt(0).toUpperCase() + key.slice(1);

export const readQuery = (query: LocationQuery, grid: string, keys: readonly string[]): GridValues => {
    const values: GridValues = {};

    for (const key of keys) {
        const value = query[queryKey(grid, key)];

        if (typeof value === 'string') {
            values[key] = value;
        }
    }

    return values;
};

export const writeQuery = (
    query: LocationQuery,
    grid: string,
    values: GridValues,
    defaults: GridValues,
): LocationQuery => {
    const own = new Set(Object.keys(values).map((key) => queryKey(grid, key)));
    const kept = Object.entries(query).filter(([name]) => own.has(name) === false);
    const changed = Object.entries(values)
        .filter(([key, value]) => value !== undefined && value !== defaults[key])
        .map(([key, value]): QueryEntry => [
            queryKey(grid, key),
            value ?? null,
        ]);

    return Object.fromEntries<LocationQuery[string]>([
        ...kept,
        ...changed,
    ]);
};

export const registerGridKeys = (grid: string, keys: readonly string[]): void => {
    for (const key of keys) {
        gridKeys.add(queryKey(grid, key));
    }
};

export const changesOnlyGrids = (from: LocationQuery, to: LocationQuery): boolean => [
    ...Object.keys(from),
    ...Object.keys(to),
].every((name) => gridKeys.has(name) || JSON.stringify(from[name]) === JSON.stringify(to[name]));

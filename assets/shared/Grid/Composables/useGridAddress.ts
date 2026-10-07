import { watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import type { GridFilters } from '@/shared/Grid/types/GridFilters';
import type { GridState } from '@/shared/Grid/types/GridState';
import { isSortDirection } from '@/shared/Grid/types/SortDirection';
import { type GridValues, readQuery, registerGridKeys, writeQuery } from '@/shared/Grid/Url/gridQuery';
import type { Guard } from '@/shared/TypeGuards/typeGuards';

type AddressOptions<S extends string, F extends GridFilters> = {
    name: string;
    isSort: Guard<S>;
    filters: F;
    isFilter: { [K in keyof F]: (value: string) => boolean };
};

type GridAddress<S extends string> = {
    read: () => GridState<S>;
    write: (state: GridState<S>) => void;
    follow: (change: (state: GridState<S>) => Promise<void>) => void;
};

const sortKeys = [
    'page',
    'sortBy',
    'sortDir',
];

const pageOf = (value: string | undefined): number | undefined =>
    value !== undefined && /^[1-9]\d{0,4}$/u.test(value) ? Number(value) : undefined;

const valuesOf = <S extends string>(state: GridState<S>): GridValues => ({
    page: String(state.page),
    sortBy: state.sort.sortBy,
    sortDir: state.sort.sortDir,
    ...state.filters,
});

export const useGridAddress = <S extends string, F extends GridFilters>(
    options: AddressOptions<S, F>,
    defaults: GridState<S>,
): GridAddress<S> => {
    const route = useRoute();
    const router = useRouter();
    const screen = route.name;
    const checks: Partial<Record<string, (value: string) => boolean>> = options.isFilter;
    const filterKeys = Object.keys(options.filters);
    const urlKeys = [
        ...sortKeys,
        ...filterKeys,
    ];

    registerGridKeys(options.name, urlKeys);

    const filterOf = (key: string, value: string | undefined): string => {
        const wanted = value?.trim();

        return wanted !== undefined && checks[key]?.(wanted) === true ? wanted : defaults.filters[key] ?? '';
    };

    const stateOf = (values: GridValues): GridState<S> => ({
        page: pageOf(values['page']) ?? defaults.page,
        sort: {
            sortBy: options.isSort(values['sortBy']) ? values['sortBy'] : defaults.sort.sortBy,
            sortDir: isSortDirection(values['sortDir']) ? values['sortDir'] : defaults.sort.sortDir,
        },
        filters: Object.fromEntries(
            filterKeys.map((key) => [
                key,
                filterOf(key, values[key]),
            ]),
        ),
    });

    const read = (): GridState<S> => stateOf(readQuery(route.query, options.name, urlKeys));

    const write = (state: GridState<S>): void => {
        void router.replace({
            query: writeQuery(route.query, options.name, valuesOf(state), valuesOf(defaults)),
            hash: route.hash,
        });
    };

    const follow = (change: (state: GridState<S>) => Promise<void>): void => {
        watch(() => route.query, async () => {
            if (route.name !== screen) {
                return;
            }

            await change(read());
        });
    };

    return {
        read,
        write,
        follow,
    };
};

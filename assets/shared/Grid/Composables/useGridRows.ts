import { onScopeDispose, type Ref, ref, shallowRef } from 'vue';
import { authFetch } from '@/shared/Auth/authFetch';
import { type GridResult, isGridResult } from '@/shared/Grid/types/GridResult';
import type { GridState } from '@/shared/Grid/types/GridState';
import { showToast } from '@/shared/Toast/Queue/toastQueue';
import type { Guard } from '@/shared/TypeGuards/typeGuards';

type RowsOptions<T> = {
    url: string;
    item: Guard<T>;
    perPage: number;
};

type LoadOutcome = { kind: 'shown' | 'failed' | 'stale' } | {
    kind: 'past';
    lastPage: number;
};

type GridRows<T> = {
    items: Ref<T[]>;
    total: Ref<number>;
    selectable: Ref<number>;
    loading: Ref<boolean>;
    failed: Ref<boolean>;
    load: (state: GridState<string>) => Promise<LoadOutcome>;
};

export const useGridRows = <T>(options: RowsOptions<T>): GridRows<T> => {
    const items = shallowRef<T[]>([]);
    const total = ref(0);
    const selectable = ref(0);
    const loading = ref(false);
    const failed = ref(false);
    let latest = 0;

    const query = (state: GridState<string>): string => new URLSearchParams({
        page: String(state.page),
        perPage: String(options.perPage),
        sortBy: state.sort.sortBy,
        sortDir: state.sort.sortDir,
        ...Object.fromEntries(Object.entries(state.filters).filter(([, value]) => value !== '')),
    }).toString();

    const load = async (state: GridState<string>): Promise<LoadOutcome> => {
        latest += 1;
        const request = latest;

        loading.value = true;
        const result = await authFetch<GridResult<T>>('GET', `${options.url}?${query(state)}`, {
            validate: isGridResult(options.item),
        });

        if (request !== latest) {
            return { kind: 'stale' };
        }

        loading.value = false;
        failed.value = result.success === false;

        if (result.success === false) {
            showToast('error', 'toast.error_title', result.data.general);

            return { kind: 'failed' };
        }

        const lastPage = Math.max(1, Math.ceil(result.data.total / options.perPage));

        if (state.page > lastPage) {
            return {
                kind: 'past',
                lastPage,
            };
        }

        items.value = result.data.items;
        total.value = result.data.total;
        selectable.value = result.data.selectable ?? result.data.total;

        return { kind: 'shown' };
    };

    onScopeDispose(() => {
        latest += 1;
    });

    return {
        items,
        total,
        selectable,
        loading,
        failed,
        load,
    };
};

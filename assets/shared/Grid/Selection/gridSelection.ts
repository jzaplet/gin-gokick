import { computed, reactive, type Ref, ref, shallowRef } from 'vue';
import type { GridSelection } from '@/shared/Grid/types/GridSelection';

export const createSelection = <T extends { id: string }>(
    items: Readonly<Ref<readonly T[]>>,
    total: Readonly<Ref<number>>,
    selectable: (item: T) => boolean,
): GridSelection<T> => {
    const ids = shallowRef<ReadonlySet<string>>(new Set());
    const all = ref(false);
    const choosable = computed(() => items.value.filter(selectable));

    const selected = (item: T): boolean => selectable(item) && (all.value || ids.value.has(item.id));

    const count = computed(() => (all.value ? total.value : ids.value.size));

    const pageSelectable = computed(() => choosable.value.length > 0);

    const pageSelected = computed(() => pageSelectable.value && choosable.value.every(selected));

    const clear = (): void => {
        ids.value = new Set();
        all.value = false;
    };

    const toggle = (item: T): void => {
        if (all.value) {
            clear();

            return;
        }

        const next = new Set(ids.value);

        if (next.has(item.id)) {
            next.delete(item.id);
        } else if (selectable(item)) {
            next.add(item.id);
        }

        ids.value = next;
    };

    const togglePage = (): void => {
        if (all.value) {
            clear();

            return;
        }

        const next = new Set(ids.value);
        const unselect = pageSelected.value;

        for (const item of choosable.value) {
            if (unselect) {
                next.delete(item.id);
            } else {
                next.add(item.id);
            }
        }

        ids.value = next;
    };

    const selectAll = (): void => {
        ids.value = new Set();
        all.value = true;
    };

    const deselect = (id: string): void => {
        const next = new Set(ids.value);

        next.delete(id);
        ids.value = next;
    };

    const chosen = (): string[] => [...ids.value];

    return reactive({
        count,
        total,
        all,
        pageSelectable,
        pageSelected,
        selectable,
        selected,
        ids: chosen,
        toggle,
        togglePage,
        selectAll,
        deselect,
        clear,
    });
};

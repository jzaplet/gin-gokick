import { readonly, type Ref, ref, watch } from 'vue';
import { appStorage } from '@/shared/Storage/appStorage';

type FilterPanel = {
    open: Readonly<Ref<boolean>>;
    toggle: () => void;
};

export const useFilterPanel = (name: string, active: () => boolean): FilterPanel => {
    const storageKey = `filterPanel.${name}`;
    const open = ref(appStorage().read(storageKey) === true || active());

    watch(active, (filtered) => {
        if (filtered) {
            open.value = true;
        }
    });

    const toggle = (): void => {
        open.value = open.value === false;
        appStorage().write(storageKey, open.value);
    };

    return {
        open: readonly(open),
        toggle,
    };
};

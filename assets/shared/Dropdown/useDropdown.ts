import { onMounted, onUnmounted, readonly, type Ref, ref, useTemplateRef } from 'vue';

type Dropdown = {
    open: Readonly<Ref<boolean>>;
    toggle: () => void;
    close: () => void;
};

export const useDropdown = (): Dropdown => {
    const root = useTemplateRef<HTMLElement>('root');
    const open = ref(false);

    const close = (): void => {
        open.value = false;
    };

    const toggle = (): void => {
        open.value = open.value === false;
    };

    const closeOutside = (event: MouseEvent): void => {
        if (event.target instanceof Node && root.value?.contains(event.target) !== true) {
            close();
        }
    };

    const closeOnEscape = (event: KeyboardEvent): void => {
        if (event.key === 'Escape') {
            close();
        }
    };

    onMounted(() => {
        document.addEventListener('click', closeOutside);
        document.addEventListener('keydown', closeOnEscape);
    });
    onUnmounted(() => {
        document.removeEventListener('click', closeOutside);
        document.removeEventListener('keydown', closeOnEscape);
    });

    return {
        open: readonly(open),
        toggle,
        close,
    };
};

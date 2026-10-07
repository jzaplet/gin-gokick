import { onMounted, type Ref, useTemplateRef, watch } from 'vue';

type ModalDialog = {
    closeOnBackdrop: (event: MouseEvent) => void;
    cancelled: (event: Event) => void;
    closed: () => void;
};

const transitionsOf = async (dialog: HTMLDialogElement): Promise<void> => {
    await Promise.allSettled(dialog.getAnimations({ subtree: true }).map(async (animation) => animation.finished));
};

export const useModalDialog = (open: Ref<boolean>): ModalDialog => {
    const dialog = useTemplateRef<HTMLDialogElement>('dialog');
    let change = 0;

    const show = (element: HTMLDialogElement): void => {
        element.toggleAttribute('data-closing', false);

        if (element.open) {
            return;
        }

        element.showModal();
    };

    const hide = async (element: HTMLDialogElement): Promise<void> => {
        const current = change;

        element.toggleAttribute('data-closing', true);
        await transitionsOf(element);

        if (current !== change) {
            return;
        }

        element.close();
        element.toggleAttribute('data-closing', false);
    };

    const follow = async (shown: boolean): Promise<void> => {
        change += 1;

        const element = dialog.value;

        if (element === null) {
            return;
        }

        if (shown) {
            show(element);

            return;
        }

        if (element.open) {
            await hide(element);
        }
    };

    const closeOnBackdrop = (event: MouseEvent): void => {
        if (event.target === dialog.value) {
            open.value = false;
        }
    };

    const cancelled = (event: Event): void => {
        event.preventDefault();
        open.value = false;
    };

    const closed = (): void => {
        open.value = false;
    };

    onMounted(async () => {
        await follow(open.value);
    });
    watch(open, follow, { flush: 'post' });

    return {
        closeOnBackdrop,
        cancelled,
        closed,
    };
};

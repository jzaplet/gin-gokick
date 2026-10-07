import { onBeforeUnmount, onMounted, readonly, type Ref, ref, useTemplateRef } from 'vue';
import { edgeState } from '@/shared/ScrollShadow/edgeState';

type ScrollShadow = {
    left: Readonly<Ref<boolean>>;
    right: Readonly<Ref<boolean>>;
    update: () => void;
};

export const useScrollShadow = (): ScrollShadow => {
    const viewport = useTemplateRef<HTMLElement>('viewport');
    const left = ref(false);
    const right = ref(false);
    let observer: ResizeObserver | undefined;

    const update = (): void => {
        if (viewport.value === null) {
            return;
        }

        const edges = edgeState(viewport.value.scrollLeft, viewport.value.scrollWidth, viewport.value.clientWidth);

        left.value = edges.left;
        right.value = edges.right;
    };

    onMounted(() => {
        update();

        if (typeof ResizeObserver === 'undefined' || viewport.value === null) {
            return;
        }

        observer = new ResizeObserver(update);
        observer.observe(viewport.value);

        for (const child of viewport.value.children) {
            observer.observe(child);
        }
    });
    onBeforeUnmount(() => {
        observer?.disconnect();
    });

    return {
        left: readonly(left),
        right: readonly(right),
        update,
    };
};

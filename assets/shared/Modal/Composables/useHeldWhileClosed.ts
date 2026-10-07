import { type Ref, shallowRef, watchEffect } from 'vue';

export const useHeldWhileClosed = <T>(open: Readonly<Ref<boolean>>, value: () => T): Readonly<Ref<T>> => {
    const held = shallowRef(value());

    watchEffect(() => {
        if (open.value) {
            held.value = value();
        }
    });

    return held;
};

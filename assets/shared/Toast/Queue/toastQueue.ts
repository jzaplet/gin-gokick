import { shallowRef } from 'vue';
import type { PlainMessageKey } from '@/shared/I18n/Texts/translate';
import type { Toast } from '@/shared/Toast/types/Toast';
import type { ToastKind } from '@/shared/Toast/types/ToastKind';
import type { ToastMessage } from '@/shared/Toast/types/ToastMessage';

const lifetime = 5000;

const toasts = shallowRef<readonly Toast[]>([]);

const timers = new Map<number, ReturnType<typeof setTimeout>>();

const holds = new Map<number, number>();

let lastId = 0;

export const shownToasts = (): readonly Toast[] => toasts.value;

const isGone = (id: number): boolean => toasts.value.every((toast) => toast.id !== id);

const stopTimer = (id: number): void => {
    clearTimeout(timers.get(id));
    timers.delete(id);
};

export const closeToast = (id: number): void => {
    stopTimer(id);
    holds.delete(id);
    toasts.value = toasts.value.filter((toast) => toast.id !== id);
};

const closeLater = (id: number): void => {
    const timer = setTimeout(() => {
        closeToast(id);
    }, lifetime);

    timers.set(id, timer);
};

export const showToast = (kind: ToastKind, title: PlainMessageKey, message: ToastMessage): void => {
    lastId += 1;

    const id = lastId;

    toasts.value = [...toasts.value, {
        id,
        kind,
        title,
        message,
    }];
    closeLater(id);
};

export const holdToast = (id: number): void => {
    if (isGone(id)) {
        return;
    }

    holds.set(id, (holds.get(id) ?? 0) + 1);
    stopTimer(id);
};

export const releaseToast = (id: number): void => {
    if (isGone(id)) {
        return;
    }

    const left = Math.max(0, (holds.get(id) ?? 0) - 1);

    holds.set(id, left);

    if (left > 0) {
        return;
    }

    stopTimer(id);
    closeLater(id);
};

export const clearToasts = (): void => {
    for (const toast of toasts.value) {
        closeToast(toast.id);
    }
};

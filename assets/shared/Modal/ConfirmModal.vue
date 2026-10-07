<script setup lang="ts">
import { useId } from 'vue';
import BaseButton from '@/shared/Buttons/BaseButton.vue';
import { t } from '@/shared/I18n/Texts/translate';
import IconTriangleAlert from '@/shared/Icons/IconTriangleAlert.vue';
import { useHeldWhileClosed } from '@/shared/Modal/Composables/useHeldWhileClosed';
import { useModalDialog } from '@/shared/Modal/Composables/useModalDialog';

const { title, message, confirmText = null, cancelText = null, busy } = defineProps<{
    title: string;
    message: string;
    confirmText?: string | null;
    cancelText?: string | null;
    busy?: boolean;
}>();

defineEmits<{ confirm: [] }>();

const open = defineModel<boolean>('open', { required: true });

const titleId = useId();

const { closeOnBackdrop, cancelled, closed } = useModalDialog(open);

const shown = useHeldWhileClosed(open, () => ({
    title,
    message,
    confirmText: confirmText ?? t('modal.confirm'),
    cancelText: cancelText ?? t('modal.cancel'),
}));
</script>

<template>
    <dialog
        ref="dialog"
        :aria-labelledby="titleId"
        class="m-auto w-[calc(100%-2rem)] max-w-md scale-95 rounded-2xl bg-white p-0 text-ink-900 opacity-0 shadow-xl
            transition-[opacity,scale] duration-200 ease-in
            open:scale-100 open:opacity-100 open:ease-out starting:open:scale-95 starting:open:opacity-0
            data-closing:scale-95 data-closing:opacity-0 data-closing:ease-in
            backdrop:bg-ink-950/0 backdrop:transition-colors backdrop:duration-200 backdrop:ease-in
            open:backdrop:bg-ink-950/50 open:backdrop:ease-out starting:open:backdrop:bg-ink-950/0
            data-closing:backdrop:bg-ink-950/0 data-closing:backdrop:ease-in
            motion-reduce:transition-none motion-reduce:backdrop:transition-none"
        @click="closeOnBackdrop"
        @cancel="cancelled"
        @close="closed"
    >
        <div class="flex gap-4 p-6">
            <span class="flex size-10 shrink-0 items-center justify-center rounded-full bg-red-50 text-red-600">
                <IconTriangleAlert class="size-5" />
            </span>
            <div class="min-w-0 space-y-2">
                <h2
                    :id="titleId"
                    class="text-lg font-semibold tracking-tight"
                >
                    {{ shown.title }}
                </h2>
                <p class="text-sm text-slate-600">
                    {{ shown.message }}
                </p>
            </div>
        </div>
        <footer class="flex flex-col-reverse gap-2 rounded-b-2xl bg-slate-50 px-6 py-4 sm:flex-row sm:justify-end">
            <BaseButton
                variant="secondary"
                size="sm"
                :disabled="busy"
                @click="open = false"
            >
                {{ shown.cancelText }}
            </BaseButton>
            <BaseButton
                variant="danger"
                size="sm"
                :loading="busy"
                @click="$emit('confirm')"
            >
                {{ shown.confirmText }}
            </BaseButton>
        </footer>
    </dialog>
</template>

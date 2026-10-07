<script setup lang="ts">
import type { Component } from 'vue';
import { t, tm } from '@/shared/I18n/Texts/translate';
import IconCircleCheck from '@/shared/Icons/IconCircleCheck.vue';
import IconCircleX from '@/shared/Icons/IconCircleX.vue';
import IconClose from '@/shared/Icons/IconClose.vue';
import IconInfo from '@/shared/Icons/IconInfo.vue';
import IconTriangleAlert from '@/shared/Icons/IconTriangleAlert.vue';
import type { Toast } from '@/shared/Toast/types/Toast';
import type { ToastKind } from '@/shared/Toast/types/ToastKind';

type Tone = Record<'box' | 'icon' | 'title' | 'text' | 'close', string>;

defineProps<{ toast: Toast }>();

defineEmits<{
    close: [];
    hold: [];
    release: [];
}>();

const icons = {
    success: IconCircleCheck,
    error: IconCircleX,
    warning: IconTriangleAlert,
    info: IconInfo,
} satisfies Record<ToastKind, Component>;

const tones = {
    success: {
        box: 'border-emerald-300 bg-emerald-50',
        icon: 'text-emerald-600',
        title: 'text-emerald-900',
        text: 'text-emerald-800',
        close: 'text-emerald-700 hover:bg-emerald-100',
    },
    error: {
        box: 'border-red-300 bg-red-50',
        icon: 'text-red-600',
        title: 'text-red-900',
        text: 'text-red-800',
        close: 'text-red-700 hover:bg-red-100',
    },
    warning: {
        box: 'border-amber-300 bg-amber-50',
        icon: 'text-amber-600',
        title: 'text-amber-900',
        text: 'text-amber-800',
        close: 'text-amber-700 hover:bg-amber-100',
    },
    info: {
        box: 'border-brand-300 bg-brand-50',
        icon: 'text-brand-600',
        title: 'text-brand-800',
        text: 'text-brand-700',
        close: 'text-brand-700 hover:bg-brand-100',
    },
} satisfies Record<ToastKind, Tone>;
</script>

<template>
    <li
        class="pointer-events-auto flex w-80 max-w-full items-start gap-3 rounded-xl border p-4 shadow-lg"
        :class="tones[toast.kind].box"
        @mouseenter="$emit('hold')"
        @mouseleave="$emit('release')"
        @focusin="$emit('hold')"
        @focusout="$emit('release')"
    >
        <component
            :is="icons[toast.kind]"
            class="size-5 shrink-0"
            :class="tones[toast.kind].icon"
        />
        <div class="min-w-0 flex-1">
            <p
                class="text-sm font-semibold"
                :class="tones[toast.kind].title"
            >
                {{ t(toast.title) }}
            </p>
            <p
                class="mt-1 text-sm"
                :class="tones[toast.kind].text"
            >
                {{ typeof toast.message === 'string' ? t(toast.message) : tm(toast.message) }}
            </p>
        </div>
        <button
            type="button"
            :aria-label="t('toast.close')"
            class="-mt-1 -mr-1 flex size-7 shrink-0 cursor-pointer items-center justify-center rounded-md
                transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-ink-900"
            :class="tones[toast.kind].close"
            @click="$emit('close')"
        >
            <IconClose class="size-4" />
        </button>
    </li>
</template>

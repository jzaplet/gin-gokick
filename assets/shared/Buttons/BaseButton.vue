<script setup lang="ts">
import BaseSpinner from '@/shared/Loading/BaseSpinner.vue';

type Variant = 'primary' | 'secondary' | 'danger' | 'ghost';

type Size = 'lg' | 'md' | 'sm' | 'xs';

const { type = 'button', variant = 'primary', size = 'lg', disabled, loading } = defineProps<{
    type?: 'button' | 'submit';
    variant?: Variant;
    disabled?: boolean;
    size?: Size;
    loading?: boolean;
}>();

const variants = {
    primary: 'bg-brand-600 text-white hover:bg-brand-700',
    secondary: 'bg-slate-200/60 text-slate-700 hover:bg-slate-200 hover:text-ink-900',
    danger: 'bg-red-600 text-white hover:bg-red-700',
    ghost: 'bg-transparent text-slate-600 hover:bg-slate-100 hover:text-ink-900',
} satisfies Record<Variant, string>;

const sizes = {
    lg: 'px-6 py-3 text-sm',
    md: 'px-4 py-3 text-sm',
    sm: 'px-4 py-2 text-sm',
    xs: 'px-1.5 py-1 text-xs',
} satisfies Record<Size, string>;
</script>

<template>
    <button
        :type="type"
        :disabled="disabled || loading"
        :aria-busy="loading"
        class="inline-flex cursor-pointer items-center justify-center gap-2 rounded-full border border-transparent
            font-semibold whitespace-nowrap
            focus:outline-none focus:ring-2 focus:ring-ink-900 focus:ring-offset-2
            disabled:cursor-not-allowed disabled:opacity-50"
        :class="[sizes[size], variants[variant]]"
    >
        <BaseSpinner
            v-if="loading"
            class="size-4"
        />
        <slot />
    </button>
</template>

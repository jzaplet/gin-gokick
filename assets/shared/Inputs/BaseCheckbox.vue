<script setup lang="ts">
import IconCheck from '@/shared/Icons/IconCheck.vue';

const { label, name = null, disabled, hideLabel } = defineProps<{
    label: string;
    name?: string | null;
    disabled?: boolean;
    hideLabel?: boolean;
}>();

const model = defineModel<boolean>({ required: true });
</script>

<template>
    <label
        class="inline-flex cursor-pointer items-center gap-2 text-sm text-ink-900
            has-disabled:cursor-not-allowed has-disabled:opacity-40"
    >
        <span class="relative inline-flex shrink-0">
            <input
                v-model="model"
                type="checkbox"
                :name="name ?? undefined"
                :disabled="disabled"
                class="peer size-4 cursor-pointer appearance-none rounded border border-slate-300 bg-white shadow-sm
                    checked:border-ink-900 checked:bg-ink-900 focus:outline-none
                    focus-visible:ring-2 focus-visible:ring-ink-900 focus-visible:ring-offset-2
                    disabled:cursor-not-allowed"
            >
            <IconCheck
                class="pointer-events-none absolute inset-0 size-4 p-0.5 text-white opacity-0 peer-checked:opacity-100"
            />
        </span>
        <span :class="hideLabel ? 'sr-only' : undefined">{{ label }}</span>
    </label>
</template>

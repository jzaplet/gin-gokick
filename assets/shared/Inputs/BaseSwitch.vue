<script setup lang="ts">
import IconCheck from '@/shared/Icons/IconCheck.vue';
import IconClose from '@/shared/Icons/IconClose.vue';

const { label, name, disabled } = defineProps<{
    label: string;
    name: string;
    disabled?: boolean;
}>();

const model = defineModel<boolean>({ required: true });
</script>

<template>
    <label class="group/switch inline-flex shrink-0 cursor-pointer has-disabled:cursor-not-allowed">
        <input
            v-model="model"
            type="checkbox"
            role="switch"
            :name="name"
            :disabled="disabled"
            class="peer sr-only"
        >
        <span class="sr-only">{{ label }}</span>
        <span
            class="flex h-6 w-11 items-center rounded-full bg-slate-500 p-0.5 transition-colors
                group-has-checked/switch:bg-ink-900 group-has-disabled/switch:opacity-40
                peer-focus-visible:ring-2 peer-focus-visible:ring-ink-900 peer-focus-visible:ring-offset-2"
        >
            <span
                class="flex size-5 items-center justify-center rounded-full bg-white text-slate-500 shadow-sm
                    transition-transform group-has-checked/switch:translate-x-5 group-has-checked/switch:text-ink-900
                    motion-reduce:transition-none"
            >
                <IconCheck
                    v-if="model"
                    class="size-3"
                />
                <IconClose
                    v-else
                    class="size-3"
                />
            </span>
        </span>
    </label>
</template>

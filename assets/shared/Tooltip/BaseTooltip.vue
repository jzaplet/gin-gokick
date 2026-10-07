<script setup lang="ts">
type Align = 'center' | 'left' | 'right';

const { text, position = 'top', align = 'center', disabled } = defineProps<{
    text: string;
    position?: 'top' | 'bottom';
    align?: Align;
    disabled?: boolean;
}>();

const bubbleAlign: Record<Align, string> = {
    center: 'left-1/2 -translate-x-1/2',
    left: 'left-0',
    right: 'right-0',
};

const arrowAlign: Record<Align, string> = {
    center: 'left-1/2 -translate-x-1/2',
    left: 'left-3',
    right: 'right-3',
};
</script>

<template>
    <span class="group/tooltip relative inline-flex">
        <slot />
        <span
            v-if="disabled !== true"
            role="tooltip"
            :class="[
                'pointer-events-none absolute z-30 w-max max-w-50 rounded-md bg-ink-900 px-2.5 py-1.5 shadow-lg',
                'text-center text-xs font-medium text-white opacity-0 transition-opacity duration-150',
                'group-hover/tooltip:opacity-100 group-has-focus-visible/tooltip:opacity-100',
                bubbleAlign[align],
                position === 'top' ? 'bottom-full mb-2' : 'top-full mt-2',
            ]"
        >
            {{ text }}
            <span
                :class="[
                    'absolute size-2 rotate-45 bg-ink-900',
                    arrowAlign[align],
                    position === 'top' ? 'top-full -mt-1' : 'bottom-full -mb-1',
                ]"
            />
        </span>
    </span>
</template>

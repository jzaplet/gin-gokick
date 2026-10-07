<script setup lang="ts">
import { useScrollShadow } from '@/shared/ScrollShadow/Composables/useScrollShadow';

const { left, right, update } = useScrollShadow();
</script>

<template>
    <div class="relative">
        <div
            v-if="left"
            class="pointer-events-none absolute inset-y-0 left-0 z-10 w-6
                bg-linear-to-r from-slate-900/10 to-transparent"
            aria-hidden="true"
        />
        <div
            ref="viewport"
            class="overflow-x-auto"
            @scroll.passive="update"
        >
            <slot />
        </div>
        <div
            v-if="right"
            class="pointer-events-none absolute inset-y-0 right-0 z-10 w-6
                bg-linear-to-l from-slate-900/10 to-transparent"
            aria-hidden="true"
        />
    </div>
</template>

<script setup lang="ts">
import { useDropdown } from '@/shared/Dropdown/useDropdown';

const { align = 'right' } = defineProps<{ align?: 'left' | 'right' }>();

const { open, toggle, close } = useDropdown();
</script>

<template>
    <div
        ref="root"
        class="relative"
    >
        <slot
            name="trigger"
            :open="open"
            :toggle="toggle"
        />
        <Transition
            enter-active-class="transition duration-100 ease-out"
            enter-from-class="scale-95 opacity-0"
            leave-active-class="transition duration-75 ease-in"
            leave-to-class="scale-95 opacity-0"
        >
            <div
                v-if="open"
                class="absolute z-50 mt-2 w-max min-w-36 rounded-lg border border-slate-200 bg-white py-1 shadow-lg"
                :class="align === 'left' ? 'left-0' : 'right-0'"
                @click="close"
            >
                <slot />
            </div>
        </Transition>
    </div>
</template>

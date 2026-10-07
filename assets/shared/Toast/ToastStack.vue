<script setup lang="ts">
import { onMounted } from 'vue';
import ToastItem from '@/shared/Toast/Components/ToastItem.vue';
import { preloadToastIcons } from '@/shared/Toast/Icons/preloadToastIcons';
import { closeToast, holdToast, releaseToast, shownToasts } from '@/shared/Toast/Queue/toastQueue';

onMounted(preloadToastIcons);
</script>

<template>
    <TransitionGroup
        tag="ul"
        aria-live="polite"
        class="pointer-events-none fixed inset-x-4 bottom-4 z-50 flex flex-col-reverse items-end gap-3"
        enter-active-class="transition duration-300 ease-out motion-reduce:transition-none"
        enter-from-class="translate-y-2 opacity-0"
        leave-active-class="transition duration-200 ease-in motion-reduce:transition-none"
        leave-to-class="opacity-0"
        move-class="transition duration-300 ease-out motion-reduce:transition-none"
    >
        <ToastItem
            v-for="toast in shownToasts()"
            :key="toast.id"
            :toast="toast"
            @close="closeToast(toast.id)"
            @hold="holdToast(toast.id)"
            @release="releaseToast(toast.id)"
        />
    </TransitionGroup>
</template>

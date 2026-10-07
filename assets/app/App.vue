<script setup lang="ts">
import { useLayout } from '@/shared/Layout/Composables/useLayout';
import LoadingScreen from '@/shared/Loading/LoadingScreen.vue';
import ToastStack from '@/shared/Toast/ToastStack.vue';

const layout = useLayout();
</script>

<template>
    <div class="min-h-screen bg-slate-50 text-ink-900">
        <component
            :is="layout"
            v-if="layout !== undefined"
        >
            <RouterView />
        </component>
        <LoadingScreen v-else />
        <ToastStack />
    </div>
</template>

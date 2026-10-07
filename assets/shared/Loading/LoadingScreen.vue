<script setup lang="ts">
import { t } from '@/shared/I18n/Texts/translate';
import BaseSpinner from '@/shared/Loading/BaseSpinner.vue';
</script>

<template>
    <div
        role="status"
        class="grid min-h-screen place-items-center"
    >
        <div class="animate-appear">
            <BaseSpinner class="size-8 text-ink-900" />
            <span class="sr-only">{{ t('app.loading') }}</span>
        </div>
    </div>
</template>

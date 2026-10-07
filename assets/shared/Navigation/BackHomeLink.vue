<script setup lang="ts">
import { t } from '@/shared/I18n/Texts/translate';
import IconChevronLeft from '@/shared/Icons/IconChevronLeft.vue';
import { useHome } from '@/shared/Navigation/Composables/useHome';

const home = useHome();
</script>

<template>
    <a
        :href="home"
        class="inline-flex items-center gap-1 rounded-full px-4 py-1.5 text-sm font-medium whitespace-nowrap
            text-slate-600 transition-colors hover:bg-slate-100 hover:text-ink-900"
    >
        <IconChevronLeft class="size-4" />
        {{ t('navigation.back_home') }}
    </a>
</template>

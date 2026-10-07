<!-- The Go routes show the same page from views/errors/not_found.html. Change both together. -->
<script setup lang="ts">
import { t } from '@/shared/I18n/Texts/translate';
import { useNoIndex } from '@/shared/Seo/useNoIndex';

useNoIndex();
</script>

<template>
    <div class="text-center">
        <p class="text-9xl leading-none font-extrabold tracking-tighter text-ink-900 sm:text-[12rem] lg:text-[15rem]">
            {{ t('not_found.status') }}
        </p>
        <h1 class="mt-4 text-2xl font-semibold tracking-tight text-slate-600 sm:text-3xl">
            {{ t('not_found.title') }}
        </h1>
    </div>
</template>

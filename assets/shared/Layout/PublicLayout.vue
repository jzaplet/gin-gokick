<script setup lang="ts">
import BrandLogo from '@/shared/Brand/BrandLogo.vue';
import LanguageSwitcher from '@/shared/I18n/Components/LanguageSwitcher.vue';
import AppFooter from '@/shared/Layout/Components/AppFooter.vue';
import BackHomeLink from '@/shared/Navigation/BackHomeLink.vue';
import { useHome } from '@/shared/Navigation/Composables/useHome';

const home = useHome();
</script>

<template>
    <div class="isolate flex min-h-screen flex-col">
        <div data-constellation />
        <header class="flex items-center justify-between px-4 py-5 sm:px-6 lg:px-10">
            <a
                :href="home"
                class="inline-flex"
            >
                <BrandLogo />
            </a>
            <LanguageSwitcher />
        </header>
        <main class="grid flex-1 place-items-center px-4 py-6 sm:py-10">
            <div class="flex w-full flex-col items-center gap-6">
                <slot />
                <BackHomeLink />
            </div>
        </main>
        <AppFooter />
    </div>
</template>

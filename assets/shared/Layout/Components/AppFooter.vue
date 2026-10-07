<!-- The Go pages draw the same footer from views/shared/footer.html. Change both together. -->
<script setup lang="ts">
import { type PlainMessageKey, t } from '@/shared/I18n/Texts/translate';
import IconGitHub from '@/shared/Icons/IconGitHub.vue';
import { needsConsent } from '@/shared/Tracking/enabledTools';

const year = String(new Date().getFullYear());
const consent = needsConsent();
const repositories: {
    url: string;
    name: PlainMessageKey;
}[] = [
    {
        url: 'https://github.com/jzaplet/gin-gokick',
        name: 'brand.name',
    },
    {
        url: 'https://github.com/gin-gonic/gin',
        name: 'footer.gin',
    },
];
</script>

<template>
    <footer class="px-4 py-6 text-xs text-slate-500">
        <div class="mx-auto flex max-w-2xl flex-col items-center gap-2 sm:flex-row sm:justify-between">
            <div class="flex items-center gap-4">
                <p>{{ t('footer.copyright', { year }) }}</p>
                <button
                    v-if="consent"
                    type="button"
                    class="cursor-pointer underline underline-offset-4 hover:text-ink-900"
                    data-consent-settings
                >
                    {{ t('consent.settings') }}
                </button>
            </div>
            <nav class="flex items-center gap-4">
                <a
                    v-for="repository in repositories"
                    :key="repository.url"
                    :href="repository.url"
                    target="_blank"
                    rel="noopener"
                    class="inline-flex items-center gap-1.5 underline-offset-4 hover:text-ink-900 hover:underline"
                >
                    <IconGitHub class="size-3.5" />
                    {{ t(repository.name) }}
                </a>
            </nav>
        </div>
    </footer>
</template>

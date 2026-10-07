<script setup lang="ts">
import ConsentBanner from '@/shared/Consent/Components/ConsentBanner.vue';
import ConsentSettings from '@/shared/Consent/Components/ConsentSettings.vue';
import { useCookieConsent } from '@/shared/Consent/Composables/useCookieConsent';
import type { EnabledTools } from '@/shared/Tracking/types/EnabledTools';

const { tools } = defineProps<{ tools: EnabledTools }>();

const {
    bannerShown,
    selected,
    rows,
    acceptAll,
    acceptNecessary,
    save,
    openSettings,
    closeSettings,
    closeOnBackdrop,
    cancelled,
    closed,
} = useCookieConsent(tools);
</script>

<template>
    <Transition
        appear
        enter-from-class="translate-y-4 opacity-0"
        enter-active-class="transition duration-300 ease-out motion-reduce:transition-none"
        leave-active-class="transition duration-200 ease-in motion-reduce:transition-none"
        leave-to-class="translate-y-4 opacity-0"
    >
        <ConsentBanner
            v-if="bannerShown"
            @details="openSettings"
            @necessary="acceptNecessary"
            @accept="acceptAll"
        />
    </Transition>
    <dialog
        ref="dialog"
        aria-labelledby="consent-settings-title"
        class="fixed inset-y-0 right-0 left-auto m-0 h-full max-h-none w-full max-w-lg translate-x-full flex-col
            bg-white p-0 text-ink-900 shadow-xl transition-[translate] duration-300 ease-in
            open:flex open:translate-x-0 open:ease-out starting:open:translate-x-full
            data-closing:translate-x-full data-closing:ease-in
            backdrop:bg-ink-950/0 backdrop:transition-colors backdrop:duration-300 backdrop:ease-in
            open:backdrop:bg-ink-950/50 open:backdrop:ease-out starting:open:backdrop:bg-ink-950/0
            data-closing:backdrop:bg-ink-950/0 data-closing:backdrop:ease-in
            motion-reduce:transition-none motion-reduce:backdrop:transition-none"
        @click="closeOnBackdrop"
        @cancel="cancelled"
        @close="closed"
    >
        <ConsentSettings
            v-model:selected="selected"
            :rows="rows"
            @close="closeSettings"
            @save="save"
            @necessary="acceptNecessary"
            @accept="acceptAll"
        />
    </dialog>
</template>

<script setup lang="ts">
import BaseButton from '@/shared/Buttons/BaseButton.vue';
import { t } from '@/shared/I18n/Texts/translate';

const emit = defineEmits<{
    details: [];
    necessary: [];
    accept: [];
}>();
</script>

<template>
    <section
        aria-labelledby="consent-title"
        class="fixed inset-x-4 bottom-4 z-40 mx-auto max-w-3xl space-y-4 rounded-xl border border-slate-200 bg-white p-5
            text-ink-900 shadow-lg"
    >
        <div class="space-y-1">
            <h2
                id="consent-title"
                class="text-base font-semibold"
            >
                {{ t('consent.title') }}
            </h2>
            <p class="text-sm text-slate-600">
                {{ t('consent.lead', { button: t('consent.understood') }) }}
            </p>
        </div>
        <div class="flex flex-col gap-2 sm:flex-row sm:justify-between">
            <BaseButton
                variant="secondary"
                @click="emit('details')"
            >
                {{ t('consent.details') }}
            </BaseButton>
            <div class="flex flex-col gap-2 sm:flex-row">
                <BaseButton
                    variant="secondary"
                    @click="emit('necessary')"
                >
                    {{ t('consent.necessary_only') }}
                </BaseButton>
                <BaseButton @click="emit('accept')">
                    {{ t('consent.understood') }}
                </BaseButton>
            </div>
        </div>
    </section>
</template>

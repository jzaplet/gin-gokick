<script setup lang="ts">
import SessionsSection from '@/app/Auth/Components/SessionsSection.vue';
import { t } from '@/shared/I18n/Texts/translate';
</script>

<template>
    <div class="space-y-12">
        <div class="space-y-8">
            <div class="space-y-2">
                <h1 class="text-3xl font-semibold tracking-tight text-ink-900">
                    {{ t('dashboard.title') }}
                </h1>
                <p class="text-slate-600">
                    {{ t('dashboard.lead') }}
                </p>
            </div>
            <p class="rounded-lg border-2 border-dashed border-slate-200 px-6 py-16 text-center text-sm text-slate-500">
                {{ t('dashboard.empty') }}
            </p>
        </div>
        <SessionsSection />
    </div>
</template>

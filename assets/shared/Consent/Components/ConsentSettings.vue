<script setup lang="ts">
import BaseButton from '@/shared/Buttons/BaseButton.vue';
import ConsentCategoryRow from '@/shared/Consent/Components/ConsentCategoryRow.vue';
import type { CategoryRow } from '@/shared/Consent/types/CategoryRow';
import type { ConsentCategory } from '@/shared/Consent/types/ConsentCategory';
import { t } from '@/shared/I18n/Texts/translate';
import IconClose from '@/shared/Icons/IconClose.vue';

const { rows } = defineProps<{ rows: CategoryRow[] }>();

const emit = defineEmits<{
    close: [];
    save: [];
    necessary: [];
    accept: [];
}>();

const selected = defineModel<Record<ConsentCategory, boolean>>('selected', { required: true });
</script>

<template>
    <div class="flex min-h-0 flex-1 flex-col">
        <header class="flex items-center justify-between gap-4 border-b border-slate-200 px-6 py-4">
            <h2
                id="consent-settings-title"
                class="text-lg font-semibold"
            >
                {{ t('consent.settings') }}
            </h2>
            <button
                type="button"
                class="cursor-pointer rounded-lg bg-slate-100 p-2 text-slate-700 hover:bg-slate-200 hover:text-ink-900
                    focus:outline-none focus-visible:ring-2 focus-visible:ring-ink-900"
                :aria-label="t('consent.close')"
                @click="emit('close')"
            >
                <IconClose
                    class="size-4"
                    stroke-width="1.75"
                />
            </button>
        </header>
        <div class="flex-1 space-y-5 overflow-y-auto px-6 py-5">
            <p class="text-sm text-slate-600">
                {{ t('consent.settings_lead') }}
            </p>
            <ul class="space-y-2">
                <ConsentCategoryRow
                    model-value
                    name="necessary"
                    :title="t('consent.necessary')"
                    :description="t('consent.necessary_text')"
                    locked
                />
                <ConsentCategoryRow
                    v-for="row in rows"
                    :key="row.category"
                    v-model="selected[row.category]"
                    :name="row.category"
                    :title="t(row.title)"
                    :description="t(row.text, { tools: row.tools })"
                />
            </ul>
        </div>
        <footer class="flex flex-col gap-2 border-t border-slate-200 px-6 py-4 sm:flex-row sm:justify-between">
            <BaseButton
                variant="secondary"
                size="md"
                @click="emit('save')"
            >
                {{ t('consent.save') }}
            </BaseButton>
            <div class="flex flex-col gap-2 sm:flex-row">
                <BaseButton
                    variant="secondary"
                    size="md"
                    @click="emit('necessary')"
                >
                    {{ t('consent.necessary_only') }}
                </BaseButton>
                <BaseButton
                    size="md"
                    @click="emit('accept')"
                >
                    {{ t('consent.accept_all') }}
                </BaseButton>
            </div>
        </footer>
    </div>
</template>

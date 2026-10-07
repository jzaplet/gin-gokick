<script setup lang="ts" generic="K extends string">
import { useBulkActionBar } from '@/shared/BulkActions/Composables/useBulkActionBar';
import type { BulkAction } from '@/shared/BulkActions/types/BulkAction';
import BaseDropdown from '@/shared/Dropdown/BaseDropdown.vue';
import { t } from '@/shared/I18n/Texts/translate';
import IconChevronDown from '@/shared/Icons/IconChevronDown.vue';

const props = defineProps<{
    count: number;
    total: number;
    all: boolean;
    actions: readonly BulkAction<K>[];
}>();

defineEmits<{
    action: [key: K];
    selectAll: [];
    clear: [];
}>();

const { count: shownCount, selected, selectAll, labelOf } = useBulkActionBar<K>(props);
</script>

<template>
    <p
        class="sr-only"
        aria-live="polite"
    >
        {{ count > 0 ? selected : '' }}
    </p>
    <div
        v-if="count > 0"
        class="flex flex-wrap items-center gap-x-3 gap-y-2 rounded-lg border border-brand-200 bg-brand-50 px-4 py-2"
    >
        <BaseDropdown align="left">
            <template #trigger="{ open, toggle }">
                <button
                    type="button"
                    :aria-label="t('bulk.actions')"
                    :aria-expanded="open"
                    class="inline-flex cursor-pointer items-center gap-1.5 rounded-full bg-brand-600 px-3 py-1.5 text-sm
                        font-medium text-white tabular-nums transition-colors hover:bg-brand-700 focus:outline-none
                        focus-visible:ring-2 focus-visible:ring-ink-900 focus-visible:ring-offset-2"
                    @click="toggle"
                >
                    {{ shownCount }}
                    <IconChevronDown
                        class="size-3.5 transition-transform"
                        :class="open ? 'rotate-180' : undefined"
                    />
                </button>
            </template>
            <button
                v-for="action in actions"
                :key="action.key"
                type="button"
                class="block w-full cursor-pointer px-4 py-2 text-left text-sm whitespace-nowrap text-slate-700
                    transition-colors hover:bg-slate-50 hover:text-ink-900 focus:outline-none
                    focus-visible:bg-slate-50"
                @click="$emit('action', action.key)"
            >
                {{ labelOf(action) }}
            </button>
            <div class="my-1 border-t border-slate-200" />
            <button
                type="button"
                class="block w-full cursor-pointer px-4 py-2 text-left text-sm whitespace-nowrap text-slate-500
                    transition-colors hover:bg-slate-50 hover:text-ink-900 focus:outline-none
                    focus-visible:bg-slate-50"
                @click="$emit('clear')"
            >
                {{ t('bulk.clear') }}
            </button>
        </BaseDropdown>
        <span
            class="text-sm text-slate-700 tabular-nums"
            aria-hidden="true"
        >
            {{ selected }}
        </span>
        <button
            v-if="all === false && count < total"
            type="button"
            class="cursor-pointer text-sm text-brand-700 underline underline-offset-2 hover:text-brand-800"
            @click="$emit('selectAll')"
        >
            {{ selectAll }}
        </button>
        <button
            type="button"
            class="cursor-pointer text-sm text-slate-500 underline underline-offset-2 hover:text-ink-900"
            @click="$emit('clear')"
        >
            {{ t('bulk.clear') }}
        </button>
    </div>
</template>

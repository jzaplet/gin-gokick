<script setup lang="ts">
import { useGridPagination } from '@/shared/Grid/Composables/useGridPagination';
import { t } from '@/shared/I18n/Texts/translate';
import IconChevronLeft from '@/shared/Icons/IconChevronLeft.vue';

const props = defineProps<{
    page: number;
    perPage: number;
    total: number;
}>();

defineEmits<{ goTo: [page: number] }>();

const { range, pages, numbers } = useGridPagination(props);
</script>

<template>
    <div
        v-if="total > 0"
        class="flex flex-wrap items-center justify-between gap-3 border-t border-slate-200 px-4 py-3"
    >
        <p class="text-sm text-slate-500">
            {{ range }}
        </p>
        <nav
            v-if="pages > 1"
            :aria-label="t('grid.pages')"
            class="flex gap-1"
        >
            <button
                type="button"
                :disabled="page <= 1"
                :aria-label="t('grid.previous')"
                class="flex size-8 cursor-pointer items-center justify-center rounded-md border border-slate-200
                    text-ink-900 hover:bg-slate-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-ink-900
                    disabled:cursor-not-allowed disabled:opacity-40"
                @click="$emit('goTo', page - 1)"
            >
                <IconChevronLeft class="size-4" />
            </button>
            <button
                v-for="number in numbers"
                :key="number"
                type="button"
                :aria-label="t('grid.page', { page: String(number) })"
                :aria-current="number === page ? 'page' : undefined"
                class="flex h-8 min-w-8 cursor-pointer items-center justify-center rounded-md border px-2 text-sm
                    tabular-nums focus:outline-none focus-visible:ring-2 focus-visible:ring-ink-900"
                :class="number === page
                    ? 'border-brand-600 bg-brand-600 text-white'
                    : 'border-slate-200 text-ink-900 hover:bg-slate-50'"
                @click="$emit('goTo', number)"
            >
                {{ number }}
            </button>
            <button
                type="button"
                :disabled="page >= pages"
                :aria-label="t('grid.next')"
                class="flex size-8 cursor-pointer items-center justify-center rounded-md border border-slate-200
                    text-ink-900 hover:bg-slate-50 focus:outline-none focus-visible:ring-2 focus-visible:ring-ink-900
                    disabled:cursor-not-allowed disabled:opacity-40"
                @click="$emit('goTo', page + 1)"
            >
                <IconChevronLeft class="size-4 rotate-180" />
            </button>
        </nav>
    </div>
</template>

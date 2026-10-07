<script setup lang="ts" generic="S extends string, T extends { id: string }">
import type { DeepReadonly } from 'vue';
import type { GridColumn } from '@/shared/Grid/types/GridColumn';
import type { PageSelection } from '@/shared/Grid/types/GridSelection';
import type { GridSort } from '@/shared/Grid/types/GridSort';
import { ariaSort, sortMark } from '@/shared/Grid/Sort/sortState';
import { type PlainMessageKey, t } from '@/shared/I18n/Texts/translate';
import IconChevronDown from '@/shared/Icons/IconChevronDown.vue';
import BaseButton from '@/shared/Buttons/BaseButton.vue';
import BaseCheckbox from '@/shared/Inputs/BaseCheckbox.vue';
import BaseSpinner from '@/shared/Loading/BaseSpinner.vue';
import ScrollShadow from '@/shared/ScrollShadow/ScrollShadow.vue';

const { columns, items, sort, loading, failed, filtered, emptyText, selection = null } = defineProps<{
    columns: readonly GridColumn<S>[];
    items: readonly T[];
    sort: DeepReadonly<GridSort<S>>;
    loading: boolean;
    failed: boolean;
    filtered: boolean;
    emptyText: PlainMessageKey;
    selection?: PageSelection | null;
}>();

defineEmits<{
    sort: [column: S];
    reload: [];
}>();

defineSlots<{ row: (props: { item: T }) => unknown }>();
</script>

<template>
    <ScrollShadow>
        <table
            class="w-full text-sm"
            :aria-busy="loading"
        >
            <thead class="border-b border-slate-200 bg-slate-100">
                <tr>
                    <th
                        v-if="selection !== null"
                        scope="col"
                        class="w-10 py-3 pr-2 pl-4"
                    >
                        <BaseCheckbox
                            :model-value="selection.pageSelected"
                            :label="t('grid.select_page')"
                            :disabled="selection.pageSelectable === false"
                            hide-label
                            @update:model-value="selection.togglePage()"
                        />
                    </th>
                    <th
                        v-for="column in columns"
                        :key="column.key"
                        scope="col"
                        :aria-sort="ariaSort(column, sort)"
                        class="px-4 py-3 text-xs font-semibold tracking-wide whitespace-nowrap text-slate-600 uppercase"
                        :class="column.align === 'right' ? 'text-right' : 'text-left'"
                    >
                        <button
                            v-if="column.sort !== undefined"
                            type="button"
                            class="inline-flex cursor-pointer items-center gap-1 rounded-sm uppercase
                                hover:text-ink-900 focus:outline-none focus-visible:ring-2
                                focus-visible:ring-ink-900"
                            @click="$emit('sort', column.sort)"
                        >
                            {{ t(column.label) }}
                            <IconChevronDown
                                class="size-3.5"
                                :class="sortMark(column, sort)"
                            />
                        </button>
                        <span
                            v-else-if="column.hideLabel"
                            class="sr-only"
                        >
                            {{ t(column.label) }}
                        </span>
                        <template v-else>
                            {{ t(column.label) }}
                        </template>
                    </th>
                </tr>
            </thead>
            <tbody
                class="divide-y divide-slate-100 transition-opacity"
                :class="loading && items.length > 0 ? 'opacity-50' : undefined"
            >
                <tr v-if="items.length === 0">
                    <td
                        :colspan="selection === null ? columns.length : columns.length + 1"
                        class="px-4 py-10 text-center text-slate-500"
                    >
                        <template v-if="loading">
                            <BaseSpinner class="mx-auto size-6 text-ink-900" />
                            <span class="sr-only">{{ t('grid.loading') }}</span>
                        </template>
                        <span
                            v-else-if="failed"
                            class="flex flex-col items-center gap-3"
                        >
                            {{ t('grid.failed') }}
                            <BaseButton
                                variant="secondary"
                                size="sm"
                                @click="$emit('reload')"
                            >
                                {{ t('grid.retry') }}
                            </BaseButton>
                        </span>
                        <template v-else>
                            {{ filtered ? t('grid.no_match') : t(emptyText) }}
                        </template>
                    </td>
                </tr>
                <template
                    v-for="item in items"
                    :key="item.id"
                >
                    <slot
                        name="row"
                        :item="item"
                    />
                </template>
            </tbody>
        </table>
    </ScrollShadow>
</template>

<script setup lang="ts">
import SessionRow from '@/app/Auth/Components/SessionRow.vue';
import { useSessionsGrid } from '@/app/Auth/Composables/useSessionsGrid';
import { useSignOutSessions } from '@/app/Auth/Composables/useSignOutSessions';
import BulkActionBar from '@/shared/BulkActions/BulkActionBar.vue';
import FilterPanel from '@/shared/FilterPanel/FilterPanel.vue';
import DataGrid from '@/shared/Grid/Components/DataGrid.vue';
import GridFrame from '@/shared/Grid/Components/GridFrame.vue';
import GridPagination from '@/shared/Grid/Components/GridPagination.vue';
import { t } from '@/shared/I18n/Texts/translate';
import BaseInput from '@/shared/Inputs/BaseInput.vue';
import ConfirmModal from '@/shared/Modal/ConfirmModal.vue';

const { grid, columns, maxIp } = useSessionsGrid();
const {
    items,
    total,
    page,
    perPage,
    sort,
    loading,
    failed,
    filters,
    filtered,
    filled,
    selection,
    reload,
    sortBy,
    goTo,
    clearFilters,
} = grid;
const { asking, question, busy, actions, askOne, askSelected, signOut } = useSignOutSessions(grid);
</script>

<template>
    <section class="space-y-4">
        <div class="space-y-1">
            <h2 class="text-xl font-semibold tracking-tight text-ink-900">
                {{ t('sessions.title') }}
            </h2>
            <p class="text-sm text-slate-600">
                {{ t('sessions.lead') }}
            </p>
        </div>
        <FilterPanel
            name="sessions"
            :active="filled"
            @clear="clearFilters"
        >
            <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
                <BaseInput
                    v-model="filters.ip"
                    name="sessionsIp"
                    :label="t('sessions.ip')"
                    :placeholder="t('sessions.ip_filter')"
                    :maxlength="maxIp"
                    autocomplete="off"
                    size="sm"
                    :active="filters.ip.trim() !== ''"
                />
            </div>
        </FilterPanel>
        <BulkActionBar
            :count="selection.count"
            :total="selection.total"
            :all="selection.all"
            :actions="actions"
            @action="askSelected"
            @select-all="selection.selectAll()"
            @clear="selection.clear()"
        />
        <GridFrame>
            <DataGrid
                :columns="columns"
                :items="items"
                :sort="sort"
                :loading="loading"
                :failed="failed"
                :filtered="filtered"
                :selection="selection"
                empty-text="sessions.empty"
                @sort="sortBy"
                @reload="reload"
            >
                <template #row="{ item }">
                    <SessionRow
                        :session="item"
                        :selection="selection"
                        @sign-out="askOne"
                    />
                </template>
            </DataGrid>
            <GridPagination
                :page="page"
                :per-page="perPage"
                :total="total"
                @go-to="goTo"
            />
        </GridFrame>
        <ConfirmModal
            v-model:open="asking"
            :title="question.title"
            :message="question.message"
            :confirm-text="t('sessions.sign_out')"
            :busy="busy"
            @confirm="signOut"
        />
    </section>
</template>

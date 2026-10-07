<script setup lang="ts">
import { useId } from 'vue';
import { useFilterPanel } from '@/shared/FilterPanel/Composables/useFilterPanel';
import { t } from '@/shared/I18n/Texts/translate';
import IconChevronDown from '@/shared/Icons/IconChevronDown.vue';

const { name, active } = defineProps<{
    name: string;
    active: boolean;
}>();

defineEmits<{ clear: [] }>();

defineSlots<{ default: () => unknown }>();

const { open, toggle } = useFilterPanel(name, () => active);
const panelId = useId();
</script>

<template>
    <div>
        <button
            type="button"
            :aria-expanded="open"
            :aria-controls="panelId"
            :aria-label="active ? t('filters.active') : undefined"
            class="inline-flex cursor-pointer items-center gap-1.5 rounded-full px-3 py-1.5 text-sm font-medium
                transition-colors focus:outline-none focus-visible:ring-2 focus-visible:ring-ink-900"
            :class="active
                ? 'bg-brand-600 text-white hover:bg-brand-700'
                : 'bg-slate-200/60 text-slate-700 hover:bg-slate-200 hover:text-ink-900'"
            @click="toggle"
        >
            {{ t('filters.title') }}
            <IconChevronDown
                class="size-3.5 transition-transform duration-200"
                :class="open ? 'rotate-180' : undefined"
            />
        </button>
        <div
            v-show="open"
            :id="panelId"
            class="mt-3 rounded-lg bg-white p-4 shadow-sm"
        >
            <slot />
            <div class="mt-3 ml-1 flex">
                <button
                    type="button"
                    :disabled="active === false"
                    class="text-xs underline underline-offset-2 transition-colors focus:outline-none
                        focus-visible:ring-2 focus-visible:ring-ink-900"
                    :class="active
                        ? 'cursor-pointer text-red-400 hover:text-red-600'
                        : 'cursor-default text-slate-300'"
                    @click="$emit('clear')"
                >
                    {{ t('filters.clear') }}
                </button>
            </div>
        </div>
    </div>
</template>

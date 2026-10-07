<script setup lang="ts">
import LanguageSwitcher from '@/shared/I18n/Components/LanguageSwitcher.vue';
import type { SaveLanguage } from '@/shared/I18n/Composables/useLanguageSwitcher';
import { t } from '@/shared/I18n/Texts/translate';
import IconLogout from '@/shared/Icons/IconLogout.vue';
import BaseTooltip from '@/shared/Tooltip/BaseTooltip.vue';

defineProps<{
    email: string;
    saveLanguage: SaveLanguage;
}>();

defineEmits<{ signOut: [] }>();
</script>

<template>
    <div class="flex items-center justify-between gap-4 rounded-lg bg-white px-4 py-3 shadow-sm">
        <span class="min-w-0 truncate text-sm font-medium text-ink-900">{{ email }}</span>
        <div class="flex shrink-0 items-center gap-1">
            <LanguageSwitcher
                tooltip-position="top"
                :save="saveLanguage"
            />
            <BaseTooltip
                :text="t('account.logout')"
                align="right"
            >
                <button
                    type="button"
                    :aria-label="t('account.logout')"
                    class="flex size-9 cursor-pointer items-center justify-center rounded-md text-slate-500
                        transition-colors hover:bg-red-50 hover:text-red-600 focus:outline-none
                        focus-visible:ring-2 focus-visible:ring-ink-900"
                    @click="$emit('signOut')"
                >
                    <IconLogout
                        class="size-5"
                        stroke-width="1.75"
                    />
                </button>
            </BaseTooltip>
        </div>
    </div>
</template>

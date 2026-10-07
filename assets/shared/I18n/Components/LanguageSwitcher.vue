<!-- The Go pages draw the same switcher from views/shared/languages.html. Change both together. -->
<script setup lang="ts">
import BaseDropdown from '@/shared/Dropdown/BaseDropdown.vue';
import { t } from '@/shared/I18n/Texts/translate';
import { type SaveLanguage, saveNothing, useLanguageSwitcher } from '@/shared/I18n/Composables/useLanguageSwitcher';
import IconChevronDown from '@/shared/Icons/IconChevronDown.vue';
import BaseTooltip from '@/shared/Tooltip/BaseTooltip.vue';

const { tooltipPosition = 'bottom', save = saveNothing } = defineProps<{
    tooltipPosition?: 'top' | 'bottom';
    save?: SaveLanguage;
}>();

const { languages, active, pick, preload } = useLanguageSwitcher(save);
</script>

<template>
    <BaseDropdown v-if="active">
        <template #trigger="{ open, toggle }">
            <BaseTooltip
                :text="t('language.change')"
                :position="tooltipPosition"
                align="right"
                :disabled="open"
            >
                <button
                    type="button"
                    :aria-label="t('language.change')"
                    :aria-expanded="open"
                    class="flex h-9 cursor-pointer items-center gap-1 rounded-md px-2 text-slate-500 transition-colors
                        hover:bg-slate-100 hover:text-slate-700 focus:outline-none focus-visible:ring-2
                        focus-visible:ring-ink-900"
                    @click="toggle"
                    @focus="preload"
                    @pointerenter="preload"
                >
                    <img
                        :src="active.flag"
                        alt=""
                        width="24"
                        height="16"
                        class="h-4 w-6 rounded-xs ring-1 ring-black/10"
                    >
                    <IconChevronDown class="size-4" />
                </button>
            </BaseTooltip>
        </template>
        <ul>
            <template
                v-for="language in languages"
                :key="language.code"
            >
                <li
                    v-if="language.active"
                    :lang="language.code"
                    aria-current="true"
                    class="flex items-center gap-3 px-4 py-2 text-sm font-semibold text-ink-900"
                >
                    <img
                        :src="language.flag"
                        alt=""
                        width="24"
                        height="16"
                        class="h-4 w-6 shrink-0 rounded-xs ring-1 ring-black/10"
                    >
                    {{ language.name }}
                </li>
                <li
                    v-else
                    :lang="language.code"
                >
                    <a
                        :href="language.href"
                        class="flex items-center gap-3 px-4 py-2 text-sm text-slate-700 transition-colors
                            hover:bg-slate-50 hover:text-ink-900 focus:outline-none focus-visible:bg-slate-50"
                        @click="pick(language, $event)"
                    >
                        <img
                            :src="language.flag"
                            alt=""
                            width="24"
                            height="16"
                            class="h-4 w-6 shrink-0 rounded-xs ring-1 ring-black/10 grayscale-[70%]"
                        >
                        {{ language.name }}
                    </a>
                </li>
            </template>
        </ul>
    </BaseDropdown>
</template>

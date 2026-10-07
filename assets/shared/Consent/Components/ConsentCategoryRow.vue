<script setup lang="ts">
import { t } from '@/shared/I18n/Texts/translate';
import IconChevronDown from '@/shared/Icons/IconChevronDown.vue';
import BaseSwitch from '@/shared/Inputs/BaseSwitch.vue';

const { name, title, description, locked } = defineProps<{
    name: string;
    title: string;
    description: string;
    locked?: boolean;
}>();

const model = defineModel<boolean>({ required: true });
</script>

<template>
    <li class="relative rounded-lg bg-slate-100">
        <details class="group/row">
            <summary
                class="flex cursor-pointer list-none items-center gap-3 py-3.5 pr-20 pl-4 text-sm font-semibold
                    [&::-webkit-details-marker]:hidden"
                :class="{ 'sm:pr-44': locked }"
            >
                <IconChevronDown
                    class="size-5 shrink-0 rounded-full bg-slate-200 p-1 transition-transform group-open/row:rotate-180
                        motion-reduce:transition-none"
                />
                {{ title }}
            </summary>
            <p class="pr-4 pb-4 pl-12 text-sm text-slate-600">
                {{ description }}
            </p>
        </details>
        <div class="absolute top-3 right-4 flex items-center gap-3">
            <span
                v-if="locked"
                class="hidden rounded-md bg-slate-200 px-2 py-0.5 text-xs font-medium text-slate-700 sm:inline"
            >
                {{ t('consent.always_on') }}
            </span>
            <BaseSwitch
                v-model="model"
                :name="name"
                :label="title"
                :disabled="locked"
            />
        </div>
    </li>
</template>

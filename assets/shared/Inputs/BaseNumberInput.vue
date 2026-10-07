<script setup lang="ts">
import { computed } from 'vue';
import type { InputProps } from '@/shared/Inputs/types/InputProps';
import { numberCodec } from '@/shared/Inputs/Field/fieldCodecs';
import FieldInput from '@/shared/Inputs/Field/FieldInput.vue';

type Props = InputProps & {
    min?: number | null;
    max?: number | null;
    step?: number | 'any' | null;
    valueOnClear?: number | null;
};

const { min = null, max = null, step = null, valueOnClear = null, ...props } = defineProps<Props>();

const model = defineModel<number | null>({ required: true });

const codec = computed(() => numberCodec(valueOnClear));
</script>

<template>
    <FieldInput
        v-model="model"
        v-bind="props"
        type="number"
        :min="min"
        :max="max"
        :step="step"
        :codec="codec"
    />
</template>

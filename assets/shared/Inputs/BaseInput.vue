<script setup lang="ts">
import type { InputProps } from '@/shared/Inputs/types/InputProps';
import type { TextType } from '@/shared/Inputs/types/TextType';
import { textCodec } from '@/shared/Inputs/Field/fieldCodecs';
import FieldInput from '@/shared/Inputs/Field/FieldInput.vue';

type Props = InputProps & {
    type?: TextType;
    maxlength?: number | null;
};

const props = defineProps<Props>();

const model = defineModel<string>({ required: true });
</script>

<template>
    <FieldInput
        v-model="model"
        v-bind="props"
        :type="props.type ?? 'text'"
        :codec="textCodec"
    />
</template>

<script setup lang="ts" generic="T">
import type { FieldCodec } from '@/shared/Inputs/types/FieldCodec';
import type { InputProps } from '@/shared/Inputs/types/InputProps';
import type { TextType } from '@/shared/Inputs/types/TextType';
import { useFieldText } from '@/shared/Inputs/Composables/useFieldText';
import { controlClasses, sideMarkClasses, sideRoom } from '@/shared/Inputs/Field/fieldClasses';
import FieldFrame from '@/shared/Inputs/Field/FieldFrame.vue';
import BaseSpinner from '@/shared/Loading/BaseSpinner.vue';

type Props = InputProps & {
    type: TextType | 'number';
    codec: FieldCodec<T>;
    maxlength?: number | null | undefined;
    min?: number | null | undefined;
    max?: number | null | undefined;
    step?: number | 'any' | null | undefined;
};

const {
    name,
    label,
    type,
    autocomplete,
    codec,
    error = null,
    status = null,
    statusVariant = 'info',
    required,
    disabled,
    loading,
    size = 'md',
    active,
    placeholder = null,
    maxlength = null,
    min = null,
    max = null,
    step = null,
} = defineProps<Props>();

const model = defineModel<T>({ required: true });

const { text, typed, committed } = useFieldText(model, () => codec);
</script>

<template>
    <FieldFrame
        v-slot="{ describedBy }"
        :name="name"
        :label="label"
        :size="size"
        :required="required"
        :error="error"
        :status="status"
        :status-variant="statusVariant"
    >
        <div class="relative">
            <input
                :id="name"
                :value="text"
                :name="name"
                :type="type"
                :autocomplete="autocomplete"
                :placeholder="placeholder ?? undefined"
                :maxlength="maxlength ?? undefined"
                :min="min ?? undefined"
                :max="max ?? undefined"
                :step="step ?? undefined"
                :required="required"
                :disabled="disabled"
                :aria-invalid="error !== null"
                :aria-describedby="describedBy"
                :aria-busy="loading"
                :class="[controlClasses(size, error, active), loading ? sideRoom(size) : undefined]"
                @input="typed"
                @change="committed"
            >
            <BaseSpinner
                v-if="loading"
                :class="sideMarkClasses(size)"
            />
        </div>
    </FieldFrame>
</template>

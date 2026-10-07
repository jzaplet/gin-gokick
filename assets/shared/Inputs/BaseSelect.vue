<script setup lang="ts">
import type { FieldSize } from '@/shared/Inputs/types/FieldSize';
import type { SelectOption } from '@/shared/Inputs/types/SelectOption';
import type { StatusVariant } from '@/shared/Inputs/types/StatusVariant';
import IconChevronDown from '@/shared/Icons/IconChevronDown.vue';
import { controlClasses, sideMarkClasses, sideRoom } from '@/shared/Inputs/Field/fieldClasses';
import FieldFrame from '@/shared/Inputs/Field/FieldFrame.vue';
import BaseSpinner from '@/shared/Loading/BaseSpinner.vue';

type Props = {
    name: string;
    label: string;
    options: readonly SelectOption[];
    placeholder?: string | null;
    error?: string | null;
    status?: string | null;
    statusVariant?: StatusVariant;
    required?: boolean;
    disabled?: boolean;
    loading?: boolean;
    size?: FieldSize;
    active?: boolean;
};

const {
    name,
    label,
    options,
    placeholder = null,
    error = null,
    status = null,
    statusVariant = 'info',
    required,
    disabled,
    loading,
    size = 'md',
    active,
} = defineProps<Props>();

const model = defineModel<string>({ required: true });
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
            <select
                :id="name"
                v-model="model"
                :name="name"
                :required="required"
                :disabled="disabled"
                :aria-invalid="error !== null"
                :aria-describedby="describedBy"
                :aria-busy="loading"
                class="cursor-pointer appearance-none has-[option[value='']:disabled:checked]:text-slate-500"
                :class="[controlClasses(size, error, active), sideRoom(size)]"
            >
                <option
                    v-if="placeholder !== null"
                    value=""
                    disabled
                >
                    {{ placeholder }}
                </option>
                <option
                    v-for="option in options"
                    :key="option.value"
                    :value="option.value"
                    class="text-ink-900"
                >
                    {{ option.label }}
                </option>
            </select>
            <BaseSpinner
                v-if="loading"
                :class="sideMarkClasses(size)"
            />
            <IconChevronDown
                v-else
                :class="sideMarkClasses(size)"
            />
        </div>
    </FieldFrame>
</template>

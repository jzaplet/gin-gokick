<script setup lang="ts">
import type { FieldSize } from '@/shared/Inputs/types/FieldSize';
import type { StatusVariant } from '@/shared/Inputs/types/StatusVariant';
import { groupClasses, labelClasses, statusClasses } from '@/shared/Inputs/Field/fieldClasses';
import { describedBy, errorId, statusId } from '@/shared/Inputs/Field/fieldMessages';

const { name, label, size, required, error, status, statusVariant } = defineProps<{
    name: string;
    label: string;
    size: FieldSize;
    required: boolean;
    error: string | null;
    status: string | null;
    statusVariant: StatusVariant;
}>();

defineSlots<{ default: (props: { describedBy: string | undefined }) => unknown }>();
</script>

<template>
    <div :class="groupClasses(size)">
        <label
            :for="name"
            :class="labelClasses(size)"
        >
            {{ label }}
            <span
                v-if="required"
                class="text-red-600"
                aria-hidden="true"
            >*</span>
        </label>
        <slot :described-by="describedBy(name, error, status)" />
        <p
            v-if="error !== null"
            :id="errorId(name)"
            class="text-sm text-red-600"
        >
            {{ error }}
        </p>
        <p
            v-else-if="status !== null"
            :id="statusId(name)"
            :class="statusClasses(statusVariant)"
        >
            {{ status }}
        </p>
    </div>
</template>

<script setup lang="ts">
import BaseCheckbox from '@/shared/Inputs/BaseCheckbox.vue';

defineProps<{
    checked: boolean;
    disabled: boolean;
    label: string;
}>();

defineEmits<{ toggle: [] }>();
</script>

<template>
    <td class="w-10 py-3 pr-2 pl-4">
        <BaseCheckbox
            :model-value="checked"
            :label="label"
            :disabled="disabled"
            hide-label
            @update:model-value="$emit('toggle')"
        />
    </td>
</template>

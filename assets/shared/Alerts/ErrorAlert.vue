<script setup lang="ts">
const { message } = defineProps<{ message: string | null }>();
</script>

<template>
    <p
        v-if="message !== null"
        role="alert"
        class="rounded-xl border border-red-200 bg-red-50 p-3 text-sm text-red-800"
    >
        {{ message }}
    </p>
</template>

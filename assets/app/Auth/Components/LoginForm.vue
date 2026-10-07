<script setup lang="ts">
import { useLoginForm } from '@/app/Auth/Composables/useLoginForm';
import ErrorAlert from '@/shared/Alerts/ErrorAlert.vue';
import BaseButton from '@/shared/Buttons/BaseButton.vue';
import { t, tm } from '@/shared/I18n/Texts/translate';
import BaseInput from '@/shared/Inputs/BaseInput.vue';

const { form, errors, sending, submit } = useLoginForm();
</script>

<template>
    <form
        class="space-y-4"
        @submit.prevent="submit"
    >
        <BaseInput
            v-model="form.email"
            name="email"
            :label="t('form.email')"
            type="email"
            autocomplete="email"
            :error="tm(errors.email)"
            :disabled="sending"
            required
        />
        <BaseInput
            v-model="form.password"
            name="password"
            :label="t('form.password')"
            type="password"
            autocomplete="current-password"
            :error="tm(errors.password)"
            :disabled="sending"
            required
        />
        <ErrorAlert :message="tm(errors.general)" />
        <BaseButton
            type="submit"
            class="w-full"
            :loading="sending"
        >
            {{ t('login.submit') }}
        </BaseButton>
    </form>
</template>

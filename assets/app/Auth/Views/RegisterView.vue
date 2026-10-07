<script setup lang="ts">
import RegisterForm from '@/app/Auth/Components/RegisterForm.vue';
import { useAuthPageLink } from '@/app/Auth/Composables/useAuthPageLink';
import { t } from '@/shared/I18n/Texts/translate';
import AuthCard from '@/shared/Layout/Components/AuthCard.vue';
import AuthTitle from '@/shared/Layout/Components/AuthTitle.vue';

const { link } = useAuthPageLink('login');
</script>

<template>
    <AuthCard>
        <AuthTitle>{{ t('register.title') }}</AuthTitle>
        <RegisterForm />
        <p class="text-center text-sm text-slate-600">
            {{ t('register.has_account') }}
            <RouterLink
                :to="link"
                class="font-medium text-ink-900 underline"
            >
                {{ t('register.to_login') }}
            </RouterLink>
        </p>
    </AuthCard>
</template>

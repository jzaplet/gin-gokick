<script setup lang="ts">
import LoginForm from '@/app/Auth/Components/LoginForm.vue';
import { useAuthPageLink } from '@/app/Auth/Composables/useAuthPageLink';
import { t } from '@/shared/I18n/Texts/translate';
import AuthCard from '@/shared/Layout/Components/AuthCard.vue';
import AuthTitle from '@/shared/Layout/Components/AuthTitle.vue';

const { link } = useAuthPageLink('register');
</script>

<template>
    <AuthCard>
        <AuthTitle>{{ t('login.title') }}</AuthTitle>
        <LoginForm />
        <p class="text-center text-sm text-slate-600">
            {{ t('login.no_account') }}
            <RouterLink
                :to="link"
                class="font-medium text-ink-900 underline"
            >
                {{ t('login.to_register') }}
            </RouterLink>
        </p>
    </AuthCard>
</template>

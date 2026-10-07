<script setup lang="ts">
import ErrorAlert from '@/shared/Alerts/ErrorAlert.vue';
import BrandLogo from '@/shared/Brand/BrandLogo.vue';
import BaseButton from '@/shared/Buttons/BaseButton.vue';
import { t, tm } from '@/shared/I18n/Texts/translate';
import AccountBar from '@/app/User/Components/AccountBar.vue';
import { useAuthLayout } from '@/shared/Layout/Composables/useAuthLayout';
import LoadingScreen from '@/shared/Loading/LoadingScreen.vue';

const { account, load, signOut, saveLanguage } = useAuthLayout();
</script>

<template>
    <LoadingScreen v-if="account.status === 'loading'" />
    <div
        v-else-if="account.status === 'failed'"
        class="mx-auto flex min-h-screen w-full max-w-md flex-col items-center justify-center gap-4 px-4"
    >
        <ErrorAlert :message="tm(account.error)" />
        <BaseButton @click="load">
            {{ t('account.retry') }}
        </BaseButton>
    </div>
    <div
        v-else
        class="mx-auto w-full max-w-5xl px-4 sm:px-6"
    >
        <header class="py-5">
            <RouterLink
                :to="{ name: 'dashboard' }"
                class="inline-flex"
            >
                <BrandLogo />
            </RouterLink>
        </header>
        <main class="space-y-8 pb-10">
            <AccountBar
                :email="account.user.email"
                :save-language="saveLanguage"
                @sign-out="signOut"
            />
            <slot />
        </main>
    </div>
</template>

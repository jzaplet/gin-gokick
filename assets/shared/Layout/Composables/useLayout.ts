import { type Component, computed, type ComputedRef } from 'vue';
import { useRoute } from 'vue-router';
import { Access } from '@/router/types/Access';
import AuthLayout from '@/shared/Layout/AuthLayout.vue';
import PublicLayout from '@/shared/Layout/PublicLayout.vue';

export const useLayout = (): ComputedRef<Component | undefined> => {
    const route = useRoute();

    return computed(() => {
        const access = route.meta['access'];

        if (access === undefined) {
            return undefined;
        }

        return access === Access.User ? AuthLayout : PublicLayout;
    });
};

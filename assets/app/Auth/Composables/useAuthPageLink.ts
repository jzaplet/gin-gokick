import { computed, type ComputedRef } from 'vue';
import { type RouteLocationRaw, useRoute } from 'vue-router';

type AuthPageLink = {
    link: ComputedRef<RouteLocationRaw>;
};

export const useAuthPageLink = (name: 'login' | 'register'): AuthPageLink => {
    const route = useRoute();
    const link = computed((): RouteLocationRaw => ({
        name,
        query: route.query,
    }));

    return { link };
};

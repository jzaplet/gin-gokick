import type { LocationQueryValue, RouteLocationNormalized, RouteLocationRaw } from 'vue-router';
import { screenIn, screenOf } from '@/shared/Navigation/screenPath';

export type RedirectQuery = LocationQueryValue | LocationQueryValue[] | undefined;

const appScreen = /^\/(?![/\\])/;

export const loginReturningTo = (route: Pick<RouteLocationNormalized, 'params' | 'fullPath'>): RouteLocationRaw => ({
    name: 'login',
    query: { redirect: screenOf(route) },
});

export const redirectAfterLogin = (redirect: RedirectQuery, language: string): RouteLocationRaw =>
    typeof redirect === 'string' && appScreen.test(redirect)
        ? screenIn(language, redirect)
        : {
                name: 'dashboard',
                params: { lang: language },
            };

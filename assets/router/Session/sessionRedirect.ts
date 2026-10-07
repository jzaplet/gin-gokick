import type { RouteLocationNormalized, RouteLocationRaw } from 'vue-router';
import { Access } from '@/router/types/Access';
import { signedIn } from '@/shared/Auth/Session/currentSession';
import { routeLanguage } from '@/shared/Navigation/screenPath';

export const sessionRedirect = (
    route: Pick<RouteLocationNormalized, 'meta' | 'params'>,
): RouteLocationRaw | undefined => {
    const access = route.meta['access'];
    const params = { lang: routeLanguage(route) };

    if (access === Access.Guest && signedIn()) {
        return {
            name: 'dashboard',
            params,
        };
    }

    if (access === Access.User && signedIn() === false) {
        return {
            name: 'login',
            params,
        };
    }

    return undefined;
};

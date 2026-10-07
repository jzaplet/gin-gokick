import type { RouteLocationNormalized } from 'vue-router';

export const routeLanguage = (route: Pick<RouteLocationNormalized, 'params'>): string => {
    const language = route.params['lang'];

    return typeof language === 'string' ? language : '';
};

export const screenIn = (language: string, screen: string): string => `/${language}/app${screen}`;

export const screenOf = (route: Pick<RouteLocationNormalized, 'params' | 'fullPath'>): string =>
    route.fullPath.slice(screenIn(routeLanguage(route), '').length) || '/';

import { createRouter, createWebHistory, isNavigationFailure } from 'vue-router';
import { routes } from '@/router/routes';
import { changesOnlyGrids } from '@/shared/Grid/Url/gridQuery';
import { followSessionOfOtherTabs } from '@/router/Session/followSessionOfOtherTabs';
import { sessionRedirect } from '@/router/Session/sessionRedirect';
import { offeredLocale } from '@/shared/I18n/Page/localeMeta';
import { loadDictionary, showDictionary } from '@/shared/I18n/Texts/pageDictionary';
import { routeLanguage } from '@/shared/Navigation/screenPath';
import { showPageMeta } from '@/shared/Seo/showPageMeta';
import { trackPageView } from '@/shared/Tracking/trackPageView';

export const router = createRouter({
    history: createWebHistory(),
    routes,
});

router.beforeEach(async (to) => {
    const locale = offeredLocale(routeLanguage(to));

    if (locale === undefined || (await loadDictionary(locale)) === false) {
        return false;
    }

    return sessionRedirect(to);
});

router.afterEach((to, from, failure) => {
    const locale = offeredLocale(routeLanguage(to));

    if (isNavigationFailure(failure) || locale === undefined) {
        return;
    }

    showDictionary(locale);

    if (to.path === from.path && changesOnlyGrids(from.query, to.query)) {
        return;
    }

    showPageMeta(to);
    trackPageView();
});

followSessionOfOtherTabs(router);

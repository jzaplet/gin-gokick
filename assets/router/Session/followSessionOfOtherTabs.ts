import type { Router } from 'vue-router';
import { sessionRedirect } from '@/router/Session/sessionRedirect';
import { followOtherTabs } from '@/shared/Auth/Session/followOtherTabs';

export const followSessionOfOtherTabs = (router: Router): void => {
    followOtherTabs(() => {
        const target = sessionRedirect(router.currentRoute.value);

        if (target !== undefined) {
            void router.replace(target);
        }
    });
};

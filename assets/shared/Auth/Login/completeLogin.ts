import { isNavigationFailure, NavigationFailureType } from 'vue-router';
import { router } from '@/router';
import { type RedirectQuery, redirectAfterLogin } from '@/shared/Auth/Login/loginRedirect';
import type { SignedInUser } from '@/shared/Auth/Login/types/SignedInUser';
import { rememberSession } from '@/shared/Auth/Session/currentSession';
import { languageOfLocale, offeredLocale } from '@/shared/I18n/Page/localeMeta';
import { shownLanguage } from '@/shared/I18n/Texts/pageDictionary';

export const completeLogin = async (user: SignedInUser, redirect: RedirectQuery): Promise<void> => {
    const language = languageOfLocale(user.locale);
    const offered = offeredLocale(language) === undefined ? shownLanguage() : language;

    rememberSession(user);

    const failure = await router.push(redirectAfterLogin(redirect, offered));

    if (offered !== shownLanguage() && isNavigationFailure(failure, NavigationFailureType.aborted)) {
        await router.push(redirectAfterLogin(redirect, shownLanguage()));
    }
};

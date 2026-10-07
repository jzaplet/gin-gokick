import { type DeepReadonly, onMounted, readonly, type Ref, shallowRef } from 'vue';
import { authFetch } from '@/shared/Auth/authFetch';
import { logout } from '@/shared/Auth/logout';
import { type CurrentUser, isCurrentUser } from '@/app/User/types/CurrentUser';
import type { LocaleRequest } from '@/app/User/types/LocaleRequest';
import type { ApiGeneralError, ApiMessage } from '@/shared/Fetch';
import { isNull } from '@/shared/TypeGuards/typeGuards';
import type { SaveLanguage, SwitcherLanguage } from '@/shared/I18n/Composables/useLanguageSwitcher';
import { showToast } from '@/shared/Toast/Queue/toastQueue';

type Loading = {
    status: 'loading';
};

type Failed = {
    status: 'failed';
    error: ApiMessage | undefined;
};

type Ready = {
    status: 'ready';
    user: CurrentUser;
};

type AuthLayout = {
    account: DeepReadonly<Ref<Loading | Failed | Ready>>;
    load: () => Promise<void>;
    signOut: () => Promise<void>;
    saveLanguage: SaveLanguage;
};

export const useAuthLayout = (): AuthLayout => {
    const account = shallowRef<Loading | Failed | Ready>({ status: 'loading' });

    const load = async (): Promise<void> => {
        account.value = { status: 'loading' };

        const result = await authFetch<CurrentUser>('GET', '/api/user/me', { validate: isCurrentUser });

        account.value = result.success
            ? {
                    status: 'ready',
                    user: result.data,
                }
            : {
                    status: 'failed',
                    error: result.data.general,
                };
    };

    const signOut = async (): Promise<void> => {
        const result = await logout();

        if (result.success) {
            showToast('success', 'account.signed_out_title', 'account.signed_out');

            return;
        }

        showToast('error', 'toast.error_title', result.data.general);
    };

    const saveLanguage = async (language: SwitcherLanguage): Promise<boolean> => {
        const result = await authFetch<null, ApiGeneralError, LocaleRequest>('PUT', '/api/user/locale', {
            body: { locale: language.locale },
            validate: isNull,
        });

        if (result.success) {
            return true;
        }

        showToast('error', 'toast.error_title', result.data.general);

        return false;
    };

    onMounted(load);

    return {
        account: readonly(account),
        load,
        signOut,
        saveLanguage,
    };
};

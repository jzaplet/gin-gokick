import { type DeepReadonly, reactive, readonly, type Ref, ref } from 'vue';
import { useRoute } from 'vue-router';
import {
    isRegisterRequestErrors,
    type RegisterRequest,
    type RegisterRequestErrors,
} from '@/app/Auth/types/RegisterRequest';
import { completeLogin } from '@/shared/Auth/Login/completeLogin';
import { isSignedInUser, type SignedInUser } from '@/shared/Auth/Login/types/SignedInUser';
import { apiFetch } from '@/shared/Fetch';
import { shownLocale } from '@/shared/I18n/Texts/pageDictionary';
import { showToast } from '@/shared/Toast/Queue/toastQueue';
import { trackEvent } from '@/shared/Tracking/trackEvent';
import { TrackedEvent } from '@/shared/Tracking/types/TrackedEvent';

type RegisterForm = {
    form: Omit<RegisterRequest, 'locale'>;
    errors: DeepReadonly<Ref<RegisterRequestErrors>>;
    sending: DeepReadonly<Ref<boolean>>;
    submit: () => Promise<void>;
};

export const useRegisterForm = (): RegisterForm => {
    const route = useRoute();
    const form = reactive<Omit<RegisterRequest, 'locale'>>({
        email: '',
        password: '',
    });
    const errors = ref<RegisterRequestErrors>({});
    const sending = ref(false);

    const submit = async (): Promise<void> => {
        sending.value = true;
        errors.value = {};

        const result = await apiFetch<SignedInUser, RegisterRequestErrors, RegisterRequest>(
            'POST',
            '/api/auth/register',
            {
                body: {
                    ...form,
                    locale: shownLocale(),
                },
                validate: isSignedInUser,
                validateError: isRegisterRequestErrors,
            },
        );

        sending.value = false;

        if (result.success) {
            await trackEvent(TrackedEvent.SignUp, { email: form.email });
            await completeLogin(result.data, route.query['redirect']);
            showToast('success', 'register.signed_up_title', 'register.signed_up');

            return;
        }

        errors.value = result.data;
    };

    return {
        form,
        errors: readonly(errors),
        sending: readonly(sending),
        submit,
    };
};

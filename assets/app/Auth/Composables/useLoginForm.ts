import { type DeepReadonly, reactive, readonly, type Ref, ref } from 'vue';
import { useRoute } from 'vue-router';
import { isLoginRequestErrors, type LoginRequest, type LoginRequestErrors } from '@/app/Auth/types/LoginRequest';
import { completeLogin } from '@/shared/Auth/Login/completeLogin';
import { isSignedInUser, type SignedInUser } from '@/shared/Auth/Login/types/SignedInUser';
import { apiFetch } from '@/shared/Fetch';
import { showToast } from '@/shared/Toast/Queue/toastQueue';
import { trackEvent } from '@/shared/Tracking/trackEvent';
import { TrackedEvent } from '@/shared/Tracking/types/TrackedEvent';

type LoginForm = {
    form: LoginRequest;
    errors: DeepReadonly<Ref<LoginRequestErrors>>;
    sending: DeepReadonly<Ref<boolean>>;
    submit: () => Promise<void>;
};

export const useLoginForm = (): LoginForm => {
    const route = useRoute();
    const form = reactive<LoginRequest>({
        email: '',
        password: '',
    });
    const errors = ref<LoginRequestErrors>({});
    const sending = ref(false);

    const submit = async (): Promise<void> => {
        sending.value = true;
        errors.value = {};

        const result = await apiFetch<SignedInUser, LoginRequestErrors, LoginRequest>('POST', '/api/auth/login', {
            body: form,
            validate: isSignedInUser,
            validateError: isLoginRequestErrors,
        });

        sending.value = false;

        if (result.success) {
            await trackEvent(TrackedEvent.Login);
            await completeLogin(result.data, route.query['redirect']);
            showToast('success', 'login.signed_in_title', 'login.signed_in');

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

import { flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import type { Component } from 'vue';
import LoginForm from '@/app/Auth/Components/LoginForm.vue';
import RegisterForm from '@/app/Auth/Components/RegisterForm.vue';
import { router } from '@/router';
import { forgetSession, signedIn } from '@/shared/Auth/Session/currentSession';
import { t, tm } from '@/shared/I18n/Texts/translate';
import { clearToasts } from '@/shared/Toast/Queue/toastQueue';
import { json } from '../fetch/fixtures';
import { offerEnglish } from '../i18n/languages';

const jan = {
    id: '0192f3a4-5b6c-7d8e-9f01-23456789abcd',
    locale: 'cs_CZ',
};

type Form = {
    component: Component;
    start: string;
    url: string;
    success: number;
    body: string;
};

const credentials = '"email":"jan@example.com","password":"correct horse battery"';

const loginForm: Form = {
    component: LoginForm,
    start: '/cs/app/login',
    url: '/api/auth/login',
    success: 200,
    body: `{${credentials}}`,
};

const registerForm: Form = {
    component: RegisterForm,
    start: '/cs/app/register',
    url: '/api/auth/register',
    success: 201,
    body: `{${credentials},"locale":"cs_CZ"}`,
};

const submit = async (form: Form, response: Response, query = '', app = router): Promise<VueWrapper> => {
    await app.push(`${form.start}${query}`);
    const wrapper = mount(form.component, { global: { plugins: [app] } });

    vi.spyOn(globalThis, 'fetch').mockResolvedValue(response);
    await wrapper.get('input[name="email"]').setValue('jan@example.com');
    await wrapper.get('input[name="password"]').setValue('correct horse battery');
    await wrapper.get('form').trigger('submit');
    await flushPromises();

    return wrapper;
};

afterEach(forgetSession);

afterEach(clearToasts);

const isArguments = (value: unknown): value is IArguments =>
    Object.prototype.toString.call(value) === '[object Arguments]';

const gtagCommands = (): unknown[][] => {
    const layer: unknown = Reflect.get(window, 'dataLayer');

    return Array.isArray(layer)
        ? layer.filter(isArguments).map((entry) => Array.from(entry, (item: unknown) => item))
        : [];
};

describe.each<Form>([
    loginForm,
    registerForm,
])('$url form', (form) => {
    it('sends the credentials, remembers the session and goes to the dashboard', async () => {
        await submit(form, json(jan, form.success));

        const [url, init] = vi.mocked(fetch).mock.calls[0] ?? [];

        expect(url).toBe(form.url);
        expect(init?.method).toBe('POST');
        expect(init?.body).toBe(form.body);
        expect(router.currentRoute.value.name).toBe('dashboard');
        expect(signedIn()).toBe(true);
        expect(localStorage.getItem('session')).toBe(`{"id":"${jan.id}"}`);
    });

    it('returns to the page that asked for the login', async () => {
        await submit(form, json(jan, form.success), '?redirect=/settings%3Ftab%3Dmail');

        expect(router.currentRoute.value.fullPath).toBe('/cs/app/settings?tab=mail');
    });

    it('shows each field error under its input', async () => {
        const errors = {
            email: { key: 'validation.email' },
            password: {
                key: 'validation.min_length',
                params: { min: 12 },
            },
        };
        const wrapper = await submit(form, json(errors, 422));

        expect(wrapper.get('#email-error').text()).toBe(tm(errors.email));
        expect(wrapper.get('#password-error').text()).toBe(tm(errors.password));
        expect(wrapper.get('input[name="email"]').attributes('aria-describedby')).toBe('email-error');
        expect(wrapper.find('[role="alert"]').exists()).toBe(false);
    });

    it('locks the form while sending and clears the old errors first', async () => {
        const wrapper = await submit(form, json({ email: { key: 'validation.email' } }, 422));
        const { promise, resolve } = Promise.withResolvers<Response>();

        vi.mocked(fetch).mockReturnValue(promise);
        await wrapper.get('form').trigger('submit');

        expect(wrapper.get('button[type="submit"]').attributes()).toHaveProperty('disabled');
        expect(wrapper.get('button[type="submit"]').attributes('aria-busy')).toBe('true');
        expect(wrapper.get('input[name="email"]').attributes()).toHaveProperty('disabled');
        expect(wrapper.find('#email-error').exists()).toBe(false);

        resolve(json({ general: { key: 'request.internal' } }, 500));
        await flushPromises();

        expect(wrapper.get('button[type="submit"]').attributes()).not.toHaveProperty('disabled');
        expect(wrapper.get('[role="alert"]').text()).toBe(t('request.internal'));
    });
});

describe('the login form', () => {
    it('shows wrong credentials instead of sending the user to the login again', async () => {
        const wrapper = await submit(
            loginForm,
            json({ general: { key: 'auth.login_failed' } }, 401),
            '?redirect=/settings',
        );

        expect(wrapper.get('[role="alert"]').text()).toBe(tm({ key: 'auth.login_failed' }));
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/login?redirect=/settings');
    });
});

describe('the register form', () => {
    afterEach(() => {
        document.head.replaceChildren();
        Reflect.deleteProperty(window, 'dataLayer');
    });

    it('sends the language the page shows', async () => {
        offerEnglish();

        await submit(
            {
                ...registerForm,
                start: '/en/app/register',
            },
            json({
                ...jan,
                locale: 'en_US',
            }, 201),
        );

        expect(vi.mocked(fetch).mock.calls[0]?.[1]?.body).toBe(`{${credentials},"locale":"en_US"}`);
    });

    it('sends the sign-up with the hashed email of the form to Google Ads', async () => {
        vi.resetModules();

        const meta = document.createElement('meta');

        meta.name = 'tracking';
        Object.assign(meta.dataset, {
            googleAds: 'AW-123456789',
            googleAdsConversions: 'sign_up=AbC',
        });
        document.head.append(meta);

        const { startTracking } = await import('@/shared/Tracking/startTracking');
        const { setCurrentChoice } = await import('@/shared/Consent/Choice/currentChoice');
        const { router: app } = await import('@/router');
        const { default: component } = await import('@/app/Auth/Components/RegisterForm.vue');

        startTracking();
        setCurrentChoice({ marketing: true });

        await submit(
            {
                ...registerForm,
                component,
            },
            json(jan, 201),
            '',
            app,
        );

        await vi.waitFor(() => {
            expect(gtagCommands()).toContainEqual([
                'set',
                'user_data',
                { sha256_email_address: 'b82dbe5433cfaeb5238ada81eadf911798ab04ba8ed69e004ec907826efa758e' },
            ]);
        });
    });
});

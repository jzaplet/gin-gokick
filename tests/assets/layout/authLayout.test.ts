import { enableAutoUnmount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { router } from '@/router';
import { forgetSession, signedIn } from '@/shared/Auth/Session/currentSession';
import { t } from '@/shared/I18n/Texts/translate';
import { clearToasts } from '@/shared/Toast/Queue/toastQueue';
import { answerApi, calls, jan, openSignedIn } from '../app/openApp';
import { json } from '../fetch/fixtures';
import { offerEnglish } from '../i18n/languages';
import { toastTexts } from '../toast/toasts';

enableAutoUnmount(afterEach);

const internal = (): Response => json({ general: { key: 'request.internal' } }, 500);

const chooseEnglish = async (wrapper: VueWrapper): Promise<void> => {
    await wrapper.get(`button[aria-label="${t('language.change')}"]`).trigger('click');
    await wrapper.get('a[href="/en/app/dashboard"]').trigger('click');
    await vi.waitFor(() => {
        expect(calls('/api/user/locale')).toHaveLength(1);
    });
    await flushPromises();
};

describe('the auth layout', () => {
    afterEach(forgetSession);
    afterEach(clearToasts);

    it('stores the chosen language of the user before it shows the dashboard in it', async () => {
        offerEnglish();
        const wrapper = await openSignedIn('/cs/app/dashboard', {
            '/api/user/locale': () => new Response(null, { status: 204 }),
        });
        const czech = t('dashboard.title');

        await chooseEnglish(wrapper);
        const [, init] = calls('/api/user/locale')[0] ?? [];

        expect(init?.method).toBe('PUT');
        expect(init?.body).toBe('{"locale":"en_US"}');
        expect(router.currentRoute.value.fullPath).toBe('/en/app/dashboard');
        expect(wrapper.get('h1').text()).toBe(t('dashboard.title'));
        expect(t('dashboard.title')).not.toBe(czech);
    });

    it('says in a toast why the language was not stored and stays', async () => {
        offerEnglish();
        const wrapper = await openSignedIn('/cs/app/dashboard', { '/api/user/locale': internal });
        const error = [
            t('toast.error_title'),
            t('request.internal'),
        ];
        const title = t('dashboard.title');

        await chooseEnglish(wrapper);

        expect(toastTexts(wrapper)).toEqual([error]);
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/dashboard');
        expect(wrapper.get('h1').text()).toBe(title);
    });

    it('sends a signed-out user to the login and back', async () => {
        await openSignedIn('/cs/app/dashboard', {
            '/api/user/me': () => json({ general: { key: 'auth.sign_in_required' } }, 401),
        });

        expect(router.currentRoute.value.name).toBe('login');
        expect(router.currentRoute.value.query).toEqual({ redirect: '/dashboard' });
        expect(signedIn()).toBe(false);
    });

    it('shows a spinner until the user loads and only then the frame with the screen', async () => {
        const pending = Promise.withResolvers<Response>();
        const wrapper = await openSignedIn('/cs/app/dashboard', { '/api/user/me': async () => pending.promise });

        expect(wrapper.get('[role="status"]').text()).toBe(t('app.loading'));
        expect(wrapper.find('main').exists()).toBe(false);

        pending.resolve(json(jan, 200));
        await flushPromises();

        expect(wrapper.find('[role="status"]').exists()).toBe(false);
        expect(wrapper.get('main').text()).toContain(jan.email);
        expect(wrapper.get('h1').text()).toBe(t('dashboard.title'));
    });

    it('shows why the user did not load and tries again', async () => {
        const wrapper = await openSignedIn('/cs/app/dashboard', { '/api/user/me': internal });

        expect(wrapper.find('main').exists()).toBe(false);
        expect(wrapper.get('[role="alert"]').text()).toBe(t('request.internal'));
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/dashboard');

        answerApi();
        await wrapper.get('button').trigger('click');
        await flushPromises();

        expect(wrapper.find('[role="alert"]').exists()).toBe(false);
        expect(wrapper.get('main').text()).toContain(jan.email);
    });

    it('says in a toast why the session did not end and stays', async () => {
        const wrapper = await openSignedIn('/cs/app/dashboard', { '/api/auth/logout': internal });
        const signOut = wrapper.get(`button[aria-label="${t('account.logout')}"]`);

        expect(signOut.text()).toBe('');

        await signOut.trigger('click');
        await flushPromises();

        expect(toastTexts(wrapper)).toEqual([[
            t('toast.error_title'),
            t('request.internal'),
        ]]);
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/dashboard');
    });
});

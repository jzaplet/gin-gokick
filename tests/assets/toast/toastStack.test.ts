import { mount, type VueWrapper } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { nextTick } from 'vue';
import { t, tm } from '@/shared/I18n/Texts/translate';
import { clearToasts, showToast } from '@/shared/Toast/Queue/toastQueue';
import ToastStack from '@/shared/Toast/ToastStack.vue';
import { toastTexts } from './toasts';

const signedIn = (): string[] => [
    t('login.signed_in_title'),
    t('login.signed_in'),
];

const signedOut = (): string[] => [
    t('account.signed_out_title'),
    t('account.signed_out'),
];

const stack = (): VueWrapper => mount(ToastStack);

describe('the toast stack', () => {
    afterEach(() => {
        clearToasts();
        vi.useRealTimers();
    });

    it('stays mounted as a polite live region, also without a toast', () => {
        const wrapper = stack();

        expect(wrapper.find('[aria-live="polite"]').exists()).toBe(true);
        expect(toastTexts(wrapper)).toEqual([]);
    });

    it('shows a title over the text of a key or of a message from the API, the newest last', async () => {
        const wrapper = stack();

        showToast('success', 'login.signed_in_title', 'login.signed_in');
        showToast('error', 'toast.error_title', { key: 'request.internal' });
        await nextTick();

        expect(toastTexts(wrapper)).toEqual([signedIn(), [
            t('toast.error_title'),
            tm({ key: 'request.internal' }),
        ]]);
    });

    it(
        'closes a toast after 5 s, but holds one under the pointer and gives it 5 s again once the pointer leaves',
        async () => {
            vi.useFakeTimers();
            const wrapper = stack();

            showToast('success', 'login.signed_in_title', 'login.signed_in');
            showToast('success', 'account.signed_out_title', 'account.signed_out');
            await nextTick();
            vi.advanceTimersByTime(4000);
            await wrapper.get('li').trigger('mouseenter');
            vi.advanceTimersByTime(999);
            await nextTick();

            expect(toastTexts(wrapper)).toEqual([
                signedIn(),
                signedOut(),
            ]);

            vi.advanceTimersByTime(1);
            await nextTick();

            expect(toastTexts(wrapper)).toEqual([signedIn()]);

            vi.advanceTimersByTime(10000);
            await nextTick();

            expect(toastTexts(wrapper)).toEqual([signedIn()]);

            await wrapper.get('li').trigger('mouseleave');
            vi.advanceTimersByTime(4999);
            await nextTick();

            expect(toastTexts(wrapper)).toEqual([signedIn()]);

            vi.advanceTimersByTime(1);
            await nextTick();

            expect(toastTexts(wrapper)).toEqual([]);
        },
    );

    it('holds a toast while the focus is inside it, also after the pointer left', async () => {
        vi.useFakeTimers();
        const wrapper = stack();

        showToast('success', 'login.signed_in_title', 'login.signed_in');
        await nextTick();
        await wrapper.get('li').trigger('focusin');
        await wrapper.get('li').trigger('mouseenter');
        await wrapper.get('li').trigger('mouseleave');
        vi.advanceTimersByTime(10000);
        await nextTick();

        expect(toastTexts(wrapper)).toEqual([signedIn()]);

        await wrapper.get('li').trigger('focusout');
        vi.advanceTimersByTime(5000);
        await nextTick();

        expect(toastTexts(wrapper)).toEqual([]);
    });

    it('closes only the toast whose button was clicked', async () => {
        const wrapper = stack();

        showToast('success', 'login.signed_in_title', 'login.signed_in');
        showToast('success', 'account.signed_out_title', 'account.signed_out');
        await nextTick();
        await wrapper.get(`li button[aria-label="${t('toast.close')}"]`).trigger('click');

        expect(toastTexts(wrapper)).toEqual([signedOut()]);
    });
});

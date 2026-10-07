import { type DOMWrapper, enableAutoUnmount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { router } from '@/router';
import { forgetSession } from '@/shared/Auth/Session/currentSession';
import { appStorage } from '@/shared/Storage/appStorage';
import { clearToasts } from '@/shared/Toast/Queue/toastQueue';
import { t } from '@/shared/I18n/Texts/translate';
import { json } from '../fetch/fixtures';
import { toastTexts } from '../toast/toasts';
import { firefox, open, page, requests, rows, session } from './sessionsPage';

enableAutoUnmount(afterEach);

const filterToggle = (wrapper: VueWrapper): Omit<DOMWrapper<HTMLButtonElement>, 'exists'> =>
    wrapper.get<HTMLButtonElement>('button[aria-controls]');

const clearFilters = async (wrapper: VueWrapper): Promise<void> => {
    const button = wrapper.findAll('button').find((found) => found.text() === t('filters.clear'));

    expect(button?.attributes('disabled')).toBeUndefined();
    await button?.trigger('click');
    await flushPromises();
};

const ipFilter = (wrapper: VueWrapper): Omit<DOMWrapper<HTMLInputElement>, 'exists'> =>
    wrapper.get<HTMLInputElement>('input[name="sessionsIp"]');

const type = async (wrapper: VueWrapper, value: string, pause: number): Promise<void> => {
    await ipFilter(wrapper).setValue(value);
    await vi.advanceTimersByTimeAsync(pause);
    await flushPromises();
};

const filtered = '/api/auth/sessions?page=1&perPage=25&sortBy=lastSeenAt&sortDir=desc&ip=203.0';

const unfiltered = '/api/auth/sessions?page=1&perPage=25&sortBy=lastSeenAt&sortDir=desc';

describe('the filter of the sessions', () => {
    afterEach(forgetSession);
    afterEach(clearToasts);
    afterEach(() => {
        appStorage().remove('filterPanel.sessions');
        vi.useRealTimers();
    });

    it('marks its toggle as soon as a field holds more than spaces, before the filter applies', async () => {
        vi.useFakeTimers({
            toFake: [
                'setTimeout',
                'clearTimeout',
            ],
        });
        const wrapper = await open(() => page([session('a', firefox, '198.51.100.4', true)]));

        await filterToggle(wrapper).trigger('click');
        await type(wrapper, '  ', 0);

        expect(filterToggle(wrapper).attributes('aria-label')).toBeUndefined();

        await type(wrapper, '2', 0);

        expect(filterToggle(wrapper).attributes('aria-label')).toBe(t('filters.active'));
        expect(requests()).toHaveLength(1);

        await clearFilters(wrapper);
        await vi.advanceTimersByTimeAsync(400);

        expect(ipFilter(wrapper).element.value).toBe('');
        expect(filterToggle(wrapper).attributes('aria-label')).toBeUndefined();
        expect(requests()).toHaveLength(1);
    });

    it('filters once the typing stops, from the first page, and keeps the filter in the address', async () => {
        vi.useFakeTimers({
            toFake: [
                'setTimeout',
                'clearTimeout',
            ],
        });
        let response = page([session('a', firefox, '198.51.100.4', true)], 60);
        const wrapper = await open(() => response, '/cs/app/dashboard?sessionsPage=2');

        document.title = 'Rozepsáno';
        await filterToggle(wrapper).trigger('click');
        response = page([]);
        await type(wrapper, '20', 200);
        await type(wrapper, ' 203.0 ', 399);

        expect(requests()).toHaveLength(1);

        await type(wrapper, ' 203.0 ', 1);

        expect(requests().slice(1)).toEqual([filtered]);
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/dashboard?sessionsIp=203.0');
        expect(document.title).toBe('Rozepsáno');
        expect(filterToggle(wrapper).attributes('aria-label')).toBe(t('filters.active'));
        expect(rows(wrapper)).toEqual([t('grid.no_match')]);

        await type(wrapper, '203.0  ', 400);

        expect(requests()).toHaveLength(2);
    });

    it('opens its panel for the filter of its address and clears the filter at once', async () => {
        const wrapper = await open(
            () => page([session('a', firefox, '203.0.113.7', true)]),
            '/cs/app/dashboard?ref=newsletter&sessionsIp=203.0',
        );

        expect(requests()).toEqual([filtered]);
        expect(filterToggle(wrapper).attributes('aria-expanded')).toBe('true');
        expect(ipFilter(wrapper).element.value).toBe('203.0');

        await clearFilters(wrapper);

        expect(requests().slice(1)).toEqual([unfiltered]);
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/dashboard?ref=newsletter');
        expect(ipFilter(wrapper).element.value).toBe('');
        expect(filterToggle(wrapper).attributes('aria-label')).toBeUndefined();
    });

    it('ignores a filter of its address longer than an address and follows a new one', async () => {
        const wrapper = await open(
            () => page([session('a', firefox, null, true)]),
            `/cs/app/dashboard?sessionsIp=${'1'.repeat(46)}`,
        );

        expect(requests()).toEqual([unfiltered]);
        expect(ipFilter(wrapper).element.value).toBe('');

        await router.push('/cs/app/dashboard?sessionsIp=203.0');
        await flushPromises();

        expect(requests().slice(1)).toEqual([filtered]);
        expect(ipFilter(wrapper).element.value).toBe('203.0');
    });

    it(
        'keeps the typed filter, still marked and clearable, but shows the rows it has when they do not load',
        async () => {
            vi.useFakeTimers({
                toFake: [
                    'setTimeout',
                    'clearTimeout',
                ],
            });
            let response = page([session('a', firefox, '198.51.100.4', true)]);
            const wrapper = await open(() => response);

            await filterToggle(wrapper).trigger('click');
            response = json({ general: { key: 'request.internal' } }, 500);
            await type(wrapper, '203.0', 400);

            expect(requests().slice(1)).toEqual([filtered]);
            expect(toastTexts(wrapper)).toHaveLength(1);
            expect(ipFilter(wrapper).element.value).toBe('203.0');
            expect(rows(wrapper)[0]).toContain('198.51.100.4');
            expect(router.currentRoute.value.fullPath).toBe('/cs/app/dashboard');
            expect(filterToggle(wrapper).attributes('aria-label')).toBe(t('filters.active'));

            await clearFilters(wrapper);

            expect(ipFilter(wrapper).element.value).toBe('');
            expect(requests()).toHaveLength(2);
        },
    );

    it('drops a filter still typed when the dashboard closes', async () => {
        vi.useFakeTimers({
            toFake: [
                'setTimeout',
                'clearTimeout',
            ],
        });
        const wrapper = await open(() => page([session('a', firefox, null, true)]));

        await filterToggle(wrapper).trigger('click');
        await type(wrapper, '203.0', 100);
        await router.push('/cs/app/nothing');
        await vi.advanceTimersByTimeAsync(400);
        await flushPromises();

        expect(requests()).toHaveLength(1);
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/nothing');
    });
});

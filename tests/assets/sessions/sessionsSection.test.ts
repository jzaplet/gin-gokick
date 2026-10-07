import { enableAutoUnmount, flushPromises } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { router } from '@/router';
import { forgetSession } from '@/shared/Auth/Session/currentSession';
import { clearToasts } from '@/shared/Toast/Queue/toastQueue';
import { t } from '@/shared/I18n/Texts/translate';
import { json } from '../fetch/fixtures';
import { toastTexts } from '../toast/toasts';
import {
    click,
    firefox,
    header,
    open,
    page,
    requests,
    rows,
    safari,
    session,
    sortBy,
    sortButton,
} from './sessionsPage';

enableAutoUnmount(afterEach);

describe('the sessions on the dashboard', () => {
    afterEach(forgetSession);
    afterEach(clearToasts);

    it('asks for the first page of the last used sessions and names the device of each', async () => {
        const wrapper = await open(() => page([
            session('a', firefox, '198.51.100.4', true),
            session('b', safari, null),
            session('c', '', '2001:db8::1'),
        ]));

        expect(requests()).toEqual(['/api/auth/sessions?page=1&perPage=25&sortBy=lastSeenAt&sortDir=desc']);
        expect(header(wrapper, t('sessions.last_seen_at'))).toBe('descending');
        expect(rows(wrapper)[0]).toContain('Firefox · macOS');
        expect(rows(wrapper)[0]).toContain(t('sessions.current'));
        expect(rows(wrapper)[1]).toContain('Safari · iOS');
        expect(rows(wrapper)[1]).not.toContain(t('sessions.current'));
        expect(rows(wrapper)[2]).toContain(t('sessions.unknown_device'));
    });

    it('sorts by the column of a header and turns the direction on the next click', async () => {
        const wrapper = await open(() => page([session('a', firefox, null, true)]));

        await sortBy(wrapper, t('sessions.created_at'));
        expect(header(wrapper, t('sessions.created_at'))).toBe('ascending');
        await sortBy(wrapper, t('sessions.created_at'));

        expect(requests().slice(1)).toEqual([
            '/api/auth/sessions?page=1&perPage=25&sortBy=createdAt&sortDir=asc',
            '/api/auth/sessions?page=1&perPage=25&sortBy=createdAt&sortDir=desc',
        ]);
        expect(header(wrapper, t('sessions.created_at'))).toBe('descending');
        expect(header(wrapper, t('sessions.last_seen_at'))).toBeUndefined();
    });

    it('goes back to the last page when the page it asked for is gone', async () => {
        let total = 60;
        const wrapper = await open(() => page([session('a', firefox, null, true)], total));

        total = 30;
        await click(wrapper, `button[aria-label="${t('grid.page', { page: '3' })}"]`);

        expect(requests().slice(1)).toEqual([
            '/api/auth/sessions?page=3&perPage=25&sortBy=lastSeenAt&sortDir=desc',
            '/api/auth/sessions?page=2&perPage=25&sortBy=lastSeenAt&sortDir=desc',
        ]);
        expect(wrapper.text()).toContain(
            t('grid.range', {
                from: '26',
                to: '30',
                total: '30',
            }),
        );
    });

    it('keeps the rows of the last request when an older one answers later', async () => {
        const wrapper = await open(() => page([session('a', firefox, null, true)]));
        const answers: ((response: Response) => void)[] = [];

        vi.mocked(fetch).mockImplementation(async () => new Promise((resolve) => {
            answers.push(resolve);
        }));
        await sortButton(wrapper, t('sessions.created_at')).trigger('click');
        await sortButton(wrapper, t('sessions.created_at')).trigger('click');
        await vi.waitFor(() => {
            expect(answers).toHaveLength(2);
        });
        answers[1]?.(page([session('b', safari, null)]));
        await flushPromises();
        answers[0]?.(page([session('c', '', null)]));
        await flushPromises();

        expect(rows(wrapper)).toHaveLength(1);
        expect(rows(wrapper)[0]).toContain('Safari · iOS');
        expect(header(wrapper, t('sessions.created_at'))).toBe('descending');
    });

    it('stays on the page and the sort it shows when the next ones do not load', async () => {
        let response = page([session('a', firefox, null, true)], 60);
        const wrapper = await open(() => response);

        response = json({ general: { key: 'request.internal' } }, 500);
        await click(wrapper, `button[aria-label="${t('grid.page', { page: '2' })}"]`);
        response = json({ general: { key: 'request.internal' } }, 500);
        await sortBy(wrapper, t('sessions.created_at'));

        expect(wrapper.get('button[aria-current="page"]').text()).toBe('1');
        expect(wrapper.text()).toContain(
            t('grid.range', {
                from: '1',
                to: '25',
                total: '60',
            }),
        );
        expect(header(wrapper, t('sessions.last_seen_at'))).toBe('descending');
        expect(header(wrapper, t('sessions.created_at'))).toBeUndefined();
        expect(toastTexts(wrapper)).toHaveLength(2);
    });

    it('keeps its page and sort in the address under its own name, without the defaults', async () => {
        const wrapper = await open(
            () => page([session('a', firefox, null, true)], 60),
            '/cs/app/dashboard?ref=newsletter',
        );

        await sortBy(wrapper, t('sessions.created_at'));
        await click(wrapper, `button[aria-label="${t('grid.page', { page: '2' })}"]`);

        expect(router.currentRoute.value.fullPath)
            .toBe('/cs/app/dashboard?ref=newsletter&sessionsPage=2&sessionsSortBy=createdAt&sessionsSortDir=asc');

        await sortBy(wrapper, t('sessions.last_seen_at'));
        await sortBy(wrapper, t('sessions.last_seen_at'));

        expect(router.currentRoute.value.fullPath).toBe('/cs/app/dashboard?ref=newsletter');
        expect(requests().at(-1)).toBe('/api/auth/sessions?page=1&perPage=25&sortBy=lastSeenAt&sortDir=desc');
    });

    it('opens the page and the sort of its address and ignores values it does not know', async () => {
        const wrapper = await open(
            () => page([session('a', firefox, null, true)], 60),
            '/cs/app/dashboard?sessionsPage=2&sessionsSortBy=createdAt&sessionsSortDir=asc',
        );

        expect(requests()).toEqual(['/api/auth/sessions?page=2&perPage=25&sortBy=createdAt&sortDir=asc']);
        expect(wrapper.get('button[aria-current="page"]').text()).toBe('2');

        await router.push('/cs/app/dashboard?sessionsPage=0&sessionsSortBy=email&sessionsSortDir=up');
        await flushPromises();

        expect(requests().at(-1)).toBe('/api/auth/sessions?page=1&perPage=25&sortBy=lastSeenAt&sortDir=desc');
        expect(wrapper.get('button[aria-current="page"]').text()).toBe('1');
        expect(header(wrapper, t('sessions.last_seen_at'))).toBe('descending');
    });

    it('says when the sessions do not load and loads them again on request', async () => {
        let response = json({ general: { key: 'request.internal' } }, 500);
        const wrapper = await open(() => response);

        expect(toastTexts(wrapper)).toEqual([[
            t('toast.error_title'),
            t('request.internal'),
        ]]);
        expect(rows(wrapper)[0]).toContain(t('grid.failed'));

        response = page([]);
        await click(wrapper, 'tbody button');

        expect(rows(wrapper)).toEqual([t('sessions.empty')]);
        expect(wrapper.text()).not.toContain(
            t('grid.range', {
                from: '1',
                to: '0',
                total: '0',
            }),
        );
    });
});

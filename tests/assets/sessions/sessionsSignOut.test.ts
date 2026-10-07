import { type DOMWrapper, enableAutoUnmount, flushPromises, type VueWrapper } from '@vue/test-utils';
import { afterEach, describe, expect, it, vi } from 'vitest';
import { router } from '@/router';
import { forgetSession } from '@/shared/Auth/Session/currentSession';
import { clearToasts } from '@/shared/Toast/Queue/toastQueue';
import { t } from '@/shared/I18n/Texts/translate';
import { json } from '../fetch/fixtures';
import { toastTexts } from '../toast/toasts';
import { click, firefox, open, page, requests, safari, session, signOuts } from './sessionsPage';

enableAutoUnmount(afterEach);

const sessions = [
    session('a', firefox, '203.0.113.7', true),
    session('b', safari, '203.0.113.9'),
    session('c', '', '198.51.100.4'),
];

const server = (ended: () => Response, listed = (): Response => page(sessions, 30, 29)) => (url: string): Response =>
    url === '/api/auth/sessions/end' ? ended() : listed();

const checkboxes = (wrapper: VueWrapper): DOMWrapper<HTMLInputElement>[] =>
    wrapper.findAll<HTMLInputElement>('tbody input[type="checkbox"]');

const toggle = async (box: Pick<DOMWrapper<HTMLInputElement>, 'element' | 'setValue'> | undefined): Promise<void> => {
    await box?.setValue(box.element.checked === false);
    await flushPromises();
};

const toggleRow = async (wrapper: VueWrapper, row: number): Promise<void> => {
    await toggle(checkboxes(wrapper)[row]);
};

const togglePage = async (wrapper: VueWrapper): Promise<void> => {
    await toggle(wrapper.get<HTMLInputElement>('thead input[type="checkbox"]'));
};

const selected = (wrapper: VueWrapper): string => wrapper.get('p[aria-live]').text();

const selection = (count: string, total: string): string => t('bulk.selected', {
    count,
    total,
});

const signOutLabel = (count: string): string => t('bulk.action', {
    action: t('sessions.sign_out'),
    count,
});

const button = (wrapper: VueWrapper, text: string, within = ''): DOMWrapper<HTMLButtonElement> | undefined =>
    wrapper.findAll<HTMLButtonElement>(`${within} button`.trim()).find((found) => found.text() === text);

const pressButton = async (wrapper: VueWrapper, text: string, within = ''): Promise<void> => {
    await button(wrapper, text, within)?.trigger('click');
    await flushPromises();
};

const bulkAction = async (wrapper: VueWrapper, label: string): Promise<void> => {
    await click(wrapper, `button[aria-label="${t('bulk.actions')}"]`);
    await pressButton(wrapper, label);
};

const question = (wrapper: VueWrapper): string | null =>
    wrapper.get<HTMLDialogElement>('dialog').element.open ? wrapper.get('dialog').text() : null;

const confirm = async (wrapper: VueWrapper): Promise<void> => {
    await pressButton(wrapper, t('sessions.sign_out'), 'dialog');
};

describe('the sign-out of the sessions', () => {
    afterEach(forgetSession);
    afterEach(clearToasts);
    afterEach(() => {
        vi.useRealTimers();
    });

    it('selects only the other devices, also a whole page at once, and says how many it holds', async () => {
        const wrapper = await open(server(() => json({ ended: 0 }, 200)));

        expect(checkboxes(wrapper).map((box) => box.element.disabled)).toEqual([
            true,
            false,
            false,
        ]);
        expect(selected(wrapper)).toBe('');

        await toggleRow(wrapper, 1);

        expect(selected(wrapper)).toBe(selection('1', '29'));
        expect(wrapper.get<HTMLInputElement>('thead input').element.checked).toBe(false);

        await togglePage(wrapper);

        expect(checkboxes(wrapper).map((box) => box.element.checked)).toEqual([
            false,
            true,
            true,
        ]);
        expect(selected(wrapper)).toBe(selection('2', '29'));
        expect(button(wrapper, t('bulk.select_all', { total: '29' }))).toBeDefined();

        await togglePage(wrapper);

        expect(selected(wrapper)).toBe('');
    });

    it('signs out the chosen devices after asking and says how many the server signed out', async () => {
        const wrapper = await open(server(() => json({ ended: 2 }, 200)));

        await toggleRow(wrapper, 1);
        await toggleRow(wrapper, 2);
        await bulkAction(wrapper, signOutLabel('2'));

        expect(question(wrapper)).toContain(t('sessions.sign_out_selected', { count: 2 }));

        await confirm(wrapper);

        expect(signOuts()).toEqual([{
            ids: [
                'b',
                'c',
            ],
            all: false,
            ip: '',
        }]);
        expect(question(wrapper)).toBeNull();
        expect(toastTexts(wrapper)).toEqual([[
            t('sessions.signed_out_title'),
            t('sessions.signed_out', { count: 2 }),
        ]]);
        expect(selected(wrapper)).toBe('');
        expect(requests()).toEqual([
            '/api/auth/sessions?page=1&perPage=25&sortBy=lastSeenAt&sortDir=desc',
            '/api/auth/sessions?page=1&perPage=25&sortBy=lastSeenAt&sortDir=desc',
        ]);
    });

    it('selects every device the shown filter lists but this one, and signs them out by the filter', async () => {
        const wrapper = await open(server(() => json({ ended: 29 }, 200)), '/cs/app/dashboard?sessionsIp=203.0');

        await toggleRow(wrapper, 1);
        await pressButton(wrapper, t('bulk.select_all', { total: '29' }));

        expect(selected(wrapper)).toBe(t('bulk.selected_all', { total: '29' }));
        expect(checkboxes(wrapper).map((box) => box.element.checked)).toEqual([
            false,
            true,
            true,
        ]);

        await bulkAction(wrapper, signOutLabel('29'));

        expect(question(wrapper)).toContain(t('sessions.sign_out_selected', { count: 29 }));

        await confirm(wrapper);

        expect(signOuts()).toEqual([{
            ids: [],
            all: true,
            ip: '203.0',
        }]);
        expect(toastTexts(wrapper)[0]).toEqual([
            t('sessions.signed_out_title'),
            t('sessions.signed_out', { count: 29 }),
        ]);
    });

    it('counts every row as selectable when the grid answers no count of its own', async () => {
        const wrapper = await open(server(() => json({ ended: 0 }, 200), () => page(sessions, 30)));

        await toggleRow(wrapper, 1);

        expect(selected(wrapper)).toBe(selection('1', '30'));
        expect(button(wrapper, t('bulk.select_all', { total: '30' }))).toBeDefined();
    });

    it('drops the selection as soon as a typed filter differs, and again when it applies', async () => {
        vi.useFakeTimers({
            toFake: [
                'setTimeout',
                'clearTimeout',
            ],
        });
        const wrapper = await open(server(() => json({ ended: 0 }, 200)));

        await toggleRow(wrapper, 1);
        await wrapper.get('input[name="sessionsIp"]').setValue('  ');

        expect(selected(wrapper)).toBe(selection('1', '29'));

        await wrapper.get('input[name="sessionsIp"]').setValue('203');

        expect(selected(wrapper)).toBe('');

        await toggleRow(wrapper, 1);
        await vi.advanceTimersByTimeAsync(400);
        await flushPromises();

        expect(selected(wrapper)).toBe('');
    });

    it('keeps the selection across pages, also when the address turns the page', async () => {
        const wrapper = await open(server(() => json({ ended: 0 }, 200)));

        await toggleRow(wrapper, 1);
        await click(wrapper, `button[aria-label="${t('grid.page', { page: '2' })}"]`);
        await router.replace('/cs/app/dashboard?sessionsPage=1');
        await flushPromises();

        expect(selected(wrapper)).toBe(selection('1', '29'));
    });

    it('drops a selection of everything as soon as one row is toggled', async () => {
        const wrapper = await open(server(() => json({ ended: 0 }, 200)));

        await toggleRow(wrapper, 1);
        await pressButton(wrapper, t('bulk.select_all', { total: '29' }));
        await toggleRow(wrapper, 2);

        expect(selected(wrapper)).toBe('');
        expect(checkboxes(wrapper).map((box) => box.element.checked)).toEqual([
            false,
            false,
            false,
        ]);
    });

    it('signs out one device from its row and keeps the rest of the selection', async () => {
        const wrapper = await open(server(() => json({ ended: 1 }, 200)));

        await toggleRow(wrapper, 1);
        await toggleRow(wrapper, 2);
        await click(wrapper, 'tbody tr:nth-child(2) button');

        expect(question(wrapper)).toContain(t('sessions.sign_out_one', { device: 'Safari · iOS' }));

        await confirm(wrapper);

        expect(signOuts()).toEqual([{
            ids: ['b'],
            all: false,
            ip: '',
        }]);
        expect(selected(wrapper)).toBe(selection('1', '29'));
    });

    it('keeps the button of this device disabled and says how to sign it out', async () => {
        const wrapper = await open(server(() => json({ ended: 0 }, 200)));
        const current = wrapper.get<HTMLButtonElement>('tbody tr:first-child button');

        expect(wrapper.findAll('th .sr-only').map((label) => label.text())).toContain(t('grid.actions'));
        expect(
            wrapper.get('tbody tr:nth-child(2) button').attributes('aria-label'),
        ).toBe(t('sessions.sign_out_device'));

        expect(current.element.disabled).toBe(true);
        expect(current.attributes('aria-label')).toBe(t('sessions.sign_out_current'));
    });

    it('says when it signed out no device, and why the sign-out failed, keeping the selection', async () => {
        let response = json({ ended: 0 }, 200);
        const wrapper = await open(server(() => response));

        await toggleRow(wrapper, 1);
        await bulkAction(wrapper, signOutLabel('1'));
        await confirm(wrapper);

        expect(toastTexts(wrapper)[0]).toEqual([
            t('sessions.signed_out_title'),
            t('sessions.signed_out_none'),
        ]);

        response = json({ general: { key: 'request.internal' } }, 500);
        await toggleRow(wrapper, 1);
        await bulkAction(wrapper, signOutLabel('1'));
        await confirm(wrapper);

        expect(toastTexts(wrapper)[1]?.[0]).toBe(t('toast.error_title'));
        expect(selected(wrapper)).toBe(selection('1', '29'));

        response = json({ ip: { key: 'validation.max_length' } }, 422);
        await bulkAction(wrapper, signOutLabel('1'));
        await confirm(wrapper);

        expect(toastTexts(wrapper)[2]).toEqual([
            t('toast.error_title'),
            t('sessions.sign_out_failed'),
        ]);
    });

    it('sends nothing when cancelled and stops asking once its selection is gone', async () => {
        const wrapper = await open(server(() => json({ ended: 0 }, 200)));

        await toggleRow(wrapper, 1);
        await bulkAction(wrapper, signOutLabel('1'));
        await pressButton(wrapper, t('modal.cancel'), 'dialog');

        expect(question(wrapper)).toBeNull();

        await bulkAction(wrapper, signOutLabel('1'));
        await router.replace('/cs/app/dashboard?sessionsIp=198');
        await flushPromises();

        expect(selected(wrapper)).toBe('');
        expect(question(wrapper)).toBeNull();
        expect(signOuts()).toEqual([]);
    });
});

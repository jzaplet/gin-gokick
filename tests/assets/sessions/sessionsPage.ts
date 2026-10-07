import { type DOMWrapper, flushPromises, type VueWrapper } from '@vue/test-utils';
import { type Answer, calls, openSignedIn } from '../app/openApp';
import { json } from '../fetch/fixtures';

type Session = {
    id: string;
    userAgent: string;
    createdAt: string;
    lastSeenAt: string;
    lastSeenIp: string | null;
    current: boolean;
};

export const firefox = 'Mozilla/5.0 (Macintosh; Intel Mac OS X 14.6; rv:131.0) Gecko/20100101 Firefox/131.0';

export const safari = 'Mozilla/5.0 (iPhone; CPU iPhone OS 18_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) '
    + 'Version/18.0 Mobile/15E148 Safari/604.1';

export const session = (id: string, userAgent: string, lastSeenIp: string | null, current = false): Session => ({
    id,
    userAgent,
    createdAt: '2026-10-01T08:30:00Z',
    lastSeenAt: '2026-10-05T17:45:00Z',
    lastSeenIp,
    current,
});

export const page = (items: Session[], total = items.length, selectable?: number): Response => json({
    items,
    total,
    ...(selectable === undefined ? {} : { selectable }),
}, 200);

export const open = async (sessions: Answer, path = '/cs/app/dashboard'): Promise<VueWrapper> => openSignedIn(path, {
    '/api/auth/sessions': sessions,
    '/api/auth/sessions/end': sessions,
});

export const requests = (): string[] => calls('/api/auth/sessions')
    .map(([input]) => (typeof input === 'string' ? input : ''));

export const signOuts = (): unknown[] => calls('/api/auth/sessions/end')
    .map(([, init]): unknown => (typeof init?.body === 'string' ? JSON.parse(init.body) : undefined));

export const rows = (wrapper: VueWrapper): string[] => wrapper.findAll('tbody tr').map((row) => row.text());

export const click = async (wrapper: VueWrapper, selector: string): Promise<void> => {
    await wrapper.get(selector).trigger('click');
    await flushPromises();
};

export const header = (wrapper: VueWrapper, name: string): string | undefined =>
    wrapper.findAll('th').find((th) => th.text() === name)?.attributes('aria-sort');

export const sortButton = (wrapper: VueWrapper, name: string): Omit<DOMWrapper<HTMLButtonElement>, 'exists'> => {
    const column = wrapper.findAll('th').find((th) => th.text() === name);

    if (column === undefined) {
        throw new Error(`no column ${name}`);
    }

    return column.get('button');
};

export const sortBy = async (wrapper: VueWrapper, name: string): Promise<void> => {
    await sortButton(wrapper, name).trigger('click');
    await flushPromises();
};

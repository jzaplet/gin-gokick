import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { startConstellation } from '@/shared/Constellation/Start/startConstellation';

let stop = (): void => undefined;
let contextRequests = (): number => 0;

const motion = (reduced: boolean): void => {
    vi.stubGlobal('matchMedia', (media: string) => ({
        media,
        matches: reduced,
    }));
};

const addHost = (): HTMLElement => {
    const host = document.createElement('div');

    host.dataset['constellation'] = '';
    document.body.append(host);

    return host;
};

const observed = async (): Promise<void> => {
    await Promise.resolve();
};

beforeEach(() => {
    vi.useFakeTimers();
    const getContext = vi.spyOn(HTMLCanvasElement.prototype, 'getContext').mockReturnValue(null);

    contextRequests = () => getContext.mock.calls.length;
});

afterEach(() => {
    stop();
    document.body.replaceChildren();
    vi.useRealTimers();
});

describe('the constellation entry', () => {
    it('puts a canvas hidden from screen readers into every element with data-constellation', () => {
        motion(false);
        const hosts = [
            addHost(),
            addHost(),
        ];

        stop = startConstellation();

        const hidden = hosts.map((host) => host.querySelector('canvas')?.getAttribute('aria-hidden'));

        expect(hidden).toEqual([
            'true',
            'true',
        ]);
    });

    it('follows elements that Vue renders and removes later', async () => {
        motion(false);
        stop = startConstellation();
        const host = addHost();

        await observed();

        expect(host.querySelector('canvas')).not.toBeNull();

        host.remove();
        await observed();
        vi.advanceTimersByTime(5000);

        expect(contextRequests()).toBe(0);
        expect(host.querySelector('canvas')).toBeNull();
    });

    it('leaves the page alone for a visitor who asked for less motion', () => {
        motion(true);
        const host = addHost();

        stop = startConstellation();
        vi.advanceTimersByTime(5000);

        expect(host.children).toHaveLength(0);
        expect(contextRequests()).toBe(0);
    });
});

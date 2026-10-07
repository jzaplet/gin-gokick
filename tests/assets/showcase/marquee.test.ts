import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { startMarquee } from '@/shared/Showcase/Marquee/startMarquee';

const ITEM = 100;
const FRAME_MS = 50;

let stop = (): void => undefined;
let wide = true;
let reduced = false;
let frameId = 0;
const queued = new Map<number, FrameRequestCallback>();
const changes = new Set<() => void>();
const offsets = [
    'offsetTop',
    'offsetLeft',
].map((name) => [
    name,
    Object.getOwnPropertyDescriptor(HTMLElement.prototype, name),
] as const);

const screen = (): void => {
    vi.stubGlobal('matchMedia', (media: string) => {
        const addEventListener = (_: string, listener: () => void): void => {
            changes.add(listener);
        };

        const removeEventListener = (_: string, listener: () => void): void => {
            changes.delete(listener);
        };

        return {
            media,
            get matches() {
                return media.includes('reduce') ? reduced : wide;
            },
            addEventListener,
            removeEventListener,
        };
    });
};

const frames = (from: number, to: number): void => {
    for (let time = from; time <= to; time += FRAME_MS) {
        const due = [...queued.values()];

        queued.clear();

        for (const frame of due) {
            frame(time);
        }
    }
};

const positionInTrack = (element: HTMLElement): number => {
    const track = element.parentElement;

    if (track === null) {
        return 0;
    }

    return Number(track.parentElement?.dataset['start'] ?? 0) + [...track.children].indexOf(element) * ITEM;
};

const addMarquee = (items: number[]): HTMLElement => {
    const marquee = document.createElement('div');

    marquee.dataset['marquee'] = '';

    for (const count of items) {
        const scroller = document.createElement('div');
        const track = document.createElement('ol');
        const entries = Array.from({ length: count }, (_, index) => {
            const item = document.createElement('li');

            item.textContent = `item ${index}`;

            return item;
        });

        track.dataset['marqueeTrack'] = '';
        track.append(...entries);
        scroller.append(track);
        marquee.append(scroller);
    }

    document.body.append(marquee);

    return marquee;
};

const tracks = (marquee: HTMLElement): HTMLElement[] =>
    [...marquee.querySelectorAll<HTMLElement>('[data-marquee-track]')];

const scrollers = (marquee: HTMLElement): HTMLElement[] =>
    tracks(marquee).map((track) => track.parentElement ?? document.body);

const tops = (marquee: HTMLElement): number[] => scrollers(marquee).map((scroller) => Math.round(scroller.scrollTop));

beforeEach(() => {
    wide = true;
    reduced = false;
    screen();
    vi.stubGlobal('requestAnimationFrame', (frame: FrameRequestCallback) => {
        frameId += 1;
        queued.set(frameId, frame);

        return frameId;
    });
    vi.stubGlobal('cancelAnimationFrame', (id: number) => queued.delete(id));

    for (const [name] of offsets) {
        Object.defineProperty(HTMLElement.prototype, name, {
            configurable: true,
            get(this: HTMLElement) {
                return positionInTrack(this);
            },
        });
    }
});

afterEach(() => {
    stop();
    queued.clear();
    changes.clear();
    document.body.replaceChildren();

    for (const [name, descriptor] of offsets) {
        if (descriptor !== undefined) {
            Object.defineProperty(HTMLElement.prototype, name, descriptor);
        }
    }
});

describe('the marquee', () => {
    it('copies the items of every track once, hidden from screen readers and out of reach', () => {
        const marquee = addMarquee([
            2,
            3,
        ]);

        stop = startMarquee();

        expect(tracks(marquee).map((track) => track.childElementCount)).toEqual([
            4,
            6,
        ]);

        const items = [...(tracks(marquee)[0]?.children ?? [])];

        expect(items.map((item) => item.getAttribute('aria-hidden'))).toEqual([
            null,
            null,
            'true',
            'true',
        ]);
        expect(items.map((item) => item.hasAttribute('inert'))).toEqual([
            false,
            false,
            true,
            true,
        ]);
        expect(items.map((item) => item.textContent)).toEqual([
            'item 0',
            'item 1',
            'item 0',
            'item 1',
        ]);
    });

    it('leaves the page alone for a visitor who asked for less motion', () => {
        reduced = true;
        const marquee = addMarquee([
            2,
            2,
        ]);

        stop = startMarquee();
        frames(0, 1000);

        expect(marquee.dataset['moving']).toBeUndefined();
        expect(tracks(marquee).map((track) => track.childElementCount)).toEqual([
            2,
            2,
        ]);
        expect(tops(marquee)).toEqual([
            0,
            0,
        ]);
    });
});

import { prefersReducedMotion } from '@/shared/Motion/prefersReducedMotion';
import type { Axis } from '@/shared/Showcase/Marquee/Axis';
import { doubleTrack } from '@/shared/Showcase/Marquee/doubleTrack';
import { scrollTrack, type TrackScroll } from '@/shared/Showcase/Marquee/scrollTrack';

const MARQUEES = '[data-marquee]';
const TRACKS = '[data-marquee-track]';
const COLUMNS = '(min-width: 80rem)';
const MAX_STEP_MS = 50;

type MovingTrack = {
    scroll: TrackScroll;
    stop: () => void;
};

const moveTrack = (track: HTMLElement, axis: Axis, clock: () => number): MovingTrack | null => {
    const first = track.firstElementChild;
    const scroller = track.parentElement;

    if ((first instanceof HTMLElement) === false || scroller === null) {
        return null;
    }

    const copies = doubleTrack(track);
    const [copy] = copies;

    if (copy === undefined) {
        return null;
    }

    const scroll = scrollTrack(scroller, first, copy, axis);
    const follow = (): void => {
        scroll.follow(clock());
    };
    const hold = (): void => {
        scroll.hold(true);
    };
    const release = (): void => {
        scroll.hold(false);
    };

    scroller.addEventListener('scroll', follow, { passive: true });
    track.addEventListener('pointerenter', hold);
    track.addEventListener('pointerleave', release);

    const stop = (): void => {
        scroller.removeEventListener('scroll', follow);
        track.removeEventListener('pointerenter', hold);
        track.removeEventListener('pointerleave', release);

        for (const item of copies) {
            item.remove();
        }
    };

    return {
        scroll,
        stop,
    };
};

const runMarquee = (marquee: HTMLElement, columns: MediaQueryList): (() => void) => {
    const axis = (): Axis => (columns.matches ? 'y' : 'x');
    let now = 0;
    let last: number | undefined;
    const tracks = [...marquee.querySelectorAll<HTMLElement>(TRACKS)]
        .map((track) => moveTrack(track, axis(), () => now))
        .filter((track): track is MovingTrack => track !== null);

    const tick = (time: number): void => {
        const ms = last === undefined ? 0 : Math.min(MAX_STEP_MS, Math.max(0, time - last));

        now = time;
        last = time;

        for (const { scroll } of tracks) {
            scroll.step(ms, now);
        }

        frame = requestAnimationFrame(tick);
    };

    const turn = (): void => {
        for (const { scroll } of tracks) {
            scroll.turn(axis());
        }
    };

    let frame = requestAnimationFrame(tick);

    marquee.dataset['moving'] = '';
    columns.addEventListener('change', turn);

    return () => {
        cancelAnimationFrame(frame);
        columns.removeEventListener('change', turn);

        for (const { stop } of tracks) {
            stop();
        }

        delete marquee.dataset['moving'];
    };
};

export const startMarquee = (): (() => void) => {
    if (prefersReducedMotion()) {
        return () => undefined;
    }

    const columns = window.matchMedia(COLUMNS);
    const stops = [...document.querySelectorAll<HTMLElement>(MARQUEES)].map((marquee) => runMarquee(marquee, columns));

    return () => {
        for (const stop of stops) {
            stop();
        }
    };
};

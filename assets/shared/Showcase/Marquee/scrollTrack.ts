import type { Axis } from '@/shared/Showcase/Marquee/Axis';

const PX_PER_MS: Record<Axis, number> = {
    x: 0.04,
    y: 0.03,
};
const IDLE_MS = 1500;

export type TrackScroll = {
    step: (ms: number, now: number) => void;
    follow: (now: number) => void;
    hold: (held: boolean) => void;
    turn: (axis: Axis) => void;
};

const wrap = (value: number, base: number, period: number): number =>
    value - Math.ceil((value - base) / period - 1) * period;

export const scrollTrack = (scroller: HTMLElement, first: HTMLElement, copy: HTMLElement, start: Axis): TrackScroll => {
    let axis = start;
    let position = 0;
    let written = 0;
    let held = false;
    let idleUntil = 0;

    const offset = (): number => (axis === 'y' ? first.offsetTop : first.offsetLeft);

    const period = (): number => (axis === 'y' ? copy.offsetTop : copy.offsetLeft) - offset();

    const read = (): number => (axis === 'y' ? scroller.scrollTop : scroller.scrollLeft);

    const place = (value: number, byHand: boolean): void => {
        const length = period();
        const base = offset();

        if (length <= 0) {
            return;
        }

        position = byHand || value > base ? wrap(value, base, length) : value;

        if (axis === 'y') {
            scroller.scrollTop = position;
        } else {
            scroller.scrollLeft = position;
        }

        written = read();
    };

    const step = (ms: number, now: number): void => {
        if (held || now < idleUntil) {
            return;
        }

        place(position + ms * PX_PER_MS[axis], false);
    };

    const follow = (now: number): void => {
        const actual = read();

        if (Math.abs(actual - written) > 1) {
            idleUntil = now + IDLE_MS;
            place(actual, true);
        }
    };

    const hold = (value: boolean): void => {
        held = value;
    };

    const turn = (next: Axis): void => {
        axis = next;
        position = read();
        written = position;
    };

    return {
        step,
        follow,
        hold,
        turn,
    };
};

import { prefersReducedMotion } from '@/shared/Motion/prefersReducedMotion';
import { fizzInside } from '@/shared/Showcase/Mark/fizzInside';
import { IMPACT_MS, pierceLime } from '@/shared/Showcase/Mark/pierceLime';
import { releaseBubble } from '@/shared/Showcase/Mark/releaseBubble';

const SVG = 'http://www.w3.org/2000/svg';
const MARKS = 'svg[data-mark-motion]';
const INTRO_MS = 600;
const BUBBLE_GAP_MS = 650;
const POUR_GAP_MS = 10000;
const BURST = 4;
const BURST_GAP_MS = 140;

const runMark = (mark: Element): (() => void) => {
    const drink = mark.querySelector('[data-mark-drink]');
    const slice = mark.querySelector('[data-mark-lime]');
    const glass = mark.querySelector('[data-mark-glass]');
    const bubbles = document.createElementNS(SVG, 'g');
    let timers: number[] = [];

    const pour = (): void => {
        if (drink !== null && slice !== null) {
            pierceLime(drink, slice);
        }

        for (let bubble = 0; bubble < BURST; bubble++) {
            releaseBubble(bubbles, IMPACT_MS + bubble * BURST_GAP_MS);
        }
    };

    const sync = (): void => {
        for (const timer of timers) {
            window.clearTimeout(timer);
        }

        timers = document.hidden
            ? []
            : [
                    window.setInterval(() => {
                        releaseBubble(bubbles);
                    }, BUBBLE_GAP_MS),
                    window.setInterval(pour, POUR_GAP_MS),
                ];
    };

    bubbles.setAttribute('fill', 'none');
    bubbles.setAttribute('stroke', 'currentColor');
    bubbles.setAttribute('stroke-width', '0.45');
    mark.prepend(bubbles);

    const stopFizz = glass === null ? () => undefined : fizzInside(glass);

    sync();
    timers.push(window.setTimeout(pour, INTRO_MS));
    document.addEventListener('visibilitychange', sync);

    return () => {
        document.removeEventListener('visibilitychange', sync);

        for (const timer of timers) {
            window.clearTimeout(timer);
        }

        bubbles.remove();
        stopFizz();
    };
};

export const startMarkMotion = (): (() => void) => {
    if (prefersReducedMotion()) {
        return () => undefined;
    }

    const stops = [...document.querySelectorAll(MARKS)].map(runMark);

    return () => {
        for (const stop of stops) {
            stop();
        }
    };
};

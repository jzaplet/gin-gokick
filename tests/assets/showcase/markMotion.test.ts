import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { startMarkMotion } from '@/shared/Showcase/Mark/startMarkMotion';
import { forgetAnimations, recordAnimations, type RecordedAnimation } from './animations';

const SVG = 'http://www.w3.org/2000/svg';

type Mark = {
    svg: SVGSVGElement;
    drink: SVGGElement;
    glass: SVGUseElement;
    lime: SVGUseElement;
};

let stop = (): void => undefined;
let animations: RecordedAnimation[] = [];
let hidden = false;

const motion = (reduced: boolean): void => {
    vi.stubGlobal('matchMedia', (media: string) => ({
        media,
        matches: reduced,
    }));
};

const addMark = (): Mark => {
    const svg = document.createElementNS(SVG, 'svg');
    const drink = document.createElementNS(SVG, 'g');
    const glass = document.createElementNS(SVG, 'use');
    const lime = document.createElementNS(SVG, 'use');

    svg.dataset['markMotion'] = '';
    drink.dataset['markDrink'] = '';
    glass.dataset['markGlass'] = '';
    lime.dataset['markLime'] = '';
    drink.append(glass, lime);
    svg.append(drink);
    document.body.append(svg);

    return {
        svg,
        drink,
        glass,
        lime,
    };
};

beforeEach(() => {
    vi.useFakeTimers();
    vi.spyOn(document, 'hidden', 'get').mockImplementation(() => hidden);
    hidden = false;
    animations = recordAnimations();
});

afterEach(() => {
    stop();
    forgetAnimations();
    document.body.replaceChildren();
    vi.useRealTimers();
});

describe('the mark motion', () => {
    it('leaves the mark still for a visitor who asked for less motion', () => {
        motion(true);
        const mark = addMark();

        stop = startMarkMotion();
        vi.advanceTimersByTime(30_000);

        expect(mark.svg.children).toHaveLength(1);
        expect(animations).toEqual([]);
    });
});

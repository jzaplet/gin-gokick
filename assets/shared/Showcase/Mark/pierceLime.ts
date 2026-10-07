import { around } from '@/shared/Showcase/Mark/around';

const LIME_X = 23;
const LIME_Y = 6;
const FOOT_X = 13;
const FOOT_Y = 29;
const PIERCE_MS = 900;
const SHAKE_MS = 1100;

export const IMPACT_MS = PIERCE_MS * 0.7;

const lime = (transform: string): string => around(LIME_X, LIME_Y, transform);

const tilt = (degrees: number): string => around(FOOT_X, FOOT_Y, `rotate(${degrees}deg)`);

export const pierceLime = (drink: Element, slice: Element): void => {
    slice.animate([
        {
            transform: lime('translate(0, 0) rotate(0deg)'),
            easing: 'cubic-bezier(.2, .7, .3, 1)',
        },
        {
            transform: lime('translate(1px, -4.5px) rotate(-14deg)'),
            offset: 0.42,
            easing: 'cubic-bezier(.7, 0, .9, .5)',
        },
        {
            transform: lime('translate(0, 0.8px) rotate(0deg)'),
            offset: 0.7,
            easing: 'ease-out',
        },
        {
            transform: lime('translate(0, -0.25px) rotate(0deg)'),
            offset: 0.85,
            easing: 'ease-in-out',
        },
        { transform: lime('translate(0, 0) rotate(0deg)') },
    ], { duration: PIERCE_MS });
    drink.animate(
        [
            tilt(0),
            tilt(-5),
            tilt(3.5),
            tilt(-2),
            tilt(0.8),
            tilt(0),
        ].map((transform) => ({ transform })),
        {
            duration: SHAKE_MS,
            delay: IMPACT_MS,
            easing: 'ease-in-out',
        },
    );
};

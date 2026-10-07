const SVG = 'http://www.w3.org/2000/svg';
const MAX_BUBBLES = 10;
const FROM_X = 12;
const SPREAD_X = 5;
const FROM_Y = 10.5;
const RISE = 15;

const at = (x: number, y: number, scale: number): string => `translate(${x}px, ${y}px) scale(${scale})`;

export const releaseBubble = (bubbles: Element, delay = 0): void => {
    if (bubbles.childElementCount >= MAX_BUBBLES) {
        return;
    }

    const bubble = document.createElementNS(SVG, 'circle');
    const x = FROM_X + Math.random() * SPREAD_X;
    const drift = (Math.random() - 0.5) * 3;
    const size = 0.7 + Math.random() * 0.5;

    bubble.setAttribute('r', '1');
    bubbles.append(bubble);
    bubble.animate(
        [
            {
                transform: at(x, FROM_Y, size * 0.4),
                opacity: 0,
            },
            {
                transform: at(x + drift * 0.4, FROM_Y - RISE * 0.3, size),
                opacity: 1,
                offset: 0.35,
            },
            {
                transform: at(x - drift * 0.2, FROM_Y - RISE * 0.65, size),
                opacity: 0.85,
                offset: 0.7,
            },
            {
                transform: at(x + drift, FROM_Y - RISE, size * 1.15),
                opacity: 0,
            },
        ],
        {
            duration: 1800 + Math.random() * 900,
            delay,
            easing: 'cubic-bezier(.33, .66, .45, 1)',
            fill: 'both',
        },
    ).onfinish = () => {
        bubble.remove();
    };
};

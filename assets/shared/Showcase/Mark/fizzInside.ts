const SVG = 'http://www.w3.org/2000/svg';
const INK = '#0b1226';
const HOLES = [
    {
        x: 9,
        y: 19,
    },
    {
        x: 11,
        y: 15,
    },
    {
        x: 9,
        y: 11,
    },
];
const RISE_MS = 5600;
const BOTTOM = 21.5;
const SURFACE = 9;
const SHOWN = [
    0.12,
    0.88,
];
const AREA = {
    x: '-8',
    y: '-8',
    width: '48',
    height: '48',
};

let masks = 0;

const svgElement = <K extends keyof SVGElementTagNameMap>(
    name: K,
    attributes: Record<string, string>,
): SVGElementTagNameMap[K] => {
    const element = document.createElementNS(SVG, name);

    for (const [key, value] of Object.entries(attributes)) {
        element.setAttribute(key, value);
    }

    return element;
};

const heightAt = (offset: number): number => BOTTOM - offset * (BOTTOM - SURFACE);

const rise = (bubble: Element, x: number, y: number): void => {
    const [appear = 0, fade = 1] = SHOWN;

    bubble.animate(
        [
            {
                transform: `translate(${x}px, ${heightAt(0)}px) scale(0.5)`,
                opacity: 0,
            },
            {
                transform: `translate(${x}px, ${heightAt(appear)}px) scale(1)`,
                opacity: 1,
                offset: appear,
            },
            {
                transform: `translate(${x}px, ${heightAt(fade)}px) scale(1)`,
                opacity: 1,
                offset: fade,
            },
            {
                transform: `translate(${x}px, ${heightAt(1)}px) scale(0.5)`,
                opacity: 0,
            },
        ],
        {
            duration: RISE_MS,
            iterations: Infinity,
            iterationStart: (BOTTOM - y) / (BOTTOM - SURFACE),
        },
    );
};

export const fizzInside = (glass: Element): (() => void) => {
    masks += 1;

    const id = `mark-fizz-${masks}`;
    const defs = svgElement('defs', {});
    const mask = svgElement('mask', {
        id,
        maskUnits: 'userSpaceOnUse',
        ...AREA,
    });
    const masked = svgElement('g', { mask: `url(#${id})` });

    mask.append(
        svgElement('rect', {
            ...AREA,
            fill: 'white',
        }),
    );
    defs.append(mask);
    glass.replaceWith(defs, masked);
    masked.append(glass);

    for (const { x, y } of HOLES) {
        const bubble = svgElement('circle', {
            r: '1',
            fill: 'black',
        });

        masked.append(
            svgElement('circle', {
                cx: String(x),
                cy: String(y),
                r: '1.1',
                fill: INK,
            }),
        );
        mask.append(bubble);
        rise(bubble, x, y);
    }

    return () => {
        masked.replaceWith(glass);
        defs.remove();
    };
};

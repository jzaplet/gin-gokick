import { describe, expect, it } from 'vitest';

const sources = import.meta.glob<string>('@/img/*.svg', {
    query: '?raw',
    import: 'default',
    eager: true,
});

const darkVariants = Object.keys(sources).filter((path) => path.endsWith('-dark.svg'));

const shapes = (source = ''): string[] => [...source.matchAll(/ d="([^"]+)"/g)].map(([, shape]) => shape ?? '');

describe('the dark variants of the brand', () => {
    it.each(darkVariants)('%s draws the shapes of its light variant', (dark) => {
        const light = sources[dark.replace(/-dark\.svg$/, '.svg')];

        expect(light).toBeDefined();
        expect(shapes(sources[dark])).toEqual(shapes(light));
    });
});

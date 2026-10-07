import { mount } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import type { Component } from 'vue';
import BaseIcon from '@/shared/Icons/BaseIcon.vue';

const files = import.meta.glob<string>('@/img/icons/*.svg', {
    query: '?raw',
    import: 'default',
    eager: true,
});

const icons = import.meta.glob<Component>('@/shared/Icons/Icon*.vue', {
    import: 'default',
    eager: true,
});

const nameOf = (path: string): string => path.replace(/^.*\/|\.svg(?:#icon)?$/g, '');

describe.each(Object.entries(icons))('%s', (_, icon) => {
    it('draws an icon file in the text color, hidden from screen readers', () => {
        const svg = mount(icon).get('svg');

        expect(svg.attributes('aria-hidden')).toBe('true');
        expect(svg.attributes('stroke')).toBe('currentColor');
        expect(svg.get('use').attributes('href')).toMatch(/\/img\/icons\/[a-z-]+\.svg#icon$/);
    });
});

describe('the icons', () => {
    it('draw every icon file, each with one component', () => {
        const drawn = Object.values(icons).map((icon) => nameOf(mount(icon).get('use').attributes('href') ?? ''));

        expect(drawn.toSorted()).toEqual(Object.keys(files).map(nameOf).toSorted());
    });

    it('take their size and stroke from the caller', () => {
        const svg = mount(BaseIcon, {
            attrs: {
                'class': 'size-4',
                'stroke-width': '1.75',
            },
        }).get('svg');

        expect(svg.classes()).toContain('size-4');
        expect(svg.attributes('stroke-width')).toBe('1.75');
    });
});

describe('the icon files', () => {
    it.each(Object.entries(files))('%s is a 16×16 SVG whose drawing is the element icon', (_, source) => {
        expect(source).toMatch(/^<svg xmlns="http:\/\/www\.w3\.org\/2000\/svg" viewBox="0 0 16 16"[ >]/);
        expect(source.match(/ id="icon"/g)).toHaveLength(1);
    });

    it('leaves the stroke width of a stroked icon to the caller', () => {
        for (const source of Object.values(files)) {
            expect(source).not.toMatch(/<path [^>]*stroke-width/);
        }
    });
});

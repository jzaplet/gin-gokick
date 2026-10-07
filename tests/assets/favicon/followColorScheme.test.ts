import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import darkMark from '@/img/mark-dark.svg';
import lightMark from '@/img/mark.svg';
import { followColorScheme } from '@/shared/Favicon/followColorScheme';

let dark = false;
let listeners: (() => void)[] = [];

const scheme = {
    get matches(): boolean {
        return dark;
    },
    addEventListener(_type: string, listener: () => void): void {
        listeners.push(listener);
    },
};

const browserIn = (mode: 'dark' | 'light'): void => {
    dark = mode === 'dark';
    listeners.forEach((listener) => {
        listener();
    });
};

const icon = (): string | null | undefined =>
    document.head.querySelector('link[type="image/svg+xml"]')?.getAttribute('href');

beforeEach(() => {
    const link = document.createElement('link');

    link.rel = 'icon';
    link.type = 'image/svg+xml';
    link.href = lightMark;
    document.head.append(link);
    dark = false;
    listeners = [];
    vi.stubGlobal('matchMedia', (media: string) => (media === '(prefers-color-scheme: dark)' ? scheme : undefined));
});

afterEach(() => {
    document.head.querySelectorAll('link[rel="icon"]').forEach((link) => {
        link.remove();
    });
});

describe('followColorScheme', () => {
    it('shows the white mark on a dark browser', () => {
        browserIn('dark');
        followColorScheme();

        expect(icon()).toBe(darkMark);
    });

    it('follows the browser when it changes its scheme', () => {
        followColorScheme();

        browserIn('dark');
        expect(icon()).toBe(darkMark);

        browserIn('light');
        expect(icon()).toBe(lightMark);
    });
});

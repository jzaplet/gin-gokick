import type { VueWrapper } from '@vue/test-utils';

export const toastTexts = (wrapper: VueWrapper): string[][] =>
    wrapper.findAll('[aria-live="polite"] > li').map((toast) => toast.findAll('p').map((line) => line.text()));

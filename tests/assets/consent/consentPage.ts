import { flushPromises } from '@vue/test-utils';
import { expect, type Mock, vi } from 'vitest';
import { followChoice } from '@/shared/Consent/Choice/currentChoice';
import { startConsent } from '@/shared/Consent/Start/startConsent';
import type { ConsentChoice } from '@/shared/Consent/types/ConsentChoice';

type Started = {
    reload: Mock<() => void>;
    choices: Readonly<ConsentChoice>[];
};

let unfollow = (): void => undefined;

export const saved = (tools: string, choice: ConsentChoice): string => JSON.stringify({
    tools,
    choice,
});

export const start = (ids: Record<string, string>, stored?: string): Started => {
    const meta = document.createElement('meta');

    meta.name = 'tracking';
    Object.assign(meta.dataset, ids);
    document.head.append(meta);

    if (stored !== undefined) {
        localStorage.setItem('consent', stored);
    }

    const link = document.createElement('a');

    link.dataset['consentSettings'] = '';
    link.href = '#';
    link.textContent = 'Nastavení cookies';
    document.body.append(link);

    const reload = vi.fn<() => void>();
    const choices: Readonly<ConsentChoice>[] = [];
    const { href, origin, pathname, search, hash } = location;

    vi.stubGlobal('location', {
        href,
        origin,
        pathname,
        search,
        hash,
        reload,
    });
    startConsent();
    unfollow = followChoice((choice) => {
        choices.push(choice);
    });

    return {
        reload,
        choices,
    };
};

export const banner = (): HTMLElement | null => document.querySelector('section[aria-labelledby="consent-title"]');

export const settingsOpen = (): boolean => document.querySelector('dialog')?.open === true;

export const button = (text: string): HTMLButtonElement => {
    const found = Array.from(document.querySelectorAll('button')).find(
        (element) => element.textContent.trim() === text,
    );

    if (found === undefined) {
        throw new Error(`no button ${text}`);
    }

    return found;
};

export const openFromLink = async (): Promise<MouseEvent> => {
    const event = new MouseEvent('click', {
        bubbles: true,
        cancelable: true,
    });

    document.querySelector('[data-consent-settings]')?.dispatchEvent(event);
    await flushPromises();

    return event;
};

export const bannerGone = async (): Promise<void> => {
    await vi.waitFor(() => {
        expect(banner()).toBeNull();
    });
};

export const toggle = (name: string): HTMLInputElement | null => document.querySelector(`dialog input[name="${name}"]`);

export const click = async (text: string): Promise<void> => {
    button(text).click();
    await flushPromises();
};

export const inSettings = async (text: string): Promise<void> => {
    const found = Array.from(document.querySelectorAll('dialog button')).find(
        (element) => element.textContent.trim() === text,
    );

    if ((found instanceof HTMLButtonElement) === false) {
        throw new Error(`no button ${text} in the settings`);
    }

    found.click();
    await flushPromises();
};

export const clearConsentPage = (): void => {
    unfollow();
    document.head.replaceChildren();
    document.body.replaceChildren();
    localStorage.clear();
};

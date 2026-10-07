import { flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import { vi } from 'vitest';
import App from '@/app/App.vue';
import { router } from '@/router';
import { rememberSession } from '@/shared/Auth/Session/currentSession';
import { json } from '../fetch/fixtures';

export type Answer = (url: string, init?: RequestInit) => Response | Promise<Response>;

export const jan = {
    id: '0192f3a4-5b6c-7d8e-9f01-23456789abcd',
    email: 'jan@example.com',
};

const urlOf = (input: unknown): string => (typeof input === 'string' ? input : '');

const pathOf = (input: unknown): string => new URL(urlOf(input), 'http://localhost').pathname;

export const answerApi = (answers: Readonly<Record<string, Answer>> = {}): void => {
    const table: Readonly<Record<string, Answer>> = {
        '/api/user/me': () => json(jan, 200),
        '/api/auth/sessions': () => json({
            items: [],
            total: 0,
        }, 200),
        ...answers,
    };

    vi.spyOn(globalThis, 'fetch').mockImplementation(async (input, init) => {
        const answer = table[pathOf(input)];

        if (answer === undefined) {
            throw new Error(`no answer for ${urlOf(input)}`);
        }

        return answer(urlOf(input), init);
    });
};

export const openApp = async (path: string): Promise<VueWrapper> => {
    await router.push(path);
    const wrapper = mount(App, { global: { plugins: [router] } });

    await flushPromises();

    return wrapper;
};

export const openSignedIn = async (
    path = '/cs/app/dashboard',
    answers: Readonly<Record<string, Answer>> = {},
): Promise<VueWrapper> => {
    rememberSession({ id: jan.id });
    answerApi(answers);

    return openApp(path);
};

export const calls = (path: string): Parameters<typeof fetch>[] => vi.mocked(fetch).mock.calls
    .filter(([input]) => pathOf(input) === path);

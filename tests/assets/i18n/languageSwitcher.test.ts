import { enableAutoUnmount, flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { nextTick } from 'vue';
import { router } from '@/router';
import type { SaveLanguage } from '@/shared/I18n/Composables/useLanguageSwitcher';
import LanguageSwitcher from '@/shared/I18n/Components/LanguageSwitcher.vue';
import { dictionaries } from '@/shared/I18n/Dictionary/dictionaries';
import { dictionary as englishDictionary } from '@/shared/I18n/Dictionary/Locales/en_US';
import { t } from '@/shared/I18n/Texts/translate';
import { offerEnglish, showEnglish } from './languages';

enableAutoUnmount(afterEach);

beforeEach(async () => {
    await router.push('/cs/app/login');
});

const make = (props: {
    tooltipPosition?: 'top' | 'bottom';
    save?: SaveLanguage;
} = {}): VueWrapper =>
    mount(LanguageSwitcher, {
        props,
        attachTo: document.body,
        global: { plugins: [router] },
    });

const openList = async (path: string, props: { save?: SaveLanguage } = {}): Promise<VueWrapper> => {
    offerEnglish();
    await router.push(path);
    const wrapper = make(props);

    await wrapper.get('button').trigger('click');

    return wrapper;
};

const clickWith = (wrapper: VueWrapper, init: MouseEventInit): boolean => {
    let prevented = false;
    const settle = (event: Event): void => {
        prevented = event.defaultPrevented;
        event.preventDefault();
    };

    document.addEventListener('click', settle, { once: true });
    wrapper.get('a').element.dispatchEvent(
        new MouseEvent('click', {
            bubbles: true,
            cancelable: true,
            ...init,
        }),
    );

    return prevented;
};

describe('the language switcher', () => {
    it('wears the flag of the shown language and names its action in it', async () => {
        offerEnglish();
        const wrapper = make();
        const czech = t('language.change');

        expect(wrapper.get('button').attributes('aria-label')).toBe(czech);
        expect(wrapper.get('button').attributes('aria-expanded')).toBe('false');
        expect(wrapper.get('button img').attributes('src')).toMatch(/\/img\/flags\/cz\.svg$/);

        await showEnglish();
        await nextTick();

        expect(wrapper.get('button').attributes('aria-label')).toBe(t('language.change'));
        expect(t('language.change')).not.toBe(czech);
        expect(wrapper.get('button img').attributes('src')).toMatch(/\/img\/flags\/us\.svg$/);
    });

    it('lists every language by its own name and marks the one of the page', async () => {
        const items = (await openList('/cs/app/login')).findAll('li');

        expect(
            items.map((item) => [
                item.attributes('lang'),
                item.text(),
                item.attributes('aria-current'),
            ]),
        ).toEqual([
            [
                'cs',
                dictionaries['cs_CZ']?.name,
                'true',
            ],
            [
                'en',
                dictionaries['en_US']?.name,
                undefined,
            ],
        ]);
        expect(items[0]?.find('a').exists()).toBe(false);
    });

    it('links the other language to the same screen and follows the route', async () => {
        const wrapper = await openList('/cs/app/login?redirect=/dashboard');

        expect(wrapper.get('a').attributes('href')).toBe('/en/app/login?redirect=/dashboard');

        await router.push('/cs/app/register?redirect=/dashboard#top');

        expect(wrapper.get('a').attributes('href')).toBe('/en/app/register?redirect=/dashboard#top');
    });

    it('switches the screen to the other language on a plain click without a reload', async () => {
        const wrapper = await openList('/cs/app/login?redirect=/dashboard#top');

        expect(clickWith(wrapper, {})).toBe(true);
        await flushPromises();

        expect(router.currentRoute.value.fullPath).toBe('/en/app/login?redirect=/dashboard#top');
        expect(t('login.title')).toBe(englishDictionary['login.title']);
        expect(document.documentElement.lang).toBe('en-US');
        expect(document.title).toBe(`${t('login.title')} | ${t('brand.name')}`);
        expect(wrapper.get('button img').attributes('src')).toMatch(/\/img\/flags\/us\.svg$/);
    });

    it('leaves a click with a modifier to the browser', async () => {
        const wrapper = await openList('/cs/app/login');

        expect(clickWith(wrapper, { ctrlKey: true })).toBe(false);
        await flushPromises();

        expect(router.currentRoute.value.fullPath).toBe('/cs/app/login');
    });

    it.each([
        [
            true,
            '/en/app/register',
        ],
        [
            false,
            '/cs/app/register',
        ],
    ])('switches only when its saver stores the language: %s', async (stored, path) => {
        const save = vi.fn<SaveLanguage>().mockResolvedValue(stored);
        const wrapper = await openList('/cs/app/register', { save });

        clickWith(wrapper, {});
        await flushPromises();

        expect(save).toHaveBeenCalledExactlyOnceWith(
            expect.objectContaining({
                locale: 'en_US',
                href: '/en/app/register',
            }),
        );
        expect(router.currentRoute.value.fullPath).toBe(path);
    });
});

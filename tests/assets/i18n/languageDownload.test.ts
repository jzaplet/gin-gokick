import { enableAutoUnmount, flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, it, type Mock, vi } from 'vitest';
import { router } from '@/router';
import LanguageSwitcher from '@/shared/I18n/Components/LanguageSwitcher.vue';
import { dictionaries } from '@/shared/I18n/Dictionary/dictionaries';
import type { DictionarySource } from '@/shared/I18n/Dictionary/types/DictionarySource';
import { appStorage } from '@/shared/Storage/appStorage';
import { offerEnglish } from './languages';

enableAutoUnmount(afterEach);

const english = (): DictionarySource => {
    const source = dictionaries['en_US'];

    if (source === undefined) {
        throw new Error('The test needs the English dictionary');
    }

    return source;
};

const openSwitcher = async (): Promise<VueWrapper> => {
    offerEnglish();
    await router.push('/cs/app/login');

    return mount(LanguageSwitcher, {
        attachTo: document.body,
        global: { plugins: [router] },
    });
};

const stubAssign = (): Mock<(url: string) => void> => {
    const assign = vi.fn<(url: string) => void>();
    const { href, origin, pathname, search, hash } = location;

    vi.stubGlobal('location', {
        href,
        origin,
        pathname,
        search,
        hash,
        assign,
    });

    return assign;
};

beforeEach(() => {
    appStorage().remove('dictionary.en_US');
});

describe('the language switcher', () => {
    it('loads the page in the other language when its dictionary does not download', async () => {
        vi.spyOn(english(), 'load').mockRejectedValue(new Error('offline'));
        const wrapper = await openSwitcher();

        await wrapper.get('button').trigger('click');
        const assign = stubAssign();

        await wrapper.get('a').trigger('click');
        await flushPromises();

        expect(assign).toHaveBeenCalledExactlyOnceWith('/en/app/login');
        expect(router.currentRoute.value.fullPath).toBe('/cs/app/login');
    });
});

import { computed, type ComputedRef } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import { screenOf, screenIn } from '@/shared/Navigation/screenPath';
import { dictionaries } from '@/shared/I18n/Dictionary/dictionaries';
import { flagOf } from '@/shared/I18n/Flags/flagOf';
import { languageOfLocale, offeredLocale, switcherLocales } from '@/shared/I18n/Page/localeMeta';
import { shownLanguage, loadDictionary, preloadDictionary } from '@/shared/I18n/Texts/pageDictionary';

export type SwitcherLanguage = {
    locale: string;
    code: string;
    name: string;
    flag: string;
    active: boolean;
    href: string;
};

export type SaveLanguage = (language: SwitcherLanguage) => Promise<boolean>;

export const saveNothing: SaveLanguage = () => Promise.resolve(true);

type LanguageSwitcher = {
    languages: ComputedRef<SwitcherLanguage[]>;
    active: ComputedRef<SwitcherLanguage | undefined>;
    pick: (language: SwitcherLanguage, event: MouseEvent) => void;
    preload: () => void;
};

const plainClick = (event: MouseEvent): boolean =>
    event.button === 0
    && event.metaKey === false
    && event.ctrlKey === false
    && event.shiftKey === false
    && event.altKey === false;

export const useLanguageSwitcher = (save: SaveLanguage): LanguageSwitcher => {
    const route = useRoute();
    const router = useRouter();
    const listed = switcherLocales().flatMap((locale) => {
        const flag = flagOf(locale);
        const name = dictionaries[locale]?.name;
        const code = languageOfLocale(locale);
        const offered = offeredLocale(code) ?? locale;

        return flag === undefined || name === undefined
            ? []
            : [{
                    locale: offered,
                    code,
                    name,
                    flag,
                }];
    });
    const languages = computed(() => listed.map((language) => ({
        ...language,
        active: language.code === shownLanguage(),
        href: screenIn(language.code, screenOf(route)),
    })));

    const switchTo = async (language: SwitcherLanguage): Promise<void> => {
        const ready = await loadDictionary(language.locale);

        if ((await save(language)) === false) {
            return;
        }

        if (ready) {
            await router.replace(language.href);
        } else {
            location.assign(language.href);
        }
    };

    const preload = (): void => {
        for (const language of languages.value) {
            if (language.active === false) {
                preloadDictionary(language.locale);
            }
        }
    };

    const pick = (language: SwitcherLanguage, event: MouseEvent): void => {
        if (plainClick(event) === false) {
            return;
        }

        event.preventDefault();
        void switchTo(language);
    };

    const active = computed(() => languages.value.find((language) => language.active));

    return {
        languages,
        active,
        pick,
        preload,
    };
};

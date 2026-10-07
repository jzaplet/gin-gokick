import { computed, type ComputedRef, onBeforeUnmount, onMounted, type Ref, ref } from 'vue';
import { currentChoice, setCurrentChoice } from '@/shared/Consent/Choice/currentChoice';
import { writeStoredChoice } from '@/shared/Consent/Choice/storedChoice';
import { askedCategories } from '@/shared/Consent/Start/askedCategories';
import type { AskedCategory } from '@/shared/Consent/types/AskedCategory';
import type { CategoryRow } from '@/shared/Consent/types/CategoryRow';
import type { ConsentCategory } from '@/shared/Consent/types/ConsentCategory';
import type { ConsentChoice } from '@/shared/Consent/types/ConsentChoice';
import { languageTag } from '@/shared/I18n/Page/localeMeta';
import { useModalDialog } from '@/shared/Modal/Composables/useModalDialog';
import { shownLocale } from '@/shared/I18n/Texts/pageDictionary';
import { type PlainMessageKey, t } from '@/shared/I18n/Texts/translate';
import type { EnabledTools } from '@/shared/Tracking/types/EnabledTools';
import type { TrackingTool } from '@/shared/Tracking/types/TrackingTool';

type CookieConsent = {
    bannerShown: Readonly<Ref<boolean>>;
    selected: Ref<Record<ConsentCategory, boolean>>;
    rows: ComputedRef<CategoryRow[]>;
    acceptAll: () => void;
    acceptNecessary: () => void;
    save: () => void;
    openSettings: () => void;
    closeSettings: () => void;
    closeOnBackdrop: (event: MouseEvent) => void;
    cancelled: (event: Event) => void;
    closed: () => void;
};

const texts = {
    analytics: {
        title: 'consent.analytics',
        text: 'consent.analytics_text',
    },
    marketing: {
        title: 'consent.marketing',
        text: 'consent.marketing_text',
    },
} satisfies Record<ConsentCategory, Pick<CategoryRow, 'title' | 'text'>>;

const toolNames = {
    ga4: 'tracking.ga4',
    google_ads: 'tracking.google_ads',
    meta_pixel: 'tracking.meta_pixel',
} satisfies Record<TrackingTool, PlainMessageKey>;

const toolList = (): Intl.ListFormat => new Intl.ListFormat(languageTag(shownLocale()), { type: 'conjunction' });

const granted = (choice: Readonly<ConsentChoice>): Record<ConsentCategory, boolean> => ({
    analytics: choice.analytics === true,
    marketing: choice.marketing === true,
});

const row = (asked: AskedCategory): CategoryRow => ({
    ...texts[asked.category],
    category: asked.category,
    tools: toolList().format(asked.tools.map((tool) => t(toolNames[tool]).replaceAll(' ', ' '))),
});

export const useCookieConsent = (tools: EnabledTools): CookieConsent => {
    const categories = askedCategories(tools);
    const undecided = ref(categories.some((asked) => currentChoice()[asked.category] === undefined));
    const settingsOpen = ref(false);
    const selected = ref(granted(currentChoice()));
    const { closeOnBackdrop, cancelled, closed } = useModalDialog(settingsOpen);

    const decide = (grant: (category: ConsentCategory) => boolean): void => {
        const before = currentChoice();
        const choice: ConsentChoice = {};

        for (const asked of categories) {
            choice[asked.category] = grant(asked.category);
        }

        undecided.value = false;
        settingsOpen.value = false;
        writeStoredChoice(tools, choice);
        setCurrentChoice(choice);

        if (categories.some((asked) => before[asked.category] === true && choice[asked.category] === false)) {
            location.reload();
        }
    };

    const openSettings = (): void => {
        selected.value = granted(currentChoice());
        settingsOpen.value = true;
    };

    const openFromLink = (event: MouseEvent): void => {
        if (event.target instanceof Element && event.target.closest('[data-consent-settings]') !== null) {
            event.preventDefault();
            openSettings();
        }
    };

    const bannerShown = computed(() => undecided.value && settingsOpen.value === false);

    const rows = computed(() => categories.map(row));

    const acceptAll = (): void => {
        decide(() => true);
    };

    const acceptNecessary = (): void => {
        decide(() => false);
    };

    const save = (): void => {
        decide((category) => selected.value[category]);
    };

    const closeSettings = (): void => {
        settingsOpen.value = false;
    };

    onMounted(() => {
        document.addEventListener('click', openFromLink);
    });
    onBeforeUnmount(() => {
        document.removeEventListener('click', openFromLink);
    });

    return {
        bannerShown,
        selected,
        rows,
        acceptAll,
        acceptNecessary,
        save,
        openSettings,
        closeSettings,
        closeOnBackdrop,
        cancelled,
        closed,
    };
};

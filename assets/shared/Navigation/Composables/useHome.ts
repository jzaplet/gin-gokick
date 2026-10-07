import { computed, type ComputedRef } from 'vue';
import { homePageOf } from '@/shared/I18n/Page/localeMeta';
import { shownLanguage } from '@/shared/I18n/Texts/pageDictionary';

export const useHome = (): ComputedRef<string> => computed(() => homePageOf(shownLanguage()));

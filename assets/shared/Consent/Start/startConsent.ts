import { createApp } from 'vue';
import { setCurrentChoice } from '@/shared/Consent/Choice/currentChoice';
import { readStoredChoice } from '@/shared/Consent/Choice/storedChoice';
import CookieConsent from '@/shared/Consent/CookieConsent.vue';
import { askedCategories } from '@/shared/Consent/Start/askedCategories';
import { readEnabledTools } from '@/shared/Tracking/enabledTools';

export const startConsent = (): void => {
    const tools = readEnabledTools();

    if (tools === undefined || askedCategories(tools).length === 0) {
        return;
    }

    setCurrentChoice(readStoredChoice(tools));

    const element = document.createElement('div');

    document.body.append(element);
    createApp(CookieConsent, { tools }).mount(element);
};

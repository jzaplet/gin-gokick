import { followChoice } from '@/shared/Consent/Choice/currentChoice';
import { ConsentCategory } from '@/shared/Consent/types/ConsentCategory';
import { loadScript } from '@/shared/ScriptLoader/loadScript';
import { ScriptURL } from '@/shared/ScriptLoader/types/ScriptURL';
import { gtag } from '@/shared/Tracking/Google/gtag';
import type { GoogleConsent } from '@/shared/Tracking/Google/types/GoogleConsent';
import type { AdsConversion } from '@/shared/Tracking/types/AdsConversion';
import type { AdsLabels } from '@/shared/Tracking/types/AdsLabels';
import type { TrackedUser } from '@/shared/Tracking/types/TrackedUser';
import { googleEmailHash } from '@/shared/Tracking/UserData/emailHashes';
import type { EnabledTools } from '@/shared/Tracking/types/EnabledTools';

const configured = new Set<string>();

let destinations = {
    ga4: '',
    ads: '',
};

let labels: AdsLabels = {};

let shownPage = '';

const withoutFragment = (url: string): string => {
    const { origin, pathname, search } = new URL(url);

    return origin + pathname + search;
};

const consentState = (granted: boolean): 'granted' | 'denied' => (granted ? 'granted' : 'denied');

const googleConsent = (analytics: boolean, ads: boolean): GoogleConsent => ({
    analytics_storage: consentState(analytics),
    ad_storage: consentState(ads),
    ad_user_data: consentState(ads),
    ad_personalization: consentState(ads),
});

const showPage = (): void => {
    shownPage = withoutFragment(location.href);
    gtag('set', { page_location: shownPage });
};

const loadGoogleTag = (id: string): void => {
    gtag('consent', 'default', googleConsent(false, false));
    gtag('js', new Date());
    loadScript(ScriptURL.GoogleTag, { id });
};

export const startGoogleTag = (tools: EnabledTools, adsLabels: AdsLabels): void => {
    const grants = [
        {
            id: tools.ga4,
            category: ConsentCategory.Analytics,
        },
        {
            id: tools.google_ads,
            category: ConsentCategory.Marketing,
        },
    ].filter(({ id }) => id !== '');

    destinations = {
        ga4: tools.ga4,
        ads: tools.google_ads,
    };
    labels = adsLabels;

    followChoice((choice) => {
        const granted = grants
            .filter(({ id, category }) => choice[category] === true && configured.has(id) === false)
            .map(({ id }) => id);
        const [first] = granted;

        if (first === undefined) {
            return;
        }

        if (configured.size === 0) {
            loadGoogleTag(first);
        }

        for (const id of granted) {
            configured.add(id);
        }

        gtag('consent', 'update', googleConsent(configured.has(tools.ga4), configured.has(tools.google_ads)));
        showPage();

        for (const id of granted) {
            gtag('config', id, { send_page_view: false });
        }

        gtag('event', 'page_view', { send_to: granted });
    });
};

export const trackGooglePageView = (): void => {
    if (configured.size > 0 && withoutFragment(location.href) !== shownPage) {
        showPage();
        gtag('event', 'page_view', { send_to: [...configured] });
    }
};

export const trackGoogleEvent = async (
    event: string | null,
    conversion: AdsConversion | null,
    user?: TrackedUser,
): Promise<void> => {
    if (event !== null && configured.has(destinations.ga4)) {
        gtag('event', event, { send_to: [destinations.ga4] });
    }

    const label = conversion === null ? undefined : labels[conversion];

    if (label === undefined || configured.has(destinations.ads) === false) {
        return;
    }

    if (user !== undefined) {
        gtag('set', 'user_data', { sha256_email_address: await googleEmailHash(user.email) });
    }

    gtag('event', 'conversion', { send_to: [`${destinations.ads}/${label}`] });
};

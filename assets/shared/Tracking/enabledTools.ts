import { isAdsConversion } from '@/shared/Tracking/types/AdsConversion';
import type { AdsLabels } from '@/shared/Tracking/types/AdsLabels';
import type { EnabledTools } from '@/shared/Tracking/types/EnabledTools';

const trackingMeta = (): HTMLMetaElement | null => document.querySelector<HTMLMetaElement>('meta[name="tracking"]');

export const needsConsent = (): boolean => trackingMeta() !== null;

export const readEnabledTools = (): EnabledTools | undefined => {
    const meta = trackingMeta();

    if (meta === null) {
        return undefined;
    }

    const { ga4 = '', googleAds = '', metaPixel = '' } = meta.dataset;

    return {
        ga4,
        google_ads: googleAds,
        meta_pixel: metaPixel,
    };
};

export const readAdsLabels = (): AdsLabels => {
    const labels: AdsLabels = {};

    for (const pair of (trackingMeta()?.dataset['googleAdsConversions'] ?? '').split(',')) {
        const [conversion, label = ''] = pair.split('=');

        if (isAdsConversion(conversion) && label !== '') {
            labels[conversion] = label;
        }
    }

    return labels;
};

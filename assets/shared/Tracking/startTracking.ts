import { startGoogleTag } from '@/shared/Tracking/Google/googleTag';
import { startMetaPixel } from '@/shared/Tracking/Meta/metaPixel';
import { readAdsLabels, readEnabledTools } from '@/shared/Tracking/enabledTools';

export const startTracking = (): void => {
    const tools = readEnabledTools();

    if (tools !== undefined) {
        startGoogleTag(tools, readAdsLabels());
        startMetaPixel(tools);
    }
};

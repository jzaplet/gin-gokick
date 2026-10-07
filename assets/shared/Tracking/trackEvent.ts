import { trackGoogleEvent } from '@/shared/Tracking/Google/googleTag';
import { trackMetaEvent } from '@/shared/Tracking/Meta/metaPixel';
import { trackedEvents } from '@/shared/Tracking/trackedEvents';
import type { TrackedEvent } from '@/shared/Tracking/types/TrackedEvent';
import type { TrackedUser } from '@/shared/Tracking/types/TrackedUser';

export const trackEvent = async (event: TrackedEvent, user?: TrackedUser): Promise<void> => {
    const { ga4, ads, meta } = trackedEvents[event];

    await trackGoogleEvent(ga4, ads, user);
    await trackMetaEvent(meta, user);
};

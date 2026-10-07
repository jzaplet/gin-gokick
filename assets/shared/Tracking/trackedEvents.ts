import { AdsConversion } from '@/shared/Tracking/types/AdsConversion';
import type { EventTargets } from '@/shared/Tracking/types/EventTargets';
import type { TrackedEvent } from '@/shared/Tracking/types/TrackedEvent';

export const trackedEvents = {
    sign_up: {
        ga4: 'sign_up',
        ads: AdsConversion.SignUp,
        meta: 'CompleteRegistration',
    },
    login: {
        ga4: 'login',
        ads: null,
        meta: null,
    },
} satisfies Record<TrackedEvent, EventTargets>;

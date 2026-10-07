import type { MetaEvent } from '@/shared/Tracking/Meta/types/MetaEvent';
import type { AdsConversion } from '@/shared/Tracking/types/AdsConversion';

export type EventTargets = {
    ga4: string | null;
    ads: AdsConversion | null;
    meta: MetaEvent | null;
};

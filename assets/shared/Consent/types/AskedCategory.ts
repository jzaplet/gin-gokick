import type { ConsentCategory } from '@/shared/Consent/types/ConsentCategory';
import type { TrackingTool } from '@/shared/Tracking/types/TrackingTool';

export type AskedCategory = {
    category: ConsentCategory;
    tools: TrackingTool[];
};

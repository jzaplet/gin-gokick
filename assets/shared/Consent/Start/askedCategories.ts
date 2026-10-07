import type { AskedCategory } from '@/shared/Consent/types/AskedCategory';
import { ConsentCategory } from '@/shared/Consent/types/ConsentCategory';
import type { EnabledTools } from '@/shared/Tracking/types/EnabledTools';
import { TrackingTool } from '@/shared/Tracking/types/TrackingTool';

const toolCategories = {
    ga4: ConsentCategory.Analytics,
    google_ads: ConsentCategory.Marketing,
    meta_pixel: ConsentCategory.Marketing,
} satisfies Record<TrackingTool, ConsentCategory>;

const toolsIn = (tools: EnabledTools, category: ConsentCategory): TrackingTool[] =>
    Object.values(TrackingTool).filter((tool) => tools[tool] !== '' && toolCategories[tool] === category);

export const askedCategories = (tools: EnabledTools): AskedCategory[] =>
    Object.values(ConsentCategory)
        .map((category) => ({
            category,
            tools: toolsIn(tools, category),
        }))
        .filter((asked) => asked.tools.length > 0);

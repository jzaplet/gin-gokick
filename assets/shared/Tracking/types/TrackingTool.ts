export const TrackingTool = {
    GA4: 'ga4',
    GoogleAds: 'google_ads',
    MetaPixel: 'meta_pixel',
} as const;

export type TrackingTool = (typeof TrackingTool)[keyof typeof TrackingTool];

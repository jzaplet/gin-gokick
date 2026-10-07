export type GoogleConsent = Record<
    'analytics_storage' | 'ad_storage' | 'ad_user_data' | 'ad_personalization',
    'granted' | 'denied'
>;

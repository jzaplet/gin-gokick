import { trackGooglePageView } from '@/shared/Tracking/Google/googleTag';

export const trackPageView = (): void => {
    trackGooglePageView();
};

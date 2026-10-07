import * as Sentry from '@sentry/vue';

export const reportUnexpected = (message: string): void => {
    Sentry.captureMessage(message, 'error');
};

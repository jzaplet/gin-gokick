import * as Sentry from '@sentry/vue';

export const identifyUser = (id: string): void => {
    Sentry.setUser({ id });
};

export const forgetUser = (): void => {
    Sentry.setUser(null);
};

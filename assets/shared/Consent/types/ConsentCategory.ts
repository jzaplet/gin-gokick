export const ConsentCategory = {
    Analytics: 'analytics',
    Marketing: 'marketing',
} as const;

export type ConsentCategory = (typeof ConsentCategory)[keyof typeof ConsentCategory];

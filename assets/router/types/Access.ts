export const Access = {
    Guest: 'guest',
    User: 'user',
    Public: 'public',
} as const;

export type Access = (typeof Access)[keyof typeof Access];

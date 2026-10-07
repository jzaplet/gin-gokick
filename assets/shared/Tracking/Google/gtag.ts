import type { GtagCommand } from '@/shared/Tracking/Google/types/GtagCommand';

type Gtag = {
    push: (...command: GtagCommand) => void;
};

const dataLayer = (): unknown[] => {
    const existing: unknown = Reflect.get(window, 'dataLayer');

    if (Array.isArray(existing)) {
        return existing;
    }

    const created: unknown[] = [];

    Reflect.set(window, 'dataLayer', created);

    return created;
};

const commands: Gtag = {
    push(): void {
        dataLayer().push(arguments);
    },
};

export const gtag = commands.push;

import type { Message, MessagePart } from '@/shared/I18n/Dictionary/types/Message';

export type ArgKind = 'text' | 'plural' | 'select';

const noArgs: ReadonlyMap<string, ArgKind> = new Map();

const known = new WeakMap<readonly MessagePart[], ReadonlyMap<string, ArgKind>>();

const collect = (message: Message | undefined, kinds: Map<string, ArgKind>): void => {
    if (message === undefined || typeof message === 'string') {
        return;
    }

    for (const part of message) {
        if (typeof part === 'string' || part.type === 'pound') {
            continue;
        }

        kinds.set(part.name, part.type === 'arg' ? 'text' : part.type);

        if (part.type === 'plural') {
            Object.values(part.exact).forEach((inner) => {
                collect(inner, kinds);
            });
        }

        if (part.type !== 'arg') {
            Object.values(part.cases).forEach((inner) => {
                collect(inner, kinds);
            });
        }
    }
};

export const argKinds = (message: Message): ReadonlyMap<string, ArgKind> => {
    if (typeof message === 'string') {
        return noArgs;
    }

    const cached = known.get(message);

    if (cached !== undefined) {
        return cached;
    }

    const kinds = new Map<string, ArgKind>();

    collect(message, kinds);
    known.set(message, kinds);

    return kinds;
};

import type { Message, MessagePart, PluralPart, SelectPart } from '@/shared/I18n/Dictionary/types/Message';
import { argKinds } from '@/shared/I18n/Format/argKinds';
import { type PluralValue, pluralValue } from '@/shared/I18n/Format/pluralValue';
import { showNumber } from '@/shared/I18n/Format/showNumber';

type Context = {
    readonly tag: string;
    readonly texts: ReadonlyMap<string, string>;
    readonly numbers: ReadonlyMap<string, PluralValue>;
};

const read = (message: Message, params: Readonly<Record<string, unknown>>, tag: string): Context => {
    const kinds = argKinds(message);
    const unknown = Object.keys(params).find((name) => kinds.has(name) === false);
    const texts = new Map<string, string>();
    const numbers = new Map<string, PluralValue>();

    if (unknown !== undefined) {
        throw new TypeError(`The message takes no argument ${unknown}`);
    }

    for (const [name, kind] of kinds) {
        const value: unknown = params[name];

        if (Object.hasOwn(params, name) === false) {
            throw new TypeError(`The message misses the argument ${name}`);
        }

        if (kind === 'plural') {
            numbers.set(name, pluralValue(name, value));
        } else if (typeof value === 'string') {
            texts.set(name, value);
        } else {
            throw new TypeError(`${name} is a ${kind}, it takes a string`);
        }
    }

    return {
        tag,
        texts,
        numbers,
    };
};

const valueOf = <T>(values: ReadonlyMap<string, T>, name: string): T => {
    const value = values.get(name);

    if (value === undefined) {
        throw new TypeError(`The message misses the argument ${name}`);
    }

    return value;
};

const text = (context: Context, name: string): string => valueOf(context.texts, name);

const selectCase = (part: SelectPart, value: string): Message =>
    (Object.hasOwn(part.cases, value) ? part.cases[value] : undefined) ?? part.cases.other;

const writePlural = (part: PluralPart, context: Context): string => {
    const amount = valueOf(context.numbers, part.name);
    const shown = showNumber(context.tag, part.name, amount, part.offset);
    const exact = Number.isInteger(amount.value) ? part.exact[amount.value] : undefined;

    return write(exact ?? part.cases[shown.category] ?? part.cases.other, context, shown.text);
};

const writePart = (part: MessagePart, context: Context, pound: string): string => {
    if (typeof part === 'string') {
        return part;
    }

    switch (part.type) {
        case 'arg':
            return text(context, part.name);
        case 'pound':
            return pound;
        case 'select':
            return write(selectCase(part, text(context, part.name)), context, pound);
        case 'plural':
            return writePlural(part, context);
    }
};

const write = (message: Message, context: Context, pound: string): string =>
    typeof message === 'string' ? message : message.map((part) => writePart(part, context, pound)).join('');

export const formatMessage = (message: Message, params: Readonly<Record<string, unknown>>, tag: string): string =>
    write(message, read(message, params, tag), '');

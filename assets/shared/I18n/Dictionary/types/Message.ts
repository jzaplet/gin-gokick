import { arrayOf, isNumber, isRecord, isString } from '@/shared/TypeGuards/typeGuards';

export type ArgPart = {
    readonly type: 'arg';
    readonly name: string;
};

export type PoundPart = {
    readonly type: 'pound';
};

export type PluralPart = {
    readonly type: 'plural';
    readonly name: string;
    readonly offset: number;
    readonly exact: Readonly<Record<number, Message>>;
    readonly cases: Readonly<Partial<Record<Intl.LDMLPluralRule, Message>>> & { readonly other: Message };
};

export type SelectPart = {
    readonly type: 'select';
    readonly name: string;
    readonly cases: Readonly<Record<string, Message>> & { readonly other: Message };
};

export type MessagePart = string | ArgPart | PoundPart | PluralPart | SelectPart;

export type Message = string | readonly MessagePart[];

const pluralCategories: readonly string[] = [
    'zero',
    'one',
    'two',
    'few',
    'many',
    'other',
];

const exactNumber = /^(?:0|[1-9]\d*)$/;

const messagesBy = (v: unknown, isKey: (key: string) => boolean): boolean =>
    isRecord(v) && Object.entries(v).every(([key, inner]) => isKey(key) && isMessage(inner));

const casesBy = (v: unknown, isKey: (key: string) => boolean): boolean =>
    isRecord(v) && Object.hasOwn(v, 'other') && messagesBy(v, isKey);

const isPart = (v: unknown): v is MessagePart => {
    if (isString(v)) {
        return true;
    }

    if (isRecord(v) === false) {
        return false;
    }

    switch (v['type']) {
        case 'arg':
            return isString(v['name']);
        case 'pound':
            return true;
        case 'plural':
            return isString(v['name'])
                && isNumber(v['offset'])
                && messagesBy(v['exact'], (key) => exactNumber.test(key))
                && casesBy(v['cases'], (key) => pluralCategories.includes(key));
        case 'select':
            return isString(v['name']) && casesBy(v['cases'], () => true);
        default:
            return false;
    }
};

export const isMessage = (v: unknown): v is Message => isString(v) || arrayOf(isPart)(v);

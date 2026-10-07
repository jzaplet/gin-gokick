import { type PluralValue, safeNumber } from '@/shared/I18n/Format/pluralValue';

type ShownNumber = {
    text: string;
    category: Intl.LDMLPluralRule;
};

type NumberFormats = {
    readonly number: Intl.NumberFormat;
    readonly plural: Intl.PluralRules;
};

const formats = new Map<string, NumberFormats>();

const formatsOf = (tag: string, amount: PluralValue): NumberFormats => {
    const { minimumFractionDigits, maximumFractionDigits } = amount;
    const key = `${tag} ${String(minimumFractionDigits)} ${String(maximumFractionDigits)}`;
    const cached = formats.get(key);

    if (cached !== undefined) {
        return cached;
    }

    const options = {
        minimumFractionDigits,
        maximumFractionDigits,
    };
    const created = {
        number: new Intl.NumberFormat(tag, options),
        plural: new Intl.PluralRules(tag, options),
    };

    formats.set(key, created);

    return created;
};

export const showNumber = (tag: string, name: string, amount: PluralValue, offset: number): ShownNumber => {
    const value = safeNumber(name, amount.value - offset);
    const { number, plural } = formatsOf(tag, amount);

    return {
        text: number.format(value),
        category: plural.select(value),
    };
};

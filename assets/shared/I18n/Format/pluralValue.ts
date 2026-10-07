import { isNumber, isRecord } from '@/shared/TypeGuards/typeGuards';
import type { Decimal } from '@/shared/I18n/Dictionary/types/Decimal';

export type PluralValue = {
    readonly value: number;
    readonly minimumFractionDigits: number;
    readonly maximumFractionDigits: number;
};

const floatDigits = 3;

const maxDecimalDigits = 15;

const isDecimal = (value: unknown): value is Decimal =>
    isRecord(value) && isNumber(value['value']) && isNumber(value['digits']);

const validDigits = (digits: number): boolean => Number.isInteger(digits) && digits >= 0 && digits <= maxDecimalDigits;

export const safeNumber = (name: string, value: number): number => {
    if (Number.isFinite(value) === false || Math.abs(value) > Number.MAX_SAFE_INTEGER) {
        throw new RangeError(`${name} is ${String(value)}, a plural takes a finite number up to 2^53`);
    }

    return value;
};

export const pluralValue = (name: string, value: unknown): PluralValue => {
    if (typeof value === 'number') {
        return {
            value: safeNumber(name, value),
            minimumFractionDigits: 0,
            maximumFractionDigits: floatDigits,
        };
    }

    if (isDecimal(value) && validDigits(value.digits)) {
        const digits = value.digits;

        return {
            value: safeNumber(name, value.value),
            minimumFractionDigits: digits,
            maximumFractionDigits: digits,
        };
    }

    throw new TypeError(
        `${name} is a plural, it takes a number or a Decimal with 0 to ${String(maxDecimalDigits)} digits`,
    );
};

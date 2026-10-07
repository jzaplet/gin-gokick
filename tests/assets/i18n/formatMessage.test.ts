import { describe, expect, it } from 'vitest';
import type { Message, PluralPart } from '@/shared/I18n/Dictionary/types/Message';
import { formatMessage } from '@/shared/I18n/Format/formatMessage';
import { plural, pound } from './plural';

const written: Message = [{
    type: 'select',
    name: 'g',
    cases: {
        female: 'Napsala',
        other: 'Napsal',
    },
}];

describe('formatMessage', () => {
    it('fills in a text', () => {
        expect(
            formatMessage(
                [
                    'Ahoj ',
                    {
                        type: 'arg',
                        name: 'name',
                    },
                    '!',
                ],
                { name: 'Jano' },
                'cs-CZ',
            ),
        ).toBe('Ahoj Jano!');
    });

    it('picks a select case and falls back to other', () => {
        expect(formatMessage(written, { g: 'female' }, 'cs-CZ')).toBe('Napsala');
        expect(formatMessage(written, { g: 'unknown' }, 'cs-CZ')).toBe('Napsal');
        expect(formatMessage(written, { g: 'toString' }, 'cs-CZ')).toBe('Napsal');
    });

    it('gives a pound the number of its own plural', () => {
        const minutes: PluralPart = {
            type: 'plural',
            name: 'm',
            offset: 0,
            exact: {},
            cases: { other: [pound] },
        };
        const nested = plural({
            other: [pound, ' ', {
                type: 'select',
                name: 'g',
                cases: { other: [minutes] },
            }],
        });

        expect(
            formatMessage(
                nested,
                {
                    n: 3,
                    g: 'a',
                    m: 7,
                },
                'cs-CZ',
            ),
        ).toBe('3 7');
    });

    it.each([
        [
            'a missing text',
            ['Ahoj ', {
                type: 'arg',
                name: 'name',
            }],
            {},
        ],
        [
            'a missing text in a case not taken',
            [{
                type: 'select',
                name: 'g',
                cases: {
                    a: [{
                        type: 'arg',
                        name: 'name',
                    }],
                    other: '',
                },
            }],
            { g: 'b' },
        ],
        [
            'an unknown argument',
            'Ahoj',
            { name: 'Jano' },
        ],
        [
            'a number as a text',
            [{
                type: 'arg',
                name: 'name',
            }],
            { name: 1 },
        ],
        [
            'undefined as a text',
            [{
                type: 'arg',
                name: 'name',
            }],
            { name: undefined },
        ],
        [
            'a number as a select',
            written,
            { g: 1 },
        ],
    ] satisfies [string, Message, Record<string, unknown>][])('refuses %s', (_, message, params) => {
        expect(() => formatMessage(message, params, 'cs-CZ')).toThrow(TypeError);
    });

    it.each([
        [
            'a text',
            '1',
            TypeError,
        ],
        [
            'null',
            null,
            TypeError,
        ],
        [
            'NaN',
            Number.NaN,
            RangeError,
        ],
        [
            'infinity',
            Number.POSITIVE_INFINITY,
            RangeError,
        ],
        [
            'a whole number beyond 2^53',
            2 ** 53,
            RangeError,
        ],
        [
            'a decimal beyond 2^53',
            1e16,
            RangeError,
        ],
        [
            'a decimal with too many digits',
            {
                value: 1,
                digits: 16,
            },
            TypeError,
        ],
        [
            'a decimal with negative digits',
            {
                value: 1,
                digits: -1,
            },
            TypeError,
        ],
        [
            'a decimal with a fraction of digits',
            {
                value: 1,
                digits: 1.5,
            },
            TypeError,
        ],
        [
            'a fixed decimal beyond 2^53',
            {
                value: 1e16,
                digits: 0,
            },
            RangeError,
        ],
    ])('refuses %s as a plural value', (_, n, error) => {
        expect(() => formatMessage(plural({ other: [pound] }), { n }, 'cs-CZ')).toThrow(error);
    });

    it('refuses a value the offset moves beyond 2^53', () => {
        const message = plural({ other: [pound] }, { offset: Number.MAX_SAFE_INTEGER });

        expect(() => formatMessage(message, { n: -Number.MAX_SAFE_INTEGER }, 'cs-CZ')).toThrow(RangeError);
    });
});

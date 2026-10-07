import { describe, expect, it } from 'vitest';
import { formatMessage } from '@/shared/I18n/Format/formatMessage';
import { plural, pound } from './plural';

const czechOthers = plural(
    {
        one: [
            'vy a ',
            pound,
            ' další',
        ],
        few: [
            'vy a ',
            pound,
            ' další',
        ],
        many: [
            'vy a ',
            pound,
            ' dalšího',
        ],
        other: [
            'vy a ',
            pound,
            ' dalších',
        ],
    },
    {
        offset: 1,
        exact: {
            0: 'nikdo',
            1: 'jen vy',
        },
    },
);

describe('formatMessage with a plural', () => {
    it.each([
        [
            0,
            'nikdo',
        ],
        [
            1,
            'jen vy',
        ],
        [
            2,
            'vy a 1 další',
        ],
        [
            3,
            'vy a 2 další',
        ],
        [
            6,
            'vy a 5 dalších',
        ],
    ])('subtracts the offset from %s after the exact cases', (n, want) => {
        expect(formatMessage(czechOthers, { n }, 'cs-CZ')).toBe(want);
    });

    it.each([
        [
            1,
            'jeden',
        ],
        [
            {
                value: 1,
                digits: 1,
            },
            'jeden',
        ],
        [
            1.0004,
            '1',
        ],
    ])('matches an exact case on %o before rounding', (n, want) => {
        expect(formatMessage(plural({ other: [pound] }, { exact: { 1: 'jeden' } }), { n }, 'cs-CZ')).toBe(want);
    });

    it('subtracts the offset before rounding', () => {
        expect(formatMessage(plural({ other: [pound] }, { offset: 1 }), { n: 2.5 }, 'cs-CZ')).toBe('1,5');
        expect(formatMessage(plural({ other: [pound] }, { offset: 1 }), { n: 1.0005 }, 'cs-CZ')).toBe('0');
    });
});

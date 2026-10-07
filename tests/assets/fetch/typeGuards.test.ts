import { describe, expect, it } from 'vitest';
import { isApiFieldErrors } from '@/shared/Fetch/Envelope/ApiFieldErrors';
import { isApiGeneralError } from '@/shared/Fetch/Envelope/ApiGeneralError';
import { isApiMessage } from '@/shared/Fetch/Envelope/ApiMessage';
import { arrayOf, isNull, isRecord, isString, nullable, optional } from '@/shared/TypeGuards/typeGuards';

describe('guards', () => {
    it('isRecord accepts objects only', () => {
        expect(
            [
                {},
                { a: 1 },
            ].every(isRecord),
        ).toBe(true);
        expect(
            [
                null,
                [],
                'a',
                1,
                undefined,
            ].some(isRecord),
        ).toBe(false);
    });

    it('arrayOf checks every element', () => {
        expect(
            arrayOf(isString)([
                'a',
                'b',
            ]),
        ).toBe(true);
        expect(
            arrayOf(isString)([
                'a',
                1,
            ]),
        ).toBe(false);
        expect(arrayOf(isString)(null)).toBe(false);
    });

    it('nullable adds null to a guard', () => {
        expect(nullable(isString)(null)).toBe(true);
        expect(nullable(isString)('a')).toBe(true);
        expect(nullable(isString)(1)).toBe(false);
        expect(isNull(undefined)).toBe(false);
    });

    it('optional adds undefined to a guard', () => {
        expect(optional(isString)(undefined)).toBe(true);
        expect(optional(isString)('a')).toBe(true);
        expect(optional(isString)(null)).toBe(false);
    });

    it('checks the API message and the general error', () => {
        expect(
            isApiMessage({
                key: 'validation.min',
                params: { min: 12 },
            }),
        ).toBe(true);
        expect(
            isApiMessage({
                key: 'validation.min',
                params: 12,
            }),
        ).toBe(false);
        expect(isApiMessage({ params: {} })).toBe(false);
        expect(isApiGeneralError({ general: { key: 'auth.required' } })).toBe(true);
        expect(isApiGeneralError({ email: { key: 'validation.email' } })).toBe(false);
    });

    it('accepts field errors only under the paths generated from Go', () => {
        const order = /^(?:general|billing|billing\.street|items|items\[\d+\]|items\[\d+\]\.name|note)$/;
        const required = { key: 'validation.required' };
        const strangers = [
            'xnote',
            'notex',
            'items[1].name.x',
            'items[x].name',
            'items[].name',
            'billing_street',
        ];

        expect(
            isApiFieldErrors(order, {
                'billing.street': required,
                'items[1].name': required,
            }),
        ).toBe(true);
        expect(isApiFieldErrors(order, { general: required })).toBe(true);
        expect(strangers.some((key) => isApiFieldErrors(order, { [key]: required }))).toBe(false);
        expect(isApiFieldErrors(order, { note: { params: {} } })).toBe(false);
        expect(
            [
                {},
                [required],
                null,
                'note',
            ].some((v) => isApiFieldErrors(order, v)),
        ).toBe(false);
    });
});

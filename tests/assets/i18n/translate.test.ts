import { beforeEach, describe, expect, expectTypeOf, it, vi } from 'vitest';
import type { MessageKey } from '@/shared/I18n/Dictionary/MessageParams';
import type { Decimal } from '@/shared/I18n/Dictionary/types/Decimal';
import { type MessageArgs, t, tm } from '@/shared/I18n/Texts/translate';
import { reportUnexpected } from '@/shared/Sentry/reportUnexpected';

vi.mock('@/shared/Sentry/reportUnexpected', () => ({ reportUnexpected: vi.fn() }));

describe('t', () => {
    it('picks the Czech form of the number', () => {
        expect(t('validation.min_length', { min: 1 })).toBe('Zadejte alespoň 1 znak.');
        expect(t('validation.min_length', { min: 3 })).toBe('Zadejte alespoň 3 znaky.');
        expect(t('validation.min_length', { min: 12 })).toBe('Zadejte alespoň 12 znaků.');
        expect(
            t('validation.min_length', {
                min: {
                    value: 1.5,
                    digits: 1,
                },
            }),
        ).toBe('Zadejte alespoň 1,5 znaku.');
    });

    it('takes exactly the params of its key', () => {
        expectTypeOf<Parameters<typeof t<'register.submit'>>>().toEqualTypeOf<['register.submit']>();
        expectTypeOf<Parameters<typeof t<'consent.lead'>>>().toEqualTypeOf<['consent.lead', { button: string }]>();
        expectTypeOf<Parameters<typeof t<'validation.min_length'>>>()
            .toEqualTypeOf<['validation.min_length', { min: Decimal | number }]>();
        expectTypeOf<MessageArgs<'register.submit' | 'consent.lead'>>().toEqualTypeOf<never>();
        expectTypeOf<'register.unknown'>().not.toExtend<MessageKey>();
    });
});

describe('tm', () => {
    beforeEach(() => {
        vi.mocked(reportUnexpected).mockClear();
    });

    it('shows and reports a message whose params do not fit', () => {
        expect(tm({ key: 'validation.min_length' })).toBe('validation.min_length');
        expect(
            tm({
                key: 'validation.min_length',
                params: { min: '12' },
            }),
        ).toBe('validation.min_length');
        expect(
            tm({
                key: 'validation.email',
                params: { min: 12 },
            }),
        ).toBe('validation.email');
        expect(reportUnexpected).toHaveBeenCalledTimes(3);
        expect(reportUnexpected).toHaveBeenCalledWith(
            'Message validation.min_length cannot be formatted: TypeError: The message misses the argument min',
        );
    });

    it('shows and reports a key without a text', () => {
        expect(tm({ key: 'user.unknown' })).toBe('user.unknown');
        expect(tm({ key: 'toString' })).toBe('toString');
        expect(reportUnexpected).toHaveBeenCalledTimes(2);
        expect(reportUnexpected).toHaveBeenCalledWith('Message key without a text: user.unknown');
    });
});

import { describe, expect, it } from 'vitest';
import { formatMessage } from '@/shared/I18n/Format/formatMessage';
import { message, type NumberSample } from './numbers/sample';

type LocaleNumbers = {
    readonly tag: string;
    readonly samples: readonly NumberSample[];
};

const locales = import.meta.glob<LocaleNumbers>('./numbers/*_*.ts', { eager: true });

describe.each(Object.values(locales))('Intl in $tag', ({ tag, samples }) => {
    it.each(samples)('shows %o like Go', (n, want) => {
        expect(formatMessage(message, { n }, tag)).toBe(want);
    });
});

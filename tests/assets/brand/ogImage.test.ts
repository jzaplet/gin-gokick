import { describe, expect, it } from 'vitest';
import { dictionaries } from '@/shared/I18n/Dictionary/dictionaries';

const images = import.meta.glob<string>('@/img/og/*.png', {
    eager: true,
    query: '?inline',
    import: 'default',
});

const byLocale = new Map(
    Object.entries(images).map(([path, data]) => [
        path.slice(path.lastIndexOf('/') + 1, -'.png'.length),
        data,
    ]),
);

const sizeOf = (data: string): number[] => {
    const bytes = Uint8Array.from(atob(data.slice(data.indexOf(',') + 1)), (char) => char.charCodeAt(0));
    const header = new DataView(bytes.buffer);

    return [
        header.getUint32(16),
        header.getUint32(20),
    ];
};

describe('the images for sharing', () => {
    it('cover exactly the locales of the dictionaries', () => {
        expect([...byLocale.keys()].toSorted()).toEqual(Object.keys(dictionaries).toSorted());
    });

    it.each([...byLocale])('%s has the size the layout declares', (_, data) => {
        expect(data).toMatch(/^data:image\/png;base64,/);
        expect(sizeOf(data)).toEqual([
            1200,
            630,
        ]);
    });
});

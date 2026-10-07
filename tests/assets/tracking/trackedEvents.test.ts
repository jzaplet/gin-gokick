import { describe, expect, it } from 'vitest';
import { AdsConversion } from '@/shared/Tracking/types/AdsConversion';
import { TrackedEvent } from '@/shared/Tracking/types/TrackedEvent';

const sources = import.meta.glob<string>('/assets/**/*.{ts,vue}', {
    query: '?raw',
    import: 'default',
    eager: true,
});

const usedIn = (reference: string): string[] =>
    Object.keys(sources).filter((path) => sources[path]?.includes(reference) === true);

describe('tracked events', () => {
    it.each(Object.keys(TrackedEvent))('are all tracked somewhere: TrackedEvent.%s', (key) => {
        expect(usedIn(`TrackedEvent.${key}`)).not.toEqual([]);
    });

    it.each(Object.keys(AdsConversion))('are all conversions of an event: AdsConversion.%s', (key) => {
        expect(usedIn(`AdsConversion.${key}`)).toEqual(['/assets/shared/Tracking/trackedEvents.ts']);
    });
});

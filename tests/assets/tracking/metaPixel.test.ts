import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ConsentChoice } from '@/shared/Consent/types/ConsentChoice';
import { ScriptURL } from '@/shared/ScriptLoader/types/ScriptURL';

type Tracking = {
    choose: (choice: ConsentChoice) => void;
    track: () => void;
};

const pixel = '1234567890123456';

const fbevents = ScriptURL.MetaPixel;

const installed = (): unknown => Reflect.get(window, 'fbq');

const queue = (): unknown => {
    const fbq = installed();

    return typeof fbq === 'function' ? Reflect.get(fbq, 'queue') : undefined;
};

const scripts = (): string[] => Array.from(document.head.querySelectorAll('script'), (script) => script.src);

const start = async (ids: Record<string, string>): Promise<Tracking> => {
    vi.resetModules();

    const meta = document.createElement('meta');

    meta.name = 'tracking';
    Object.assign(meta.dataset, ids);
    document.head.append(meta);

    const { startTracking } = await import('@/shared/Tracking/startTracking');
    const { setCurrentChoice } = await import('@/shared/Consent/Choice/currentChoice');
    const { fbq } = await import('@/shared/Tracking/Meta/fbq');

    const track = (): void => {
        fbq('track', 'PageView');
    };

    startTracking();

    return {
        choose: setCurrentChoice,
        track,
    };
};

afterEach(() => {
    document.head.replaceChildren();
    Reflect.deleteProperty(window, 'fbq');
    Reflect.deleteProperty(window, '_fbq');
    Reflect.deleteProperty(window, 'dataLayer');
});

describe('Meta Pixel', () => {
    it.each([
        [
            'without consent to marketing',
            { metaPixel: pixel },
            false,
        ],
        [
            'without its ID',
            { googleAds: 'AW-123456789' },
            true,
        ],
    ])('stays out %s', async (_, ids, marketing) => {
        const { choose } = await start(ids);

        choose({ marketing });

        expect(installed()).toBeUndefined();
        expect(scripts()).not.toContain(fbevents);
    });

    it('queues the pixel and its page view, then loads fbevents.js once', async () => {
        const { choose } = await start({ metaPixel: pixel });

        choose({ marketing: true });
        choose({ marketing: true });

        const fbq = installed();

        expect(queue()).toEqual([
            [
                'init',
                pixel,
            ],
            [
                'track',
                'PageView',
            ],
        ]);
        expect(scripts()).toEqual([fbevents]);
        expect(Reflect.get(window, '_fbq')).toBe(fbq);
        expect(fbq).toBeTypeOf('function');
        expect(typeof fbq === 'function' && Reflect.get(fbq, 'push')).toBe(fbq);
        expect(typeof fbq === 'function' && Reflect.get(fbq, 'loaded')).toBe(true);
        expect(typeof fbq === 'function' && Reflect.get(fbq, 'version')).toBe('2.0');
    });

    it('hands commands to fbevents.js once it has loaded', async () => {
        const { choose, track } = await start({ metaPixel: pixel });
        const callMethod = vi.fn();

        choose({ marketing: true });
        Reflect.set(Object(installed()), 'callMethod', callMethod);
        track();

        expect(callMethod).toHaveBeenCalledExactlyOnceWith('track', 'PageView');
        expect(queue()).toHaveLength(2);
    });
});

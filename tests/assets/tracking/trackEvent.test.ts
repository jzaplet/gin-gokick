import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ConsentChoice } from '@/shared/Consent/types/ConsentChoice';
import type { TrackedEvent } from '@/shared/Tracking/types/TrackedEvent';
import type { TrackedUser } from '@/shared/Tracking/types/TrackedUser';

type Tracking = {
    choose: (choice: ConsentChoice) => void;
    track: (event: TrackedEvent, user?: TrackedUser) => Promise<void>;
};

const ga4 = 'G-AB12CD34EF';

const ads = 'AW-123456789';

const pixel = '1234567890123456';

const allTools = {
    ga4,
    googleAds: ads,
    metaPixel: pixel,
    googleAdsConversions: 'sign_up=AbC-D_efG',
};

const isArguments = (value: unknown): value is IArguments =>
    Object.prototype.toString.call(value) === '[object Arguments]';

const commands = (): unknown[][] => {
    const layer: unknown = Reflect.get(window, 'dataLayer');

    return Array.isArray(layer)
        ? layer.map((entry: unknown) => (isArguments(entry) ? Array.from(entry, (item: unknown) => item) : []))
        : [];
};

const googleEvents = (): unknown[][] => commands().filter(([name, event]) => name === 'event' && event !== 'page_view');

const userData = (): unknown[][] => commands().filter(([name, key]) => name === 'set' && key === 'user_data');

const metaQueue = (): unknown => {
    const fbq: unknown = Reflect.get(window, 'fbq');

    return typeof fbq === 'function' ? Reflect.get(fbq, 'queue') : undefined;
};

const start = async (ids: Record<string, string>): Promise<Tracking> => {
    vi.resetModules();

    const meta = document.createElement('meta');

    meta.name = 'tracking';
    Object.assign(meta.dataset, ids);
    document.head.append(meta);

    const { startTracking } = await import('@/shared/Tracking/startTracking');
    const { setCurrentChoice } = await import('@/shared/Consent/Choice/currentChoice');
    const { trackEvent } = await import('@/shared/Tracking/trackEvent');

    startTracking();

    return {
        choose: setCurrentChoice,
        track: trackEvent,
    };
};

afterEach(() => {
    document.head.replaceChildren();
    Reflect.deleteProperty(window, 'dataLayer');
    Reflect.deleteProperty(window, 'fbq');
    Reflect.deleteProperty(window, '_fbq');
});

describe('trackEvent', () => {
    it('sends nothing before consent and keeps the consent defaults first', async () => {
        const { choose, track } = await start(allTools);

        await track('sign_up');

        expect(Reflect.get(window, 'dataLayer')).toBeUndefined();
        expect(Reflect.get(window, 'fbq')).toBeUndefined();

        choose({
            analytics: true,
            marketing: true,
        });

        expect(commands()[0]?.[0]).toBe('consent');
        expect(googleEvents()).toEqual([]);
        expect(metaQueue()).toEqual([
            [
                'init',
                pixel,
            ],
            [
                'track',
                'PageView',
            ],
        ]);
    });

    it('sends a sign-up to GA4, as a conversion to Google Ads and to Meta', async () => {
        const { choose, track } = await start(allTools);

        choose({
            analytics: true,
            marketing: true,
        });
        await track('sign_up');

        expect(googleEvents()).toEqual([
            [
                'event',
                'sign_up',
                { send_to: [ga4] },
            ],
            [
                'event',
                'conversion',
                { send_to: [`${ads}/AbC-D_efG`] },
            ],
        ]);
        expect(metaQueue()).toEqual([
            [
                'init',
                pixel,
            ],
            [
                'track',
                'PageView',
            ],
            [
                'track',
                'CompleteRegistration',
            ],
        ]);
    });

    it('sends each tool only what its category allows', async () => {
        const analytics = await start(allTools);

        analytics.choose({
            analytics: true,
            marketing: false,
        });
        await analytics.track('sign_up');

        expect(googleEvents()).toEqual([[
            'event',
            'sign_up',
            { send_to: [ga4] },
        ]]);
        expect(metaQueue()).toBeUndefined();

        document.head.replaceChildren();
        Reflect.deleteProperty(window, 'dataLayer');

        const marketing = await start(allTools);

        marketing.choose({
            analytics: false,
            marketing: true,
        });
        await marketing.track('sign_up');

        expect(googleEvents()).toEqual([[
            'event',
            'conversion',
            { send_to: [`${ads}/AbC-D_efG`] },
        ]]);
        expect(metaQueue()).toContainEqual([
            'track',
            'CompleteRegistration',
        ]);
    });

    it('sends a login to GA4 only', async () => {
        const { choose, track } = await start(allTools);

        choose({
            analytics: true,
            marketing: true,
        });
        await track('login');

        expect(googleEvents()).toEqual([[
            'event',
            'login',
            { send_to: [ga4] },
        ]]);
        expect(metaQueue()).toEqual([
            [
                'init',
                pixel,
            ],
            [
                'track',
                'PageView',
            ],
        ]);
    });

    it('adds the email, hashed for each tool, to the Ads conversion and to Meta', async () => {
        const { choose, track } = await start(allTools);

        choose({
            analytics: true,
            marketing: true,
        });
        await track('sign_up', { email: ' Jan.Novak@Gmail.com ' });

        const googleHash = '005ed88a887dbd4c32e8d7ca3665981df82512b3a7fbf451328ef0a46835d803';
        const metaHash = '7d4f91fdbfb1c424c7b2b762fb811d6493d9290dc167732be776e70c4ae72c2d';

        expect(commands().slice(-2)).toEqual([
            [
                'set',
                'user_data',
                { sha256_email_address: googleHash },
            ],
            [
                'event',
                'conversion',
                { send_to: [`${ads}/AbC-D_efG`] },
            ],
        ]);
        expect(metaQueue()).toEqual([
            [
                'init',
                pixel,
            ],
            [
                'track',
                'PageView',
            ],
            [
                'init',
                pixel,
                { em: metaHash },
            ],
            [
                'track',
                'CompleteRegistration',
            ],
        ]);
    });

    it('keeps the email out without consent to marketing', async () => {
        const { choose, track } = await start(allTools);

        choose({
            analytics: true,
            marketing: false,
        });
        await track('sign_up', { email: 'jan.novak@gmail.com' });

        expect(userData()).toEqual([]);
        expect(metaQueue()).toBeUndefined();
    });
});

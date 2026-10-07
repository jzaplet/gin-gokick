import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ConsentChoice } from '@/shared/Consent/types/ConsentChoice';
import { ScriptURL } from '@/shared/ScriptLoader/types/ScriptURL';

type Tracking = {
    choose: (choice: ConsentChoice) => void;
    navigate: (path: string) => Promise<void>;
};

const ga4 = 'G-AB12CD34EF';

const ads = 'AW-123456789';

const denied = {
    analytics_storage: 'denied',
    ad_storage: 'denied',
    ad_user_data: 'denied',
    ad_personalization: 'denied',
};

const analyticsOnly = {
    ...denied,
    analytics_storage: 'granted',
};

const adsOnly = {
    ...denied,
    ad_storage: 'granted',
    ad_user_data: 'granted',
    ad_personalization: 'granted',
};

const allGranted = {
    ...adsOnly,
    analytics_storage: 'granted',
};

const isArguments = (value: unknown): value is IArguments =>
    Object.prototype.toString.call(value) === '[object Arguments]';

const commands = (): unknown[][] => {
    const layer: unknown = Reflect.get(window, 'dataLayer');

    if (Array.isArray(layer) === false) {
        return [];
    }

    return layer.map((entry: unknown) => {
        if (isArguments(entry) === false) {
            throw new Error('gtag.js reads only arguments objects');
        }

        return Array.from(entry, (item: unknown) => item);
    });
};

const pageViews = (): unknown[][] => commands().filter(([name]) => name === 'set');

const scripts = (): string[] => Array.from(document.head.querySelectorAll('script'), (script) => script.src);

const page = (path: string): string => location.origin + path;

const start = async (ids: Record<string, string>): Promise<Tracking> => {
    vi.resetModules();

    const meta = document.createElement('meta');

    meta.name = 'tracking';
    Object.assign(meta.dataset, ids);
    document.head.append(meta);

    const { loadPageDictionary } = await import('@/shared/I18n/Texts/pageDictionary');
    const { startTracking } = await import('@/shared/Tracking/startTracking');
    const { setCurrentChoice } = await import('@/shared/Consent/Choice/currentChoice');
    const { router } = await import('@/router');

    await loadPageDictionary();

    const navigate = async (path: string): Promise<void> => {
        await router.push(path);
    };

    startTracking();

    return {
        choose: setCurrentChoice,
        navigate,
    };
};

afterEach(() => {
    document.head.replaceChildren();
    Reflect.deleteProperty(window, 'dataLayer');
    history.replaceState({}, '', '/');
});

describe('Google tag', () => {
    it('loads nothing before consent', async () => {
        const { choose, navigate } = await start({
            ga4,
            googleAds: ads,
        });

        choose({
            analytics: false,
            marketing: false,
        });
        await navigate('/cs/app/register');

        expect(commands()).toEqual([]);
        expect(scripts()).toEqual([]);
    });

    it('starts GA4 after analytics consent, with the query of the page and without its fragment', async () => {
        history.replaceState({}, '', '/cs/app/login?redirect=%2Fdashboard#form');
        const { choose } = await start({
            ga4,
            googleAds: ads,
        });

        choose({
            analytics: true,
            marketing: false,
        });
        choose({
            analytics: true,
            marketing: false,
        });

        expect(commands()).toEqual([
            [
                'consent',
                'default',
                denied,
            ],
            [
                'js',
                expect.any(Date),
            ],
            [
                'consent',
                'update',
                analyticsOnly,
            ],
            [
                'set',
                { page_location: page('/cs/app/login?redirect=%2Fdashboard') },
            ],
            [
                'config',
                ga4,
                { send_page_view: false },
            ],
            [
                'event',
                'page_view',
                { send_to: [ga4] },
            ],
        ]);
        expect(scripts()).toEqual([`${ScriptURL.GoogleTag}?id=${ga4}`]);
    });

    it('adds Google Ads to the loaded gtag.js when marketing is granted later', async () => {
        const { choose } = await start({
            ga4,
            googleAds: ads,
        });

        choose({
            analytics: true,
            marketing: false,
        });
        choose({
            analytics: true,
            marketing: true,
        });

        expect(commands().slice(6)).toEqual([
            [
                'consent',
                'update',
                allGranted,
            ],
            [
                'set',
                { page_location: page('/') },
            ],
            [
                'config',
                ads,
                { send_page_view: false },
            ],
            [
                'event',
                'page_view',
                { send_to: [ads] },
            ],
        ]);
        expect(scripts()).toEqual([`${ScriptURL.GoogleTag}?id=${ga4}`]);
    });

    it('loads gtag.js once for both tools granted together', async () => {
        const { choose } = await start({
            ga4,
            googleAds: ads,
        });

        choose({
            analytics: true,
            marketing: true,
        });

        expect(commands().slice(2)).toEqual([
            [
                'consent',
                'update',
                allGranted,
            ],
            [
                'set',
                { page_location: page('/') },
            ],
            [
                'config',
                ga4,
                { send_page_view: false },
            ],
            [
                'config',
                ads,
                { send_page_view: false },
            ],
            ['event', 'page_view', {
                send_to: [
                    ga4,
                    ads,
                ],
            }],
        ]);
        expect(scripts()).toEqual([`${ScriptURL.GoogleTag}?id=${ga4}`]);
    });

    it('loads gtag.js for Google Ads alone', async () => {
        const { choose } = await start({ googleAds: ads });

        choose({ marketing: true });

        expect(commands()).toEqual([
            [
                'consent',
                'default',
                denied,
            ],
            [
                'js',
                expect.any(Date),
            ],
            [
                'consent',
                'update',
                adsOnly,
            ],
            [
                'set',
                { page_location: page('/') },
            ],
            [
                'config',
                ads,
                { send_page_view: false },
            ],
            [
                'event',
                'page_view',
                { send_to: [ads] },
            ],
        ]);
        expect(scripts()).toEqual([`${ScriptURL.GoogleTag}?id=${ads}`]);
    });

    it('grants Google only the categories of its own tools', async () => {
        const { choose } = await start({
            ga4,
            metaPixel: '1234567890',
        });

        choose({
            analytics: false,
            marketing: true,
        });

        expect(commands()).toEqual([]);

        choose({
            analytics: true,
            marketing: true,
        });

        expect(commands()).toContainEqual([
            'consent',
            'update',
            analyticsOnly,
        ]);
    });

    it('counts every page of the SPA once', async () => {
        history.replaceState({}, '', '/cs/app/login?redirect=%2Fdashboard');
        const { choose, navigate } = await start({ ga4 });

        choose({ analytics: true });
        await navigate('/cs/app/login?redirect=%2Fdashboard');
        await navigate('/cs/app/register#form');
        await navigate('/cs/app/register');
        await navigate('/cs/app/register?ref=newsletter');

        expect(pageViews()).toEqual([
            [
                'set',
                { page_location: page('/cs/app/login?redirect=%2Fdashboard') },
            ],
            [
                'set',
                { page_location: page('/cs/app/register') },
            ],
            [
                'set',
                { page_location: page('/cs/app/register?ref=newsletter') },
            ],
        ]);
        expect(commands().at(-1)).toEqual([
            'event',
            'page_view',
            { send_to: [ga4] },
        ]);
    });
});

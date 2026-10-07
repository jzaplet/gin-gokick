const browsers: Readonly<Record<string, RegExp>> = {
    'Edge': /Edg(?:e|A|iOS)?\//u,
    'Opera': /OPR\/|Opera/u,
    'Samsung Internet': /SamsungBrowser\//u,
    'Firefox': /Firefox\/|FxiOS\//u,
    'Chrome': /Chrome\/|CriOS\//u,
    'Safari': /Version\/[\d.]+.*Safari\//u,
};

const systems: Readonly<Record<string, RegExp>> = {
    Windows: /Windows/u,
    iOS: /iPhone|iPod/u,
    iPadOS: /iPad/u,
    Android: /Android/u,
    ChromeOS: /CrOS/u,
    macOS: /Mac OS X|Macintosh/u,
    Linux: /Linux/u,
};

const first = (names: Readonly<Record<string, RegExp>>, userAgent: string): string | undefined =>
    Object.entries(names).find(([, pattern]) => pattern.test(userAgent))?.[0];

export const deviceName = (userAgent: string): string | null => {
    const parts = [
        first(browsers, userAgent),
        first(systems, userAgent),
    ].filter((part) => part !== undefined);

    return parts.length === 0 ? null : parts.join(' · ');
};

const files = import.meta.glob<string>('@/img/flags/*.svg', {
    eager: true,
    import: 'default',
});

const byRegion = new Map(
    Object.entries(files).map(([path, url]) => [
        path.slice(path.lastIndexOf('/') + 1, -'.svg'.length),
        url,
    ]),
);

export const flagOf = (locale: string): string | undefined => byRegion.get(locale.slice(3).toLowerCase());

export const doubleTrack = (track: Element): HTMLElement[] => {
    const copies: HTMLElement[] = [];

    for (const item of [...track.children]) {
        const copy = item.cloneNode(true);

        if (copy instanceof HTMLElement) {
            copy.setAttribute('aria-hidden', 'true');
            copy.setAttribute('inert', '');
            copies.push(copy);
        }
    }

    track.append(...copies);

    return copies;
};

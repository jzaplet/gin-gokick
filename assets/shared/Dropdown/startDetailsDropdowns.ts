const OPEN_DROPDOWNS = 'details[data-dropdown][open]';

const closeOpenDropdowns = (keep: (dropdown: HTMLDetailsElement) => boolean): void => {
    for (const dropdown of document.querySelectorAll<HTMLDetailsElement>(OPEN_DROPDOWNS)) {
        if (keep(dropdown) === false) {
            dropdown.open = false;
        }
    }
};

const onSummary = (dropdown: HTMLDetailsElement, target: EventTarget | null): boolean =>
    target instanceof Node && dropdown.querySelector(':scope > summary')?.contains(target) === true;

const closeOnClick = (event: MouseEvent): void => {
    closeOpenDropdowns((dropdown) => onSummary(dropdown, event.target));
};

const closeOnEscape = (event: KeyboardEvent): void => {
    if (event.key === 'Escape') {
        closeOpenDropdowns(() => false);
    }
};

export const startDetailsDropdowns = (): (() => void) => {
    document.addEventListener('click', closeOnClick);
    document.addEventListener('keydown', closeOnEscape);

    return () => {
        document.removeEventListener('click', closeOnClick);
        document.removeEventListener('keydown', closeOnEscape);
    };
};

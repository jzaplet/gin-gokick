export type Debounce = {
    run: (action: () => void) => void;
    cancel: () => void;
};

export const createDebounce = (delay: number): Debounce => {
    let timer: ReturnType<typeof setTimeout> | undefined;

    const cancel = (): void => {
        clearTimeout(timer);
        timer = undefined;
    };

    const run = (action: () => void): void => {
        cancel();
        timer = setTimeout(() => {
            timer = undefined;
            action();
        }, delay);
    };

    return {
        run,
        cancel,
    };
};

export const whenIdle = (work: () => void): (() => void) => {
    let idle: number | undefined;
    let timeout: number | undefined;

    const schedule = (): void => {
        if (typeof window.requestIdleCallback === 'function') {
            idle = window.requestIdleCallback(work, { timeout: 1500 });

            return;
        }

        timeout = window.setTimeout(work, 200);
    };

    if (document.readyState === 'complete') {
        schedule();
    } else {
        window.addEventListener('load', schedule, { once: true });
    }

    return () => {
        window.removeEventListener('load', schedule);

        if (idle !== undefined) {
            window.cancelIdleCallback(idle);
        }

        window.clearTimeout(timeout);
    };
};

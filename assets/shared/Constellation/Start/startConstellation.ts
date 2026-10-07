import { mountBackdrop } from '@/shared/Constellation/Start/mountBackdrop';
import { prefersReducedMotion } from '@/shared/Motion/prefersReducedMotion';

const HOSTS = '[data-constellation]';

export const startConstellation = (): (() => void) => {
    if (prefersReducedMotion()) {
        return () => undefined;
    }

    const running = new Map<Element, () => void>();

    const sync = (): void => {
        for (const [host, stop] of running) {
            if (host.isConnected === false) {
                stop();
                running.delete(host);
            }
        }

        for (const host of document.querySelectorAll(HOSTS)) {
            if (running.has(host) === false) {
                running.set(host, mountBackdrop(host));
            }
        }
    };

    const observer = new MutationObserver(sync);

    observer.observe(document.body, {
        childList: true,
        subtree: true,
    });
    sync();

    return () => {
        observer.disconnect();

        for (const stop of running.values()) {
            stop();
        }

        running.clear();
    };
};

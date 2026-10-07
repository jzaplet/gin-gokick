import { constellation } from '@/shared/Constellation/Effect/constellation';
import { createCanvasLoop } from '@/shared/Constellation/Loop/createCanvasLoop';
import { whenIdle } from '@/shared/Constellation/Start/whenIdle';

export const mountBackdrop = (host: Element): (() => void) => {
    const canvas = document.createElement('canvas');

    canvas.setAttribute('aria-hidden', 'true');
    canvas.className = 'pointer-events-none fixed inset-0 -z-10 size-full opacity-0 transition-opacity duration-1000 '
        + 'motion-reduce:hidden [&.is-live]:opacity-55';
    host.append(canvas);

    const loop = createCanvasLoop(canvas, constellation, () => canvas.getContext('2d'));
    const cancelStart = whenIdle(loop.start);

    return () => {
        cancelStart();
        loop.stop();
        canvas.remove();
    };
};

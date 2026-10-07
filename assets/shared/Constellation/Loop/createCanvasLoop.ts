import type { CanvasDrawing } from '@/shared/Constellation/types/CanvasDrawing';
import type { CanvasEffect } from '@/shared/Constellation/types/CanvasEffect';
import type { CanvasFxContext } from '@/shared/Constellation/types/CanvasFxContext';
import type { CanvasLoop } from '@/shared/Constellation/types/CanvasLoop';

export const FRAME_GAP_MS = 1000 / 30;

const pixelRatio = (): number => Math.min(window.devicePixelRatio === 0 ? 1 : window.devicePixelRatio, 1.5);

export const createCanvasLoop = (
    canvas: HTMLCanvasElement,
    effect: CanvasEffect,
    drawing: () => CanvasDrawing | null,
): CanvasLoop => {
    let started = false;
    let stopped = false;
    let running = false;
    let remeasuring = false;
    let timer: number | undefined;
    let frame: number | undefined;
    let context: CanvasFxContext | undefined;
    let draw = (): void => undefined;

    const measure = (): void => {
        if (context === undefined || (context.w === window.innerWidth && context.h === window.innerHeight)) {
            return;
        }

        const ratio = pixelRatio();

        context.w = window.innerWidth;
        context.h = window.innerHeight;
        canvas.width = Math.round(context.w * ratio);
        canvas.height = Math.round(context.h * ratio);
        context.ctx.setTransform(ratio, 0, 0, ratio, 0, 0);
    };

    const tick = (): void => {
        frame = undefined;

        if (running === false) {
            return;
        }

        draw();
        timer = window.setTimeout(() => {
            timer = undefined;
            frame = requestAnimationFrame(tick);
        }, FRAME_GAP_MS);
    };

    const cancel = (): void => {
        window.clearTimeout(timer);
        timer = undefined;

        if (frame !== undefined) {
            cancelAnimationFrame(frame);
            frame = undefined;
        }
    };

    const sync = (): void => {
        if (running !== document.hidden) {
            return;
        }

        running = document.hidden === false;

        if (running) {
            frame = requestAnimationFrame(tick);

            return;
        }

        cancel();
    };

    const resize = (): void => {
        if (remeasuring) {
            return;
        }

        remeasuring = true;
        requestAnimationFrame(() => {
            remeasuring = false;
            measure();

            if (running) {
                cancel();
                tick();
            }
        });
    };

    const start = (): void => {
        if (started || stopped) {
            return;
        }

        started = true;

        const ctx = drawing();

        if (ctx === null) {
            return;
        }

        context = {
            ctx,
            w: 0,
            h: 0,
        };
        draw = effect(context);
        measure();
        canvas.classList.add('is-live');
        window.addEventListener('resize', resize, { passive: true });
        document.addEventListener('visibilitychange', sync);
        sync();
    };

    const stop = (): void => {
        stopped = true;
        running = false;
        cancel();
        canvas.classList.remove('is-live');
        window.removeEventListener('resize', resize);
        document.removeEventListener('visibilitychange', sync);
    };

    return {
        start,
        stop,
    };
};

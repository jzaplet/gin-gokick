import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { createCanvasLoop, FRAME_GAP_MS } from '@/shared/Constellation/Loop/createCanvasLoop';
import type { CanvasFxContext } from '@/shared/Constellation/types/CanvasFxContext';
import type { CanvasLoop } from '@/shared/Constellation/types/CanvasLoop';
import { countingDrawing } from './drawing';

let frames = new Map<number, FrameRequestCallback>();
let lastFrame = 0;
let hidden = false;
let loops: CanvasLoop[] = [];

const runFrames = (): number => {
    const queued = [...frames.values()];

    frames = new Map();

    for (const frame of queued) {
        frame(0);
    }

    return queued.length;
};

const start = (): {
    canvas: HTMLCanvasElement;
    drawn: () => number;
    context: () => CanvasFxContext | undefined;
} => {
    const canvas = document.createElement('canvas');
    let draws = 0;
    let shown: CanvasFxContext | undefined;
    const loop = createCanvasLoop(
        canvas,
        (fx) => {
            shown = fx;

            return () => {
                draws += 1;
            };
        },
        () => countingDrawing().drawing,
    );

    const drawn = (): number => draws;

    const context = (): CanvasFxContext | undefined => shown;

    loops.push(loop);
    loop.start();

    return {
        canvas,
        drawn,
        context,
    };
};

beforeEach(() => {
    frames = new Map();
    hidden = false;
    loops = [];
    vi.useFakeTimers();
    vi.stubGlobal('requestAnimationFrame', (frame: FrameRequestCallback) => {
        lastFrame += 1;
        frames.set(lastFrame, frame);

        return lastFrame;
    });
    vi.stubGlobal('cancelAnimationFrame', (frame: number) => frames.delete(frame));
    vi.spyOn(document, 'hidden', 'get').mockImplementation(() => hidden);
});

afterEach(() => {
    for (const loop of loops) {
        loop.stop();
    }

    vi.useRealTimers();
});

describe('the canvas loop', () => {
    it('wakes the browser once per drawn frame, not once per vsync', () => {
        const { drawn } = start();

        expect(runFrames()).toBe(1);
        vi.advanceTimersByTime(FRAME_GAP_MS - 1);

        expect(frames.size).toBe(0);

        vi.advanceTimersByTime(1);
        for (let frame = 2; frame <= 5; frame++) {
            expect(runFrames()).toBe(1);
            expect(drawn()).toBe(frame);
            expect(frames.size).toBe(0);
            vi.advanceTimersByTime(FRAME_GAP_MS);
        }
    });

    it('stops while the tab is in the background and goes on when it returns', () => {
        const { drawn } = start();

        runFrames();
        hidden = true;
        document.dispatchEvent(new Event('visibilitychange'));
        vi.advanceTimersByTime(1000);

        expect(runFrames()).toBe(0);
        expect(drawn()).toBe(1);

        hidden = false;
        document.dispatchEvent(new Event('visibilitychange'));

        expect(runFrames()).toBe(1);
        expect(drawn()).toBe(2);
    });

    it('sizes the bitmap to the window at no more than 1.5 device pixels per CSS pixel', () => {
        vi.stubGlobal('devicePixelRatio', 2);
        const { canvas, context } = start();

        expect([
            canvas.width,
            canvas.height,
        ]).toEqual([
            window.innerWidth * 1.5,
            window.innerHeight * 1.5,
        ]);
        expect([
            context()?.w,
            context()?.h,
        ]).toEqual([
            window.innerWidth,
            window.innerHeight,
        ]);
    });
});

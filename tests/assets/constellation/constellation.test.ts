import { describe, expect, it } from 'vitest';
import { constellation, linkRadius, particleCount } from '@/shared/Constellation/Effect/constellation';
import type { CanvasFxContext } from '@/shared/Constellation/types/CanvasFxContext';
import { countingDrawing, type DrawCounts } from './drawing';

const field = (w: number, h: number): {
    context: CanvasFxContext;
    counts: DrawCounts;
} => {
    const { drawing, counts } = countingDrawing();

    return {
        context: {
            ctx: drawing,
            w,
            h,
        },
        counts,
    };
};

describe('constellation sizing', () => {
    it('stops at 180 particles and a reach of 150 px on a huge screen', () => {
        expect(particleCount(3840, 2160)).toBe(180);
        expect(linkRadius(4000, 4000)).toBe(150);
    });
});

describe('constellation per frame', () => {
    it('clears once and paints the whole field in at most nine styled passes', () => {
        const { context, counts } = field(1440, 620);

        constellation(context)();

        expect(counts.clearRect).toBe(1);
        expect(counts.styles.length).toBeLessThanOrEqual(9);
        expect(counts.arc).toBeGreaterThan(0);
        expect(counts.stroke).toBeGreaterThan(0);
    });
});

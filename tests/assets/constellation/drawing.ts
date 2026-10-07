import type { CanvasDrawing } from '@/shared/Constellation/types/CanvasDrawing';

export type DrawCounts = {
    clearRect: number;
    beginPath: number;
    moveTo: number;
    lineTo: number;
    arc: number;
    fill: number;
    stroke: number;
    styles: string[];
};

export const countingDrawing = (): {
    drawing: CanvasDrawing;
    counts: DrawCounts;
} => {
    const counts: DrawCounts = {
        clearRect: 0,
        beginPath: 0,
        moveTo: 0,
        lineTo: 0,
        arc: 0,
        fill: 0,
        stroke: 0,
        styles: [],
    };
    let fill: CanvasDrawing['fillStyle'] = '';
    let stroke: CanvasDrawing['strokeStyle'] = '';
    const style = (value: CanvasDrawing['fillStyle']): CanvasDrawing['fillStyle'] => {
        counts.styles.push(typeof value === 'string' ? value : '');

        return value;
    };
    const drawing: CanvasDrawing = {
        lineWidth: 1,
        get fillStyle() {
            return fill;
        },
        set fillStyle(value) {
            fill = style(value);
        },
        get strokeStyle() {
            return stroke;
        },
        set strokeStyle(value) {
            stroke = style(value);
        },
        clearRect: () => {
            counts.clearRect += 1;
        },
        beginPath: () => {
            counts.beginPath += 1;
        },
        moveTo: () => {
            counts.moveTo += 1;
        },
        lineTo: () => {
            counts.lineTo += 1;
        },
        arc: () => {
            counts.arc += 1;
        },
        fill: () => {
            counts.fill += 1;
        },
        stroke: () => {
            counts.stroke += 1;
        },
        setTransform: () => undefined,
    };

    return {
        drawing,
        counts,
    };
};

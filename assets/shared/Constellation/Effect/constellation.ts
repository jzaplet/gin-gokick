import { at } from '@/shared/Constellation/Effect/at';
import { fillGrid } from '@/shared/Constellation/Effect/particleGrid';
import { moveParticle, seedParticle, unitPolygon } from '@/shared/Constellation/Effect/particles';
import type { CanvasEffect } from '@/shared/Constellation/types/CanvasEffect';
import type { Particle } from '@/shared/Constellation/types/Particle';
import type { ParticleGrid } from '@/shared/Constellation/types/ParticleGrid';

const DOT_RGB = '61,108,255';
const LINK_RGB = '11,18,38';
const TWO_PI = Math.PI * 2;
const AREA_PER_PARTICLE = 9000;
const MIN_PARTICLES = 18;
const MAX_PARTICLES = 180;
const LINK_MIN = 110;
const LINK_MAX = 150;
const FADE_SHARE = 0.26;
const NEIGHBOUR_X = [
    0,
    1,
    -1,
    0,
    1,
];
const NEIGHBOUR_Y = [
    0,
    0,
    1,
    1,
    1,
];
const BUCKETS = 4;
const DOT_ALPHA_STEP = 0.26;
const LINK_CAPACITY = 1024;

export const particleCount = (w: number, h: number): number =>
    Math.max(MIN_PARTICLES, Math.min(MAX_PARTICLES, Math.round((w * h) / AREA_PER_PARTICLE)));

export const linkRadius = (w: number, h: number): number =>
    Math.max(LINK_MIN, Math.min(LINK_MAX, Math.sqrt(w * h) / 9));

const bucketIndex = (value: number, step: number): number => Math.min(BUCKETS - 1, Math.floor(value / step));

export const constellation: CanvasEffect = (context) => {
    const { ctx } = context;
    const linkColors = Array.from({ length: BUCKETS }, (_, bucket) => `rgba(${LINK_RGB},${0.1 + bucket * 0.1})`);
    const dotColors = Array.from({ length: BUCKETS }, (_, bucket) => `rgba(${DOT_RGB},${0.42 + bucket * 0.17})`);
    const polygonColor = `rgba(${DOT_RGB},0.7)`;
    const links = Array.from({ length: BUCKETS }, () => new Float64Array(LINK_CAPACITY));
    const linkLength = new Int32Array(BUCKETS);
    const dotLength = new Int32Array(BUCKETS);
    let dots = Array.from({ length: BUCKETS }, () => new Float64Array(0));
    let particles: Particle[] = [];
    let grid: ParticleGrid = {
        cell: 1,
        cols: 1,
        rows: 1,
        start: new Int32Array(2),
        cursor: new Int32Array(2),
        items: new Int32Array(0),
    };
    let lastW = -1;
    let lastH = -1;
    let link = 0;
    let fadeEnd = 0;

    const seed = (w: number, h: number): void => {
        const count = particleCount(w, h);

        link = linkRadius(w, h);
        fadeEnd = h * FADE_SHARE;
        particles = Array.from({ length: count }, (_, index) => seedParticle(index, w, h));

        const cols = Math.max(1, Math.ceil(w / link));
        const rows = Math.max(1, Math.ceil(h / link));

        grid = {
            cell: link,
            cols,
            rows,
            start: new Int32Array(cols * rows + 1),
            cursor: new Int32Array(cols * rows + 1),
            items: new Int32Array(count),
        };
        dots = Array.from({ length: BUCKETS }, () => new Float64Array(count * 3));
    };

    const ramp = (y: number): number => {
        if (y >= fadeEnd) {
            return 1;
        }

        return y <= 0 ? 0 : y / fadeEnd;
    };

    const pushLink = (bucket: number, a: Particle, b: Particle): void => {
        const slot = at(linkLength, bucket);
        let data = links[bucket] ?? new Float64Array(LINK_CAPACITY);

        if (slot + 4 > data.length) {
            const grown = new Float64Array(data.length * 2);

            grown.set(data);
            data = grown;
        }

        links[bucket] = data;
        data[slot] = a.x;
        data[slot + 1] = a.y;
        data[slot + 2] = b.x;
        data[slot + 3] = b.y;
        linkLength[bucket] = slot + 4;
    };

    const linkPair = (a: Particle, b: Particle): void => {
        const dx = a.x - b.x;
        const dy = a.y - b.y;

        if (Math.abs(a.layer - b.layer) > 1 || dx * dx + dy * dy >= link * link) {
            return;
        }

        const strength = (1 - Math.sqrt(dx * dx + dy * dy) / link) * ramp((a.y + b.y) * 0.5);

        if (strength > 0.03) {
            pushLink(bucketIndex(strength, 1 / BUCKETS), a, b);
        }
    };

    const linkCells = (here: number, there: number, same: boolean): void => {
        const hereTo = at(grid.start, here + 1);
        const thereTo = at(grid.start, there + 1);

        for (let first = at(grid.start, here); first < hereTo; first++) {
            const a = particles[at(grid.items, first)];

            for (let second = same ? first + 1 : at(grid.start, there); a !== undefined && second < thereTo; second++) {
                const b = particles[at(grid.items, second)];

                if (b !== undefined) {
                    linkPair(a, b);
                }
            }
        }
    };

    const linkNeighbours = (cx: number, cy: number): void => {
        const here = cy * grid.cols + cx;

        for (let neighbour = 0; neighbour < NEIGHBOUR_X.length; neighbour++) {
            const nx = cx + at(NEIGHBOUR_X, neighbour);
            const ny = cy + at(NEIGHBOUR_Y, neighbour);

            if (nx >= 0 && nx < grid.cols && ny >= 0 && ny < grid.rows) {
                linkCells(here, ny * grid.cols + nx, neighbour === 0);
            }
        }
    };

    const strokeLinks = (): void => {
        linkLength.fill(0);

        for (let cell = 0; cell < grid.cols * grid.rows; cell++) {
            if (at(grid.start, cell) !== at(grid.start, cell + 1)) {
                linkNeighbours(cell % grid.cols, Math.floor(cell / grid.cols));
            }
        }

        ctx.lineWidth = 0.9;

        for (let bucket = 0; bucket < BUCKETS; bucket++) {
            const used = at(linkLength, bucket);
            const data = links[bucket];

            if (used > 0 && data !== undefined) {
                ctx.strokeStyle = linkColors[bucket] ?? '';
                ctx.beginPath();

                for (let slot = 0; slot < used; slot += 4) {
                    ctx.moveTo(at(data, slot), at(data, slot + 1));
                    ctx.lineTo(at(data, slot + 2), at(data, slot + 3));
                }

                ctx.stroke();
            }
        }
    };

    const collectDots = (): void => {
        dotLength.fill(0);

        for (const particle of particles) {
            const alpha = particle.sides === 0 ? (0.5 + Math.sin(particle.tw) * 0.16) * ramp(particle.y) : 0;
            const bucket = bucketIndex(alpha, DOT_ALPHA_STEP);
            const data = dots[bucket];

            if (alpha > 0.03 && data !== undefined) {
                const slot = at(dotLength, bucket);

                data[slot] = particle.x;
                data[slot + 1] = particle.y;
                data[slot + 2] = particle.r;
                dotLength[bucket] = slot + 3;
            }
        }
    };

    const fillDots = (): void => {
        collectDots();

        for (let bucket = 0; bucket < BUCKETS; bucket++) {
            const used = at(dotLength, bucket);
            const data = dots[bucket];

            if (used > 0 && data !== undefined) {
                ctx.fillStyle = dotColors[bucket] ?? '';
                ctx.beginPath();

                for (let slot = 0; slot < used; slot += 3) {
                    ctx.moveTo(at(data, slot) + at(data, slot + 2), at(data, slot + 1));
                    ctx.arc(at(data, slot), at(data, slot + 1), at(data, slot + 2), 0, TWO_PI);
                }

                ctx.fill();
            }
        }
    };

    const tracePolygon = (particle: Particle): void => {
        const cos = Math.cos(particle.rot);
        const sin = Math.sin(particle.rot);
        const vertices = unitPolygon(particle.sides);

        for (let side = 0; side <= particle.sides; side++) {
            const ux = at(vertices, side * 2);
            const uy = at(vertices, side * 2 + 1);
            const x = particle.x + (ux * cos - uy * sin) * particle.r;
            const y = particle.y + (ux * sin + uy * cos) * particle.r;

            if (side === 0) {
                ctx.moveTo(x, y);
            } else {
                ctx.lineTo(x, y);
            }
        }
    };

    const strokePolygons = (): void => {
        ctx.strokeStyle = polygonColor;
        ctx.lineWidth = 1.1;
        ctx.beginPath();

        for (const particle of particles) {
            if (particle.sides > 0 && ramp(particle.y) >= 0.7) {
                tracePolygon(particle);
            }
        }

        ctx.stroke();
    };

    return () => {
        const { w, h } = context;

        if (w !== lastW || h !== lastH) {
            seed(w, h);
            lastW = w;
            lastH = h;
        }

        ctx.clearRect(0, 0, w, h);

        for (const particle of particles) {
            moveParticle(particle, w, h);
        }

        fillGrid(grid, particles);
        strokeLinks();
        fillDots();
        strokePolygons();
    };
};

import { at } from '@/shared/Constellation/Effect/at';
import type { Particle } from '@/shared/Constellation/types/Particle';

const TWO_PI = Math.PI * 2;
const DRIFT_SPEED = 0.96;
const WRAP_MARGIN = 12;
const POLYGON_SIDES = [
    3,
    4,
    6,
];

const polygons = new Map<number, number[]>();

export const unitPolygon = (sides: number): number[] => {
    const cached = polygons.get(sides);

    if (cached !== undefined) {
        return cached;
    }

    const vertices: number[] = [];

    for (let i = 0; i <= sides; i++) {
        vertices.push(Math.cos((i / sides) * TWO_PI), Math.sin((i / sides) * TWO_PI));
    }

    polygons.set(sides, vertices);

    return vertices;
};

export const seedParticle = (index: number, w: number, h: number): Particle => {
    const sides = index % 6 === 0 ? at(POLYGON_SIDES, Math.floor(Math.random() * 3)) : 0;

    return {
        x: Math.random() * w,
        y: Math.random() * h,
        vx: (Math.random() - 0.5) * DRIFT_SPEED,
        vy: (Math.random() - 0.5) * DRIFT_SPEED,
        r: sides > 0 ? 3 + Math.random() * 2.6 : 0.7 + Math.random() * 1.5,
        sides,
        rot: Math.random() * TWO_PI,
        spin: (Math.random() - 0.5) * 0.02,
        layer: (index % 3) + 1,
        tw: Math.random() * TWO_PI,
    };
};

const wrap = (value: number, size: number): number => {
    if (value < -WRAP_MARGIN) {
        return size + WRAP_MARGIN;
    }

    return value > size + WRAP_MARGIN ? -WRAP_MARGIN : value;
};

export const moveParticle = (particle: Particle, w: number, h: number): void => {
    const speed = 4 - particle.layer;

    particle.x = wrap(particle.x + particle.vx * speed, w);
    particle.y = wrap(particle.y + particle.vy * speed, h);
    particle.tw += 0.015;

    if (particle.sides > 0) {
        particle.rot += particle.spin;
    }
};

import { at } from '@/shared/Constellation/Effect/at';
import type { Particle } from '@/shared/Constellation/types/Particle';
import type { ParticleGrid } from '@/shared/Constellation/types/ParticleGrid';

const cellOf = (grid: ParticleGrid, particle: Particle): number =>
    Math.min(grid.rows - 1, Math.max(0, Math.floor(particle.y / grid.cell))) * grid.cols
    + Math.min(grid.cols - 1, Math.max(0, Math.floor(particle.x / grid.cell)));

export const fillGrid = (grid: ParticleGrid, particles: Particle[]): void => {
    grid.start.fill(0);

    for (const particle of particles) {
        const next = cellOf(grid, particle) + 1;

        grid.start[next] = at(grid.start, next) + 1;
    }

    for (let cell = 1; cell < grid.start.length; cell++) {
        grid.start[cell] = at(grid.start, cell) + at(grid.start, cell - 1);
    }

    grid.cursor.set(grid.start);

    for (let index = 0; index < particles.length; index++) {
        const particle = particles[index];

        if (particle !== undefined) {
            const cell = cellOf(grid, particle);
            const slot = at(grid.cursor, cell);

            grid.items[slot] = index;
            grid.cursor[cell] = slot + 1;
        }
    }
};

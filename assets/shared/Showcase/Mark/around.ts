export const around = (x: number, y: number, transform: string): string =>
    `translate(${x}px, ${y}px) ${transform} translate(${-x}px, ${-y}px)`;

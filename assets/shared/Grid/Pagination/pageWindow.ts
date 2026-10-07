const windowSize = 5;

export const pageWindow = (page: number, pages: number): number[] => {
    const last = Math.min(pages, Math.max(1, page - 2) + windowSize - 1);
    const first = Math.max(1, last - windowSize + 1);

    return Array.from({ length: last - first + 1 }, (_, index) => first + index);
};

export type EdgeState = {
    left: boolean;
    right: boolean;
};

export const edgeState = (scrollLeft: number, scrollWidth: number, clientWidth: number): EdgeState => {
    const maxScroll = scrollWidth - clientWidth;

    return {
        left: maxScroll > 0 && scrollLeft > 1,
        right: maxScroll > 0 && scrollLeft < maxScroll - 1,
    };
};

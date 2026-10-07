export type PageSelection = Readonly<{
    pageSelectable: boolean;
    pageSelected: boolean;
    togglePage: () => void;
}>;

export type GridSelection<T> = PageSelection & Readonly<{
    count: number;
    total: number;
    all: boolean;
    selectable: (item: T) => boolean;
    selected: (item: T) => boolean;
    ids: () => string[];
    toggle: (item: T) => void;
    selectAll: () => void;
    deselect: (id: string) => void;
    clear: () => void;
}>;

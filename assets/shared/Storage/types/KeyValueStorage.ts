export type KeyValueStorage = {
    read: (key: string) => unknown;
    write: (key: string, value: unknown) => boolean;
    remove: (key: string) => boolean;
    follow: (key: string, listener: () => void) => () => void;
};

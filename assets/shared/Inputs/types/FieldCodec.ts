export type FieldCodec<T> = {
    parse: (text: string) => T;
    format: (value: T) => string;
};

import type { FieldCodec } from '@/shared/Inputs/types/FieldCodec';

export const textCodec: FieldCodec<string> = {
    parse: (text) => text,
    format: (value) => value,
};

export const nullableTextCodec: FieldCodec<string | null> = {
    parse: (text) => (text === '' ? null : text),
    format: (value) => value ?? '',
};

const formatNumber = (value: number | null): string => (value === null ? '' : String(value));

export const numberCodec = (valueOnClear: number | null): FieldCodec<number | null> => {
    const parse = (text: string): number | null => (text === '' ? valueOnClear : Number(text));

    return {
        parse,
        format: formatNumber,
    };
};

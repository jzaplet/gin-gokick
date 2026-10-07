import { readonly, type Ref, ref, watch } from 'vue';
import type { FieldCodec } from '@/shared/Inputs/types/FieldCodec';

type FieldText = {
    text: Readonly<Ref<string>>;
    typed: (event: Event) => void;
    committed: () => void;
};

export const useFieldText = <T>(model: Ref<T>, codec: () => FieldCodec<T>): FieldText => {
    const text = ref(codec().format(model.value));

    const typed = (event: Event): void => {
        if (event.target instanceof HTMLInputElement) {
            text.value = event.target.value;
            model.value = codec().parse(event.target.value);
        }
    };

    const committed = (): void => {
        if (text.value !== '') {
            return;
        }

        const { parse, format } = codec();

        text.value = format(parse(''));
    };

    const follow = (value: T): void => {
        if (codec().parse(text.value) === value) {
            return;
        }

        text.value = codec().format(value);
    };

    watch(model, follow);

    return {
        text: readonly(text),
        typed,
        committed,
    };
};

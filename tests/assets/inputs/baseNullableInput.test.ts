import { mount, type VueWrapper } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import BaseNullableInput from '@/shared/Inputs/BaseNullableInput.vue';

const nullableInput = (modelValue: string | null): VueWrapper => mount(BaseNullableInput, {
    props: {
        name: 'note',
        label: 'Poznámka',
        autocomplete: 'off',
        modelValue,
    },
});

describe('the base nullable input', () => {
    it.each([
        [
            'Volat odpoledne',
            'Volat odpoledne',
        ],
        [
            '',
            null,
        ],
    ] as const)('sends %j as %j', async (typed, sent) => {
        const wrapper = nullableInput('Poznámka');

        await wrapper.get('input').setValue(typed);

        expect(wrapper.emitted('update:modelValue')).toEqual([[sent]]);
    });
});

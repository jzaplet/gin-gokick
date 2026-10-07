import { mount, type VueWrapper } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import BaseNumberInput from '@/shared/Inputs/BaseNumberInput.vue';

type Props = {
    modelValue?: number | null;
    min?: number;
    max?: number;
    step?: number | 'any';
    valueOnClear?: number;
};

const numberInput = (props: Props = {}): VueWrapper => mount(BaseNumberInput, {
    props: {
        name: 'count',
        label: 'Počet',
        autocomplete: 'off',
        modelValue: null,
        ...props,
    },
});

const typeWithoutLeaving = async (wrapper: VueWrapper, text: string): Promise<void> => {
    const field = wrapper.get('input');

    field.element.value = text;
    await field.trigger('input');
};

describe('the base number input', () => {
    it.each([
        [
            '5',
            5,
        ],
        [
            '1.5',
            1.5,
        ],
        [
            '-3',
            -3,
        ],
        [
            '',
            null,
        ],
    ] as const)('sends %j as %j', async (typed, sent) => {
        const wrapper = numberInput({ modelValue: 8 });

        await wrapper.get('input').setValue(typed);

        expect(wrapper.emitted('update:modelValue')).toEqual([[sent]]);
    });

    it('keeps what was typed when the model comes back with its number', async () => {
        const wrapper = numberInput();

        await wrapper.get('input').setValue('1.50');
        await wrapper.setProps({ modelValue: 1.5 });

        expect(wrapper.emitted('update:modelValue')).toEqual([[1.5]]);
        expect(wrapper.get('input').element.value).toBe('1.50');
    });

    it('shows a number the model gets from elsewhere', async () => {
        const wrapper = numberInput({ modelValue: 3 });

        await wrapper.setProps({ modelValue: 7 });

        expect(wrapper.get('input').element.value).toBe('7');

        await wrapper.setProps({ modelValue: null });

        expect(wrapper.get('input').element.value).toBe('');
    });

    it('sends its value on clear for an emptied field and shows it once the visitor leaves', async () => {
        const wrapper = numberInput({
            modelValue: 30,
            valueOnClear: 25,
        });
        const field = wrapper.get('input');

        await typeWithoutLeaving(wrapper, '');
        await wrapper.setProps({ modelValue: 25 });

        expect(wrapper.emitted('update:modelValue')).toEqual([[25]]);
        expect(field.element.value).toBe('');

        await field.trigger('change');

        expect(field.element.value).toBe('25');
    });

    it('lets the visitor type a new number into the emptied field', async () => {
        const wrapper = numberInput({
            modelValue: 30,
            valueOnClear: 25,
        });
        const field = wrapper.get('input');

        await typeWithoutLeaving(wrapper, '');
        await wrapper.setProps({ modelValue: 25 });
        await typeWithoutLeaving(wrapper, '4');
        await field.trigger('change');

        expect(wrapper.emitted('update:modelValue')).toEqual([
            [25],
            [4],
        ]);
        expect(field.element.value).toBe('4');
    });

    it('follows a value on clear that changes', async () => {
        const wrapper = numberInput({
            modelValue: 30,
            valueOnClear: 25,
        });

        await wrapper.setProps({ valueOnClear: 50 });
        await wrapper.get('input').setValue('');

        expect(wrapper.emitted('update:modelValue')).toEqual([[50]]);
    });
});

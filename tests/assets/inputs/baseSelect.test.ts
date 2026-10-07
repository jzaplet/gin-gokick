import { mount, type VueWrapper } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import BaseSelect from '@/shared/Inputs/BaseSelect.vue';

type Props = {
    modelValue?: string;
    placeholder?: string;
    error?: string;
    status?: string;
    required?: boolean;
    loading?: boolean;
};

const select = (props: Props = {}): VueWrapper => mount(BaseSelect, {
    props: {
        name: 'language',
        label: 'Jazyk',
        options: [
            {
                value: 'cs',
                label: 'Čeština',
            },
            {
                value: 'en',
                label: 'English',
            },
        ],
        modelValue: '',
        ...props,
    },
});

const options = (wrapper: VueWrapper): string[] => wrapper.findAll('option').map((option) => option.text());

describe('the base select', () => {
    it('shows the chosen option under its label and sends the one picked', async () => {
        const wrapper = select({ modelValue: 'en' });

        expect(wrapper.get('label').attributes('for')).toBe('language');
        expect(wrapper.get('select').attributes('id')).toBe('language');
        expect(options(wrapper)).toEqual([
            'Čeština',
            'English',
        ]);
        expect(wrapper.get('select').element.value).toBe('en');

        await wrapper.get('select').setValue('cs');

        expect(wrapper.emitted('update:modelValue')).toEqual([['cs']]);
    });

    it('ties its error to the select', () => {
        expect(select().get('select').attributes('aria-invalid')).toBe('false');

        const wrapper = select({ error: 'Vyberte jazyk.' });

        expect(wrapper.get('select').attributes('aria-invalid')).toBe('true');
        expect(wrapper.get('select').attributes('aria-describedby')).toBe('language-error');
        expect(wrapper.get('#language-error').text()).toBe('Vyberte jazyk.');
    });

    it('marks a required select with an asterisk a screen reader skips', () => {
        const wrapper = select({ required: true });

        expect(wrapper.get('select').attributes()).toHaveProperty('required');
        expect(wrapper.get('label span').attributes('aria-hidden')).toBe('true');
    });

    it('shows a spinner and is busy while it loads', async () => {
        const wrapper = select();

        expect(wrapper.find('.animate-spin').exists()).toBe(false);

        await wrapper.setProps({ loading: true });

        expect(wrapper.find('.animate-spin').exists()).toBe(true);
        expect(wrapper.get('select').attributes('aria-busy')).toBe('true');
    });
});

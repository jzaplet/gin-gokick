import { mount, type VueWrapper } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import type { TextType } from '@/shared/Inputs/types/TextType';
import BaseInput from '@/shared/Inputs/BaseInput.vue';

type Props = {
    modelValue?: string;
    type?: TextType;
    error?: string;
    status?: string;
    required?: boolean;
    loading?: boolean;
};

const input = (props: Props = {}): VueWrapper => mount(BaseInput, {
    props: {
        name: 'email',
        label: 'E-mail',
        autocomplete: 'email',
        modelValue: '',
        ...props,
    },
});

describe('the base input', () => {
    it('shows its text under its label and sends what is typed', async () => {
        const wrapper = input({ modelValue: 'jana@example.com' });

        expect(wrapper.get('label').attributes('for')).toBe('email');
        expect(wrapper.get('input').attributes('id')).toBe('email');
        expect(wrapper.get('input').attributes('type')).toBe('text');
        expect(wrapper.get('input').element.value).toBe('jana@example.com');

        await wrapper.get('input').setValue('petr@example.com');

        expect(wrapper.emitted('update:modelValue')).toEqual([['petr@example.com']]);
    });

    it('takes the type password', () => {
        expect(input({ type: 'password' }).get('input').attributes('type')).toBe('password');
    });

    it('marks a required field with an asterisk a screen reader skips', () => {
        const wrapper = input({ required: true });
        const asterisk = wrapper.get('label span');

        expect(wrapper.get('input').attributes()).toHaveProperty('required');
        expect(asterisk.text()).toBe('*');
        expect(asterisk.attributes('aria-hidden')).toBe('true');
    });

    it('ties a status to the input', () => {
        const wrapper = input({ status: 'E-mail je volný.' });

        expect(wrapper.get('input').attributes('aria-describedby')).toBe('email-status');
        expect(wrapper.get('input').attributes('aria-invalid')).toBe('false');
        expect(wrapper.get('#email-status').text()).toBe('E-mail je volný.');
    });

    it('shows the error in place of the status', () => {
        const wrapper = input({
            error: 'Tento e-mail už má účet.',
            status: 'E-mail je volný.',
        });

        expect(wrapper.get('input').attributes('aria-invalid')).toBe('true');
        expect(wrapper.get('input').attributes('aria-describedby')).toBe('email-error');
        expect(wrapper.get('#email-error').text()).toBe('Tento e-mail už má účet.');
        expect(wrapper.find('#email-status').exists()).toBe(false);
    });

    it('shows a spinner in the field while it loads and stays editable', async () => {
        const idle = input();

        expect(idle.get('input').attributes('aria-busy')).toBe('false');
        expect(idle.find('.animate-spin').exists()).toBe(false);

        const wrapper = input({ loading: true });

        expect(wrapper.get('input').attributes('aria-busy')).toBe('true');
        expect(wrapper.get('input').attributes()).not.toHaveProperty('disabled');
        expect(wrapper.get('.animate-spin').attributes('aria-hidden')).toBe('true');

        await wrapper.get('input').setValue('jana@example.com');

        expect(wrapper.emitted('update:modelValue')).toEqual([['jana@example.com']]);
    });
});

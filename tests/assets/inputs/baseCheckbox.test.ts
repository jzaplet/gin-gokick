import { mount, type VueWrapper } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import BaseCheckbox from '@/shared/Inputs/BaseCheckbox.vue';

type Props = {
    modelValue?: boolean;
    disabled?: boolean;
    hideLabel?: boolean;
};

const checkbox = (props: Props = {}): VueWrapper => mount(BaseCheckbox, {
    props: {
        label: 'Vybrat stránku',
        modelValue: false,
        ...props,
    },
});

describe('the base checkbox', () => {
    it('hides its label from the eye, not from a screen reader', () => {
        const wrapper = checkbox({ hideLabel: true });

        expect(wrapper.get('label input').attributes('type')).toBe('checkbox');
        expect(wrapper.get('label').text()).toBe('Vybrat stránku');
        expect(wrapper.get('label > span:last-child').classes()).toContain('sr-only');
    });
});

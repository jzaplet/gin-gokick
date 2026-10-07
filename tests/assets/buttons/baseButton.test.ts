import { mount, type VueWrapper } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import BaseButton from '@/shared/Buttons/BaseButton.vue';

const button = (props: InstanceType<typeof BaseButton>['$props'] = {}): VueWrapper => mount(BaseButton, {
    props,
    slots: { default: 'Uložit' },
});

describe('the base button', () => {
    it.each([
        false,
        true,
    ])(
        'is a plain button that locks itself with a spinner hidden from screen readers only while loading: %s',
        (loading) => {
            const wrapper = button({ loading });

            expect(wrapper.attributes('type')).toBe('button');
            expect(wrapper.attributes('disabled') !== undefined).toBe(loading);
            expect(wrapper.attributes('aria-busy')).toBe(String(loading));
            expect(wrapper.text()).toBe('Uložit');
            expect(
                wrapper.findAll('.animate-spin').map((spinner) => spinner.attributes('aria-hidden')),
            ).toEqual(loading ? ['true'] : []);
        },
    );
});

import { enableAutoUnmount, mount, type VueWrapper } from '@vue/test-utils';
import { afterEach, describe, expect, it } from 'vitest';
import BaseDropdown from '@/shared/Dropdown/BaseDropdown.vue';

enableAutoUnmount(afterEach);

const make = (props: InstanceType<typeof BaseDropdown>['$props'] = {}): VueWrapper => mount(BaseDropdown, {
    props,
    attachTo: document.body,
    slots: {
        trigger: `<template #trigger="{ open, toggle }">
            <button type="button" class="trigger" :aria-expanded="open" @click="toggle">open</button>
        </template>`,
        default: '<p class="item">item</p>',
    },
});

const opened = async (): Promise<VueWrapper> => {
    const wrapper = make();

    await wrapper.get('.trigger').trigger('click');

    return wrapper;
};

describe('the dropdown', () => {
    it('closes on Escape and stays open on another key', async () => {
        const wrapper = await opened();

        expect(wrapper.get('.trigger').attributes('aria-expanded')).toBe('true');

        document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));
        await wrapper.vm.$nextTick();

        expect(wrapper.find('.item').exists()).toBe(true);

        document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));
        await wrapper.vm.$nextTick();

        expect(wrapper.find('.item').exists()).toBe(false);
    });
});

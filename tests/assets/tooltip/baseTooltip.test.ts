import { mount, type VueWrapper } from '@vue/test-utils';
import { describe, expect, it } from 'vitest';
import BaseTooltip from '@/shared/Tooltip/BaseTooltip.vue';

type Placement = {
    position?: 'top' | 'bottom';
    align?: 'center' | 'left' | 'right';
    disabled?: boolean;
};

const make = (placement: Placement = {}): VueWrapper => mount(BaseTooltip, {
    props: {
        text: 'Odhlásit se',
        ...placement,
    },
    slots: { default: '<button type="button">trigger</button>' },
});

const bubble = (wrapper: VueWrapper): string[] => wrapper.get('[role="tooltip"]').classes();

describe('the tooltip', () => {
    it('wraps its trigger and keeps the bubble hidden until hover or keyboard focus', () => {
        const wrapper = make();
        const hidden = [
            'pointer-events-none',
            'opacity-0',
            'group-hover/tooltip:opacity-100',
            'group-has-focus-visible/tooltip:opacity-100',
        ];

        expect(wrapper.get('button').text()).toBe('trigger');
        expect(wrapper.get('[role="tooltip"]').text()).toBe('Odhlásit se');
        expect(bubble(wrapper)).toEqual(expect.arrayContaining(hidden));
        expect(wrapper.classes()).toContain('group/tooltip');
    });
});

import { enableAutoUnmount, flushPromises, mount, type VueWrapper } from '@vue/test-utils';
import { afterEach, describe, expect, it } from 'vitest';
import { t } from '@/shared/I18n/Texts/translate';
import ConfirmModal from '@/shared/Modal/ConfirmModal.vue';
import { showEnglish } from '../i18n/languages';

type Props = {
    open?: boolean;
    busy?: boolean;
};

enableAutoUnmount(afterEach);

const modal = (props: Props = {}): VueWrapper => {
    const wrapper: VueWrapper = mount(ConfirmModal, {
        attachTo: document.body,
        props: {
            'title': 'Smazat účet',
            'message': 'Účet se smaže i se všemi daty.',
            'open': true,
            'onUpdate:open': async (open: boolean) => {
                await wrapper.setProps({ open });
            },
            ...props,
        },
    });

    return wrapper;
};

const dialog = (wrapper: VueWrapper): HTMLDialogElement => wrapper.get<HTMLDialogElement>('dialog').element;

const buttons = (wrapper: VueWrapper): string[] => wrapper.findAll('button').map((button) => button.text());

const animations = (value: () => unknown[]): void => {
    Object.defineProperty(HTMLDialogElement.prototype, 'getAnimations', {
        configurable: true,
        writable: true,
        value,
    });
};

const fading = (): (() => void) => {
    let finish = (): void => undefined;
    const finished = new Promise<void>((resolve) => {
        finish = resolve;
    });

    animations(() => [{ finished }]);

    return finish;
};

describe('the confirm modal', () => {
    afterEach(() => {
        animations(() => []);
    });

    it('opens as a modal dialog named by its title', () => {
        const wrapper = modal();
        const title = wrapper.get('h2');

        expect(dialog(wrapper).open).toBe(true);
        expect(wrapper.get('dialog').attributes('aria-labelledby')).toBe(title.attributes('id'));
        expect(title.text()).toBe('Smazat účet');
    });

    it('emits confirm and stays open for the caller to close it', async () => {
        const wrapper = modal();

        await wrapper.findAll('button')[1]?.trigger('click');

        expect(wrapper.emitted('confirm')).toHaveLength(1);
        expect(dialog(wrapper).open).toBe(true);
    });

    it('closes on cancel without confirming', async () => {
        const wrapper = modal();

        await wrapper.findAll('button')[0]?.trigger('click');

        expect(wrapper.emitted('update:open')).toEqual([[false]]);
        expect(wrapper.emitted('confirm')).toBeUndefined();
        expect(dialog(wrapper).open).toBe(false);
    });

    it('closes on Escape, which closes the dialog itself', async () => {
        const wrapper = modal();

        dialog(wrapper).close();
        await wrapper.vm.$nextTick();

        expect(wrapper.emitted('update:open')).toEqual([[false]]);
    });

    it('stays open when it opens again while it fades out', async () => {
        const finish = fading();
        const wrapper = modal();

        await wrapper.setProps({ open: false });
        await flushPromises();
        await wrapper.setProps({ open: true });
        finish();
        await flushPromises();

        expect(dialog(wrapper).open).toBe(true);
        expect(dialog(wrapper).hasAttribute('data-closing')).toBe(false);
    });

    it('follows the shown language with its default labels', async () => {
        const wrapper = modal();
        const czech = [
            t('modal.cancel'),
            t('modal.confirm'),
        ];

        expect(buttons(wrapper)).toEqual(czech);

        await showEnglish();

        expect(buttons(wrapper)).toEqual([
            t('modal.cancel'),
            t('modal.confirm'),
        ]);
        expect(buttons(wrapper)).not.toEqual(czech);
    });

    it('keeps its texts while it fades out and takes the new ones when it opens again', async () => {
        const wrapper = modal();

        await wrapper.setProps({
            open: false,
            title: 'Smazat projekt',
            message: 'Projekt se smaže.',
        });

        expect(wrapper.get('h2').text()).toBe('Smazat účet');
        expect(wrapper.get('p').text()).toBe('Účet se smaže i se všemi daty.');

        await wrapper.setProps({ open: true });

        expect(wrapper.get('h2').text()).toBe('Smazat projekt');
        expect(wrapper.get('p').text()).toBe('Projekt se smaže.');
    });

    it('shows a spinner on confirm and locks cancel while busy', () => {
        const wrapper = modal({ busy: true });
        const [cancel, confirm] = wrapper.findAll('button');

        expect(confirm?.attributes('aria-busy')).toBe('true');
        expect(confirm?.find('.animate-spin').exists()).toBe(true);
        expect(cancel?.attributes()).toHaveProperty('disabled');
    });
});

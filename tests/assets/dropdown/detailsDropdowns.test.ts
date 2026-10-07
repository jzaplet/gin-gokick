import { afterEach, beforeEach, describe, expect, it } from 'vitest';
import { startDetailsDropdowns } from '@/shared/Dropdown/startDetailsDropdowns';

type Details = {
    element: HTMLDetailsElement;
    label: HTMLElement;
    item: HTMLElement;
};

let stop = (): void => undefined;

const details = (dropdown: boolean): Details => {
    const element = document.createElement('details');
    const summary = document.createElement('summary');
    const label = document.createElement('span');
    const list = document.createElement('ul');
    const item = document.createElement('li');

    if (dropdown) {
        element.dataset['dropdown'] = '';
    }

    summary.append(label);
    list.append(item);
    element.append(summary, list);
    document.body.append(element);

    return {
        element,
        label,
        item,
    };
};

const opened = (): Details => {
    const dropdown = details(true);

    dropdown.label.click();

    return dropdown;
};

beforeEach(() => {
    stop = startDetailsDropdowns();
});

afterEach(() => {
    stop();
    document.body.replaceChildren();
});

describe('a details dropdown of a Go page', () => {
    it('closes on Escape and stays open on another key', () => {
        const dropdown = opened();

        document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter' }));

        expect(dropdown.element.open).toBe(true);

        document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));

        expect(dropdown.element.open).toBe(false);
    });

    it('leaves a details without data-dropdown alone', () => {
        const plain = details(false);

        plain.label.click();
        document.body.click();
        document.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }));

        expect(plain.element.open).toBe(true);
    });
});

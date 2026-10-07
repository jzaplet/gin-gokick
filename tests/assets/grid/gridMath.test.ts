import { describe, expect, it } from 'vitest';
import { pageWindow } from '@/shared/Grid/Pagination/pageWindow';

describe('the page window', () => {
    it.each([
        [
            1,
            1,
            [1],
        ],
        [1, 3, [
            1,
            2,
            3,
        ]],
        [1, 9, [
            1,
            2,
            3,
            4,
            5,
        ]],
        [5, 9, [
            3,
            4,
            5,
            6,
            7,
        ]],
        [9, 9, [
            5,
            6,
            7,
            8,
            9,
        ]],
    ])('shows around page %i of %i the pages %j', (page, pages, numbers) => {
        expect(pageWindow(page, pages)).toEqual(numbers);
    });
});

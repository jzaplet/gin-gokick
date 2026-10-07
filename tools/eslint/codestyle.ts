import type { ESLint } from 'eslint';
import { listLayout } from './listLayout/listLayout';
import { returnByName } from './returnByName/returnByName';

export const codestyle: ESLint.Plugin = {
    rules: {
        'list-layout': listLayout,
        'return-by-name': returnByName,
    },
};

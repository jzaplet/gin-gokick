import { RuleTester } from 'eslint';
import tseslint from 'typescript-eslint';
import { describe, it } from 'vitest';
import { returnByName } from '../../../tools/eslint/returnByName/returnByName';

RuleTester.describe = describe;
RuleTester.it = it;
RuleTester.itOnly = it.only;

const tester = new RuleTester({ languageOptions: { parser: tseslint.parser } });

const lines = (...rows: string[]): string => rows.join('\n');

const inline = [{ messageId: 'inline' }];

tester.run('return-by-name', returnByName, {
    valid: [
        lines('const use = () => {', '    const run = () => go();', '', '    return { run };', '};'),
        lines('const use = () => {', '    const total = computed(() => 1);', '', '    return { total };', '};'),
        'const use = () => ({ open: readonly(open) });',
        'const use = () => reactive({ count, clear });',
        'const point = () => ({ rows: Object.fromEntries(keys.map((key) => [key, 1])) });',
        'const ids = () => items.map((item) => ({ id: item.id }));',
        'const handler = () => () => undefined;',
        'const fake = () => ({ get matches() { return true; }, set onfinish(value) { last = value; } });',
        lines('const fake = () => {', '    const value = { run: () => 1 };', '', '    return value;', '};'),
        'const options = { run: () => 1 };',
    ],
    invalid: [
        {
            code: lines('const use = () => {', '    return { run: () => { go(); } };', '};'),
            errors: inline,
        },
        {
            code: 'const use = () => ({ run() { go(); } });',
            errors: inline,
        },
        {
            code: 'const use = () => ({ total: computed(() => count.value) });',
            errors: inline,
        },
        {
            code: lines('const use = () => {', '    return reactive({ clear: () => undefined });', '};'),
            errors: inline,
        },
        {
            code: 'const use = () => ({ nested: { run: function () { go(); } } });',
            errors: inline,
        },
        {
            code: 'const use = (): Use => ({ run: (() => undefined) as Runner } satisfies Use);',
            errors: inline,
        },
        {
            code: 'const use = () => ({ run: () => 1, stop: () => 2 });',
            errors: [
                { messageId: 'inline' },
                { messageId: 'inline' },
            ],
        },
    ],
});

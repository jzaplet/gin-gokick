import { Linter } from 'eslint';
import pluginVue from 'eslint-plugin-vue';
import tseslint from 'typescript-eslint';
import { describe, expect, it } from 'vitest';
import config from '../../../eslint.config';

const bans = new Set([
    'no-restricted-syntax',
    'vue/no-restricted-v-bind',
    'vue/no-restricted-syntax',
    'vue/no-v-html',
]);

const projectBans = config.map(({ files, rules = {} }) => ({
    ...(files === undefined ? {} : { files }),
    rules: Object.fromEntries(Object.entries(rules).filter(([name]) => bans.has(name))),
}));

const linter = new Linter();

const refusals = (code: string, filename: string): (string | null)[] => linter.verify(
    code,
    [
        ...pluginVue.configs['flat/base'],
        {
            files: ['**/*.ts'],
            languageOptions: { parser: tseslint.parser },
        },
        ...projectBans,
    ],
    filename,
).map(({ ruleId }) => ruleId);

describe('the ban of innerHTML', () => {
    it.each([
        [
            'v-html',
            '<template><div v-html="html" /></template>',
            'Probe.vue',
            'vue/no-v-html',
        ],
        [
            'a binding',
            '<template><div :innerHTML="html" /></template>',
            'Probe.vue',
            'vue/no-restricted-v-bind',
        ],
        [
            'an object binding',
            '<template><div v-bind="{ innerHTML: html }" /></template>',
            'Probe.vue',
            'vue/no-restricted-syntax',
        ],
        [
            'a render function',
            'h(\'div\', { innerHTML: html });',
            'probe.ts',
            'no-restricted-syntax',
        ],
    ])('refuses %s', (_, code, filename, rule) => {
        expect(refusals(code, filename)).toContain(rule);
    });
});

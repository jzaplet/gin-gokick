import eslint from '@eslint/js';
import stylistic from '@stylistic/eslint-plugin';
import { defineConfig } from 'eslint/config';
import sonarjs from 'eslint-plugin-sonarjs';
import pluginVue from 'eslint-plugin-vue';
import tseslint from 'typescript-eslint';
import { codestyle } from './tools/eslint/codestyle';
import {
    componentLogic,
    functionKeyword,
    innerHtml,
    negation,
    relativeImports,
    webStorages,
} from './tools/eslint/restrictions';
import { vueRules } from './tools/eslint/vueRules';

export default defineConfig(
    { ignores: ['public/**'] },
    eslint.configs.recommended,
    tseslint.configs.strictTypeChecked,
    tseslint.configs.stylisticTypeChecked,
    pluginVue.configs['flat/recommended-error'],
    stylistic.configs.customize({
        indent: 4,
        quotes: 'single',
        semi: true,
        jsx: false,
        braceStyle: '1tbs',
        arrowParens: true,
        quoteProps: 'consistent-as-needed',
    }),
    {
        languageOptions: {
            parserOptions: {
                parser: tseslint.parser,
                projectService: true,
                tsconfigRootDir: import.meta.dirname,
                extraFileExtensions: ['.vue'],
            },
        },
    },
    {
        plugins: {
            sonarjs,
            codestyle,
        },
        rules: {
            '@typescript-eslint/consistent-type-definitions': [
                'error',
                'type',
            ],
            '@typescript-eslint/no-unnecessary-boolean-literal-compare': 'off',
            '@typescript-eslint/explicit-function-return-type': ['error', {
                allowExpressions: true,
                allowTypedFunctionExpressions: true,
            }],
            '@typescript-eslint/explicit-module-boundary-types': 'error',
            '@typescript-eslint/consistent-type-imports': ['error', {
                prefer: 'type-imports',
                fixStyle: 'inline-type-imports',
            }],
            '@typescript-eslint/consistent-type-exports': [
                'error',
                { fixMixedExportsWithInlineTypeSpecifier: true },
            ],
            '@typescript-eslint/consistent-type-assertions': [
                'error',
                { assertionStyle: 'never' },
            ],
            '@typescript-eslint/no-import-type-side-effects': 'error',
            '@typescript-eslint/strict-boolean-expressions': [
                'error',
                { allowNullableBoolean: true },
            ],
            '@typescript-eslint/restrict-template-expressions': ['error', {
                allowAny: false,
                allowBoolean: false,
                allowNever: false,
                allowNullish: false,
                allowNumber: true,
                allowRegExp: false,
            }],
            '@typescript-eslint/switch-exhaustiveness-check': 'error',
            '@typescript-eslint/prefer-readonly': 'error',
            '@typescript-eslint/require-array-sort-compare': 'error',
            '@typescript-eslint/no-misused-promises': [
                'error',
                { checksVoidReturn: { arguments: false } },
            ],

            'no-restricted-syntax': [
                'error',
                innerHtml,
                negation,
                ...functionKeyword,
            ],
            'arrow-body-style': [
                'error',
                'as-needed',
            ],
            'no-console': 'error',
            'no-alert': 'error',
            'no-var': 'error',
            'prefer-const': 'error',
            'prefer-template': 'error',
            'object-shorthand': 'error',
            'no-param-reassign': 'error',
            'no-nested-ternary': 'error',
            'no-else-return': 'error',
            'eqeqeq': [
                'error',
                'always',
            ],
            'curly': [
                'error',
                'all',
            ],
            'max-lines': ['error', {
                max: 300,
                skipBlankLines: true,
                skipComments: true,
            }],
            'max-depth': [
                'error',
                4,
            ],
            'sonarjs/cognitive-complexity': [
                'error',
                15,
            ],

            '@stylistic/max-len': ['error', {
                code: 120,
                ignoreUrls: true,
                ignoreRegExpLiterals: true,
            }],
            'codestyle/list-layout': 'error',
            'codestyle/return-by-name': 'error',
            '@stylistic/padding-line-between-statements': [
                'error',
                {
                    blankLine: 'always',
                    prev: '*',
                    next: 'return',
                },
                {
                    blankLine: 'always',
                    prev: [
                        'const',
                        'let',
                    ],
                    next: '*',
                },
                {
                    blankLine: 'any',
                    prev: [
                        'const',
                        'let',
                    ],
                    next: [
                        'const',
                        'let',
                    ],
                },
                {
                    blankLine: 'always',
                    prev: '*',
                    next: [
                        'interface',
                        'type',
                    ],
                },
            ],
            ...vueRules,
        },
    },
    {
        files: ['**/*.vue'],
        rules: {
            'no-undef': 'off',
            'no-restricted-syntax': [
                'error',
                innerHtml,
                negation,
                ...functionKeyword,
                componentLogic,
            ],
        },
    },
    {
        files: ['assets/shared/Icons/**'],
        rules: { 'vue/no-restricted-html-elements': 'off' },
    },
    {
        files: ['assets/shared/Tracking/Google/gtag.ts'],
        rules: { 'prefer-rest-params': 'off' },
    },
    {
        files: ['assets/shared/I18n/Dictionary/Locales/*.ts'],
        rules: { 'max-lines': 'off' },
    },
    {
        files: [
            'assets/shared/I18n/Dictionary/Locales/*.ts',
            'assets/shared/I18n/Dictionary/dictionaries.ts',
            'tests/assets/i18n/numbers/sample.ts',
        ],
        rules: {
            '@stylistic/max-len': ['error', {
                code: 120,
                ignoreUrls: true,
                ignoreStrings: true,
                ignoreTemplateLiterals: true,
                ignoreRegExpLiterals: true,
            }],
        },
    },
    {
        files: ['assets/**'],
        rules: {
            'no-restricted-imports': [
                'error',
                { patterns: [relativeImports] },
            ],
            'no-restricted-globals': [
                'error',
                ...webStorages,
            ],
            'no-restricted-properties': [
                'error',
                ...webStorages.map(({ name, message }) => ({
                    object: 'window',
                    property: name,
                    message,
                })),
            ],
        },
    },
    {
        files: ['assets/shared/Storage/Adapters/**'],
        rules: {
            'no-restricted-globals': 'off',
            'no-restricted-properties': 'off',
        },
    },
);

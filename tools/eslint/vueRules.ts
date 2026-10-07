import type { Linter } from 'eslint';
import { inlineSvg, innerHtml, negation, textAttributes } from './restrictions';

export const vueRules: Linter.RulesRecord = {
    'vue/html-indent': [
        'error',
        4,
    ],
    'vue/block-lang': [
        'error',
        { script: { lang: 'ts' } },
    ],
    'vue/block-order': ['error', {
        order: [
            'script',
            'template',
            'style',
        ],
    }],
    'vue/component-api-style': [
        'error',
        ['script-setup'],
    ],
    'vue/component-name-in-template-casing': [
        'error',
        'PascalCase',
    ],
    'vue/custom-event-name-casing': [
        'error',
        'camelCase',
    ],
    'vue/define-emits-declaration': [
        'error',
        'type-based',
    ],
    'vue/define-props-declaration': [
        'error',
        'type-based',
    ],
    'vue/define-macros-order': ['error', {
        order: [
            'defineProps',
            'defineEmits',
            'defineSlots',
        ],
        defineExposeLast: true,
    }],
    'vue/html-button-has-type': 'error',
    'vue/no-bare-strings-in-template': [
        'error',
        { attributes: { '/.+/': textAttributes } },
    ],
    'vue/no-empty-component-block': 'error',
    'vue/no-ref-object-reactivity-loss': 'error',
    'vue/no-static-inline-styles': 'error',
    'vue/no-useless-mustaches': 'error',
    'vue/no-useless-v-bind': 'error',
    'vue/no-restricted-html-elements': [
        'error',
        inlineSvg,
    ],
    'vue/no-restricted-v-bind': [
        'error',
        'innerHTML',
        'outerHTML',
    ],
    'vue/no-restricted-syntax': [
        'error',
        innerHtml,
        negation,
    ],
    'vue/no-v-text': 'error',
    'vue/padding-line-between-blocks': 'error',
    'vue/prefer-separate-static-class': 'off',
    'vue/prefer-true-attribute-shorthand': 'error',
    'vue/require-macro-variable-name': 'error',
    'vue/require-typed-ref': 'error',
    'vue/v-for-delimiter-style': [
        'error',
        'in',
    ],
};

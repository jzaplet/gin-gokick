export const relativeImports = {
    group: [
        './*',
        '../*',
    ],
    message: 'Import from @/ instead of a relative path.',
};

export const negation = {
    selector: 'UnaryExpression[operator="!"]',
    message: 'Compare with === false, !== true or === undefined instead of !, which is easy to overlook.',
};

export const innerHtml = {
    selector: 'Property[key.name=/^(?:inner|outer)HTML$/]',
    message: 'Vue writes innerHTML through its Trusted Types policy, so nothing but this rule stops it.',
};

export const textAttributes = [
    'title',
    'alt',
    'label',
    'placeholder',
    'aria-label',
    'aria-placeholder',
    'aria-roledescription',
    'aria-valuetext',
];

export const componentLogic = {
    selector: 'Program > VariableDeclaration > VariableDeclarator'
        + ' > ArrowFunctionExpression[body.type="BlockStatement"]',
    message: 'Move the state and its logic to a composable in app/<Domain>/Composables.',
};

export const inlineSvg = {
    element: 'svg',
    message: 'Add the icon to @/shared/Icons and use its component.',
};

export const webStorages = [
    'localStorage',
    'sessionStorage',
].map((name) => ({
    name,
    message: 'Store through appStorage() from @/shared/Storage/appStorage.',
}));

const arrowFunction = 'Declare an arrow function instead: const name = (…): Type => ….';

export const functionKeyword = [
    {
        selector: 'FunctionDeclaration',
        message: arrowFunction,
    },
    {
        selector: 'TSDeclareFunction',
        message: arrowFunction,
    },
    {
        selector: ':not(MethodDefinition, Property[method=true], Property[kind="get"], Property[kind="set"])'
            + ' > FunctionExpression',
        message: arrowFunction,
    },
];

import type { Rule } from 'eslint';
import { child, children, isSyntaxNode, type SyntaxNode } from '../syntax';

const functionTypes = new Set([
    'ArrowFunctionExpression',
    'FunctionExpression',
]);

const wrapperTypes = new Set([
    'TSAsExpression',
    'TSSatisfiesExpression',
    'TSNonNullExpression',
]);

const reactiveWrappers = new Set([
    'reactive',
    'readonly',
    'shallowReactive',
    'shallowReadonly',
]);

const unwrapped = (node: SyntaxNode): SyntaxNode => {
    const inner = wrapperTypes.has(node.type) ? child(node, 'expression') : undefined;

    return inner === undefined ? node : unwrapped(inner);
};

const argumentsOf = (call: SyntaxNode): SyntaxNode[] => (children(call, 'arguments') ?? []).map(unwrapped);

const wrapsReactive = (call: SyntaxNode): boolean => {
    const callee = child(call, 'callee');

    return callee?.type === 'Identifier' && typeof callee['name'] === 'string' && reactiveWrappers.has(callee['name']);
};

const returnedObjects = (returned: SyntaxNode | undefined): SyntaxNode[] => {
    if (returned === undefined) {
        return [];
    }

    const value = unwrapped(returned);

    if (value.type === 'ObjectExpression') {
        return [value];
    }

    if (value.type === 'CallExpression' && wrapsReactive(value)) {
        return argumentsOf(value).filter((argument) => argument.type === 'ObjectExpression');
    }

    return [];
};

const inlineFunction = (value: SyntaxNode): SyntaxNode | undefined => {
    const node = unwrapped(value);

    if (functionTypes.has(node.type)) {
        return node;
    }

    if (node.type === 'CallExpression') {
        return argumentsOf(node).find((argument) => functionTypes.has(argument.type));
    }

    return undefined;
};

const inlineFunctions = (object: SyntaxNode): SyntaxNode[] => {
    const inProperty = (property: SyntaxNode): SyntaxNode[] => {
        if (property.type !== 'Property' || property['kind'] !== 'init') {
            return [];
        }

        const value = child(property, 'value');

        if (value === undefined) {
            return [];
        }

        if (unwrapped(value).type === 'ObjectExpression') {
            return inlineFunctions(unwrapped(value));
        }

        const found = inlineFunction(value);

        return found === undefined ? [] : [found];
    };

    return (children(object, 'properties') ?? []).flatMap(inProperty);
};

export const returnByName: Rule.RuleModule = {
    meta: {
        type: 'suggestion',
        schema: [],
        messages: { inline: 'Declare this function as a const above the return and return the const by its name.' },
    },
    create(context) {
        const check = (returned: SyntaxNode | undefined): void => {
            for (const object of returnedObjects(returned)) {
                for (const found of inlineFunctions(object)) {
                    context.report({
                        loc: context.sourceCode.getLocFromIndex(found.range[0]),
                        messageId: 'inline',
                    });
                }
            }
        };

        const returnStatement = (node: unknown): void => {
            if (isSyntaxNode(node)) {
                check(child(node, 'argument'));
            }
        };

        const arrowFunction = (node: unknown): void => {
            if (isSyntaxNode(node) && node['expression'] === true) {
                check(child(node, 'body'));
            }
        };

        return {
            ReturnStatement: returnStatement,
            ArrowFunctionExpression: arrowFunction,
        };
    },
};

import type { AST } from 'eslint';
import { child, children, isPunctuator, isSyntaxNode, type SyntaxNode, type Tokens } from '../syntax';

export type Item = {
    node: SyntaxNode;
    first: AST.Token;
    last: AST.Token;
};

export type List = {
    node: SyntaxNode;
    open: AST.Token;
    close: AST.Token;
    items: Item[];
    delimiter: ',' | ';';
    padding: '' | ' ';
    forced: 'object' | 'array' | 'type' | undefined;
    hugs: 'none' | 'only' | 'last' | 'both';
    typed: boolean;
};

const functions = [
    'ArrowFunctionExpression',
    'FunctionExpression',
    'FunctionDeclaration',
    'TSDeclareFunction',
    'TSEmptyBodyFunctionExpression',
    'TSFunctionType',
    'TSConstructorType',
    'TSMethodSignature',
    'TSCallSignatureDeclaration',
    'TSConstructSignatureDeclaration',
];

const calls = [
    'CallExpression',
    'NewExpression',
];

const typeLists = [
    'TSTypeParameterInstantiation',
    'TSTypeParameterDeclaration',
];

const openedBeforeItems = [
    ...functions,
    'ImportDeclaration',
    'ExportNamedDeclaration',
];

const itemKeys: Record<string, string> = {
    ObjectExpression: 'properties',
    ObjectPattern: 'properties',
    ArrayExpression: 'elements',
    ArrayPattern: 'elements',
    TSTypeLiteral: 'members',
    TSTupleType: 'elementTypes',
    TSTypeParameterInstantiation: 'params',
    TSTypeParameterDeclaration: 'params',
    CallExpression: 'arguments',
    NewExpression: 'arguments',
    ImportDeclaration: 'specifiers',
    ExportNamedDeclaration: 'specifiers',
    ...Object.fromEntries(
        functions.map((type) => [
            type,
            'params',
        ]),
    ),
};

const closing: Record<string, string> = {
    '{': '}',
    '[': ']',
    '(': ')',
    '<': '>',
};

export const listTypes = Object.keys(itemKeys);

const elementsOf = (node: SyntaxNode): SyntaxNode[] | undefined => {
    const key = itemKeys[node.type];
    const elements = key === undefined ? undefined : children(node, key);

    return node.type === 'ImportDeclaration'
        ? elements?.filter((element) => element.type === 'ImportSpecifier')
        : elements;
};

const afterCallee = (node: SyntaxNode, tokens: Tokens): AST.Token | undefined => {
    const callee = child(node, 'typeArguments') ?? child(node, 'callee');
    const end = callee === undefined ? undefined : tokens.last(callee);
    let token = end === undefined ? undefined : tokens.after(end);

    while (isPunctuator(token, ')') || isPunctuator(token, '?.')) {
        token = tokens.after(token);
    }

    return token;
};

const openOf = (node: SyntaxNode, first: SyntaxNode, tokens: Tokens): AST.Token | undefined => {
    if (calls.includes(node.type)) {
        return afterCallee(node, tokens);
    }

    if (openedBeforeItems.includes(node.type)) {
        const token = tokens.first(first);

        return token === undefined ? undefined : tokens.before(token);
    }

    return tokens.first(node);
};

const itemOf = (node: SyntaxNode, open: AST.Token, delimiter: List['delimiter'], tokens: Tokens): Item | undefined => {
    let first = tokens.first(node);
    let last = tokens.last(node);

    if (delimiter === ';' && (isPunctuator(last, ';') || isPunctuator(last, ','))) {
        last = tokens.before(last);
    }

    while (first !== undefined && last !== undefined) {
        const before = tokens.before(first);
        const after = tokens.after(last);

        if (before === open || isPunctuator(before, '(') === false || isPunctuator(after, ')') === false) {
            return {
                node,
                first,
                last,
            };
        }

        first = before;
        last = after;
    }

    return undefined;
};

const isDelimiter = (token: AST.Token | undefined, delimiter: List['delimiter']): token is AST.Token =>
    isPunctuator(token, ',') || (delimiter === ';' && isPunctuator(token, ';'));

const closeOf = (
    open: AST.Token,
    items: Item[],
    delimiter: List['delimiter'],
    tokens: Tokens,
): AST.Token | undefined => {
    let expected = tokens.after(open);

    for (const [index, item] of items.entries()) {
        const next = tokens.after(item.last);

        if (item.first !== expected || (index < items.length - 1 && isDelimiter(next, delimiter) === false)) {
            return undefined;
        }

        expected = isDelimiter(next, delimiter) ? tokens.after(next) : next;
    }

    return isPunctuator(expected, closing[open.value] ?? '') ? expected : undefined;
};

const declaresType = (node: SyntaxNode): boolean => {
    const parent = node['parent'];
    const grandparent = isSyntaxNode(parent) ? parent['parent'] : undefined;

    return isSyntaxNode(parent)
        && (parent.type === 'TSTypeAliasDeclaration'
            || (parent.type === 'TSIntersectionType' && isSyntaxNode(
                grandparent,
            ) && grandparent.type === 'TSTypeAliasDeclaration'));
};

const forcedOf = (node: SyntaxNode, items: Item[]): List['forced'] => {
    if (node.type === 'ObjectExpression' && items.length >= 2) {
        return 'object';
    }

    if (node.type === 'ArrayExpression' && items.length >= 2) {
        return 'array';
    }

    return node.type === 'TSTypeLiteral' && (items.length >= 2 || declaresType(node)) ? 'type' : undefined;
};

const hugsOf = (node: SyntaxNode): List['hugs'] => {
    if (calls.includes(node.type)) {
        return 'both';
    }

    if (functions.includes(node.type)) {
        return 'only';
    }

    return node.type === 'TSTypeParameterInstantiation' || node.type === 'ArrayExpression' ? 'last' : 'none';
};

export const listOf = (node: SyntaxNode, tokens: Tokens): List | undefined => {
    const elements = elementsOf(node);
    const first = elements?.[0];
    const open = first === undefined ? undefined : openOf(node, first, tokens);

    if (elements === undefined || open?.type !== 'Punctuator' || closing[open.value] === undefined) {
        return undefined;
    }

    const delimiter = node.type === 'TSTypeLiteral' ? ';' : ',';
    const items = elements
        .map((element) => itemOf(element, open, delimiter, tokens))
        .filter((item) => item !== undefined);
    const close = items.length === elements.length ? closeOf(open, items, delimiter, tokens) : undefined;

    if (close === undefined) {
        return undefined;
    }

    return {
        node,
        open,
        close,
        items,
        delimiter,
        padding: open.value === '{' ? ' ' : '',
        forced: forcedOf(node, items),
        hugs: hugsOf(node),
        typed: typeLists.includes(node.type),
    };
};

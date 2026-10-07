import {
    expanded,
    indentOf,
    prefixOf,
    shift,
    shiftable,
    singleLine,
    type Source,
    startLine,
    suffixOf,
    textOf,
    width,
} from './layout';
import type { Item, List } from './lists';
import { child, type SyntaxNode } from '../syntax';

export type Hug = {
    width: number;
    text: string | undefined;
};

type Opened = {
    text: string | undefined;
    firstLine: string;
    lastLine: string;
};

const bracketed = [
    'ObjectExpression',
    'ArrayExpression',
    'TSTypeLiteral',
    'ObjectPattern',
];

const functions = [
    'ArrowFunctionExpression',
    'FunctionExpression',
];

const assertions = [
    'TSAsExpression',
    'TSSatisfiesExpression',
];

const annotatedOf = (node: SyntaxNode): SyntaxNode | undefined => {
    const target = node.type === 'AssignmentPattern' ? child(node, 'left') : node;
    const annotation = target?.type === 'Identifier' ? child(target, 'typeAnnotation') : undefined;
    const type = annotation === undefined ? undefined : child(annotation, 'typeAnnotation');

    return type?.type === 'TSTypeLiteral' ? type : undefined;
};

const bracketedOf = (node: SyntaxNode): SyntaxNode | undefined => {
    if (bracketed.includes(node.type)) {
        return node;
    }

    const expression = child(node, 'expression');

    return assertions.includes(node.type) && expression !== undefined && bracketed.includes(expression.type)
        ? expression
        : annotatedOf(node);
};

const openedMultiline = (source: Source, list: List, item: Item): Opened | undefined => {
    if (bracketedOf(item.node) === undefined && functions.includes(item.node.type) === false) {
        return undefined;
    }

    const text = shift(
        textOf(source, item),
        indentOf(source, startLine(list.open)) - indentOf(source, startLine(item.first)),
    );
    const rows = text.split('\n');

    return {
        text: shiftable(source, item) ? text : undefined,
        firstLine: rows[0] ?? '',
        lastLine: rows.at(-1) ?? '',
    };
};

const opened = (source: Source, list: List, item: Item, lists: Map<SyntaxNode, List>): Opened | undefined => {
    if (singleLine(item) === false) {
        return openedMultiline(source, list, item);
    }

    const node = bracketedOf(item.node);
    const inner = node === undefined ? undefined : lists.get(node);

    if (inner === undefined || (list.forced !== undefined && inner.forced === undefined)) {
        return undefined;
    }

    const base = indentOf(source, startLine(list.open));
    const head = source.text.slice(item.first.range[0], inner.open.range[1]);
    const tail = source.text.slice(inner.close.range[0], item.last.range[1]);
    const body = expanded(source, inner, base);

    return {
        text: body === undefined ? undefined : head + body + tail,
        firstLine: head,
        lastLine: ' '.repeat(base) + tail,
    };
};

export const hugLast = (source: Source, list: List, lists: Map<SyntaxNode, List>): Hug | undefined => {
    const last = list.items.at(-1);
    const others = list.items.slice(0, -1);
    const hugs = list.hugs === 'only' ? others.length === 0 : list.hugs !== 'none';
    const item = hugs === false || last === undefined || others.every(singleLine) === false
        ? undefined
        : opened(source, list, last, lists);

    if (item === undefined) {
        return undefined;
    }

    const lead = others.map((other) => `${textOf(source, other)}, `).join('');

    return {
        width: width(prefixOf(source, list.open) + lead + item.firstLine),
        text: item.text === undefined ? undefined : lead + item.text,
    };
};

export const hugFirst = (source: Source, list: List, lists: Map<SyntaxNode, List>): Hug | undefined => {
    const [first, second] = list.items;
    const item = list.hugs !== 'both' || list.items.length !== 2 || first === undefined
        ? undefined
        : opened(source, list, first, lists);

    if (item === undefined || second === undefined || singleLine(second) === false) {
        return undefined;
    }

    const rest = `, ${textOf(source, second)}`;
    const firstWidth = width(prefixOf(source, list.open) + item.firstLine);
    const lastWidth = width(item.lastLine + rest + suffixOf(source, list.close));

    return {
        width: Math.max(firstWidth, lastWidth),
        text: item.text === undefined ? undefined : item.text + rest,
    };
};

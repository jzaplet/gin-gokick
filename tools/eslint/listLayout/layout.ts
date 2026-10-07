import type { AST } from 'eslint';
import type { Item, List } from './lists';
import type { Tokens } from '../syntax';

export type Source = {
    text: string;
    lines: string[];
    tokens: Tokens;
};

export type Layout = 'flat' | 'lastHugged' | 'firstHugged' | 'expanded' | 'messy';

export const maxWidth = 120;

const indentWidth = 4;

export const startLine = (token: AST.Token): number => token.loc.start.line;

export const endLine = (token: AST.Token): number => token.loc.end.line;

export const width = (text: string): number => Array.from(text).length;

export const singleLine = (item: Item): boolean => startLine(item.first) === endLine(item.last);

const lineOf = (source: Source, line: number): string => source.lines[line - 1] ?? '';

export const indentOf = (source: Source, line: number): number => {
    const text = lineOf(source, line);

    return text.length - text.trimStart().length;
};

export const textOf = (source: Source, item: Item): string => source.text.slice(
    item.first.range[0],
    item.last.range[1],
);

export const prefixOf = (source: Source, token: AST.Token): string =>
    lineOf(source, startLine(token)).slice(0, token.loc.end.column);

export const suffixOf = (source: Source, token: AST.Token): string =>
    lineOf(source, endLine(token)).slice(token.loc.start.column);

export const contentWidth = (source: Source, list: List): number => {
    const line = lineOf(source, startLine(list.open));
    const end = startLine(list.close) === startLine(list.open) ? list.close.loc.start.column : line.length;

    return width(line.slice(list.open.loc.end.column, end));
};

const isExpanded = (list: List): boolean => {
    let previous = endLine(list.open);

    for (const item of list.items) {
        if (startLine(item.first) <= previous) {
            return false;
        }

        previous = endLine(item.last);
    }

    return startLine(list.close) > previous;
};

const isLastHugged = (list: List): boolean => {
    const line = startLine(list.open);
    const last = list.items.at(-1);

    return last !== undefined
        && list.items.slice(0, -1).every((item) => startLine(item.first) === line && endLine(item.last) === line)
        && startLine(last.first) === line
        && endLine(last.last) > line
        && startLine(list.close) === endLine(last.last);
};

const isFirstHugged = (list: List): boolean => {
    const [first, second] = list.items;
    const line = startLine(list.open);

    return list.items.length === 2
        && first !== undefined
        && second !== undefined
        && startLine(first.first) === line
        && endLine(first.last) > line
        && startLine(second.first) === endLine(first.last)
        && startLine(list.close) === endLine(second.last)
        && singleLine(second);
};

export const layoutOf = (list: List): Layout => {
    if (endLine(list.close) === startLine(list.open)) {
        return 'flat';
    }

    if (isExpanded(list)) {
        return 'expanded';
    }

    if (isLastHugged(list)) {
        return 'lastHugged';
    }

    return isFirstHugged(list) ? 'firstHugged' : 'messy';
};

export const collapsed = (source: Source, list: List): string =>
    list.padding + list.items.map((item) => textOf(source, item)).join(`${list.delimiter} `) + list.padding;

export const joinedWidth = (source: Source, list: List): number =>
    width(prefixOf(source, list.open) + collapsed(source, list) + suffixOf(source, list.close));

const shiftLine = (line: string, delta: number): string => {
    const content = line.trimStart();

    return content === '' ? '' : ' '.repeat(Math.max(line.length - content.length + delta, 0)) + content;
};

export const shift = (text: string, delta: number): string => {
    const [head = '', ...rest] = text.split('\n');

    return [
        head,
        ...rest.map((line) => shiftLine(line, delta)),
    ].join('\n');
};

export const shiftable = (source: Source, item: Item): boolean =>
    source.tokens.from(item.first, item.last).every(
        (token) => token.type !== 'Template' || startLine(token) === endLine(token),
    );

const trailing = (list: List, item: Item): string => {
    if (list.delimiter === ';') {
        return ';';
    }

    return item.node.type === 'RestElement' ? '' : ',';
};

export const expanded = (
    source: Source,
    list: List,
    base = indentOf(source, startLine(list.open)),
): string | undefined => {
    if (list.items.every((item) => shiftable(source, item)) === false) {
        return undefined;
    }

    const indent = base + indentWidth;
    const items = list.items.map((item) => {
        const text = shift(textOf(source, item), indent - indentOf(source, startLine(item.first)));

        return ' '.repeat(indent) + text + trailing(list, item);
    });

    return `\n${items.join('\n')}\n${' '.repeat(base)}`;
};

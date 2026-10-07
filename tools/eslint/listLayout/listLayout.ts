import type { Rule } from 'eslint';
import { hugFirst, hugLast } from './hugs';
import {
    collapsed,
    contentWidth,
    expanded,
    joinedWidth,
    type Layout,
    layoutOf,
    maxWidth,
    singleLine,
    type Source,
} from './layout';
import { type List, listOf, listTypes } from './lists';
import { isSyntaxNode, type SyntaxNode, tokensOf } from '../syntax';

type Verdict = {
    list: List;
    messageId: 'collapse' | 'expand' | 'hugLast' | 'hugFirst' | 'object' | 'array' | 'type';
    text: string | undefined;
    wide: boolean;
};

type Target = {
    layout: Layout;
    messageId: Verdict['messageId'];
    text: string | undefined;
    byWidth: boolean;
};

const generated = /^\/\/ Code generated .* DO NOT EDIT\.$/mu;

const targetOf = (source: Source, list: List, lists: Map<SyntaxNode, List>): Target => {
    if (list.forced !== undefined) {
        const hug = hugLast(source, list, lists);

        return hug !== undefined && hug.width <= maxWidth
            ? {
                    layout: 'lastHugged',
                    messageId: list.forced,
                    text: hug.text,
                    byWidth: false,
                }
            : {
                    layout: 'expanded',
                    messageId: list.forced,
                    text: expanded(source, list),
                    byWidth: false,
                };
    }

    if (list.items.every(singleLine) && joinedWidth(source, list) <= maxWidth) {
        return {
            layout: 'flat',
            messageId: 'collapse',
            text: collapsed(source, list),
            byWidth: false,
        };
    }

    const last = hugLast(source, list, lists);

    if (last !== undefined && last.width <= maxWidth) {
        return {
            layout: 'lastHugged',
            messageId: 'hugLast',
            text: last.text,
            byWidth: true,
        };
    }

    const first = hugFirst(source, list, lists);

    if (first !== undefined && first.width <= maxWidth) {
        return {
            layout: 'firstHugged',
            messageId: 'hugFirst',
            text: first.text,
            byWidth: true,
        };
    }

    const byWidth = last !== undefined || first !== undefined;

    return {
        layout: 'expanded',
        messageId: 'expand',
        text: expanded(source, list),
        byWidth,
    };
};

const verdictOf = (source: Source, list: List, lists: Map<SyntaxNode, List>): Verdict | undefined => {
    const layout = layoutOf(list);
    const target = targetOf(source, list, lists);

    if (target.layout === layout) {
        return undefined;
    }

    const hugged = layout === 'lastHugged' || layout === 'firstHugged';
    const wide = layout === 'flat' ? list.forced === undefined : hugged && target.byWidth;

    return {
        list,
        messageId: target.messageId,
        text: target.text,
        wide,
    };
};

const contains = (outer: List, inner: List): boolean =>
    outer !== inner && outer.open.range[0] <= inner.open.range[0] && outer.close.range[1] >= inner.close.range[1];

const lineOf = (verdict: Verdict): number => verdict.list.open.loc.start.line;

const rank = (source: Source, verdict: Verdict): number[] => [
    verdict.list.typed ? 0 : 1,
    verdict.messageId === 'expand' ? 0 : 1,
    contentWidth(source, verdict.list),
    verdict.list.open.range[0],
];

const outranks = (ranks: number[], others: number[]): boolean => {
    const index = ranks.findIndex((value, at) => value !== others[at]);

    return index !== -1 && (ranks[index] ?? 0) > (others[index] ?? 0);
};

const oneOnEachLine = (source: Source, verdicts: Verdict[]): Verdict[] => {
    const outermost = verdicts.filter((verdict) => verdicts.some((other) =>
        lineOf(other) === lineOf(verdict) && contains(other.list, verdict.list)) === false);
    const lines = new Map<number, Verdict>();

    for (const verdict of outermost) {
        const current = lines.get(lineOf(verdict));

        if (current === undefined || outranks(rank(source, verdict), rank(source, current))) {
            lines.set(lineOf(verdict), verdict);
        }
    }

    return [...lines.values()];
};

const check = (context: Rule.RuleContext, nodes: SyntaxNode[]): void => {
    const { sourceCode } = context;

    if (generated.test(sourceCode.text)) {
        return;
    }

    const tokens = tokensOf(sourceCode);
    const source = {
        text: sourceCode.text,
        lines: sourceCode.lines,
        tokens,
    };
    const comments = sourceCode.getAllComments().flatMap(
        (comment) => (comment.range === undefined ? [] : [comment.range]),
    );
    const lists = new Map<SyntaxNode, List>();

    for (const node of nodes) {
        const list = listOf(node, tokens);

        if (list !== undefined && comments.some(
            ([start, end]) => start > list.open.range[0] && end < list.close.range[1],
        ) === false) {
            lists.set(node, list);
        }
    }

    const verdicts = [...lists.values()]
        .map((list) => verdictOf(source, list, lists))
        .filter((verdict) => verdict !== undefined);
    const wide = oneOnEachLine(source, verdicts.filter((verdict) => verdict.wide));

    for (const { list, messageId, text } of [
        ...verdicts.filter((verdict) => verdict.wide === false),
        ...wide,
    ]) {
        context.report({
            loc: list.open.loc,
            messageId,
            data: { max: String(maxWidth) },
            fix: text === undefined
                ? null
                : (fixer) => fixer.replaceTextRange([
                        list.open.range[1],
                        list.close.range[0],
                    ], text),
        });
    }
};

export const listLayout: Rule.RuleModule = {
    meta: {
        type: 'layout',
        fixable: 'whitespace',
        schema: [],
        messages: {
            collapse: 'This list fits on one line of {{max}} columns: put it on one line.',
            expand: 'This list does not fit on one line of {{max}} columns: put each item on its own line.',
            hugLast: 'Start the last item on the line of the bracket, after the others.',
            hugFirst: 'Start the first item on the line of the bracket and put the second one after its end.',
            object: 'Put each property of an object with two or more on its own line.',
            array: 'Put each item of an array with two or more on its own line.',
            type: 'Put each member of this type on its own line.',
        },
    },
    create(context) {
        const nodes: SyntaxNode[] = [];
        const collect = (node: unknown): void => {
            if (isSyntaxNode(node)) {
                nodes.push(node);
            }
        };

        const programExit = (): void => {
            check(context, nodes);
        };

        return {
            ...Object.fromEntries(
                listTypes.map((type) => [
                    type,
                    collect,
                ]),
            ),
            'Program:exit': programExit,
        };
    },
};

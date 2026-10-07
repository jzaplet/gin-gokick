import type { AST, SourceCode } from 'eslint';

export type SyntaxNode = {
    type: string;
    range: AST.Range;
    [key: string]: unknown;
};

export type Tokens = {
    first: (node: SyntaxNode) => AST.Token | undefined;
    last: (node: SyntaxNode) => AST.Token | undefined;
    before: (token: AST.Token) => AST.Token | undefined;
    after: (token: AST.Token) => AST.Token | undefined;
    from: (first: AST.Token, last: AST.Token) => AST.Token[];
};

const isRange = (value: unknown): value is AST.Range =>
    Array.isArray(value) && value.length === 2 && typeof value[0] === 'number' && typeof value[1] === 'number';

export const isSyntaxNode = (value: unknown): value is SyntaxNode =>
    typeof value === 'object'
    && value !== null
    && 'type' in value
    && typeof value.type === 'string'
    && 'range' in value
    && isRange(value.range);

export const child = (node: SyntaxNode, key: string): SyntaxNode | undefined => {
    const value = node[key];

    return isSyntaxNode(value) ? value : undefined;
};

export const children = (node: SyntaxNode, key: string): SyntaxNode[] | undefined => {
    const value = node[key];

    if (Array.isArray(value) === false) {
        return undefined;
    }

    const nodes = value.filter(isSyntaxNode);

    return nodes.length === value.length ? nodes : undefined;
};

export const isPunctuator = (token: AST.Token | undefined, value: string): token is AST.Token =>
    token?.type === 'Punctuator' && token.value === value;

export const tokensOf = (sourceCode: SourceCode): Tokens => {
    const tokens = sourceCode.ast.tokens;
    const starts = new Map<number, number>(
        tokens.map((token, index) => [
            token.range[0],
            index,
        ]),
    );
    const ends = new Map<number, number>(
        tokens.map((token, index) => [
            token.range[1],
            index,
        ]),
    );
    const at = (index: number | undefined, offset = 0): AST.Token | undefined =>
        index === undefined ? undefined : tokens[index + offset];

    const first = (node: SyntaxNode): AST.Token | undefined => at(starts.get(node.range[0]));

    const last = (node: SyntaxNode): AST.Token | undefined => at(ends.get(node.range[1]));

    const before = (token: AST.Token): AST.Token | undefined => at(starts.get(token.range[0]), -1);

    const after = (token: AST.Token): AST.Token | undefined => at(starts.get(token.range[0]), 1);

    const from = (start: AST.Token, end: AST.Token): AST.Token[] =>
        tokens.slice(starts.get(start.range[0]), (starts.get(end.range[0]) ?? -1) + 1);

    return {
        first,
        last,
        before,
        after,
        from,
    };
};

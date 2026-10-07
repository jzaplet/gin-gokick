import { arrayOf, type Guard, isNumber, isRecord, optional } from '@/shared/TypeGuards/typeGuards';

export type GridResult<T> = {
    items: T[];
    total: number;
    selectable?: number;
};

export const isGridResult = <T>(item: Guard<T>): Guard<GridResult<T>> => (v: unknown): v is GridResult<T> =>
    isRecord(v)
    && arrayOf(item)(v['items'])
    && isNumber(v['total'])
    && optional(isNumber)(v['selectable']);

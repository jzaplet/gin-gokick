import { isListedSession, type ListedSession } from '@/app/Auth/types/ListedSession';
import { isSessionSort, SessionSort } from '@/app/Auth/types/SessionSort';
import { type Grid, useGrid } from '@/shared/Grid/Composables/useGrid';
import type { GridColumn } from '@/shared/Grid/types/GridColumn';
import { SortDirection } from '@/shared/Grid/types/SortDirection';

export type SessionFilters = {
    ip: string;
};

type SessionsGrid = {
    grid: Grid<SessionSort, ListedSession, SessionFilters>;
    columns: readonly GridColumn<SessionSort>[];
    maxIp: number;
};

const maxIp = 45;

const columns: readonly GridColumn<SessionSort>[] = [
    {
        key: 'device',
        label: 'sessions.device',
    },
    {
        key: 'ip',
        label: 'sessions.ip',
    },
    {
        key: 'createdAt',
        label: 'sessions.created_at',
        sort: SessionSort.CreatedAt,
    },
    {
        key: 'lastSeenAt',
        label: 'sessions.last_seen_at',
        sort: SessionSort.LastSeenAt,
    },
    {
        key: 'actions',
        label: 'grid.actions',
        align: 'right',
        hideLabel: true,
    },
];

export const useSessionsGrid = (): SessionsGrid => ({
    grid: useGrid<SessionSort, ListedSession, SessionFilters>({
        name: 'sessions',
        url: '/api/auth/sessions',
        item: isListedSession,
        isSort: isSessionSort,
        sort: {
            sortBy: SessionSort.LastSeenAt,
            sortDir: SortDirection.Descending,
        },
        filters: { ip: '' },
        isFilter: { ip: (value) => value.length <= maxIp },
        perPage: 25,
        selectable: (session) => session.current === false,
    }),
    columns,
    maxIp,
});

import type { FieldSize } from '@/shared/Inputs/types/FieldSize';
import type { StatusVariant } from '@/shared/Inputs/types/StatusVariant';

const groupSizes = {
    sm: 'space-y-2',
    md: 'space-y-1',
    lg: 'space-y-1.5',
    xl: 'space-y-2',
} satisfies Record<FieldSize, string>;

const controlSizes = {
    sm: 'rounded-lg px-3 py-1.5 text-sm',
    md: 'rounded-xl px-4 py-2 shadow-sm',
    lg: 'rounded-xl px-4 py-3 shadow-sm',
    xl: 'rounded-2xl px-6 py-4 text-lg shadow-sm',
} satisfies Record<FieldSize, string>;

const labelSizes = {
    sm: 'text-xs',
    md: 'text-sm',
    lg: 'text-sm',
    xl: 'text-base',
} satisfies Record<FieldSize, string>;

const sideRooms = {
    sm: 'pr-10',
    md: 'pr-10',
    lg: 'pr-11',
    xl: 'pr-14',
} satisfies Record<FieldSize, string>;

const sideMarks = {
    sm: 'right-3 size-4',
    md: 'right-3 size-4',
    lg: 'right-4 size-4',
    xl: 'right-5 size-5',
} satisfies Record<FieldSize, string>;

const statusColors = {
    success: 'text-emerald-700',
    info: 'text-brand-600',
} satisfies Record<StatusVariant, string>;

const borderClasses = (error: string | null, active: boolean): string => {
    if (error !== null) {
        return 'border-red-400 focus:ring-red-500';
    }

    return active ? 'border-brand-500 focus:ring-ink-900' : 'border-slate-300 focus:ring-ink-900';
};

export const groupClasses = (size: FieldSize): string => groupSizes[size];

export const labelClasses = (size: FieldSize): string[] => [
    'block font-medium text-slate-700',
    labelSizes[size],
];

export const controlClasses = (size: FieldSize, error: string | null, active: boolean): string[] => [
    'w-full border bg-white text-ink-900 transition-colors focus:outline-none focus:ring-2',
    'disabled:cursor-not-allowed disabled:bg-slate-50',
    controlSizes[size],
    borderClasses(error, active),
];

export const sideRoom = (size: FieldSize): string => sideRooms[size];

export const sideMarkClasses = (size: FieldSize): string[] => [
    'pointer-events-none absolute top-1/2 -translate-y-1/2 text-slate-500',
    sideMarks[size],
];

export const statusClasses = (variant: StatusVariant): string[] => [
    'text-sm',
    statusColors[variant],
];

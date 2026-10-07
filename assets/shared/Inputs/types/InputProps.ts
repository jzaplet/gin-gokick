import type { FieldSize } from '@/shared/Inputs/types/FieldSize';
import type { StatusVariant } from '@/shared/Inputs/types/StatusVariant';

export type InputProps = {
    name: string;
    label: string;
    autocomplete: string;
    error?: string | null | undefined;
    status?: string | null | undefined;
    statusVariant?: StatusVariant | undefined;
    required?: boolean;
    disabled?: boolean;
    loading?: boolean;
    size?: FieldSize | undefined;
    active?: boolean;
    placeholder?: string | null | undefined;
};

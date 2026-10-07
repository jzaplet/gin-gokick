import type { PlainMessageKey } from '@/shared/I18n/Texts/translate';
import type { ToastKind } from '@/shared/Toast/types/ToastKind';
import type { ToastMessage } from '@/shared/Toast/types/ToastMessage';

export type Toast = {
    id: number;
    kind: ToastKind;
    title: PlainMessageKey;
    message: ToastMessage;
};

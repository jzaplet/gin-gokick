import type { ApiMessage } from '@/shared/Fetch';
import type { PlainMessageKey } from '@/shared/I18n/Texts/translate';

export type ToastMessage = PlainMessageKey | ApiMessage;

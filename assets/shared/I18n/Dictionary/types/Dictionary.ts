import type { MessageKey } from '@/shared/I18n/Dictionary/MessageParams';
import type { Message } from '@/shared/I18n/Dictionary/types/Message';

export type Dictionary = Readonly<Record<MessageKey, Message>>;

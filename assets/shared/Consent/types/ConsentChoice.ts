import type { ConsentCategory } from '@/shared/Consent/types/ConsentCategory';

export type ConsentChoice = Partial<Record<ConsentCategory, boolean>>;

import type { MetaEvent } from '@/shared/Tracking/Meta/types/MetaEvent';

type Init = ['init', string, { em: string }?];

type Track = ['track', MetaEvent];

export type FbqCommand = Init | Track;

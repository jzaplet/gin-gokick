import type { GoogleConsent } from '@/shared/Tracking/Google/types/GoogleConsent';

type Consent = ['consent', 'default' | 'update', GoogleConsent];

type Start = ['js', Date];

type Config = ['config', string, { send_page_view: false }];

type Page = ['set', { page_location: string }];

type UserData = ['set', 'user_data', { sha256_email_address: string }];

type Event = ['event', string, { send_to: string[] }];

export type GtagCommand = Consent | Start | Config | Page | UserData | Event;

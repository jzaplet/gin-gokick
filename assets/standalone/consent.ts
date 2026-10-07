import { startConsent } from '@/shared/Consent/Start/startConsent';
import { loadPageDictionary } from '@/shared/I18n/Texts/pageDictionary';
import { startTracking } from '@/shared/Tracking/startTracking';

await loadPageDictionary();
startConsent();
startTracking();

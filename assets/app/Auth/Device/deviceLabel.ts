import { deviceName } from '@/app/Auth/Device/deviceName';
import { t } from '@/shared/I18n/Texts/translate';

export const deviceLabel = (userAgent: string): string => deviceName(userAgent) ?? t('sessions.unknown_device');

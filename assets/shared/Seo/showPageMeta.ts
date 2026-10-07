import type { RouteMeta } from 'vue-router';
import { isPlainMessageKey, t } from '@/shared/I18n/Texts/translate';

const text = (key: unknown): string | undefined => (isPlainMessageKey(key) ? t(key) : undefined);

export const showPageMeta = (route: { meta: RouteMeta }): void => {
    const title = text(route.meta['title']);

    document.title = title === undefined ? t('brand.name') : `${title} | ${t('brand.name')}`;
    document.querySelector('meta[name="description"]')?.setAttribute('content', text(route.meta['description']) ?? '');
};

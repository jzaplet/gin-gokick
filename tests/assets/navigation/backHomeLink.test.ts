import { enableAutoUnmount } from '@vue/test-utils';
import { afterEach, describe, expect, it } from 'vitest';
import { t } from '@/shared/I18n/Texts/translate';
import { openApp } from '../app/openApp';
import { setHomes } from '../i18n/languages';

enableAutoUnmount(afterEach);

describe('the login page', () => {
    it('leads back to the home page of its language, also from its named logo', async () => {
        setHomes('cs=/cs en=/');
        const links = (await openApp('/cs/app/login')).findAll('a[href="/cs"]');

        expect(links.map((link) => link.text())).toContain(t('navigation.back_home'));
        const logos = links.flatMap((link) => link.findAll('img'));

        expect(logos.map((logo) => logo.attributes('alt'))).toEqual([t('brand.name')]);
    });
});

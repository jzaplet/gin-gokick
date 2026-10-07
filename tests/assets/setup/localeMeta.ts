import { beforeEach } from 'vitest';

const resetPageLocale = (): void => {
    const meta = document.querySelector<HTMLMetaElement>('meta[name="locale"]') ?? document.createElement('meta');

    meta.name = 'locale';
    meta.content = 'cs_CZ';
    meta.dataset['homes'] = 'cs=/';
    delete meta.dataset['languages'];
    document.head.append(meta);
};

resetPageLocale();

beforeEach(resetPageLocale);

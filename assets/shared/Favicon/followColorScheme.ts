import darkMark from '@/img/mark-dark.svg';
import lightMark from '@/img/mark.svg';

export const followColorScheme = (): void => {
    const icon = document.querySelector<HTMLLinkElement>('link[rel="icon"][type="image/svg+xml"]');

    if (icon === null) {
        return;
    }

    const dark = window.matchMedia('(prefers-color-scheme: dark)');
    const show = (): void => {
        icon.href = dark.matches ? darkMark : lightMark;
    };

    show();
    dark.addEventListener('change', show);
};

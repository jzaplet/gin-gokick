import circleCheck from '@/img/icons/circle-check.svg';
import circleX from '@/img/icons/circle-x.svg';
import info from '@/img/icons/info.svg';
import triangleAlert from '@/img/icons/triangle-alert.svg';

export const preloadToastIcons = (): void => {
    for (const icon of [
        circleCheck,
        circleX,
        triangleAlert,
        info,
    ]) {
        const image = new Image();

        image.src = icon;
    }
};

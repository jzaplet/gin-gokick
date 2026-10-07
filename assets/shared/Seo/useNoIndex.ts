import { onMounted, onUnmounted } from 'vue';

export const useNoIndex = (): void => {
    const robots = document.createElement('meta');

    robots.name = 'robots';
    robots.content = 'noindex';

    onMounted(() => {
        document.head.append(robots);
    });
    onUnmounted(() => {
        robots.remove();
    });
};

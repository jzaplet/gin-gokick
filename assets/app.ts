import '@/img/icons/chevron-down.svg';
import '@/img/icons/chevron-left.svg';
import '@/img/icons/github.svg';
import '@/img/logo.svg';
import '@/img/mail/mark.png';
import '@/img/mark.svg';

import { createApp } from 'vue';
import { router } from '@/router';
import { restoreSession } from '@/shared/Auth/Session/restoreSession';
import { loadPageDictionary } from '@/shared/I18n/Texts/pageDictionary';
import { startSentry } from '@/shared/Sentry/startSentry';
import App from '@/app/App.vue';

import.meta.glob('@/img/og/*.png', { eager: true });

const app = createApp(App);

startSentry(app);
await loadPageDictionary();
void restoreSession();
app.use(router);
app.mount('#app');

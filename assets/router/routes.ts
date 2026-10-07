import type { RouteRecordSingleView } from 'vue-router';
import LoginView from '@/app/Auth/Views/LoginView.vue';
import RegisterView from '@/app/Auth/Views/RegisterView.vue';
import DashboardView from '@/app/Dashboard/Views/DashboardView.vue';
import NotFoundView from '@/app/NotFound/Views/NotFoundView.vue';
import { Access } from '@/router/types/Access';
import type { PlainMessageKey } from '@/shared/I18n/Texts/translate';

export type AppRoute = RouteRecordSingleView & {
    path: `/:lang/app${string}`;
    name: string;
    meta: {
        access: Access;
        title: PlainMessageKey;
        description: PlainMessageKey;
    };
};

export const routes: AppRoute[] = [
    {
        path: '/:lang/app/login',
        name: 'login',
        component: LoginView,
        meta: {
            access: Access.Guest,
            title: 'login.title',
            description: 'login.description',
        },
    },
    {
        path: '/:lang/app/register',
        name: 'register',
        component: RegisterView,
        meta: {
            access: Access.Guest,
            title: 'register.title',
            description: 'register.description',
        },
    },
    {
        path: '/:lang/app/dashboard',
        name: 'dashboard',
        component: DashboardView,
        meta: {
            access: Access.User,
            title: 'dashboard.title',
            description: 'dashboard.description',
        },
    },
    {
        path: '/:lang/app/:pathMatch(.*)*',
        name: 'not-found',
        component: NotFoundView,
        meta: {
            access: Access.Public,
            title: 'not_found.title',
            description: 'not_found.description',
        },
    },
];

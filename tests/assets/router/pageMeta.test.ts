import { flushPromises } from '@vue/test-utils';
import { afterEach, beforeEach, describe, expect, expectTypeOf, it } from 'vitest';
import type { RouteComponent } from 'vue-router';
import { router } from '@/router';
import { type AppRoute, routes } from '@/router/routes';
import type { Access } from '@/router/types/Access';
import { forgetSession } from '@/shared/Auth/Session/currentSession';
import { registerGridKeys } from '@/shared/Grid/Url/gridQuery';
import { type PlainMessageKey, t } from '@/shared/I18n/Texts/translate';

const description = (): string | null | undefined =>
    document.head.querySelector('meta[name="description"]')?.getAttribute('content');

describe('the title and description of a page', () => {
    beforeEach(() => {
        const meta = document.createElement('meta');

        meta.name = 'description';
        meta.content = 'V aplikaci se přihlásíte, zaregistrujete a uvidíte přehled svého účtu v Gin GoKick.';
        document.head.append(meta);
    });

    afterEach(() => {
        document.head.querySelector('meta[name="description"]')?.remove();
        forgetSession();
    });

    it.each([
        '/cs/app/login',
        '/cs/app/register',
        '/cs/app/nothing',
    ])('come from the route of %s', async (path) => {
        await router.push(path);
        await flushPromises();

        const route = routes.find(({ name }) => name === router.currentRoute.value.name);

        if (route === undefined) {
            throw new Error(`no route for ${path}`);
        }

        expect(router.currentRoute.value.fullPath).toBe(path);
        expect(document.title).toBe(`${t(route.meta.title)} | ${t('brand.name')}`);
        expect(description()).toBe(t(route.meta.description));
    });

    it('stay as they are when only the query of a grid changes, since that is no new page', async () => {
        registerGridKeys('sessions', ['page']);
        await router.push('/cs/app/register');
        await flushPromises();
        document.title = 'Rozepsáno';

        await router.push('/cs/app/register?sessionsPage=2');
        await flushPromises();

        expect(document.title).toBe('Rozepsáno');

        await router.push('/cs/app/register?sessionsPage=2&ref=newsletter');
        await flushPromises();

        expect(document.title).toBe(`${t('register.title')} | ${t('brand.name')}`);
    });

    it('are required on every route, which writes its whole path', () => {
        type Meta = {
            access: Access;
            title: PlainMessageKey;
            description: PlainMessageKey;
        };

        type Untitled = {
            path: '/:lang/app/login';
            name: string;
            component: RouteComponent;
            meta: { access: Access };
        };

        type Unprefixed = {
            path: '/login';
            name: string;
            component: RouteComponent;
            meta: Meta;
        };

        type Complete = {
            path: '/:lang/app/login';
            name: string;
            component: RouteComponent;
            meta: Meta;
        };

        expectTypeOf<Untitled>().not.toExtend<AppRoute>();
        expectTypeOf<Unprefixed>().not.toExtend<AppRoute>();
        expectTypeOf<Complete>().toExtend<AppRoute>();
    });
});

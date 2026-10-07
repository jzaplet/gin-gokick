import * as Sentry from '@sentry/vue';
import type { App } from 'vue';

const extension = /^(?:chrome|moz|safari-web)-extension:/;

const withoutFragment = (url: string): string => url.replace(/#.*$/s, '');

const withoutFragments = (event: Sentry.ErrorEvent): Sentry.ErrorEvent => {
    const request = event.request;

    if (request?.url !== undefined) {
        request.url = withoutFragment(request.url);
    }

    const referer = request?.headers?.['Referer'];

    if (request?.headers !== undefined && referer !== undefined) {
        request.headers['Referer'] = withoutFragment(referer);
    }

    return event;
};

const breadcrumbWithoutFragments = (breadcrumb: Sentry.Breadcrumb): Sentry.Breadcrumb => {
    const data = breadcrumb.data;

    if (data === undefined) {
        return breadcrumb;
    }

    for (const key of [
        'url',
        'from',
        'to',
    ]) {
        const value: unknown = data[key];

        if (typeof value === 'string') {
            data[key] = withoutFragment(value);
        }
    }

    return breadcrumb;
};

const blocked = (uri: string): string => {
    try {
        const url = new URL(uri);

        return url.protocol.startsWith('http') ? url.origin : url.protocol;
    } catch {
        return uri;
    }
};

const reportViolation = (event: SecurityPolicyViolationEvent): void => {
    if (extension.test(event.sourceFile) || extension.test(event.blockedURI)) {
        return;
    }

    const directive = event.effectiveDirective;
    const source = directive === 'trusted-types' ? event.sample : blocked(event.blockedURI);

    Sentry.captureMessage(`CSP ${directive}: ${source}`, {
        level: 'warning',
        fingerprint: [
            'csp',
            directive,
            source,
        ],
        tags: { csp_directive: directive },
        extra: {
            source_file: withoutFragment(event.sourceFile),
            line: event.lineNumber,
        },
    });
};

export const startSentry = (app: App): void => {
    const meta = document.querySelector<HTMLMetaElement>('meta[name="sentry"]');

    if (meta === null) {
        return;
    }

    const { dsn = '', environment = 'production', release = 'dev' } = meta.dataset;

    Sentry.init({
        app,
        dsn,
        environment,
        release,
        attachProps: false,
        dataCollection: { userInfo: false },
        integrations: [Sentry.browserTracingIntegration()],
        beforeSend: withoutFragments,
        beforeBreadcrumb: breadcrumbWithoutFragments,
    });
    document.addEventListener('securitypolicyviolation', reportViolation);
};

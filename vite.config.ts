import { globSync, rmSync, writeFileSync } from 'node:fs';
import { constants } from 'node:os';
import { fileURLToPath } from 'node:url';

import { sentryVitePlugin } from '@sentry/vite-plugin';
import tailwindcss from '@tailwindcss/vite';
import vue from '@vitejs/plugin-vue';
import type { Plugin, PluginOption, ViteDevServer } from 'vite';
import { defineConfig } from 'vitest/config';

const hotFile = 'public/.hot';

const outDir = 'public/build';

const fileOnly = /\.(?:woff2?|ttf|otf|eot|svg|png)$/;

const {
    SENTRY_UPLOAD_TOKEN: sentryToken = '',
    SENTRY_RELEASE: release = '',
    SENTRY_REPOSITORY: repository = '',
} = process.env;

const hot = (): Plugin => {
    const configureServer = (server: ViteDevServer): void => {
        const httpServer = server.httpServer;

        if (httpServer === null) {
            return;
        }

        const remove = (): void => {
            rmSync(hotFile, { force: true });
        };
        const exit = (signal: NodeJS.Signals): void => {
            process.exit(128 + constants.signals[signal]);
        };

        httpServer.once('listening', () => {
            const address = httpServer.address();

            if (typeof address === 'object' && address !== null) {
                writeFileSync(hotFile, `http://localhost:${String(address.port)}`);
            }
        });
        httpServer.once('close', () => {
            remove();
            process.off('exit', remove);
            process.off('SIGINT', exit);
            process.off('SIGHUP', exit);
        });
        process.on('exit', remove);
        process.on('SIGINT', exit);
        process.on('SIGHUP', exit);
    };

    return {
        name: 'hot-file',
        apply: 'serve',
        configureServer,
    };
};

const removeFailedBuild = (): Plugin => {
    const renderError = (): void => {
        rmSync(outDir, {
            recursive: true,
            force: true,
        });
    };

    return {
        name: 'remove-failed-build',
        apply: 'build',
        renderError,
    };
};

const sentry = (): PluginOption[] => {
    if (sentryToken === '') {
        return [];
    }

    return sentryVitePlugin({
        authToken: sentryToken,
        telemetry: false,
        release: {
            name: release,
            inject: false,
            setCommits: repository === ''
                ? false
                : {
                        repo: repository,
                        commit: release,
                        ignoreMissing: true,
                    },
            deploy: { env: 'production' },
        },
        sourcemaps: { filesToDeleteAfterUpload: [`${outDir}/**/*.map`] },
        errorHandler: (error) => {
            throw error;
        },
    });
};

const embeddable = (name: string): string => name.normalize('NFC').replace(/[^\p{L}\d._\-/]/gu, '_');

export default defineConfig({
    plugins: [
        tailwindcss(),
        vue(),
        hot(),
        removeFailedBuild(),
        sentry(),
    ],
    resolve: { alias: { '@': fileURLToPath(new URL('assets', import.meta.url)) } },
    appType: 'custom',
    base: '/build/',
    publicDir: false,
    server: { host: 'localhost' },
    test: {
        include: [
            'tests/assets/**/*.test.ts',
            'tests/tools/**/*.test.ts',
        ],
        setupFiles: [
            'tests/assets/setup/dialogs.ts',
            'tests/assets/setup/localeMeta.ts',
            'tests/assets/setup/pageDictionary.ts',
        ],
        environment: 'jsdom',
        restoreMocks: true,
        unstubGlobals: true,
    },
    build: {
        outDir,
        manifest: 'manifest.json',
        sourcemap: sentryToken === '' ? false : 'hidden',
        assetsInlineLimit: (file) => (fileOnly.test(file) ? false : undefined),
        rolldownOptions: {
            input: [
                'assets/app.css',
                'assets/app.ts',
                ...globSync('assets/standalone/*.ts').toSorted(),
            ],
            output: { sanitizeFileName: embeddable },
        },
    },
});

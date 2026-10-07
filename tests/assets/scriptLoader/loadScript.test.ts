import { afterEach, describe, expect, it, type Mock, vi } from 'vitest';
import { loadScript } from '@/shared/ScriptLoader/loadScript';
import { ScriptURL } from '@/shared/ScriptLoader/types/ScriptURL';
import { TrustedTypesPolicy } from '@/shared/ScriptLoader/types/TrustedTypesPolicy';

type Rules = {
    createScriptURL: (input: string) => string;
};

type TrustedURL = {
    toString: () => string;
};

type TrustedTypesStub = {
    createPolicy: Mock<(name: string, rules: Rules) => { createScriptURL: (input: string) => TrustedURL }>;
    created: TrustedURL[];
};

const loadedScripts = (): string[] => Array.from(document.head.querySelectorAll('script'), (script) => script.src);

const stubTrustedTypes = (): TrustedTypesStub => {
    const created: TrustedURL[] = [];
    const createPolicy = vi.fn((_name: string, rules: Rules) => {
        const createScriptURL = (input: string): TrustedURL => {
            const url = rules.createScriptURL(input);
            const trusted = { toString: (): string => url };

            created.push(trusted);

            return trusted;
        };

        return { createScriptURL };
    });

    vi.stubGlobal('trustedTypes', { createPolicy });

    return {
        createPolicy,
        created,
    };
};

const policyRules = ({ createPolicy }: TrustedTypesStub): Rules => {
    const rules = createPolicy.mock.calls[0]?.[1];

    if (rules === undefined) {
        throw new Error('the loader created no policy');
    }

    return rules;
};

afterEach(() => {
    for (const script of document.head.querySelectorAll('script')) {
        script.remove();
    }
});

describe('loadScript', () => {
    it('loads the same script only once', () => {
        loadScript(ScriptURL.MetaPixel);
        loadScript(ScriptURL.MetaPixel);
        loadScript(ScriptURL.GoogleTag, { id: 'G-TEST' });

        expect(loadedScripts()).toEqual([
            ScriptURL.MetaPixel,
            `${ScriptURL.GoogleTag}?id=G-TEST`,
        ]);
    });

    it('refuses a foreign URL from an untyped caller without Trusted Types too', () => {
        expect(() => {
            Reflect.apply(loadScript, undefined, ['https://evil.example/gtag/js']);
        }).toThrow(TypeError);
        expect(loadedScripts()).toEqual([]);
    });

    it('hands the script its URL as a trusted value of one policy of its own', () => {
        const { createPolicy, created } = stubTrustedTypes();
        const setSrc = vi.spyOn(HTMLScriptElement.prototype, 'src', 'set');

        loadScript(ScriptURL.MetaPixel);
        loadScript(ScriptURL.GoogleTag, { id: 'G-TEST' });

        expect(createPolicy).toHaveBeenCalledOnce();
        expect(createPolicy.mock.calls[0]?.[0]).toBe(TrustedTypesPolicy.ScriptLoader);
        expect(created).toHaveLength(2);
        expect(setSrc.mock.calls).toEqual([
            [created[0]],
            [created[1]],
        ]);
        expect(loadedScripts()).toEqual([
            ScriptURL.MetaPixel,
            `${ScriptURL.GoogleTag}?id=G-TEST`,
        ]);
    });

    it('lets its policy create only the URLs of the allowed scripts', () => {
        const trustedTypes = stubTrustedTypes();

        loadScript(ScriptURL.MetaPixel);
        const rules = policyRules(trustedTypes);

        expect(rules.createScriptURL(`${ScriptURL.GoogleTag}?id=AW-TEST`)).toBe(`${ScriptURL.GoogleTag}?id=AW-TEST`);

        for (const refused of [
            'https://evil.example/gtag/js',
            'http://www.googletagmanager.com/gtag/js',
            'https://www.googletagmanager.com:8443/gtag/js',
            'https://www.googletagmanager.com/gtm.js',
            'https://www.googletagmanager.com/gtag/js/../../gtm.js',
            'https://user@www.googletagmanager.com/gtag/js',
            'https://www.googletagmanager.com/gtag/js#x',
            '/gtag/js',
            'javascript:alert(1)',
        ]) {
            expect(() => rules.createScriptURL(refused), refused).toThrow(TypeError);
        }
    });
});

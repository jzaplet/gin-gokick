import { isScriptURL, type ScriptURL } from '@/shared/ScriptLoader/types/ScriptURL';
import { TrustedTypesPolicy } from '@/shared/ScriptLoader/types/TrustedTypesPolicy';

type ScriptURLRules = {
    createScriptURL: (input: string) => string;
};

type ScriptURLPolicy = {
    createScriptURL: (input: string) => unknown;
};

type PolicyFactory = {
    createPolicy: (name: string, rules: ScriptURLRules) => ScriptURLPolicy;
};

const policies = new WeakMap<PolicyFactory, ScriptURLPolicy>();

const isPolicyFactory = (value: unknown): value is PolicyFactory =>
    typeof value === 'object' && value !== null && 'createPolicy' in value && typeof value.createPolicy === 'function';

const allowedScriptURL = (input: string): string => {
    const url = new URL(input);
    const script = url.origin + url.pathname;

    if (isScriptURL(script) === false || url.username !== '' || url.password !== '' || url.hash !== '') {
        throw new TypeError(`The script loader refuses ${script}`);
    }

    return url.href;
};

const trustedScriptURL = (input: string): unknown => {
    const factory: unknown = Reflect.get(globalThis, 'trustedTypes');

    if (isPolicyFactory(factory) === false) {
        return allowedScriptURL(input);
    }

    const policy = policies.get(factory)
        ?? factory.createPolicy(TrustedTypesPolicy.ScriptLoader, { createScriptURL: allowedScriptURL });

    policies.set(factory, policy);

    return policy.createScriptURL(input);
};

export const loadScript = (script: ScriptURL, query: Record<string, string> = {}): void => {
    const url = new URL(script);

    url.search = new URLSearchParams(query).toString();

    if (Array.from(document.scripts).some((loaded) => loaded.src === url.href)) {
        return;
    }

    const element = document.createElement('script');

    Reflect.set(element, 'src', trustedScriptURL(url.href));
    document.head.append(element);
};

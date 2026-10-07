import { followChoice } from '@/shared/Consent/Choice/currentChoice';
import { loadScript } from '@/shared/ScriptLoader/loadScript';
import { ScriptURL } from '@/shared/ScriptLoader/types/ScriptURL';
import { fbq } from '@/shared/Tracking/Meta/fbq';
import type { MetaEvent } from '@/shared/Tracking/Meta/types/MetaEvent';
import type { EnabledTools } from '@/shared/Tracking/types/EnabledTools';
import type { TrackedUser } from '@/shared/Tracking/types/TrackedUser';
import { metaEmailHash } from '@/shared/Tracking/UserData/emailHashes';

let started = false;

let pixel = '';

export const startMetaPixel = (tools: EnabledTools): void => {
    pixel = tools.meta_pixel;

    if (pixel === '') {
        return;
    }

    followChoice((choice) => {
        if (started || choice.marketing !== true) {
            return;
        }

        started = true;
        fbq('init', pixel);
        fbq('track', 'PageView');
        loadScript(ScriptURL.MetaPixel);
    });
};

export const trackMetaEvent = async (event: MetaEvent | null, user?: TrackedUser): Promise<void> => {
    if (started === false || event === null) {
        return;
    }

    if (user !== undefined) {
        fbq('init', pixel, { em: await metaEmailHash(user.email) });
    }

    fbq('track', event);
};

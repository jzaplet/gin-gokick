import { ConsentCategory } from '@/shared/Consent/types/ConsentCategory';
import type { ConsentChoice } from '@/shared/Consent/types/ConsentChoice';
import { isRecord } from '@/shared/TypeGuards/typeGuards';
import { appStorage } from '@/shared/Storage/appStorage';
import type { EnabledTools } from '@/shared/Tracking/types/EnabledTools';
import { TrackingTool } from '@/shared/Tracking/types/TrackingTool';

const storageKey = 'consent';

const toolsKey = (tools: EnabledTools): string => new URLSearchParams(
    Object.values(TrackingTool)
        .filter((tool) => tools[tool] !== '')
        .map((tool) => [
            tool,
            tools[tool],
        ]),
).toString();

export const readStoredChoice = (tools: EnabledTools): ConsentChoice => {
    const value = appStorage().read(storageKey);
    const choice: ConsentChoice = {};
    const stored = isRecord(value) && value['tools'] === toolsKey(tools) ? value['choice'] : undefined;

    if (isRecord(stored) === false) {
        return choice;
    }

    for (const category of Object.values(ConsentCategory)) {
        const granted = stored[category];

        if (typeof granted === 'boolean') {
            choice[category] = granted;
        }
    }

    return choice;
};

export const writeStoredChoice = (tools: EnabledTools, choice: ConsentChoice): boolean =>
    appStorage().write(storageKey, {
        tools: toolsKey(tools),
        choice,
    });

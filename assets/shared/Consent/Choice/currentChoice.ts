import type { ConsentChoice } from '@/shared/Consent/types/ConsentChoice';

type ChoiceFollower = (choice: Readonly<ConsentChoice>) => void;

let current: Readonly<ConsentChoice> = {};

const followers = new Set<ChoiceFollower>();

export const currentChoice = (): Readonly<ConsentChoice> => current;

export const setCurrentChoice = (choice: ConsentChoice): void => {
    current = { ...choice };

    for (const follower of followers) {
        follower(current);
    }
};

export const followChoice = (follower: ChoiceFollower): (() => void) => {
    followers.add(follower);
    follower(current);

    return () => {
        followers.delete(follower);
    };
};

export type RecordedAnimation = {
    element: Element;
    keyframes: Keyframe[];
    options: KeyframeAnimationOptions;
    state: 'running' | 'paused' | 'cancelled';
    finish: () => void;
};

export const recordAnimations = (): RecordedAnimation[] => {
    const recorded: RecordedAnimation[] = [];

    Object.defineProperty(Element.prototype, 'animate', {
        configurable: true,
        value(this: Element, keyframes: Keyframe[], options: KeyframeAnimationOptions) {
            let finished: (() => void) | null = null;
            const record: RecordedAnimation = {
                element: this,
                keyframes,
                options,
                state: 'running',
                finish: () => {
                    finished?.();
                },
            };

            const cancel = (): void => {
                record.state = 'cancelled';
            };

            const pause = (): void => {
                record.state = 'paused';
            };

            const play = (): void => {
                record.state = 'running';
            };

            recorded.push(record);

            return {
                cancel,
                pause,
                play,
                set onfinish(handler: (() => void) | null) {
                    finished = handler;
                },
            };
        },
    });

    return recorded;
};

export const forgetAnimations = (): void => {
    Reflect.deleteProperty(Element.prototype, 'animate');
};

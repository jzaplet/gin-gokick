import type { FbqCommand } from '@/shared/Tracking/Meta/types/FbqCommand';

type Fbq = (...command: FbqCommand) => void;

const queuedFbq = (): Fbq => {
    const queue: FbqCommand[] = [];
    const stub = (...command: FbqCommand): void => {
        const callMethod: unknown = Reflect.get(stub, 'callMethod');

        if (typeof callMethod === 'function') {
            Reflect.apply(callMethod, stub, command);
        } else {
            queue.push(command);
        }
    };

    Object.assign(stub, {
        push: stub,
        loaded: true,
        version: '2.0',
        queue,
    });
    Reflect.set(window, 'fbq', stub);
    Reflect.set(window, '_fbq', stub);

    return stub;
};

export const fbq: Fbq = (...command) => {
    const installed: unknown = Reflect.get(window, 'fbq');

    Reflect.apply(typeof installed === 'function' ? installed : queuedFbq(), window, command);
};

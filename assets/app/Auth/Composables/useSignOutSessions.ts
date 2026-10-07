import { computed, type ComputedRef, readonly, type Ref, ref, shallowRef, watch, type WritableComputedRef } from 'vue';
import type { SessionFilters } from '@/app/Auth/Composables/useSessionsGrid';
import { deviceLabel } from '@/app/Auth/Device/deviceLabel';
import { type EndedSessions, isEndedSessions } from '@/app/Auth/types/EndedSessions';
import {
    type EndSessionsRequest,
    type EndSessionsRequestErrors,
    isEndSessionsRequestErrors,
} from '@/app/Auth/types/EndSessionsRequest';
import type { ListedSession } from '@/app/Auth/types/ListedSession';
import type { SessionSort } from '@/app/Auth/types/SessionSort';
import { authFetch } from '@/shared/Auth/authFetch';
import type { BulkAction } from '@/shared/BulkActions/types/BulkAction';
import type { Grid } from '@/shared/Grid/Composables/useGrid';
import { t, uiMessage } from '@/shared/I18n/Texts/translate';
import { showToast } from '@/shared/Toast/Queue/toastQueue';

type Pending = {
    kind: 'one';
    session: ListedSession;
} | { kind: 'selection' };

type Question = {
    title: string;
    message: string;
};

type SignOutSessions = {
    asking: WritableComputedRef<boolean>;
    question: ComputedRef<Question>;
    busy: Readonly<Ref<boolean>>;
    actions: readonly BulkAction<'signOut'>[];
    askOne: (session: ListedSession) => void;
    askSelected: () => void;
    signOut: () => Promise<void>;
};

const actions: readonly BulkAction<'signOut'>[] = [{
    key: 'signOut',
    label: 'sessions.sign_out',
}];

const report = (ended: number): void => {
    if (ended === 0) {
        showToast('info', 'sessions.signed_out_title', 'sessions.signed_out_none');

        return;
    }

    showToast('success', 'sessions.signed_out_title', uiMessage('sessions.signed_out', { count: ended }));
};

export const useSignOutSessions = (grid: Grid<SessionSort, ListedSession, SessionFilters>): SignOutSessions => {
    const { selection } = grid;
    const pending = shallowRef<Pending | null>(null);
    const busy = ref(false);

    const bodyOf = (target: Pending): EndSessionsRequest => {
        if (target.kind === 'one') {
            return {
                ids: [target.session.id],
                all: false,
                ip: '',
            };
        }

        return selection.all
            ? {
                    ids: [],
                    all: true,
                    ip: grid.appliedFilters.value.ip,
                }
            : {
                    ids: selection.ids(),
                    all: false,
                    ip: '',
                };
    };

    const asking = computed({
        get: () => pending.value !== null,
        set: (open) => {
            if (open === false) {
                pending.value = null;
            }
        },
    });

    const question = computed((): Question => {
        const target = pending.value;

        return target?.kind === 'one'
            ? {
                    title: t('sessions.sign_out_device'),
                    message: t('sessions.sign_out_one', { device: deviceLabel(target.session.userAgent) }),
                }
            : {
                    title: t('sessions.sign_out_selected_title'),
                    message: t('sessions.sign_out_selected', { count: selection.count }),
                };
    });

    const askOne = (session: ListedSession): void => {
        pending.value = {
            kind: 'one',
            session,
        };
    };

    const askSelected = (): void => {
        pending.value = { kind: 'selection' };
    };

    const signOut = async (): Promise<void> => {
        const target = pending.value;

        if (target === null) {
            return;
        }

        busy.value = true;
        const result = await authFetch<EndedSessions, EndSessionsRequestErrors, EndSessionsRequest>(
            'POST',
            '/api/auth/sessions/end',
            {
                body: bodyOf(target),
                validate: isEndedSessions,
                validateError: isEndSessionsRequestErrors,
            },
        );

        busy.value = false;
        pending.value = null;

        if (result.success === false) {
            showToast('error', 'toast.error_title', result.data.general ?? 'sessions.sign_out_failed');

            return;
        }

        report(result.data.ended);

        if (target.kind === 'one') {
            selection.deselect(target.session.id);
        } else {
            selection.clear();
        }

        await grid.reload();
    };

    watch(() => selection.count, (count) => {
        if (count === 0 && pending.value?.kind === 'selection') {
            pending.value = null;
        }
    });

    return {
        asking,
        question,
        busy: readonly(busy),
        actions,
        askOne,
        askSelected,
        signOut,
    };
};

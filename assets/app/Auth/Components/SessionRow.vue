<script setup lang="ts">
import { deviceLabel } from '@/app/Auth/Device/deviceLabel';
import type { ListedSession } from '@/app/Auth/types/ListedSession';
import BaseButton from '@/shared/Buttons/BaseButton.vue';
import GridCheckCell from '@/shared/Grid/Components/GridCheckCell.vue';
import type { GridSelection } from '@/shared/Grid/types/GridSelection';
import { formatDateTime } from '@/shared/I18n/Format/formatDateTime';
import { t } from '@/shared/I18n/Texts/translate';
import IconLogout from '@/shared/Icons/IconLogout.vue';
import BaseTooltip from '@/shared/Tooltip/BaseTooltip.vue';

defineProps<{
    session: ListedSession;
    selection: GridSelection<ListedSession>;
}>();

defineEmits<{ signOut: [session: ListedSession] }>();
</script>

<template>
    <tr :class="selection.selected(session) ? 'bg-brand-50/60' : undefined">
        <GridCheckCell
            :checked="selection.selected(session)"
            :disabled="selection.selectable(session) === false"
            :label="t('sessions.select', { device: deviceLabel(session.userAgent) })"
            @toggle="selection.toggle(session)"
        />
        <td class="px-4 py-3">
            <span class="flex flex-wrap items-center gap-2">
                <span
                    class="font-medium whitespace-nowrap text-ink-900"
                    :title="session.userAgent"
                >
                    {{ deviceLabel(session.userAgent) }}
                </span>
                <span
                    v-if="session.current"
                    class="rounded-full bg-brand-50 px-2 py-0.5 text-xs font-medium whitespace-nowrap text-brand-700"
                >
                    {{ t('sessions.current') }}
                </span>
            </span>
        </td>
        <td class="px-4 py-3 font-mono text-xs whitespace-nowrap text-slate-600">
            {{ session.lastSeenIp ?? t('sessions.unknown_ip') }}
        </td>
        <td class="px-4 py-3 whitespace-nowrap text-slate-600 tabular-nums">
            {{ formatDateTime(session.createdAt) }}
        </td>
        <td class="px-4 py-3 whitespace-nowrap text-slate-600 tabular-nums">
            {{ formatDateTime(session.lastSeenAt) }}
        </td>
        <td class="px-4 py-3 text-right">
            <BaseTooltip
                :text="session.current ? t('sessions.sign_out_current') : t('sessions.sign_out_device')"
                align="right"
            >
                <BaseButton
                    variant="secondary"
                    size="xs"
                    :aria-label="session.current ? t('sessions.sign_out_current') : t('sessions.sign_out_device')"
                    :disabled="session.current"
                    @click="$emit('signOut', session)"
                >
                    <IconLogout class="size-4" />
                </BaseButton>
            </BaseTooltip>
        </td>
    </tr>
</template>

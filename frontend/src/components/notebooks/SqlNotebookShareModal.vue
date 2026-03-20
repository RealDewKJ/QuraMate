<script lang="ts" setup>
import type { SqlNotebookLiveSessionStatus, SqlNotebookShareView } from '../../types/sqlNotebook';

const props = defineProps<{
    isOpen: boolean;
    targetLabel: string;
    targetDescription: string;
    selectedView: SqlNotebookShareView;
    allowLiveSession?: boolean;
    hasLiveCode?: boolean;
    previewText: string;
    expiresAt?: string;
    liveStatus?: SqlNotebookLiveSessionStatus;
    livePeerCount?: number;
    errorMessage?: string;
    isExporting?: boolean;
}>();

const emit = defineEmits<{
    close: [];
    copy: [];
    export: [];
    disconnectLive: [];
    'update:view': [value: SqlNotebookShareView];
}>();
</script>

<template>
    <div v-if="props.isOpen" class="fixed inset-0 z-[80] flex items-center justify-center bg-black/55 p-4">
        <div class="w-full max-w-3xl rounded-lg border border-border bg-card p-6 shadow-2xl">
            <div class="flex items-start justify-between gap-4">
                <div>
                    <h3 class="text-lg font-semibold text-foreground">Share {{ props.targetLabel }}</h3>
                    <p class="mt-1 text-sm text-muted-foreground">
                        {{ props.targetDescription }}
                    </p>
                </div>
                <button
                    type="button"
                    class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('close')"
                >
                    Close
                </button>
            </div>

            <div class="mt-6 flex flex-wrap gap-2">
                <button
                    type="button"
                    class="inline-flex h-9 items-center justify-center rounded-md border px-3 text-sm font-medium transition-colors"
                    :class="props.selectedView === 'summary' ? 'border-primary bg-primary/10 text-primary' : 'border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground'"
                    @click="emit('update:view', 'summary')"
                >
                    Summary
                </button>
                <button
                    v-if="props.allowLiveSession"
                    type="button"
                    class="inline-flex h-9 items-center justify-center rounded-md border px-3 text-sm font-medium transition-colors"
                    :class="props.selectedView === 'live' ? 'border-primary bg-primary/10 text-primary' : 'border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground'"
                    @click="emit('update:view', 'live')"
                >
                    SessionShare
                </button>
            </div>

            <div class="mt-4 rounded-lg border border-border bg-muted/10 p-3">
                <div class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Preview</div>
                <div class="mt-1 text-xs text-muted-foreground">
                    {{ props.selectedView === 'live'
                        ? 'Start a temporary SessionShare so connected peers see notebook updates while the session is active.'
                        : 'Summary export is easy to paste into chat, docs, or tickets.' }}
                </div>
                <textarea
                    :value="props.previewText"
                    readonly
                    :rows="props.selectedView === 'summary' ? 18 : 5"
                    class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-xs text-foreground outline-none"
                />
                <div v-if="props.expiresAt" class="mt-2 text-xs text-muted-foreground">
                    Expires: {{ props.expiresAt }}
                </div>
                <div v-if="props.selectedView === 'live' && props.liveStatus" class="mt-2 text-xs text-muted-foreground">
                    Status: {{ props.liveStatus }}
                    <span v-if="typeof props.livePeerCount === 'number'">
                        | {{ props.livePeerCount }} peer{{ props.livePeerCount === 1 ? '' : 's' }}
                    </span>
                </div>
                <div v-if="props.errorMessage" class="mt-3 rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
                    {{ props.errorMessage }}
                </div>
            </div>

            <div class="mt-4 flex flex-wrap justify-end gap-3">
                <button
                    type="button"
                    class="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('copy')"
                >
                    Copy
                </button>
                <button
                    v-if="props.selectedView === 'live' && (props.liveStatus === 'hosting' || props.liveStatus === 'connected')"
                    type="button"
                    class="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 text-sm font-medium text-destructive transition-colors hover:bg-destructive/10"
                    @click="emit('disconnectLive')"
                >
                    End SessionShare
                </button>
                <button
                    type="button"
                    class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                    :disabled="props.isExporting"
                    @click="emit('export')"
                >
                    {{ props.isExporting ? 'Saving...' : props.selectedView === 'live' ? (props.hasLiveCode ? 'Reconnect SessionShare' : 'Start SessionShare') : 'Export Markdown' }}
                </button>
            </div>
        </div>
    </div>
</template>

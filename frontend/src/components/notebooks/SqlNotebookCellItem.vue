<script lang="ts" setup>
import { computed } from 'vue';

import MarkdownCellEditor from './cells/MarkdownCellEditor.vue';
import RunbookCellEditor from './cells/RunbookCellEditor.vue';
import SqlNotebookCellResults from './SqlNotebookCellResults.vue';
import SqlCellEditor from './cells/SqlCellEditor.vue';

import { parseRunbookContent } from '../../types/sqlNotebook';
import type { SqlNotebookCell, SqlNotebookCellRunResult, SqlNotebookEmbeddedImage } from '../../types/sqlNotebook';

const props = defineProps<{
    cell: SqlNotebookCell;
    index: number;
    totalCells: number;
    result: SqlNotebookCellRunResult | null;
    isActive: boolean;
    tables: string[];
    getColumns: (tableName: string) => Promise<string[]>;
    editorSettings: { fontFamily: string; fontSize: number };
    isReadOnly?: boolean;
}>();

const emit = defineEmits<{
    'update-title': [payload: { cellId: string; value: string }];
    'update-content': [payload: { cellId: string; value: string }];
    'update-embedded-images': [payload: { cellId: string; value: SqlNotebookEmbeddedImage[] }];
    'update-collapsed': [payload: { cellId: string; value: boolean }];
    'update-execution-state': [payload: { cellId: string; value: 'idle' | 'running' | 'success' | 'error' | 'verified' | 'skipped' }];
    'delete': [cellId: string];
    'duplicate': [cellId: string];
    'move': [payload: { cellId: string; direction: 'up' | 'down' }];
    'run': [cellId: string];
    'save-snapshot': [cellId: string];
    share: [cellId: string];
    activate: [cellId: string];
}>();

const typeLabelByCellType = {
    sql: 'SQL',
    markdown: 'Notes',
    runbook: 'Runbook',
} as const;

const typeToneByCellType = {
    sql: 'border-sky-500/30 bg-sky-500/10 text-sky-700 dark:text-sky-300',
    markdown: 'border-violet-500/30 bg-violet-500/10 text-violet-700 dark:text-violet-300',
    runbook: 'border-amber-500/30 bg-amber-500/10 text-amber-700 dark:text-amber-300',
} as const;

const statusToneByExecutionState = {
    idle: 'border-border bg-background text-muted-foreground',
    running: 'border-primary/30 bg-primary/5 text-primary',
    success: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
    error: 'border-destructive/30 bg-destructive/10 text-destructive',
    verified: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
    skipped: 'border-amber-500/30 bg-amber-500/10 text-amber-700',
} as const;

const collapseLabel = computed(() => {
    const typeLabel = props.cell.type === 'sql' ? 'SQL' : props.cell.type === 'runbook' ? 'Runbook' : 'Notes';
    return props.cell.collapsed ? `Expand ${typeLabel}` : `Collapse ${typeLabel}`;
});

const collapsedPreview = computed(() => {
    if (props.cell.type === 'runbook') {
        return parseRunbookContent(props.cell.content).objective || 'Collapsed runbook step';
    }

    return props.cell.content.slice(0, 180) || (props.cell.type === 'sql' ? 'Collapsed SQL cell' : 'Collapsed notes cell');
});

const sqlRunContext = computed(() => {
    if (props.cell.type !== 'sql') {
        return null;
    }

    const latestTimestamp = props.result?.completedAt || props.result?.startedAt || props.cell.lastRunAt;
    const timestampLabel = latestTimestamp
        ? new Intl.DateTimeFormat(undefined, {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        }).format(new Date(latestTimestamp))
        : 'Not run yet';

    if (!props.result) {
        return {
            summary: props.cell.lastRunAt ? `Last run ${timestampLabel}` : 'Ready to run in notebook canvas',
            timestampLabel,
        };
    }

    if (props.result.status === 'running') {
        return {
            summary: 'Running now',
            timestampLabel,
        };
    }

    if (props.result.status === 'error') {
        return {
            summary: props.result.errorMessage || 'Last run failed',
            timestampLabel,
        };
    }

    const resultSetCount = props.result.resultSets.length;
    const rowCount = props.result.totalRows ?? 0;
    return {
        summary: `${resultSetCount} result set${resultSetCount === 1 ? '' : 's'} • ${rowCount} row${rowCount === 1 ? '' : 's'}`,
        timestampLabel,
    };
});

const toggleCollapsed = () => {
    emit('update-collapsed', {
        cellId: props.cell.id,
        value: !props.cell.collapsed,
    });
};

const iconButtonClass = 'inline-flex h-8 w-8 items-center justify-center rounded-md border border-input text-muted-foreground transition-colors hover:bg-accent hover:text-foreground disabled:cursor-not-allowed disabled:opacity-40';
const deleteButtonClass = 'inline-flex h-8 w-8 items-center justify-center rounded-md border border-input text-destructive transition-colors hover:bg-destructive/10 disabled:cursor-not-allowed disabled:opacity-40';
</script>

<template>
    <article
        :id="`sql-notebook-cell-${props.cell.id}`"
        tabindex="-1"
        class="min-w-0 w-full max-w-full overflow-visible rounded-xl border border-border bg-card shadow-sm outline-none transition-colors"
        :class="props.isActive ? 'bg-muted/10' : ''"
        @click="emit('activate', props.cell.id)"
    >
        <header class="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3">
            <div class="flex items-center gap-3">
                <span
                    class="rounded-full border px-2.5 py-1 text-[11px] font-semibold uppercase tracking-[0.14em]"
                    :class="typeToneByCellType[props.cell.type]"
                >
                    {{ typeLabelByCellType[props.cell.type] }}
                </span>
                <span
                    v-if="props.cell.type === 'sql' || props.cell.type === 'runbook'"
                    class="rounded-full border px-2.5 py-1 text-[11px] font-medium capitalize"
                    :class="statusToneByExecutionState[props.cell.executionState]"
                >
                    {{ props.cell.executionState }}
                </span>
                <span class="text-xs text-muted-foreground">Cell {{ props.index + 1 }} of {{ props.totalCells }}</span>
                <button
                    class="inline-flex h-8 items-center justify-center rounded-md border border-input px-2.5 text-xs text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                    @click.stop="toggleCollapsed"
                >
                    {{ collapseLabel }}
                </button>
            </div>

            <div class="flex items-center gap-2">
                <button
                    v-if="props.cell.type === 'sql'"
                    :class="iconButtonClass"
                    title="Share cell"
                    @click.stop="emit('share', props.cell.id)"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <circle cx="18" cy="5" r="3" />
                        <circle cx="6" cy="12" r="3" />
                        <circle cx="18" cy="19" r="3" />
                        <path d="m8.59 13.51 6.83 3.98" />
                        <path d="m15.41 6.51-6.82 3.98" />
                    </svg>
                </button>
                <button
                    :class="iconButtonClass"
                    :disabled="props.index === 0"
                    title="Move up"
                    @click.stop="emit('move', { cellId: props.cell.id, direction: 'up' })"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="m18 15-6-6-6 6" />
                    </svg>
                </button>
                <button
                    :class="iconButtonClass"
                    :disabled="props.index === props.totalCells - 1"
                    title="Move down"
                    @click.stop="emit('move', { cellId: props.cell.id, direction: 'down' })"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="m6 9 6 6 6-6" />
                    </svg>
                </button>
                <button
                    :class="iconButtonClass"
                    title="Duplicate cell"
                    @click.stop="emit('duplicate', props.cell.id)"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                        <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                    </svg>
                </button>
                <button
                    :class="deleteButtonClass"
                    :disabled="props.totalCells <= 1"
                    title="Delete cell"
                    @click.stop="emit('delete', props.cell.id)"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M3 6h18" />
                        <path d="M8 6V4h8v2" />
                        <path d="M19 6l-1 14H6L5 6" />
                        <path d="M10 11v6" />
                        <path d="M14 11v6" />
                    </svg>
                </button>
            </div>
        </header>

        <div v-if="!props.cell.collapsed" class="min-w-0 w-full max-w-full p-4">
            <div
                v-if="props.cell.type === 'sql' && sqlRunContext"
                class="mb-4 flex flex-wrap items-center justify-between gap-3 rounded-lg border border-border bg-muted/20 px-3 py-2 text-xs"
            >
                <div class="font-medium text-foreground">
                    {{ sqlRunContext.summary }}
                </div>
                <div class="text-muted-foreground">
                    {{ sqlRunContext.timestampLabel }}
                </div>
            </div>

            <SqlCellEditor
                v-if="props.cell.type === 'sql'"
                :title="props.cell.title"
                :content="props.cell.content"
                :tables="props.tables"
                :get-columns="props.getColumns"
                :editor-settings="props.editorSettings"
                :is-read-only="props.isReadOnly"
                @update:title="emit('update-title', { cellId: props.cell.id, value: $event })"
                @update:content="emit('update-content', { cellId: props.cell.id, value: $event })"
                @run="emit('run', props.cell.id)"
            />
    <MarkdownCellEditor
            v-else-if="props.cell.type === 'markdown'"
            :title="props.cell.title"
            :content="props.cell.content"
            :embedded-images="props.cell.embeddedImages"
            @update:title="emit('update-title', { cellId: props.cell.id, value: $event })"
            @update:content="emit('update-content', { cellId: props.cell.id, value: $event })"
            @update:embedded-images="emit('update-embedded-images', { cellId: props.cell.id, value: $event })"
        />
            <RunbookCellEditor
                v-else
                :title="props.cell.title"
                :content="props.cell.content"
                :execution-state="props.cell.executionState"
                @update:title="emit('update-title', { cellId: props.cell.id, value: $event })"
                @update:content="emit('update-content', { cellId: props.cell.id, value: $event })"
                @update:status="emit('update-execution-state', { cellId: props.cell.id, value: $event })"
                @focus="emit('activate', props.cell.id)"
            />

            <SqlNotebookCellResults
                v-if="props.cell.type === 'sql' && props.result"
                :result="props.result"
                :source-label="props.cell.title"
                class="mt-4"
                @focus-source="emit('activate', props.cell.id)"
                @rerun="emit('run', props.cell.id)"
                @save-snapshot="emit('save-snapshot', props.cell.id)"
            />
        </div>

        <div
            v-else
            class="px-4 py-3 text-sm text-muted-foreground"
        >
            {{ collapsedPreview }}
        </div>
    </article>
</template>

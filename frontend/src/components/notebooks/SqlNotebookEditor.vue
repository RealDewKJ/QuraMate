<script lang="ts" setup>
import { computed } from 'vue';

import SqlNotebookCellList from './SqlNotebookCellList.vue';
import SqlNotebookHeader from './SqlNotebookHeader.vue';

import type { SqlNotebook, SqlNotebookCellRunResult, SqlNotebookEmbeddedImage } from '../../types/sqlNotebook';

const props = defineProps<{
    notebook: SqlNotebook | null;
    resultsByCellId: Record<string, SqlNotebookCellRunResult>;
    activeCellId: string | null;
    tables: string[];
    getColumns: (tableName: string) => Promise<string[]>;
    editorSettings: { fontFamily: string; fontSize: number };
    isDirty: boolean;
    isSaving: boolean;
    lastSavedAt: string | null;
    isReadOnly?: boolean;
}>();

const emit = defineEmits<{
    'update-title': [payload: { notebookId: string; value: string }];
    'update-description': [payload: { notebookId: string; value: string }];
    'update-tags': [value: string[]];
    'update-metadata': [value: { environment?: string; purpose?: string; owner?: string }];
    'toggle-favorite': [];
    'add-cell': [type: 'sql' | 'markdown' | 'runbook'];
    share: [];
    'update-cell-title': [payload: { cellId: string; value: string }];
    'update-cell-content': [payload: { cellId: string; value: string }];
    'update-cell-embedded-images': [payload: { cellId: string; value: SqlNotebookEmbeddedImage[] }];
    'update-cell-collapsed': [payload: { cellId: string; value: boolean }];
    'update-cell-execution-state': [payload: { cellId: string; value: 'idle' | 'running' | 'success' | 'error' | 'verified' | 'skipped' }];
    'delete-cell': [cellId: string];
    'duplicate-cell': [cellId: string];
    'move-cell': [payload: { cellId: string; direction: 'up' | 'down' }];
    'run-cell': [cellId: string];
    'save-snapshot': [cellId: string];
    'share-cell': [cellId: string];
    'activate-cell': [cellId: string];
    save: [];
}>();

const sqlCellCount = computed(() => {
    return props.notebook?.cells.filter((cell) => cell.type === 'sql').length ?? 0;
});

const markdownCellCount = computed(() => {
    return props.notebook?.cells.filter((cell) => cell.type === 'markdown').length ?? 0;
});

const runbookCellCount = computed(() => {
    return props.notebook?.cells.filter((cell) => cell.type === 'runbook').length ?? 0;
});
</script>

<template>
    <div v-if="props.notebook" class="flex min-h-0 min-w-0 w-full max-w-full flex-1 flex-col gap-4 overflow-auto p-4">
        <SqlNotebookHeader
            :notebook="props.notebook"
            :sql-cell-count="sqlCellCount"
            :markdown-cell-count="markdownCellCount"
            :runbook-cell-count="runbookCellCount"
            :is-dirty="props.isDirty"
            :is-saving="props.isSaving"
            :last-saved-at="props.lastSavedAt"
            :is-read-only="props.isReadOnly"
            @update-title="emit('update-title', { notebookId: props.notebook.id, value: $event })"
            @update-description="emit('update-description', { notebookId: props.notebook.id, value: $event })"
            @update-tags="emit('update-tags', $event)"
            @update-metadata="emit('update-metadata', $event)"
            @toggle-favorite="emit('toggle-favorite')"
            @add-cell="emit('add-cell', $event)"
            @share="emit('share')"
            @save="emit('save')"
        />

        <SqlNotebookCellList
            :cells="props.notebook.cells"
            :results-by-cell-id="props.resultsByCellId"
            :active-cell-id="props.activeCellId"
            :tables="props.tables"
            :get-columns="props.getColumns"
            :editor-settings="props.editorSettings"
            :is-read-only="props.isReadOnly"
            @update-title="emit('update-cell-title', $event)"
            @update-content="emit('update-cell-content', $event)"
            @update-embedded-images="emit('update-cell-embedded-images', $event)"
            @update-collapsed="emit('update-cell-collapsed', $event)"
            @update-execution-state="emit('update-cell-execution-state', $event)"
            @delete-cell="emit('delete-cell', $event)"
            @duplicate-cell="emit('duplicate-cell', $event)"
            @move-cell="emit('move-cell', $event)"
            @run-cell="emit('run-cell', $event)"
            @save-snapshot="emit('save-snapshot', $event)"
            @share-cell="emit('share-cell', $event)"
            @activate-cell="emit('activate-cell', $event)"
        />
    </div>

    <div v-else class="flex flex-1 items-center justify-center p-8 text-sm text-muted-foreground">
        Select or create a notebook to start structuring repeatable SQL work.
    </div>
</template>

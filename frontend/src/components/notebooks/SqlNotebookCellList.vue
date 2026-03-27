<script lang="ts" setup>
import SqlNotebookCellItem from './SqlNotebookCellItem.vue';

import type { SqlNotebookCell, SqlNotebookCellRunResult, SqlNotebookEmbeddedImage } from '../../types/sqlNotebook';

const props = defineProps<{
    cells: SqlNotebookCell[];
    resultsByCellId: Record<string, SqlNotebookCellRunResult>;
    activeCellId: string | null;
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
    'delete-cell': [cellId: string];
    'duplicate-cell': [cellId: string];
    'move-cell': [payload: { cellId: string; direction: 'up' | 'down' }];
    'run-cell': [cellId: string];
    'save-snapshot': [cellId: string];
    'share-cell': [cellId: string];
    'activate-cell': [cellId: string];
}>();
</script>

<template>
    <div class="min-w-0 w-full max-w-full space-y-4">
        <SqlNotebookCellItem
            v-for="(cell, index) in props.cells"
            :key="cell.id"
            :cell="cell"
            :index="index"
            :total-cells="props.cells.length"
            :result="props.resultsByCellId[cell.id] ?? null"
            :is-active="props.activeCellId === cell.id"
            :tables="props.tables"
            :get-columns="props.getColumns"
            :editor-settings="props.editorSettings"
            :is-read-only="props.isReadOnly"
            @update-title="emit('update-title', $event)"
            @update-content="emit('update-content', $event)"
            @update-embedded-images="emit('update-embedded-images', $event)"
            @update-collapsed="emit('update-collapsed', $event)"
            @update-execution-state="emit('update-execution-state', $event)"
            @delete="emit('delete-cell', $event)"
            @duplicate="emit('duplicate-cell', $event)"
            @move="emit('move-cell', $event)"
            @run="emit('run-cell', $event)"
            @save-snapshot="emit('save-snapshot', $event)"
            @share="emit('share-cell', $event)"
            @activate="emit('activate-cell', $event)"
        />
    </div>
</template>

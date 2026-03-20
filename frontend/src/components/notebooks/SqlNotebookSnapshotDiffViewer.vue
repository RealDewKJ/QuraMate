<script lang="ts" setup>
import { computed, ref } from 'vue';

import type { SqlNotebookResultSet, SqlNotebookResultSnapshot } from '../../types/sqlNotebook';

const props = defineProps<{
    baseSnapshot: SqlNotebookResultSnapshot | null;
    compareSnapshot: SqlNotebookResultSnapshot | null;
}>();

interface ResultSetRowValueDiff {
    column: string;
    baseValue: string;
    compareValue: string;
}

interface ResultSetRowPreview {
    keyLabel: string;
    values: Record<string, string>;
}

interface ResultSetChangedRowPreview {
    keyLabel: string;
    values: ResultSetRowValueDiff[];
}

interface ResultSetDiffSummary {
    index: number;
    addedColumns: string[];
    removedColumns: string[];
    changedRows: number;
    addedRows: number;
    removedRows: number;
    matchMode: string;
    inferredKey: string | null;
    activeKey: string | null;
    changedRowPreviews: ResultSetChangedRowPreview[];
    addedRowPreviews: ResultSetRowPreview[];
    removedRowPreviews: ResultSetRowPreview[];
    baseChangedCells: string[];
    compareChangedCells: string[];
    baseRemovedRows: number[];
    compareAddedRows: number[];
}

const MAX_ROW_PREVIEW_COUNT = 3;
const ignoreRowOrder = ref(true);
const selectedKeyByResultSet = ref<Record<number, string>>({});
const expandedResultSetIds = ref<number[]>([]);

const preferredKeyNames = ['id', 'uuid', 'guid', 'key', 'code', 'slug', 'name'];

const stableValue = (value: unknown): string => {
    if (Array.isArray(value)) {
        return `[${value.map((item) => stableValue(item)).join(',')}]`;
    }

    if (value && typeof value === 'object') {
        const entries = Object.entries(value as Record<string, unknown>)
            .sort(([left], [right]) => left.localeCompare(right))
            .map(([key, item]) => `${key}:${stableValue(item)}`);
        return `{${entries.join(',')}}`;
    }

    return JSON.stringify(value);
};

const formatValue = (value: unknown): string => {
    if (value === null) {
        return 'null';
    }
    if (typeof value === 'undefined') {
        return '-';
    }
    if (typeof value === 'string') {
        return value;
    }

    return stableValue(value);
};

const inferPrimaryKeyColumn = (baseResultSet?: SqlNotebookResultSet, compareResultSet?: SqlNotebookResultSet): string | null => {
    const baseColumns = baseResultSet?.columns || [];
    const compareColumns = compareResultSet?.columns || [];
    const commonColumns = baseColumns.filter((column) => compareColumns.includes(column));
    if (commonColumns.length === 0) {
        return null;
    }

    const candidates = [
        ...preferredKeyNames.filter((name) => commonColumns.includes(name)),
        ...commonColumns.filter((column) => column.endsWith('_id')),
    ];

    for (const candidate of candidates) {
        const baseValues = (baseResultSet?.rows || []).map((row) => stableValue(row[candidate]));
        const compareValues = (compareResultSet?.rows || []).map((row) => stableValue(row[candidate]));
        const uniqueInBase = baseValues.length > 0 && new Set(baseValues).size === baseValues.length;
        const uniqueInCompare = compareValues.length > 0 && new Set(compareValues).size === compareValues.length;
        if (uniqueInBase && uniqueInCompare) {
            return candidate;
        }
    }

    return null;
};

const getCommonColumns = (baseResultSet?: SqlNotebookResultSet, compareResultSet?: SqlNotebookResultSet): string[] => {
    const baseColumns = baseResultSet?.columns || [];
    const compareColumns = compareResultSet?.columns || [];
    return baseColumns.filter((column) => compareColumns.includes(column));
};

const buildRowSignature = (row: Record<string, unknown>, keyColumn: string | null) => {
    if (!keyColumn) {
        return stableValue(row);
    }

    return `${keyColumn}:${stableValue(row[keyColumn])}`;
};

const buildRowPreview = (row: Record<string, unknown>, columns: string[], keyColumn: string | null, fallbackIndex: number): ResultSetRowPreview => {
    const previewColumns = columns.slice(0, 4);
    const values = Object.fromEntries(previewColumns.map((column) => [column, formatValue(row[column])]));
    const keyValue = keyColumn ? formatValue(row[keyColumn]) : `Row ${fallbackIndex + 1}`;

    return {
        keyLabel: keyColumn ? `${keyColumn}: ${keyValue}` : keyValue,
        values,
    };
};

const buildChangedRowPreview = (
    baseRow: Record<string, unknown>,
    compareRow: Record<string, unknown>,
    columns: string[],
    keyColumn: string | null,
    fallbackIndex: number,
): ResultSetChangedRowPreview | null => {
    const valueDiffs = columns
        .filter((column) => stableValue(baseRow[column]) !== stableValue(compareRow[column]))
        .map((column) => ({
            column,
            baseValue: formatValue(baseRow[column]),
            compareValue: formatValue(compareRow[column]),
        }));

    if (valueDiffs.length === 0) {
        return null;
    }

    const keyValue = keyColumn ? formatValue(compareRow[keyColumn] ?? baseRow[keyColumn]) : `Row ${fallbackIndex + 1}`;
    return {
        keyLabel: keyColumn ? `${keyColumn}: ${keyValue}` : keyValue,
        values: valueDiffs,
    };
};

const buildResultSetDiff = (index: number, baseResultSet?: SqlNotebookResultSet, compareResultSet?: SqlNotebookResultSet): ResultSetDiffSummary => {
    const baseColumns = baseResultSet?.columns || [];
    const compareColumns = compareResultSet?.columns || [];
    const commonColumns = getCommonColumns(baseResultSet, compareResultSet);
    const addedColumns = compareColumns.filter((column) => !baseColumns.includes(column));
    const removedColumns = baseColumns.filter((column) => !compareColumns.includes(column));

    const baseRows = baseResultSet?.rows || [];
    const compareRows = compareResultSet?.rows || [];
    const manualKey = selectedKeyByResultSet.value[index] || '';
    const inferredKey = inferPrimaryKeyColumn(baseResultSet, compareResultSet);
    const activeKey = manualKey === '__auto__' || !manualKey ? inferredKey : manualKey;
    let changedRows = 0;
    let addedRows = 0;
    let removedRows = 0;
    let matchMode = 'Exact row compare';
    const changedRowPreviews: ResultSetChangedRowPreview[] = [];
    const addedRowPreviews: ResultSetRowPreview[] = [];
    const removedRowPreviews: ResultSetRowPreview[] = [];
    const baseChangedCells: string[] = [];
    const compareChangedCells: string[] = [];
    const baseRemovedRows: number[] = [];
    const compareAddedRows: number[] = [];

    if (activeKey) {
        const baseMap = new Map(baseRows.map((row, rowIndex) => [buildRowSignature(row, activeKey), { row, rowIndex }]));
        const compareMap = new Map(compareRows.map((row, rowIndex) => [buildRowSignature(row, activeKey), { row, rowIndex }]));

        for (const [key, baseEntry] of baseMap.entries()) {
            const compareEntry = compareMap.get(key);
            if (!compareEntry) {
                removedRows += 1;
                baseRemovedRows.push(baseEntry.rowIndex);
                if (removedRowPreviews.length < MAX_ROW_PREVIEW_COUNT) {
                    removedRowPreviews.push(buildRowPreview(baseEntry.row, baseColumns, activeKey, baseEntry.rowIndex));
                }
                continue;
            }

            const changedColumns = commonColumns.filter((column) => stableValue(baseEntry.row[column]) !== stableValue(compareEntry.row[column]));
            const preview = buildChangedRowPreview(baseEntry.row, compareEntry.row, commonColumns, activeKey, compareEntry.rowIndex);
            if (preview) {
                changedRows += 1;
                for (const column of changedColumns) {
                    baseChangedCells.push(`${baseEntry.rowIndex}:${column}`);
                    compareChangedCells.push(`${compareEntry.rowIndex}:${column}`);
                }
                if (changedRowPreviews.length < MAX_ROW_PREVIEW_COUNT) {
                    changedRowPreviews.push(preview);
                }
            }
        }

        for (const [key, compareEntry] of compareMap.entries()) {
            if (!baseMap.has(key)) {
                addedRows += 1;
                compareAddedRows.push(compareEntry.rowIndex);
                if (addedRowPreviews.length < MAX_ROW_PREVIEW_COUNT) {
                    addedRowPreviews.push(buildRowPreview(compareEntry.row, compareColumns, activeKey, compareEntry.rowIndex));
                }
            }
        }

        matchMode = `Primary key: ${activeKey}`;
    } else if (ignoreRowOrder.value) {
        const baseCounts = new Map<string, Array<{ row: Record<string, unknown>; rowIndex: number }>>();
        const compareCounts = new Map<string, Array<{ row: Record<string, unknown>; rowIndex: number }>>();
        for (const [rowIndex, row] of baseRows.entries()) {
            const signature = stableValue(row);
            const items = baseCounts.get(signature) || [];
            items.push({ row, rowIndex });
            baseCounts.set(signature, items);
        }
        for (const [rowIndex, row] of compareRows.entries()) {
            const signature = stableValue(row);
            const items = compareCounts.get(signature) || [];
            items.push({ row, rowIndex });
            compareCounts.set(signature, items);
        }

        const signatures = new Set([...baseCounts.keys(), ...compareCounts.keys()]);
        for (const signature of signatures) {
            const baseEntries = [...(baseCounts.get(signature) || [])];
            const compareEntries = [...(compareCounts.get(signature) || [])];
            const sharedCount = Math.min(baseEntries.length, compareEntries.length);
            if (compareEntries.length > sharedCount) {
                const extras = compareEntries.slice(sharedCount);
                addedRows += extras.length;
                for (const entry of extras) {
                    compareAddedRows.push(entry.rowIndex);
                    if (addedRowPreviews.length < MAX_ROW_PREVIEW_COUNT) {
                        addedRowPreviews.push(buildRowPreview(entry.row, compareColumns, null, entry.rowIndex));
                    }
                }
            }
            if (baseEntries.length > sharedCount) {
                const extras = baseEntries.slice(sharedCount);
                removedRows += extras.length;
                for (const entry of extras) {
                    baseRemovedRows.push(entry.rowIndex);
                    if (removedRowPreviews.length < MAX_ROW_PREVIEW_COUNT) {
                        removedRowPreviews.push(buildRowPreview(entry.row, baseColumns, null, entry.rowIndex));
                    }
                }
            }
        }

        changedRows = Math.min(addedRows, removedRows);
        matchMode = 'Row-order agnostic';
    } else {
        const rowPairs = Math.min(baseRows.length, compareRows.length);
        for (let rowIndex = 0; rowIndex < rowPairs; rowIndex += 1) {
            const changedColumns = commonColumns.filter((column) => stableValue(baseRows[rowIndex][column]) !== stableValue(compareRows[rowIndex][column]));
            const preview = buildChangedRowPreview(baseRows[rowIndex], compareRows[rowIndex], commonColumns, null, rowIndex);
            if (preview) {
                changedRows += 1;
                for (const column of changedColumns) {
                    baseChangedCells.push(`${rowIndex}:${column}`);
                    compareChangedCells.push(`${rowIndex}:${column}`);
                }
                if (changedRowPreviews.length < MAX_ROW_PREVIEW_COUNT) {
                    changedRowPreviews.push(preview);
                }
            }
        }

        if (compareRows.length > rowPairs) {
            const extras = compareRows.slice(rowPairs);
            addedRows += extras.length;
            extras.forEach((row, extraIndex) => {
                compareAddedRows.push(rowPairs + extraIndex);
                if (addedRowPreviews.length < MAX_ROW_PREVIEW_COUNT) {
                    addedRowPreviews.push(buildRowPreview(row, compareColumns, null, rowPairs + extraIndex));
                }
            });
        }

        if (baseRows.length > rowPairs) {
            const extras = baseRows.slice(rowPairs);
            removedRows += extras.length;
            extras.forEach((row, extraIndex) => {
                baseRemovedRows.push(rowPairs + extraIndex);
                if (removedRowPreviews.length < MAX_ROW_PREVIEW_COUNT) {
                    removedRowPreviews.push(buildRowPreview(row, baseColumns, null, rowPairs + extraIndex));
                }
            });
        }
    }

    return {
        index,
        addedColumns,
        removedColumns,
        changedRows,
        addedRows,
        removedRows,
        matchMode,
        inferredKey,
        activeKey,
        changedRowPreviews,
        addedRowPreviews,
        removedRowPreviews,
        baseChangedCells,
        compareChangedCells,
        baseRemovedRows,
        compareAddedRows,
    };
};

const commonColumnsByResultSet = computed(() => {
    if (!props.baseSnapshot || !props.compareSnapshot) {
        return {} as Record<number, string[]>;
    }

    const totalSets = Math.max(props.baseSnapshot.resultSets.length, props.compareSnapshot.resultSets.length);
    return Object.fromEntries(
        Array.from({ length: totalSets }, (_, index) => [
            index,
            getCommonColumns(props.baseSnapshot?.resultSets[index], props.compareSnapshot?.resultSets[index]),
        ]),
    );
});

const resultSetDiffs = computed(() => {
    if (!props.baseSnapshot || !props.compareSnapshot) {
        return [] as ResultSetDiffSummary[];
    }

    const totalSets = Math.max(props.baseSnapshot.resultSets.length, props.compareSnapshot.resultSets.length);
    return Array.from({ length: totalSets }, (_, index) => (
        buildResultSetDiff(index, props.baseSnapshot?.resultSets[index], props.compareSnapshot?.resultSets[index])
    ));
});

const overallDiff = computed(() => {
    return resultSetDiffs.value.reduce(
        (summary, diff) => ({
            addedColumns: summary.addedColumns + diff.addedColumns.length,
            removedColumns: summary.removedColumns + diff.removedColumns.length,
            changedRows: summary.changedRows + diff.changedRows,
            addedRows: summary.addedRows + diff.addedRows,
            removedRows: summary.removedRows + diff.removedRows,
        }),
        {
            addedColumns: 0,
            removedColumns: 0,
            changedRows: 0,
            addedRows: 0,
            removedRows: 0,
        },
    );
});

const getDisplayColumns = (index: number): string[] => {
    const baseColumns = props.baseSnapshot?.resultSets[index]?.columns || [];
    const compareColumns = props.compareSnapshot?.resultSets[index]?.columns || [];
    return Array.from(new Set([...baseColumns, ...compareColumns]));
};

const isResultSetExpanded = (index: number): boolean => {
    return expandedResultSetIds.value.includes(index);
};

const toggleResultSetExpanded = (index: number) => {
    expandedResultSetIds.value = isResultSetExpanded(index)
        ? expandedResultSetIds.value.filter((item) => item !== index)
        : [...expandedResultSetIds.value, index];
};

const getRowLabel = (row: Record<string, unknown>, rowIndex: number, keyColumn: string | null): string => {
    if (!keyColumn) {
        return `Row ${rowIndex + 1}`;
    }

    return `${keyColumn}: ${formatValue(row[keyColumn])}`;
};

const isChangedCell = (side: 'base' | 'compare', diff: ResultSetDiffSummary, rowIndex: number, column: string): boolean => {
    const key = `${rowIndex}:${column}`;
    return side === 'base'
        ? diff.baseChangedCells.includes(key)
        : diff.compareChangedCells.includes(key);
};

const isChangedRow = (side: 'base' | 'compare', diff: ResultSetDiffSummary, rowIndex: number): boolean => {
    const prefix = `${rowIndex}:`;
    return (side === 'base' ? diff.baseChangedCells : diff.compareChangedCells).some((key) => key.startsWith(prefix));
};

const isRemovedRow = (diff: ResultSetDiffSummary, rowIndex: number): boolean => {
    return diff.baseRemovedRows.includes(rowIndex);
};

const isAddedRow = (diff: ResultSetDiffSummary, rowIndex: number): boolean => {
    return diff.compareAddedRows.includes(rowIndex);
};
</script>

<template>
    <div v-if="!baseSnapshot || !compareSnapshot" class="rounded-lg border border-dashed border-border p-3 text-xs text-muted-foreground">
        Select two snapshots to compare result changes.
    </div>

    <div v-else class="space-y-3">
        <div class="flex items-center justify-between gap-3 rounded-lg border border-border bg-background/80 px-3 py-2">
            <div>
                <div class="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Comparison Mode</div>
                <div class="mt-1 text-xs text-muted-foreground">Prefer primary-key matches when possible, otherwise compare rows without depending on their order.</div>
            </div>
            <label class="inline-flex items-center gap-2 text-xs text-muted-foreground">
                <input v-model="ignoreRowOrder" type="checkbox" class="h-4 w-4 rounded border-input" />
                Ignore row order
            </label>
        </div>

        <div class="grid gap-3 sm:grid-cols-2">
            <div class="rounded-lg border border-border bg-background/80 p-3">
                <div class="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Base Snapshot</div>
                <div class="mt-1 text-sm font-medium">{{ baseSnapshot.cellTitle }}</div>
                <div class="mt-1 text-[11px] text-muted-foreground">{{ baseSnapshot.resultSets.length }} sets • {{ baseSnapshot.totalRows }} rows</div>
            </div>
            <div class="rounded-lg border border-border bg-background/80 p-3">
                <div class="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Compare Snapshot</div>
                <div class="mt-1 text-sm font-medium">{{ compareSnapshot.cellTitle }}</div>
                <div class="mt-1 text-[11px] text-muted-foreground">{{ compareSnapshot.resultSets.length }} sets • {{ compareSnapshot.totalRows }} rows</div>
            </div>
        </div>

        <div class="grid gap-3 sm:grid-cols-2 lg:grid-cols-5">
            <div class="rounded-lg border border-emerald-500/30 bg-emerald-500/10 p-3 text-center">
                <div class="text-[11px] uppercase tracking-[0.12em] text-emerald-700">Added Rows</div>
                <div class="mt-1 text-lg font-semibold text-emerald-800">{{ overallDiff.addedRows }}</div>
            </div>
            <div class="rounded-lg border border-destructive/30 bg-destructive/10 p-3 text-center">
                <div class="text-[11px] uppercase tracking-[0.12em] text-destructive">Removed Rows</div>
                <div class="mt-1 text-lg font-semibold text-destructive">{{ overallDiff.removedRows }}</div>
            </div>
            <div class="rounded-lg border border-amber-500/30 bg-amber-500/10 p-3 text-center">
                <div class="text-[11px] uppercase tracking-[0.12em] text-amber-700">Changed Rows</div>
                <div class="mt-1 text-lg font-semibold text-amber-800">{{ overallDiff.changedRows }}</div>
            </div>
            <div class="rounded-lg border border-primary/30 bg-primary/10 p-3 text-center">
                <div class="text-[11px] uppercase tracking-[0.12em] text-primary">Added Columns</div>
                <div class="mt-1 text-lg font-semibold text-primary">{{ overallDiff.addedColumns }}</div>
            </div>
            <div class="rounded-lg border border-muted bg-muted/60 p-3 text-center">
                <div class="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Removed Columns</div>
                <div class="mt-1 text-lg font-semibold text-foreground">{{ overallDiff.removedColumns }}</div>
            </div>
        </div>

        <div class="space-y-2">
            <div
                v-for="diff in resultSetDiffs"
                :key="diff.index"
                class="rounded-lg border border-border bg-background/80 p-3"
            >
                <div class="flex flex-wrap items-center justify-between gap-2">
                    <div class="text-sm font-medium">Result Set {{ diff.index + 1 }}</div>
                    <div class="text-[11px] text-muted-foreground">{{ diff.matchMode }}</div>
                </div>

                <label class="mt-2 block">
                    <span class="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Primary Key</span>
                    <select
                        v-model="selectedKeyByResultSet[diff.index]"
                        class="mt-2 h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none transition-colors focus:border-primary"
                    >
                        <option value="__auto__">Auto infer</option>
                        <option value="">No key</option>
                        <option
                            v-for="column in commonColumnsByResultSet[diff.index] || []"
                            :key="column"
                            :value="column"
                        >
                            {{ column }}
                        </option>
                    </select>
                </label>

                <div class="mt-2 grid gap-2 sm:grid-cols-2">
                    <div class="text-xs text-muted-foreground">Added rows: <span class="font-medium text-foreground">{{ diff.addedRows }}</span></div>
                    <div class="text-xs text-muted-foreground">Removed rows: <span class="font-medium text-foreground">{{ diff.removedRows }}</span></div>
                    <div class="text-xs text-muted-foreground">Changed rows: <span class="font-medium text-foreground">{{ diff.changedRows }}</span></div>
                    <div class="text-xs text-muted-foreground">Primary key: <span class="font-medium text-foreground">{{ diff.activeKey || '-' }}</span></div>
                    <div class="text-xs text-muted-foreground">
                        Added columns:
                        <span class="font-medium text-foreground">{{ diff.addedColumns.length > 0 ? diff.addedColumns.join(', ') : '-' }}</span>
                    </div>
                    <div class="text-xs text-muted-foreground sm:col-span-2">
                        Removed columns:
                        <span class="font-medium text-foreground">{{ diff.removedColumns.length > 0 ? diff.removedColumns.join(', ') : '-' }}</span>
                    </div>
                </div>

                <div class="mt-3">
                    <button
                        type="button"
                        class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium transition-colors hover:bg-accent"
                        @click="toggleResultSetExpanded(diff.index)"
                    >
                        {{ isResultSetExpanded(diff.index) ? 'Hide Full Compare' : 'Open Full Compare' }}
                    </button>
                </div>

                <div class="mt-3 space-y-3">
                    <div v-if="diff.changedRowPreviews.length > 0" class="rounded-lg border border-amber-500/30 bg-amber-500/5 p-3">
                        <div class="text-[11px] font-semibold uppercase tracking-[0.12em] text-amber-700">Changed Values Preview</div>
                        <div class="mt-2 space-y-2">
                            <div
                                v-for="preview in diff.changedRowPreviews"
                                :key="preview.keyLabel"
                                class="rounded-md border border-amber-500/20 bg-background/80 p-3"
                            >
                                <div class="text-xs font-medium text-foreground">{{ preview.keyLabel }}</div>
                                <div class="mt-2 space-y-2">
                                    <div
                                        v-for="valueDiff in preview.values"
                                        :key="`${preview.keyLabel}-${valueDiff.column}`"
                                        class="grid gap-2 rounded-md border border-border bg-card/70 p-2 text-xs md:grid-cols-[120px_minmax(0,1fr)_minmax(0,1fr)]"
                                    >
                                        <div class="font-medium text-foreground">{{ valueDiff.column }}</div>
                                        <div class="min-w-0 rounded bg-muted/70 px-2 py-1 text-muted-foreground">{{ valueDiff.baseValue }}</div>
                                        <div class="min-w-0 rounded bg-amber-500/10 px-2 py-1 text-foreground">{{ valueDiff.compareValue }}</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div v-if="diff.addedRowPreviews.length > 0 || diff.removedRowPreviews.length > 0" class="grid gap-3 lg:grid-cols-2">
                        <div v-if="diff.addedRowPreviews.length > 0" class="rounded-lg border border-emerald-500/30 bg-emerald-500/5 p-3">
                            <div class="text-[11px] font-semibold uppercase tracking-[0.12em] text-emerald-700">Added Rows Preview</div>
                            <div class="mt-2 space-y-2">
                                <div
                                    v-for="preview in diff.addedRowPreviews"
                                    :key="preview.keyLabel"
                                    class="rounded-md border border-emerald-500/20 bg-background/80 p-3"
                                >
                                    <div class="text-xs font-medium text-foreground">{{ preview.keyLabel }}</div>
                                    <div class="mt-2 grid gap-2">
                                        <div
                                            v-for="(value, column) in preview.values"
                                            :key="`${preview.keyLabel}-${column}`"
                                            class="grid gap-2 text-xs sm:grid-cols-[110px_minmax(0,1fr)]"
                                        >
                                            <div class="font-medium text-foreground">{{ column }}</div>
                                            <div class="min-w-0 rounded bg-muted/70 px-2 py-1 text-muted-foreground">{{ value }}</div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>

                        <div v-if="diff.removedRowPreviews.length > 0" class="rounded-lg border border-destructive/30 bg-destructive/5 p-3">
                            <div class="text-[11px] font-semibold uppercase tracking-[0.12em] text-destructive">Removed Rows Preview</div>
                            <div class="mt-2 space-y-2">
                                <div
                                    v-for="preview in diff.removedRowPreviews"
                                    :key="preview.keyLabel"
                                    class="rounded-md border border-destructive/20 bg-background/80 p-3"
                                >
                                    <div class="text-xs font-medium text-foreground">{{ preview.keyLabel }}</div>
                                    <div class="mt-2 grid gap-2">
                                        <div
                                            v-for="(value, column) in preview.values"
                                            :key="`${preview.keyLabel}-${column}`"
                                            class="grid gap-2 text-xs sm:grid-cols-[110px_minmax(0,1fr)]"
                                        >
                                            <div class="font-medium text-foreground">{{ column }}</div>
                                            <div class="min-w-0 rounded bg-muted/70 px-2 py-1 text-muted-foreground">{{ value }}</div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>

                    <div
                        v-if="diff.changedRows > 0 && diff.changedRowPreviews.length === 0 && !diff.activeKey && ignoreRowOrder"
                        class="rounded-lg border border-dashed border-border p-3 text-xs text-muted-foreground"
                    >
                        Changed rows were detected after row-order-agnostic matching. Pick a primary key to preview field-level changes.
                    </div>
                    <div v-if="isResultSetExpanded(diff.index)" class="rounded-lg border border-border bg-card/70 p-3">
                        <div class="mb-3 flex flex-wrap items-center justify-between gap-2">
                            <div class="text-[11px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">Full Compare Table</div>
                            <div class="text-[11px] text-muted-foreground">Amber cells changed, red rows removed, green rows added.</div>
                        </div>
                        <div class="grid gap-3 xl:grid-cols-2">
                            <div class="min-w-0 rounded-lg border border-border bg-background/80">
                                <div class="border-b border-border px-3 py-2 text-xs font-medium text-foreground">
                                    Base Rows ({{ props.baseSnapshot?.resultSets[diff.index]?.rows.length || 0 }})
                                </div>
                                <div class="max-h-96 overflow-auto">
                                    <table class="min-w-full border-separate border-spacing-0 text-xs">
                                        <thead class="sticky top-0 bg-background">
                                            <tr>
                                                <th class="border-b border-border px-3 py-2 text-left font-semibold text-muted-foreground">Row</th>
                                                <th
                                                    v-for="column in getDisplayColumns(diff.index)"
                                                    :key="`base-${diff.index}-${column}`"
                                                    class="border-b border-border px-3 py-2 text-left font-semibold text-muted-foreground"
                                                >
                                                    {{ column }}
                                                </th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            <tr
                                                v-for="(row, rowIndex) in props.baseSnapshot?.resultSets[diff.index]?.rows || []"
                                                :key="`base-row-${diff.index}-${rowIndex}`"
                                                class="align-top"
                                                :class="{
                                                    'bg-destructive/5': isRemovedRow(diff, rowIndex),
                                                    'bg-amber-500/5': isChangedRow('base', diff, rowIndex) && !isRemovedRow(diff, rowIndex),
                                                }"
                                            >
                                                <td
                                                    class="border-b border-border/70 px-3 py-2 font-medium text-foreground"
                                                    :class="{
                                                        'text-destructive': isRemovedRow(diff, rowIndex),
                                                        'text-amber-800': isChangedRow('base', diff, rowIndex) && !isRemovedRow(diff, rowIndex),
                                                    }"
                                                >
                                                    {{ getRowLabel(row, rowIndex, diff.activeKey) }}
                                                </td>
                                                <td
                                                    v-for="column in getDisplayColumns(diff.index)"
                                                    :key="`base-cell-${diff.index}-${rowIndex}-${column}`"
                                                    class="border-b border-border/70 px-3 py-2 text-muted-foreground"
                                                    :class="{
                                                        'bg-amber-500/15 font-medium text-foreground': isChangedCell('base', diff, rowIndex, column),
                                                        'bg-destructive/10 text-destructive': isRemovedRow(diff, rowIndex),
                                                    }"
                                                >
                                                    {{ formatValue(row[column]) }}
                                                </td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </div>
                            </div>

                            <div class="min-w-0 rounded-lg border border-border bg-background/80">
                                <div class="border-b border-border px-3 py-2 text-xs font-medium text-foreground">
                                    Compare Rows ({{ props.compareSnapshot?.resultSets[diff.index]?.rows.length || 0 }})
                                </div>
                                <div class="max-h-96 overflow-auto">
                                    <table class="min-w-full border-separate border-spacing-0 text-xs">
                                        <thead class="sticky top-0 bg-background">
                                            <tr>
                                                <th class="border-b border-border px-3 py-2 text-left font-semibold text-muted-foreground">Row</th>
                                                <th
                                                    v-for="column in getDisplayColumns(diff.index)"
                                                    :key="`compare-${diff.index}-${column}`"
                                                    class="border-b border-border px-3 py-2 text-left font-semibold text-muted-foreground"
                                                >
                                                    {{ column }}
                                                </th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            <tr
                                                v-for="(row, rowIndex) in props.compareSnapshot?.resultSets[diff.index]?.rows || []"
                                                :key="`compare-row-${diff.index}-${rowIndex}`"
                                                class="align-top"
                                                :class="{
                                                    'bg-emerald-500/5': isAddedRow(diff, rowIndex),
                                                    'bg-amber-500/5': isChangedRow('compare', diff, rowIndex) && !isAddedRow(diff, rowIndex),
                                                }"
                                            >
                                                <td
                                                    class="border-b border-border/70 px-3 py-2 font-medium text-foreground"
                                                    :class="{
                                                        'text-emerald-700': isAddedRow(diff, rowIndex),
                                                        'text-amber-800': isChangedRow('compare', diff, rowIndex) && !isAddedRow(diff, rowIndex),
                                                    }"
                                                >
                                                    {{ getRowLabel(row, rowIndex, diff.activeKey) }}
                                                </td>
                                                <td
                                                    v-for="column in getDisplayColumns(diff.index)"
                                                    :key="`compare-cell-${diff.index}-${rowIndex}-${column}`"
                                                    class="border-b border-border/70 px-3 py-2 text-muted-foreground"
                                                    :class="{
                                                        'bg-amber-500/15 font-medium text-foreground': isChangedCell('compare', diff, rowIndex, column),
                                                        'bg-emerald-500/10 text-emerald-700': isAddedRow(diff, rowIndex),
                                                    }"
                                                >
                                                    {{ formatValue(row[column]) }}
                                                </td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

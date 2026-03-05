<script lang="ts" setup>
import { ref } from 'vue';
import { useEventListener } from '@vueuse/core';

import DatePicker from '../DatePicker.vue';
import DbQueryResultsTabs from './DbQueryResultsTabs.vue';
import DbQueryResultGrid from './DbQueryResultGrid.vue';

interface Props {
    activeTab: any;
    isReadOnly?: boolean;
    filteredResults?: any[];
    openAiCopilot: (mode?: any) => void;
    toggleSort: (col: string) => void;
    startColumnResize: (e: MouseEvent, col: string) => void;
    handleCellClick: (item: any, col: string, rsIndex: number) => void;
    handleRowContextMenu: (e: MouseEvent, row: any, col: string, rowIndex?: number | string) => void;
    getEditorType: (col: string) => string;
    saveCellEdit: (item: any, col: string) => void | Promise<void>;
    isImageValue: (value: any, col: string) => boolean;
    openImagePreview: (url: any) => void;
    openMockDataModal: () => void;
    openInsertRowModal: () => void;
    startResultSetResize: (e: MouseEvent, resultSetIndex: number, resultSet: any) => void;
    getResultSetCardStyle: (resultSet: any, resultSetIndex: number) => Record<string, string> | string | undefined;
}

const props = defineProps<Props>();

const selectedRowIndex = defineModel<Array<number | string> | number | string | null>('selectedRowIndex', { required: true });
const selectedColumn = defineModel<string | null>('selectedColumn', { required: true });
const selectedRowData = defineModel<any>('selectedRowData');

const rootRef = ref<HTMLElement | null>(null);

const collapsedResultSets = ref<Record<number, boolean>>({});

const toggleResultSetCollapse = (index: number) => {
    collapsedResultSets.value[index] = !collapsedResultSets.value[index];
};

const lastClickedRow = ref<number | string | null>(null);

const isRowSelected = (index: number | string) => {
    if (Array.isArray(selectedRowIndex.value)) {
        return selectedRowIndex.value.includes(index);
    }
    return selectedRowIndex.value === index && !selectedColumn.value; // Only highlight full row if no specific column selected
};

// --- Multi-Cell Drag Selection ---
const dragStartCell = ref<{ row: string | number, col: string } | null>(null);
const dragCurrentCell = ref<{ row: string | number, col: string } | null>(null);
const draggedCellSelection = ref<{ row: string | number, col: string }[]>([]);

const getRowIndexNumber = (formattedIndex: string | number): { rsIndex: number, rIndex: number } => {
    if (typeof formattedIndex === 'number') {
        return { rsIndex: 0, rIndex: formattedIndex };
    }
    const parts = formattedIndex.split('-');
    return { rsIndex: parseInt(parts[1]) + 1, rIndex: parseInt(parts[2]) };
};

const calculateDragSelection = () => {
    if (!dragStartCell.value || !dragCurrentCell.value) return;

    draggedCellSelection.value = [];

    const startObj = getRowIndexNumber(dragStartCell.value.row);
    const currObj = getRowIndexNumber(dragCurrentCell.value.row);

    // Only allow drag selection within the same result set
    if (startObj.rsIndex !== currObj.rsIndex) return;

    const rsIndex = startObj.rsIndex;
    const resultSet = props.activeTab.resultSets[rsIndex];
    if (!resultSet || !resultSet.columns) return;

    const minRow = Math.min(startObj.rIndex, currObj.rIndex);
    const maxRow = Math.max(startObj.rIndex, currObj.rIndex);

    const startColIndex = resultSet.columns.indexOf(dragStartCell.value.col);
    const currColIndex = resultSet.columns.indexOf(dragCurrentCell.value.col);

    if (startColIndex === -1 || currColIndex === -1) return;

    const minCol = Math.min(startColIndex, currColIndex);
    const maxCol = Math.max(startColIndex, currColIndex);

    const newSelection = [];

    for (let r = minRow; r <= maxRow; r++) {
        const rowId = rsIndex === 0 ? r : `sub-${rsIndex - 1}-${r}`;
        for (let c = minCol; c <= maxCol; c++) {
            newSelection.push({ row: rowId, col: resultSet.columns[c] });
        }
    }

    draggedCellSelection.value = newSelection;
};

const handleCellDragStart = (rowIndex: number | string, col: string) => {
    draggedCellSelection.value = []; // Clear previous drag
    dragStartCell.value = { row: rowIndex, col };
    dragCurrentCell.value = { row: rowIndex, col };
    calculateDragSelection();
};

const handleCellDragEnter = (rowIndex: number | string, col: string) => {
    if (dragStartCell.value) {
        dragCurrentCell.value = { row: rowIndex, col };
        calculateDragSelection();
    }
};

let dragJustEnded = false;

const handleCellDragEnd = () => {
    // Commit the drag selection to the main selection model if it's more than one cell
    if (draggedCellSelection.value.length > 1) {
        dragJustEnded = true;
        setTimeout(() => { dragJustEnded = false; }, 300);

        // Collect all unique row indices
        const uniqueRows = Array.from(new Set(draggedCellSelection.value.map(cell => cell.row)));

        // Update selection state
        selectedRowIndex.value = uniqueRows;

        // We set selectedColumn to null when range selecting so we don't just highlight one specific cell's border
        selectedColumn.value = null;
        lastClickedRow.value = uniqueRows[uniqueRows.length - 1];
    } else if (draggedCellSelection.value.length === 1) {
        // Single cell click handled by original logic, but we still clean up
        const cell = draggedCellSelection.value[0];
        selectedRowIndex.value = [cell.row];
        selectedColumn.value = cell.col;
        lastClickedRow.value = cell.row;
    }

    dragStartCell.value = null;
    dragCurrentCell.value = null;
};

const isCellSelected = (rowIndex: number | string, col: string): boolean => {
    // If we have a multi-cell selection (either dragging or finalized)
    if (draggedCellSelection.value.length > 1) {
        return draggedCellSelection.value.some(cell => cell.row === rowIndex && cell.col === col);
    }

    // Otherwise rely on the committed multi-row selection + whatever was copied
    if (Array.isArray(selectedRowIndex.value)) {
        if (selectedRowIndex.value.includes(rowIndex)) {
            return !selectedColumn.value || selectedColumn.value === col;
        }
    } else if (selectedRowIndex.value === rowIndex) {
        return selectedColumn.value === col;
    }

    return false;
};

const handleRowSelectorClick = (e: MouseEvent, itemIndex: number | string) => {
    let currentSelection = Array.isArray(selectedRowIndex.value)
        ? [...selectedRowIndex.value]
        : (selectedRowIndex.value !== null ? [selectedRowIndex.value] : []);

    if (e.shiftKey && lastClickedRow.value !== null && typeof itemIndex === typeof lastClickedRow.value) {
        if (typeof itemIndex === 'number' && typeof lastClickedRow.value === 'number') {
            const start = Math.min(itemIndex, lastClickedRow.value);
            const end = Math.max(itemIndex, lastClickedRow.value);
            if (!e.metaKey && !e.ctrlKey) {
                currentSelection = [];
            }
            for (let i = start; i <= end; i++) {
                if (!currentSelection.includes(i)) currentSelection.push(i);
            }
            if (lastClickedRow.value !== null && !currentSelection.includes(lastClickedRow.value)) {
                currentSelection.push(lastClickedRow.value);
            }
        } else if (typeof itemIndex === 'string' && typeof lastClickedRow.value === 'string') {
            const parts1 = itemIndex.split('-');
            const parts2 = lastClickedRow.value.split('-');
            if (parts1[1] === parts2[1]) {
                const start = Math.min(parseInt(parts1[2]), parseInt(parts2[2]));
                const end = Math.max(parseInt(parts1[2]), parseInt(parts2[2]));
                if (!e.metaKey && !e.ctrlKey) {
                    currentSelection = [];
                }
                for (let i = start; i <= end; i++) {
                    const idxStr = `sub-${parts1[1]}-${i}`;
                    if (!currentSelection.includes(idxStr)) currentSelection.push(idxStr);
                }
                if (lastClickedRow.value !== null && !currentSelection.includes(lastClickedRow.value)) {
                    currentSelection.push(lastClickedRow.value);
                }
            }
        }
    } else if (e.metaKey || e.ctrlKey) {
        const idx = currentSelection.indexOf(itemIndex);
        if (idx > -1) {
            currentSelection.splice(idx, 1);
        } else {
            currentSelection.push(itemIndex);
        }
    } else {
        if (currentSelection.length === 1 && currentSelection[0] === itemIndex) {
            currentSelection = [];
        } else {
            currentSelection = [itemIndex];
        }
    }

    selectedRowIndex.value = currentSelection;
    if (!e.shiftKey || lastClickedRow.value === null) {
        lastClickedRow.value = itemIndex;
    }
    selectedColumn.value = null;
};

const handleCellClickCustom = (itemIndex: number | string, col: string) => {
    if (dragJustEnded) return;

    // If clicking a single cell, clear the multi-cell drag highlight cache
    draggedCellSelection.value = [];

    selectedRowIndex.value = [itemIndex];
    lastClickedRow.value = itemIndex;
    selectedColumn.value = col;
};

// --- Copy Selection (Ctrl+C) ---
const copyCurrentSelection = () => {
    if (draggedCellSelection.value.length > 0) {
        // 1. Multi-cell dragged selection
        const cells = [...draggedCellSelection.value].sort((a, b) => {
            const aObj = getRowIndexNumber(a.row);
            const bObj = getRowIndexNumber(b.row);
            if (aObj.rsIndex !== bObj.rsIndex) return aObj.rsIndex - bObj.rsIndex;
            if (aObj.rIndex !== bObj.rIndex) return aObj.rIndex - bObj.rIndex;
            return 0; // Col order was maintained during selection
        });

        const rsIndex = getRowIndexNumber(cells[0].row).rsIndex;
        const resultSet = props.activeTab.resultSets[rsIndex];
        if (!resultSet || !resultSet.rows) return;

        let currentStrRow = '';
        let currentRowId = cells[0].row;
        let rowsStrs: string[] = [];

        for (const cell of cells) {
            if (cell.row !== currentRowId) {
                rowsStrs.push(currentStrRow);
                currentStrRow = '';
                currentRowId = cell.row;
            }

            const rObj = getRowIndexNumber(cell.row);
            const rowData = resultSet.rows[rObj.rIndex];
            if (rowData) {
                const val = rowData[cell.col];
                const valStr = val === null ? 'NULL' : String(val);
                if (currentStrRow) currentStrRow += '\t';
                currentStrRow += valStr;
            }
        }
        if (currentStrRow) rowsStrs.push(currentStrRow);

        navigator.clipboard.writeText(rowsStrs.join('\n'));
    }
    else if (selectedRowIndex.value !== null) {
        // 2. Multi-row or single row/cell selection
        const indices = Array.isArray(selectedRowIndex.value)
            ? selectedRowIndex.value.slice().sort((a, b) => {
                const aObj = getRowIndexNumber(a);
                const bObj = getRowIndexNumber(b);
                if (aObj.rsIndex !== bObj.rsIndex) return aObj.rsIndex - bObj.rsIndex;
                if (aObj.rIndex !== bObj.rIndex) return aObj.rIndex - bObj.rIndex;
                return 0;
            })
            : [selectedRowIndex.value];

        let rowsStrs: string[] = [];

        for (const selVal of indices) {
            const rObj = getRowIndexNumber(selVal);
            const resultSet = props.activeTab.resultSets[rObj.rsIndex];
            if (!resultSet || !resultSet.rows) continue;

            const rowData = resultSet.rows[rObj.rIndex];
            if (!rowData) continue;

            if (selectedColumn.value) {
                // Single cell copy
                const val = rowData[selectedColumn.value];
                rowsStrs.push(val === null ? 'NULL' : String(val));
            } else {
                // Full row copy
                const columns = resultSet.columns || [];
                const valueLine = columns.map((col: string) => {
                    const val = rowData[col];
                    return val === null ? 'NULL' : String(val);
                }).join('\t');
                rowsStrs.push(valueLine);
            }
        }

        if (rowsStrs.length > 0) {
            navigator.clipboard.writeText(rowsStrs.join('\n'));
        }
    }
};

useEventListener(document, 'keydown', (e: KeyboardEvent) => {
    const target = e.target as HTMLElement;
    if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable) return;

    if (!rootRef.value || getComputedStyle(rootRef.value).display === 'none') return;

    if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'c') {
        // Only skip to native copy if they natively selected something AND we don't have grid cells selected
        const hasSelection = window.getSelection()?.toString();

        if (draggedCellSelection.value.length > 0 || selectedRowIndex.value !== null) {
            e.preventDefault();
            e.stopPropagation();
            copyCurrentSelection();
        } else if (hasSelection) {
            return; // Let native copy happen
        }
    }
});

const containsTarget = (target: EventTarget | null) => {
    const node = target as Node | null;
    return !!(node && rootRef.value?.contains(node));
};

const getSecondaryFilteredRows = (resultSet: any, filters: any) => {
    if (!resultSet || !resultSet.rows) return [];
    let rows = resultSet.rows;
    if (!filters) return rows;

    const activeFilters = Object.entries(filters).filter(([_, val]) => val !== '' && val !== null && val !== undefined);
    if (activeFilters.length > 0) {
        rows = rows.filter((row: any) => {
            return activeFilters.every(([col, val]) => {
                if (row[col] === null || row[col] === undefined) return false;
                return String(row[col]).toLowerCase().includes(String(val).toLowerCase());
            });
        });
    }
    return rows;
};

defineExpose({
    containsTarget,
    copyCurrentSelection,
    isCellSelected
});
</script>

<template>
    <div v-if="!activeTab.isERView && !activeTab.isDesignView" ref="rootRef"
        class="flex-1 overflow-hidden bg-muted/10 flex flex-col">

        <DbQueryResultsTabs :active-tab="activeTab" @change-tab="activeTab.resultViewTab = $event" />

        <!-- Data Tab Content -->
        <div v-show="activeTab.resultViewTab === 'data' || (!activeTab.queryExecuted && !activeTab.error)"
            class="flex-1 overflow-auto p-4 flex flex-col">
            <!-- Error State -->
            <div v-if="activeTab.error"
                class="bg-destructive/10 border border-destructive/20 text-destructive p-4 rounded-lg shadow-sm flex items-start gap-3 animate-in fade-in slide-in-from-top-2 shrink-0 mb-4">
                <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-alert-triangle mt-0.5">
                    <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                    <path d="M12 9v4" />
                    <path d="M12 17h.01" />
                </svg>
                <div class="flex-1">
                    <div class="text-sm font-medium break-all font-mono">{{ activeTab.error }}</div>
                    <div class="mt-3">
                        <button @click="openAiCopilot('fix_error')"
                            class="inline-flex items-center gap-1 rounded-md border border-destructive/30 bg-background/60 px-2.5 py-1 text-xs font-medium hover:bg-background">
                            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round">
                                <path
                                    d="M12 3l1.912 5.813L20 10.5l-6.088 1.688L12 18l-1.912-5.813L4 10.5l6.088-1.687z" />
                            </svg>
                            Fix Error with AI
                        </button>
                    </div>
                </div>
            </div>

            <!-- Results List (Multiple Sets) -->
            <div v-if="activeTab.resultSets && activeTab.resultSets.length > 0" class="flex flex-col gap-4 min-h-0">

                <!-- Primary Result Set (Virtual List) -->
                <div v-if="activeTab.resultSets[0]"
                    class="border border-border rounded-lg shadow-sm bg-card flex flex-col overflow-hidden shrink-0"
                    :class="collapsedResultSets[0] ? '' : 'min-h-[220px]'"
                    :style="collapsedResultSets[0] ? { flex: '0 0 auto', height: 'auto', minHeight: '0' } : getResultSetCardStyle(activeTab.resultSets[0], 0)">

                    <!-- Header -->
                    <div class="bg-muted px-4 py-2 border-b border-border flex justify-between items-center select-none cursor-pointer hover:bg-muted/80"
                        @click="toggleResultSetCollapse(0)">
                        <div class="flex items-center gap-2">
                            <span class="font-semibold text-sm text-foreground">Result 1</span>
                            <span
                                class="text-xs text-muted-foreground bg-background/50 px-2 py-0.5 rounded-full border border-border/50">
                                {{ activeTab.resultSets[0].rows ? activeTab.resultSets[0].rows.length : 0 }} rows
                            </span>
                        </div>
                        <button class="text-muted-foreground hover:text-foreground transition-colors p-1"
                            @click.stop="toggleResultSetCollapse(0)">
                            <svg v-if="collapsedResultSets[0]" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-chevron-down">
                                <path d="m6 9 6 6 6-6" />
                            </svg>
                            <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-chevron-up">
                                <path d="m18 15-6-6-6 6" />
                            </svg>
                        </button>
                    </div>

                    <template v-if="!collapsedResultSets[0]">
                        <DbQueryResultGrid :resultSet="activeTab.resultSets[0]" :resultSetIndex="0"
                            :activeTab="activeTab" :isReadOnly="isReadOnly" v-model:selectedRowIndex="selectedRowIndex"
                            v-model:selectedColumn="selectedColumn" v-model:selectedRowData="selectedRowData"
                            :lastClickedRow="lastClickedRow" :isRowSelected="isRowSelected"
                            :isCellSelected="isCellSelected" @rowSelectorClick="handleRowSelectorClick"
                            @cellClickCustom="handleCellClickCustom" @cellDragStart="handleCellDragStart"
                            @cellDragEnter="handleCellDragEnter" @cellDragEnd="handleCellDragEnd"
                            @openMockDataModal="openMockDataModal" @openInsertRowModal="openInsertRowModal"
                            @startColumnResize="startColumnResize" @handleCellClick="handleCellClick"
                            @handleRowContextMenu="handleRowContextMenu" @saveCellEdit="saveCellEdit"
                            @openImagePreview="openImagePreview" @toggleSort="toggleSort" />
                        <div class="h-2 shrink-0 border-t border-border/60 bg-muted/20 hover:bg-primary/20 cursor-row-resize transition-colors"
                            title="Drag to resize this query result"
                            @mousedown="startResultSetResize($event, 0, activeTab.resultSets[0])"></div>
                    </template>
                </div>

                <!-- Subsequent Result Sets (Standard Tables) -->
                <div v-for="(resultSet, rsIndex) in activeTab.resultSets.slice(1)" :key="Number(rsIndex) + 1"
                    class="border border-border rounded-lg shadow-sm bg-card flex flex-col overflow-hidden shrink-0"
                    :class="collapsedResultSets[Number(rsIndex) + 1] ? '' : 'min-h-[220px]'"
                    :style="collapsedResultSets[Number(rsIndex) + 1] ? { flex: '0 0 auto', height: 'auto', minHeight: '0' } : getResultSetCardStyle(resultSet, Number(rsIndex) + 1)">

                    <!-- Header -->
                    <div class="bg-muted px-4 py-2 border-b border-border flex justify-between items-center select-none cursor-pointer hover:bg-muted/80"
                        @click="toggleResultSetCollapse(Number(rsIndex) + 1)">
                        <div class="flex items-center gap-2">
                            <span class="font-semibold text-sm text-foreground">Result {{ Number(rsIndex) + 2 }}</span>
                            <span
                                class="text-xs text-muted-foreground bg-background/50 px-2 py-0.5 rounded-full border border-border/50">
                                {{ resultSet.rows ? resultSet.rows.length : 0 }} rows
                            </span>
                        </div>
                        <button class="text-muted-foreground hover:text-foreground transition-colors p-1"
                            @click.stop="toggleResultSetCollapse(Number(rsIndex) + 1)">
                            <svg v-if="collapsedResultSets[Number(rsIndex) + 1]" xmlns="http://www.w3.org/2000/svg"
                                width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                class="lucide lucide-chevron-down">
                                <path d="m6 9 6 6 6-6" />
                            </svg>
                            <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-chevron-up">
                                <path d="m18 15-6-6-6 6" />
                            </svg>
                        </button>
                    </div>

                    <template v-if="!collapsedResultSets[Number(rsIndex) + 1]">
                        <DbQueryResultGrid :resultSet="resultSet" :resultSetIndex="Number(rsIndex) + 1"
                            :activeTab="activeTab" :isReadOnly="isReadOnly" v-model:selectedRowIndex="selectedRowIndex"
                            v-model:selectedColumn="selectedColumn" v-model:selectedRowData="selectedRowData"
                            :lastClickedRow="lastClickedRow" :isRowSelected="isRowSelected"
                            :isCellSelected="isCellSelected" @rowSelectorClick="handleRowSelectorClick"
                            @cellClickCustom="handleCellClickCustom" @cellDragStart="handleCellDragStart"
                            @cellDragEnter="handleCellDragEnter" @cellDragEnd="handleCellDragEnd"
                            @openMockDataModal="openMockDataModal" @openInsertRowModal="openInsertRowModal"
                            @startColumnResize="startColumnResize" @handleCellClick="handleCellClick"
                            @handleRowContextMenu="handleRowContextMenu" @saveCellEdit="saveCellEdit"
                            @openImagePreview="openImagePreview" @toggleSort="toggleSort" />
                        <div class="h-2 shrink-0 border-t border-border/60 bg-muted/20 hover:bg-primary/20 cursor-row-resize transition-colors"
                            title="Drag to resize this query result"
                            @mousedown="startResultSetResize($event, Number(rsIndex) + 1, resultSet)"></div>
                    </template>
                </div>
            </div>

            <!-- Empty State -->
            <div v-else-if="!activeTab.error && activeTab.queryExecuted && (!activeTab.resultSets || activeTab.resultSets.length === 0)"
                class="flex flex-col items-center justify-center h-full text-muted-foreground animate-in fade-in zoom-in-95 duration-300">
                <div class="h-12 w-12 rounded-full bg-muted flex items-center justify-center mb-4">
                    <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-search-x">
                        <path d="m13.5 8.5-5 5" />
                        <path d="m8.5 8.5 5 5" />
                        <circle cx="11" cy="11" r="8" />
                        <path d="m21 21-4.3-4.3" />
                    </svg>
                </div>
                <h3 class="text-lg font-semibold text-foreground">No Results Found</h3>
                <p class="text-sm">The query executed successfully but returned no data.</p>
            </div>

            <!-- Initial State -->
            <div v-else-if="!activeTab.error && !activeTab.queryExecuted && (!activeTab.resultSets || activeTab.resultSets.length === 0)"
                class="flex flex-col items-center justify-center h-full text-muted-foreground">
                <div class="h-16 w-16 rounded-2xl bg-muted/50 flex items-center justify-center mb-4 shadow-sm">
                    <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-terminal text-primary">
                        <polyline points="4 17 10 11 4 5" />
                        <line x1="12" x2="20" y1="19" y2="19" />
                    </svg>
                </div>
                <h3 class="text-lg font-semibold text-foreground">Ready to Query</h3>
                <p class="text-sm max-w-sm text-center mt-2">Select a table from the sidebar or type a
                    custom
                    SQL query to get started.</p>

                <div class="flex gap-2 mt-6">
                    <span class="text-xs bg-muted px-2 py-1 rounded border border-border">Ctrl + Enter to
                        Run</span>
                </div>
            </div>
        </div> <!-- End of scrollable results area -->

        <!-- Global Query Stats Footer -->
        <div v-if="activeTab.resultSets && activeTab.resultSets.length > 0 && !activeTab.error && activeTab.resultViewTab === 'data'"
            class="shrink-0 bg-muted/40 border-t border-border px-4 py-2 flex justify-between text-[11px] text-muted-foreground z-10">
            <div class="flex items-center gap-4">
                <span class="flex items-center gap-1.5" title="Total rows affected/returned across all SQL statements">
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-table">
                        <rect width="18" height="18" x="3" y="3" rx="2" ry="2" />
                        <line x1="3" x2="21" y1="9" y2="9" />
                        <line x1="3" x2="21" y1="15" y2="15" />
                        <line x1="9" x2="9" y1="9" y2="21" />
                        <line x1="15" x2="15" y1="9" y2="21" />
                    </svg>
                    Total: {{activeTab.totalRowCount !== undefined ? (activeTab.totalRowCount +
                        (activeTab.isPartialStats ? '+' : '')) : (activeTab.resultSets.reduce((sum: number, rs: any) => sum
                            + (rs.rows?.length || 0), 0))}} rows
                </span>
                <span v-if="activeTab.executionTime !== undefined" class="flex items-center gap-1.5">
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-cpu">
                        <rect width="16" height="16" x="4" y="4" rx="2" ry="2" />
                        <rect width="6" height="6" x="9" y="9" rx="1" ry="1" />
                        <path d="M15 2v2" />
                        <path d="M15 20v2" />
                        <path d="M2 15h2" />
                        <path d="M2 9h2" />
                        <path d="M20 15h2" />
                        <path d="M20 9h2" />
                        <path d="M9 2v2" />
                        <path d="M9 20v2" />
                    </svg>
                    Execution: {{ activeTab.executionTime }}ms
                </span>
                <span v-if="activeTab.fetchTime !== undefined" class="flex items-center gap-1.5">
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-network">
                        <rect x="16" y="16" width="6" height="6" rx="1" />
                        <rect x="2" y="16" width="6" height="6" rx="1" />
                        <rect x="9" y="2" width="6" height="6" rx="1" />
                        <path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3" />
                        <path d="M12 12V8" />
                    </svg>
                    Fetch: {{ activeTab.fetchTime }}ms
                </span>
            </div>
            <div class="flex items-center">
                <span class="flex items-center gap-1.5 opacity-80" v-if="activeTab.completionTime">
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-check-circle">
                        <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14" />
                        <path d="m9 11 3 3L22 4" />
                    </svg>
                    {{ activeTab.completionTime }}
                </span>
            </div>
        </div>

        <!-- Messages Tab Content -->
        <div v-if="activeTab.resultViewTab === 'messages' && (activeTab.queryExecuted || activeTab.error)"
            class="flex-1 overflow-auto p-4">
            <div class="border border-border rounded-lg shadow-sm bg-card overflow-hidden">
                <div
                    class="bg-muted px-4 py-2 text-xs font-semibold text-muted-foreground uppercase border-b border-border">
                    Messages
                </div>
                <div class="p-4 font-mono text-sm text-foreground space-y-1">
                    <!-- Error message -->
                    <div v-if="activeTab.error" class="text-destructive">
                        Msg: {{ activeTab.error }}
                    </div>

                    <!-- Result set messages -->
                    <template v-if="!activeTab.error && activeTab.resultSets">
                        <div v-for="(rs, idx) in activeTab.resultSets" :key="idx" class="text-foreground">
                            <span v-if="rs.columns && rs.columns.length > 0">
                                ({{ rs.rows ? rs.rows.length : 0 }} rows affected)
                            </span>
                            <span v-else>
                                Commands completed successfully.
                            </span>
                        </div>
                    </template>

                    <!-- No results message -->
                    <div v-if="!activeTab.error && (!activeTab.resultSets || activeTab.resultSets.length === 0) && activeTab.queryExecuted"
                        class="text-foreground">
                        Commands completed successfully.
                    </div>

                    <!-- Completion time -->
                    <div v-if="activeTab.completionTime" class="text-muted-foreground mt-3 pt-2 border-t border-border">
                        Completion time: {{ activeTab.completionTime }}
                    </div>
                </div>
            </div>
        </div>
        <!-- Analysis Tab Content -->
        <div v-if="activeTab.resultViewTab === 'analysis' && (activeTab.explanation || activeTab.aiExplanation || activeTab.isAiExplaining)"
            class="flex-1 overflow-auto p-4">
            <div v-if="activeTab.explanation"
                class="border border-border rounded-lg shadow-sm bg-card overflow-hidden h-full flex flex-col">
                <div
                    class="bg-muted px-4 py-2 text-xs font-semibold text-muted-foreground uppercase border-b border-border flex justify-between items-center">
                    <span>Query Execution Plan</span>
                    <span v-if="activeTab.isExplaining" class="flex items-center gap-2 text-primary">
                        <svg class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none"
                            viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4">
                            </circle>
                            <path class="opacity-75" fill="currentColor" d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z">
                            </path>
                        </svg>
                        Analyzing...
                    </span>
                </div>
                <div class="p-4 font-mono text-xs text-foreground overflow-auto whitespace-pre flex-1">
                    {{ activeTab.explanation }}
                </div>
            </div>
            <div v-if="activeTab.aiExplanation || activeTab.isAiExplaining"
                class="border border-border rounded-lg shadow-sm bg-card overflow-hidden h-full flex flex-col"
                :class="activeTab.explanation ? 'mt-4' : ''">
                <div
                    class="bg-muted px-4 py-2 text-xs font-semibold text-muted-foreground uppercase border-b border-border flex justify-between items-center">
                    <span>AI Query Explanation</span>
                    <span v-if="activeTab.isAiExplaining" class="flex items-center gap-2 text-primary">
                        <svg class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none"
                            viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4">
                            </circle>
                            <path class="opacity-75" fill="currentColor" d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z">
                            </path>
                        </svg>
                        Explaining...
                    </span>
                </div>
                <div class="p-4 text-sm text-foreground overflow-auto whitespace-pre-wrap flex-1">
                    {{ activeTab.aiExplanation || 'Analyzing with AI...' }}
                </div>
            </div>
        </div>
    </div>

</template>

<script lang="ts" setup>
import { computed, toRef, ref } from 'vue';
import { useVirtualList } from '@vueuse/core';
import type { QueryTab } from '../../types/dashboard';
import DatePicker from '../DatePicker.vue';

interface Props {
    resultSet: any;
    resultSetIndex: number;
    activeTab: QueryTab;
    isReadOnly?: boolean;
    selectedRowIndex?: Array<number | string> | number | string | null;
    selectedColumn?: string | null;
    lastClickedRow?: number | string | null;
    isRowSelected?: (index: number | string) => boolean;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    (e: 'update:selectedRowIndex', value: Array<number | string> | number | string | null): void;
    (e: 'update:selectedColumn', value: string | null): void;
    (e: 'update:selectedRowData', value: any): void;
    (e: 'rowSelectorClick', ev: MouseEvent, itemIndex: number | string): void;
    (e: 'cellClickCustom', itemIndex: number | string, col: string): void;
    (e: 'openMockDataModal'): void;
    (e: 'openInsertRowModal'): void;
    (e: 'startColumnResize', ev: MouseEvent, col: string): void;
    (e: 'handleCellClick', item: any, col: string, rsIndex: number): void;
    (e: 'handleRowContextMenu', ev: MouseEvent, row: any, col: string, rowIndex?: number | string): void;
    (e: 'saveCellEdit', item: any, col: string): void;
    (e: 'openImagePreview', url: string): void;
    (e: 'toggleSort', col: string): void;
}>();

// Removed unused computed properties 

// Reuse logic from useQueryResultsView but localized to this grid
const getSecondaryFilteredRows = (resultSet: any, filters: any, sortColumn?: string, sortDirection?: 'asc' | 'desc' | null) => {
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

    if (sortColumn && sortDirection) {
        rows = [...rows].sort((a: any, b: any) => {
            const valA = a[sortColumn];
            const valB = b[sortColumn];

            if (valA === valB) return 0;
            if (valA === null) return 1;
            if (valB === null) return -1;

            if (valA < valB) return sortDirection === 'asc' ? -1 : 1;
            if (valA > valB) return sortDirection === 'asc' ? 1 : -1;
            return 0;
        });
    }

    return rows;
};

const localFilteredResults = computed(() => {
    return getSecondaryFilteredRows(props.resultSet, props.activeTab.filters, props.activeTab.sortColumn, props.activeTab.sortDirection);
});

const { list: virtualList, containerProps, wrapperProps } = useVirtualList(localFilteredResults, {
    itemHeight: 37,
    overscan: 10,
});

const padTop = computed(() => {
    if (virtualList.value.length === 0) return 0;
    const start = virtualList.value[0].index;
    return start * 37;
});

const padBottom = computed(() => {
    if (virtualList.value.length === 0) return 0;
    const end = virtualList.value[virtualList.value.length - 1].index;
    const total = localFilteredResults.value.length;
    return (total - end - 1) * 37;
});

const getFormattedRowIndex = (rIndex: number) => {
    if (props.resultSetIndex === 0) {
        return rIndex;
    }
    return `sub-${props.resultSetIndex - 1}-${rIndex}`;
};

const isRowSelected = (rIndex: number) => {
    if (props.isRowSelected) {
        return props.isRowSelected(getFormattedRowIndex(rIndex));
    }
    return false;
};

const selectRow = (rIndex: number, col: string) => {
    const formatted = getFormattedRowIndex(rIndex);
    // Backward compatibility for standard single row selecting if parents expect it
    if (props.selectedRowIndex === formatted && props.selectedColumn === col) {
        emit('update:selectedRowIndex', null);
        emit('update:selectedRowData', null);
    } else {
        emit('update:selectedRowIndex', formatted);
        emit('update:selectedColumn', col);
        emit('update:selectedRowData', localFilteredResults.value[rIndex]);
    }
};

const isImageValue = (value: any, col: string): boolean => {
    if (!value || typeof value !== 'string') return false;
    // Check if the column is specifically known to hold images
    const lowerCol = col.toLowerCase();
    if (lowerCol.includes('photo') || lowerCol.includes('image') || lowerCol.includes('pic') || lowerCol.includes('logo')) {
        // If it looks like a URL or a base64 string
        if (value.startsWith('http') || value.startsWith('data:image') || value.startsWith('/api/') || value.startsWith('/assets/')) {
            return true;
        }
        // Special case for our database where photos are just GUIDs or filenames
        if (value.length > 5) {
            return true;
        }
    }
    return false;
};

const getEditorType = (col: string): string => {
    // Attempt basic typing inference over column name since we don't have types here
    const lower = col.toLowerCase();
    if (lower.includes('date') || lower.includes('time')) {
        return 'datetime-local';
    }
    if (lower.includes('id') || lower.includes('count') || lower.includes('num')) {
        return 'number';
    }
    return 'text';
};

const isCellEditing = (index: number, col: string) => {
    return props.activeTab.editingCell &&
        props.activeTab.editingCell.rowId === index &&
        props.activeTab.editingCell.col === col &&
        props.activeTab.editingCell.resultSetIndex === props.resultSetIndex;
};

</script>

<template>
    <div class="flex-1 overflow-auto bg-card" v-bind="containerProps">
        <table v-if="resultSet.columns && resultSet.columns.length > 0" class="w-full text-sm text-left relative">
            <thead class="text-xs text-muted-foreground uppercase bg-muted sticky top-0 z-20 font-medium">
                <tr>
                    <th scope="col" class="w-8 min-w-8 sticky left-0 z-30 bg-muted border-b border-border"></th>
                    <th v-for="(col, index) in resultSet.columns" :key="index + '-' + col" scope="col"
                        class="px-4 py-3 whitespace-nowrap border-b border-border min-w-[50px] cursor-pointer hover:bg-muted/80 select-none relative group/th"
                        :style="{ width: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', minWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px' }"
                        @click="emit('toggleSort', col)">
                        <div class="flex flex-col gap-2">
                            <div class="flex items-center justify-between gap-2">
                                <span>{{ col }}</span>
                                <div class="flex flex-col">
                                    <svg v-if="activeTab.sortColumn === col && activeTab.sortDirection === 'asc'"
                                        xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round" class="lucide lucide-chevron-up">
                                        <path d="m18 15-6-6-6 6" />
                                    </svg>
                                    <svg v-if="activeTab.sortColumn === col && activeTab.sortDirection === 'desc'"
                                        xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round" class="lucide lucide-chevron-down">
                                        <path d="m6 9 6 6 6-6" />
                                    </svg>
                                    <svg v-if="activeTab.sortColumn !== col" xmlns="http://www.w3.org/2000/svg"
                                        width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                        stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                                        class="lucide lucide-chevrons-up-down text-muted-foreground/30">
                                        <path d="m7 15 5 5 5-5" />
                                        <path d="m7 9 5-5 5 5" />
                                    </svg>
                                </div>
                            </div>
                            <input type="text" v-model="activeTab.filters[col]" placeholder="Filter..."
                                class="w-full h-6 px-2 text-[10px] rounded border border-input bg-background focus:outline-none focus:ring-1 focus:ring-ring font-normal normal-case text-foreground cursor-text placeholder:text-muted-foreground/70"
                                @click.stop />
                        </div>
                        <!-- Column Resizer -->
                        <div class="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-primary/50 opacity-0 group-hover/th:opacity-100 transition-opacity z-20"
                            @mousedown.stop="emit('startColumnResize', $event, col)"></div>
                    </th>
                </tr>
            </thead>
            <tbody class="divide-y divide-border">
                <tr :style="{ height: `${padTop}px` }"></tr>
                <tr v-for="item in virtualList" :key="item.index" class="transition-colors h-[37px] cursor-pointer group"
                    :class="isRowSelected(item.index) ? 'bg-primary/5' : 'bg-card hover:bg-muted/50'"
                    @click="emit('update:selectedRowData', item.data); emit('cellClickCustom', getFormattedRowIndex(item.index), '')">
                    <td class="w-8 min-w-8 sticky left-0 z-20 border-r border-border cursor-pointer select-none bg-muted hover:bg-muted/80"
                        :class="isRowSelected(item.index) ? 'border-l-2 border-l-primary' : ''"
                        @click.stop="emit('update:selectedRowData', item.data); emit('rowSelectorClick', $event, getFormattedRowIndex(item.index))"
                        @mousedown="e => { if (e.shiftKey) e.preventDefault(); }">
                        <div class="flex items-center justify-center w-full h-full">
                            <span v-if="isRowSelected(item.index)" class="w-1.5 h-1.5 rounded-full bg-primary inline-block"></span>
                            <span v-else class="w-1.5 h-1.5 rounded-full bg-muted-foreground/30 inline-block opacity-0 group-hover:opacity-100 transition-opacity"></span>
                        </div>
                    </td>
                    <td v-for="(col, index) in resultSet.columns" :key="index + '-' + col"
                        class="px-4 py-2 whitespace-nowrap text-foreground font-mono text-xs border-r border-transparent hover:border-border cursor-pointer relative overflow-hidden"
                        :style="{ width: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', minWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', maxWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px' }"
                        :class="{
                            'bg-accent/50': isCellEditing(item.index, col),
                            'ring-1 ring-inset ring-primary z-10': props.lastClickedRow === getFormattedRowIndex(item.index) && selectedColumn === col && (!activeTab.editingCell || activeTab.editingCell.resultSetIndex !== resultSetIndex)
                        }" @click.stop="emit('update:selectedRowData', item.data); emit('cellClickCustom', getFormattedRowIndex(item.index), col)" @dblclick="emit('handleCellClick', item, col, resultSetIndex)"
                        @contextmenu.prevent="emit('update:selectedRowData', item.data); emit('handleRowContextMenu', $event, item.data, col, getFormattedRowIndex(item.index))">

                        <div v-if="isCellEditing(item.index, col) && activeTab.editingCell"
                            class="absolute inset-0 p-0.5">
                            <DatePicker v-if="getEditorType(col) === 'date' || getEditorType(col) === 'datetime-local'"
                                :id="`edit-input-${resultSetIndex}-${item.index}-${col}`"
                                v-model="activeTab.editingCell!.value" :type="getEditorType(col) as any"
                                @confirm="emit('saveCellEdit', item, col)" @cancel="activeTab.editingCell = null" />
                            <input v-else :id="`edit-input-${resultSetIndex}-${item.index}-${col}`"
                                v-model="activeTab.editingCell!.value" :type="getEditorType(col)"
                                :step="getEditorType(col) === 'number' ? 'any' : undefined"
                                class="w-full h-full px-2 bg-background text-foreground border border-primary focus:outline-none focus:ring-1 focus:ring-primary rounded-sm shadow-sm"
                                @blur="emit('saveCellEdit', item, col)" @keydown.enter="emit('saveCellEdit', item, col)"
                                @keydown.esc="activeTab.editingCell = null" />
                        </div>
                        <div v-else class="flex items-center gap-2 overflow-hidden w-full h-full">
                            <div v-if="isImageValue(item.data[col], col)"
                                class="shrink-0 h-7 w-7 rounded border border-border overflow-hidden bg-muted flex items-center justify-center group/img relative cursor-pointer hover:bg-muted/80"
                                @click.stop="emit('openImagePreview', item.data[col])" title="Click to view full image">
                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round"
                                    class="lucide lucide-image text-muted-foreground group-hover/img:text-foreground transition-colors">
                                    <rect width="18" height="18" x="3" y="3" rx="2" ry="2" />
                                    <circle cx="9" cy="9" r="2" />
                                    <path d="m21 15-3.086-3.086a2 2 0 0 0-2.828 0L6 21" />
                                </svg>
                            </div>
                            <span class="truncate block flex-1" :title="String(item.data[col])">
                                {{ item.data[col] === null ? 'NULL' : item.data[col] }}
                            </span>
                        </div>
                    </td>
                </tr>
                <tr :style="{ height: `${padBottom}px` }"></tr>
            </tbody>
        </table>
    </div>

    <div
        class="bg-muted/30 px-4 py-2 border-t border-border text-xs text-muted-foreground flex justify-between items-center shrink-0 h-10">
        <span>Showing {{ localFilteredResults.length }} rows {{ localFilteredResults.length !== (resultSet.rows ?
            resultSet.rows.length : 0) ? `(out of ${resultSet.rows ? resultSet.rows.length : 0})` : '' }}</span>
        <div class="flex items-center gap-3" v-if="resultSetIndex === 0">
            <button v-if="activeTab.tableName && !isReadOnly" @click="emit('openMockDataModal')"
                class="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium bg-blue-500/10 text-blue-600 dark:text-blue-400 hover:bg-blue-500/20 border border-blue-500/20 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-flask-conical">
                    <path d="M10 2v7.31" />
                    <path d="M14 9.3V2" />
                    <path d="M8.5 2h7" />
                    <path
                        d="M14 9.3a5 5 0 0 1 .6 2.4V19a3 3 0 0 1-3 3h-1.2a3 3 0 0 1-3-3v-7.3a5 5 0 0 1 .6-2.4L10 6h4Z" />
                </svg>
                Mock Data
            </button>
            <button v-if="activeTab.tableName && !isReadOnly && activeTab.primaryKeys.length > 0"
                @click="emit('openInsertRowModal')"
                class="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium bg-primary/10 text-primary hover:bg-primary/20 border border-primary/20 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-plus">
                    <path d="M5 12h14" />
                    <path d="M12 5v14" />
                </svg>
                Row
            </button>
        </div>
    </div>
</template>

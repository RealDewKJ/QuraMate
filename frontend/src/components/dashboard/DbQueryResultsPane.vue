<script lang="ts" setup>
import { ref } from 'vue';

import DatePicker from '../DatePicker.vue';
import DbQueryResultsTabs from './DbQueryResultsTabs.vue';

interface Props {
    activeTab: any;
    isReadOnly?: boolean;
    containerProps: any;
    wrapperProps?: any;
    virtualList: any[];
    padTop: number;
    padBottom: number;
    filteredResults: any[];
    openAiCopilot: (mode?: any) => void;
    toggleSort: (col: string) => void;
    startColumnResize: (e: MouseEvent, col: string) => void;
    handleCellClick: (item: any, col: string) => void;
    handleRowContextMenu: (e: MouseEvent, row: any, col: string) => void;
    getEditorType: (col: string) => string;
    saveCellEdit: (item: any, col: string) => void | Promise<void>;
    isImageValue: (value: any, col: string) => boolean;
    openImagePreview: (url: any) => void;
    openMockDataModal: () => void;
    openInsertRowModal: () => void;
    startResultSetResize: (e: MouseEvent, resultSetIndex: number, resultSet: any) => void;
    getResultSetCardStyle: (resultSet: any, resultSetIndex: number) => Record<string, string> | string | undefined;
}

defineProps<Props>();

const selectedRowIndex = defineModel<number | string | null>('selectedRowIndex', { required: true });
const selectedColumn = defineModel<string | null>('selectedColumn', { required: true });

const rootRef = ref<HTMLElement | null>(null);

const containsTarget = (target: EventTarget | null) => {
    const node = target as Node | null;
    return !!(node && rootRef.value?.contains(node));
};

defineExpose({
    containsTarget,
});
</script>

<template>
                <div v-if="!activeTab.isERView && !activeTab.isDesignView" ref="rootRef"
                    class="flex-1 overflow-hidden bg-muted/10 flex flex-col">

                    <DbQueryResultsTabs
                        :active-tab="activeTab"
                        @change-tab="activeTab.resultViewTab = $event"
                    />

                    <!-- Data Tab Content -->
                    <div v-if="activeTab.resultViewTab === 'data' || (!activeTab.queryExecuted && !activeTab.error)"
                        class="flex-1 overflow-auto p-4">
                        <!-- Error State -->
                        <div v-if="activeTab.error"
                            class="bg-destructive/10 border border-destructive/20 text-destructive p-4 rounded-lg shadow-sm flex items-start gap-3 animate-in fade-in slide-in-from-top-2">
                            <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-alert-triangle mt-0.5">
                                <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                                <path d="M12 9v4" />
                                <path d="M12 17h.01" />
                            </svg>
                            <div class="flex-1">
                                <div class="text-sm font-medium break-all font-mono">{{ activeTab.error }}</div>
                                <div class="mt-3">
                                    <button @click="openAiCopilot('fix_error')"
                                        class="inline-flex items-center gap-1 rounded-md border border-destructive/30 bg-background/60 px-2.5 py-1 text-xs font-medium hover:bg-background">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                            <path d="M12 3l1.912 5.813L20 10.5l-6.088 1.688L12 18l-1.912-5.813L4 10.5l6.088-1.687z" />
                                        </svg>
                                        Fix Error with AI
                                    </button>
                                </div>
                            </div>
                        </div>

                        <!-- Results List (Multiple Sets) -->
                        <div v-else-if="activeTab.resultSets && activeTab.resultSets.length > 0"
                            class="flex flex-col gap-4 min-h-0">

                            <!-- Primary Result Set (Virtual List) -->
                            <div v-if="activeTab.resultSets[0]"
                                class="border border-border rounded-lg shadow-sm bg-card flex flex-col min-h-[220px] overflow-hidden"
                                :style="getResultSetCardStyle(activeTab.resultSets[0], 0)">

                                <!-- Virtual Table Container -->
                                <div class="flex-1 overflow-auto bg-card" v-bind="containerProps">
                                    <table
                                        v-if="activeTab.resultSets[0].columns && activeTab.resultSets[0].columns.length > 0"
                                        class="w-full text-sm text-left relative">
                                        <thead
                                            class="text-xs text-muted-foreground uppercase bg-muted sticky top-0 z-10 font-medium">
                                            <tr>
                                                <th v-for="(col, index) in activeTab.resultSets[0].columns"
                                                    :key="index + '-' + col" scope="col"
                                                    class="px-4 py-3 whitespace-nowrap border-b border-border min-w-[50px] cursor-pointer hover:bg-muted/80 select-none relative group/th"
                                                    :style="{ width: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', minWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px' }"
                                                    @click="toggleSort(col)">
                                                    <div class="flex flex-col gap-2">
                                                        <div class="flex items-center justify-between gap-2">
                                                            <span>{{ col }}</span>
                                                            <div class="flex flex-col">
                                                                <svg v-if="activeTab.sortColumn === col && activeTab.sortDirection === 'asc'"
                                                                    xmlns="http://www.w3.org/2000/svg" width="12"
                                                                    height="12" viewBox="0 0 24 24" fill="none"
                                                                    stroke="currentColor" stroke-width="2"
                                                                    stroke-linecap="round" stroke-linejoin="round"
                                                                    class="lucide lucide-chevron-up">
                                                                    <path d="m18 15-6-6-6 6" />
                                                                </svg>
                                                                <svg v-if="activeTab.sortColumn === col && activeTab.sortDirection === 'desc'"
                                                                    xmlns="http://www.w3.org/2000/svg" width="12"
                                                                    height="12" viewBox="0 0 24 24" fill="none"
                                                                    stroke="currentColor" stroke-width="2"
                                                                    stroke-linecap="round" stroke-linejoin="round"
                                                                    class="lucide lucide-chevron-down">
                                                                    <path d="m6 9 6 6 6-6" />
                                                                </svg>
                                                                <svg v-if="activeTab.sortColumn !== col"
                                                                    xmlns="http://www.w3.org/2000/svg" width="12"
                                                                    height="12" viewBox="0 0 24 24" fill="none"
                                                                    stroke="currentColor" stroke-width="2"
                                                                    stroke-linecap="round" stroke-linejoin="round"
                                                                    class="lucide lucide-chevrons-up-down text-muted-foreground/30">
                                                                    <path d="m7 15 5 5 5-5" />
                                                                    <path d="m7 9 5-5 5 5" />
                                                                </svg>
                                                            </div>
                                                        </div>
                                                        <input v-if="activeTab.primaryKeys.length > 0 || true"
                                                            type="text" v-model="activeTab.filters[col]"
                                                            placeholder="Filter..."
                                                            class="w-full h-6 px-2 text-[10px] rounded border border-input bg-background focus:outline-none focus:ring-1 focus:ring-ring font-normal normal-case text-foreground cursor-text placeholder:text-muted-foreground/70"
                                                            @click.stop />
                                                    </div>
                                                    <!-- Column Resizer -->
                                                    <div class="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-primary/50 opacity-0 group-hover/th:opacity-100 transition-opacity z-20"
                                                        @mousedown.stop="startColumnResize($event, col)"></div>
                                                </th>
                                            </tr>
                                        </thead>
                                        <tbody class="divide-y divide-border">
                                            <tr :style="{ height: `${padTop}px` }"></tr>
                                            <tr v-for="item in virtualList" :key="item.index"
                                                class="transition-colors h-[37px] cursor-pointer"
                                                :class="selectedRowIndex === item.index ? 'bg-primary/10 border-l-2 border-l-primary' : 'bg-card hover:bg-muted/50'"
                                                @click="selectedRowIndex = selectedRowIndex === item.index ? null : item.index">
                                                <td v-for="(col, index) in activeTab.resultSets[0].columns"
                                                    :key="index + '-' + col"
                                                    class="px-4 py-2 whitespace-nowrap text-foreground font-mono text-xs border-r border-transparent hover:border-border cursor-pointer relative overflow-hidden"
                                                    :style="{ width: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', minWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', maxWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px' }"
                                                    :class="{
                                                        'bg-accent/50': activeTab.editingCell && activeTab.editingCell.rowId === item.index && activeTab.editingCell.col === col,
                                                        'ring-1 ring-inset ring-primary z-10': selectedRowIndex === item.index && selectedColumn === col && !activeTab.editingCell
                                                    }"
                                                    @click.stop="selectedRowIndex = item.index; selectedColumn = col"
                                                    @dblclick="handleCellClick(item, col)"
                                                    @contextmenu.prevent="handleRowContextMenu($event, item.data, col)">

                                                    <div v-if="activeTab.editingCell && activeTab.editingCell.rowId === item.index && activeTab.editingCell.col === col"
                                                        class="absolute inset-0 p-0.5">
                                                        <DatePicker
                                                            v-if="getEditorType(col) === 'date' || getEditorType(col) === 'datetime-local'"
                                                            :id="`edit-input-${item.index}-${col}`"
                                                            v-model="activeTab.editingCell.value"
                                                            :type="getEditorType(col) as any"
                                                            @confirm="saveCellEdit(item, col)"
                                                            @cancel="activeTab.editingCell = null" />
                                                        <input v-else :id="`edit-input-${item.index}-${col}`"
                                                            v-model="activeTab.editingCell.value"
                                                            :type="getEditorType(col)"
                                                            :step="getEditorType(col) === 'number' ? 'any' : undefined"
                                                            class="w-full h-full px-2 bg-background text-foreground border border-primary focus:outline-none focus:ring-1 focus:ring-primary rounded-sm shadow-sm"
                                                            @blur="saveCellEdit(item, col)"
                                                            @keydown.enter="saveCellEdit(item, col)"
                                                            @keydown.esc="activeTab.editingCell = null" />
                                                    </div>
                                                    <div v-else
                                                        class="flex items-center gap-2 overflow-hidden w-full h-full">
                                                        <div v-if="isImageValue(item.data[col], col)"
                                                            class="shrink-0 h-7 w-7 rounded border border-border overflow-hidden bg-muted flex items-center justify-center group/img relative"
                                                            @click.stop="openImagePreview(item.data[col])"
                                                            title="Click to view full image">
                                                            <img :src="String(item.data[col])"
                                                                class="h-full w-full object-contain cursor-pointer" />
                                                        </div>
                                                        <span class="truncate block flex-1"
                                                            :title="String(item.data[col])">
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
                                    class="bg-muted/30 px-4 py-2 border-t border-border text-xs text-muted-foreground flex justify-between items-center">
                                    <span>{{ filteredResults.length }} rows returned ({{ activeTab.totalRowCount !==
                                        undefined ?
                                        (activeTab.totalRowCount + (activeTab.isPartialStats ? '+' : '')) :
                                        (activeTab.resultSets[0].rows ?
                                            activeTab.resultSets[0].rows.length : 0) }}
                                        total)</span>
                                    <div class="flex items-center gap-3">
                                        <button v-if="activeTab.tableName && !isReadOnly"
                                            @click="openMockDataModal"
                                            class="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium bg-blue-500/10 text-blue-600 dark:text-blue-400 hover:bg-blue-500/20 border border-blue-500/20 transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                                stroke-linecap="round" stroke-linejoin="round"
                                                class="lucide lucide-flask-conical">
                                                <path d="M10 2v7.31" />
                                                <path d="M14 9.3V2" />
                                                <path
                                                    d="M8.5 2h7" />
                                                <path
                                                    d="M14 9.3a5 5 0 0 1 .6 2.4V19a3 3 0 0 1-3 3h-1.2a3 3 0 0 1-3-3v-7.3a5 5 0 0 1 .6-2.4L10 6h4Z" />
                                            </svg>
                                            Mock Data
                                        </button>
                                        <button
                                            v-if="activeTab.tableName && !isReadOnly && activeTab.primaryKeys.length > 0"
                                            @click="openInsertRowModal"
                                            class="inline-flex items-center gap-1 px-2.5 py-1 rounded-md text-xs font-medium bg-primary/10 text-primary hover:bg-primary/20 border border-primary/20 transition-colors">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                                stroke-linecap="round" stroke-linejoin="round"
                                                class="lucide lucide-plus">
                                                <path d="M5 12h14" />
                                                <path d="M12 5v14" />
                                            </svg>
                                            Row
                                        </button>
                                        <span class="font-mono text-[10px] opacity-70">Double-click to edit</span>
                                    </div>
                                </div>
                                <div
                                    class="h-2 shrink-0 border-t border-border/60 bg-muted/20 hover:bg-primary/20 cursor-row-resize transition-colors"
                                    title="Drag to resize this query result"
                                    @mousedown="startResultSetResize($event, 0, activeTab.resultSets[0])"></div>
                            </div>

                            <!-- Subsequent Result Sets (Standard Tables) -->
                            <div v-for="(resultSet, rsIndex) in activeTab.resultSets.slice(1)" :key="Number(rsIndex) + 1"
                                class="border border-border rounded-lg shadow-sm bg-card flex flex-col min-h-[220px] overflow-hidden"
                                :style="getResultSetCardStyle(resultSet, Number(rsIndex) + 1)">

                                <!-- Standard Table for subsequent result sets -->
                                <div v-if="resultSet.columns && resultSet.columns.length > 0"
                                    class="flex-1 overflow-auto bg-card">
                                    <table class="w-full text-sm text-left relative">
                                        <thead
                                            class="text-xs text-muted-foreground uppercase bg-muted sticky top-0 z-10 font-medium">
                                            <tr>
                                                <th v-for="(col, index) in resultSet.columns" :key="index + '-' + col"
                                                    class="px-4 py-3 whitespace-nowrap border-b border-border min-w-[50px] select-none relative group/th"
                                                    :style="{ width: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', minWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px' }">
                                                    {{ col }}
                                                    <!-- Column Resizer -->
                                                    <div class="absolute right-0 top-0 bottom-0 w-1 cursor-col-resize hover:bg-primary/50 opacity-0 group-hover/th:opacity-100 transition-opacity z-20"
                                                        @mousedown.stop="startColumnResize($event, col)"></div>
                                                </th>
                                            </tr>
                                        </thead>
                                        <tbody class="divide-y divide-border">
                                            <tr v-for="(row, rIndex) in resultSet.rows" :key="rIndex"
                                                class="transition-colors cursor-pointer"
                                                :class="selectedRowIndex === `sub-${rsIndex}-${rIndex}` ? 'bg-primary/10 border-l-2 border-l-primary' : 'bg-card hover:bg-muted/50'"
                                                @click="selectedRowIndex = selectedRowIndex === `sub-${rsIndex}-${rIndex}` ? null : `sub-${rsIndex}-${rIndex}`">
                                                <td v-for="(col, index) in resultSet.columns" :key="index + '-' + col"
                                                    class="px-4 py-2 whitespace-nowrap text-foreground font-mono text-xs border-r border-transparent hover:border-border overflow-hidden cursor-pointer relative"
                                                    :style="{ width: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', minWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', maxWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px' }"
                                                    :class="{ 'ring-1 ring-inset ring-primary z-10': selectedRowIndex === `sub-${rsIndex}-${rIndex}` && selectedColumn === col }"
                                                    @click.stop="selectedRowIndex = `sub-${rsIndex}-${rIndex}`; selectedColumn = col"
                                                    @contextmenu.prevent="handleRowContextMenu($event, row, col)">
                                                    <div class="flex items-center gap-2 overflow-hidden w-full h-full">
                                                        <div v-if="isImageValue(row[col], col)"
                                                            class="shrink-0 h-7 w-7 rounded border border-border overflow-hidden bg-muted flex items-center justify-center group/img relative"
                                                            @click.stop="openImagePreview(row[col])"
                                                            title="Click to view full image">
                                                            <img :src="String(row[col])"
                                                                class="h-full w-full object-contain cursor-pointer" />
                                                        </div>
                                                        <span class="truncate block flex-1" :title="String(row[col])">
                                                            {{ row[col] === null ? 'NULL' : row[col] }}
                                                        </span>
                                                    </div>
                                                </td>
                                            </tr>
                                        </tbody>
                                    </table>
                                </div>

                                <div
                                    class="bg-muted/30 px-4 py-2 border-t border-border text-xs text-muted-foreground flex justify-between items-center">
                                    <span>{{ resultSet.rows ? resultSet.rows.length : 0 }} rows</span>
                                </div>
                                <div
                                    class="h-2 shrink-0 border-t border-border/60 bg-muted/20 hover:bg-primary/20 cursor-row-resize transition-colors"
                                    title="Drag to resize this query result"
                                    @mousedown="startResultSetResize($event, Number(rsIndex) + 1, resultSet)"></div>
                            </div>
                        </div>

                        <!-- Empty State -->
                        <div v-else-if="!activeTab.error && activeTab.queryExecuted"
                            class="flex flex-col items-center justify-center h-full text-muted-foreground animate-in fade-in zoom-in-95 duration-300">
                            <div class="h-12 w-12 rounded-full bg-muted flex items-center justify-center mb-4">
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-search-x">
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
                        <div v-else class="flex flex-col items-center justify-center h-full text-muted-foreground">
                            <div
                                class="h-16 w-16 rounded-2xl bg-muted/50 flex items-center justify-center mb-4 shadow-sm">
                                <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-terminal text-primary">
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
                                <div v-if="activeTab.completionTime"
                                    class="text-muted-foreground mt-3 pt-2 border-t border-border">
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
                                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                            stroke-width="4"></circle>
                                        <path class="opacity-75" fill="currentColor"
                                            d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z">
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
                                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                            stroke-width="4"></circle>
                                        <path class="opacity-75" fill="currentColor"
                                            d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z">
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


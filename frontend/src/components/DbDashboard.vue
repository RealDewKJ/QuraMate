<template>
    <div class="flex h-full bg-background text-foreground font-sans">
        <DbSidebar :connection-name="connectionName" :db-type="dbType" :activity-task-count="activityTaskCount"
            :table-search="tableSearch" :view-search="viewSearch" :stored-procedure-search="storedProcedureSearch"
            :function-search="functionSearch" :filtered-tables="filteredTables" :filtered-views="filteredViews"
            :filtered-stored-procedures="filteredStoredProcedures" :filtered-functions="filteredFunctions"
            :open-folders="openFolders" @open-db-context-menu="openDbContextMenu" @open-history="isHistoryOpen = true"
            @open-activity-monitor="isActivityMonitorOpen = true" @open-ai-copilot="openAiCopilot"
            @open-settings="isSettingsOpen = true" @toggle-folder="toggleFolder"
            @open-folder-context-menu="openFolderContextMenu" @select-table="selectTable"
            @open-table-context-menu="openContextMenu" @select-view="selectView"
            @open-view-context-menu="openViewContextMenu" @select-routine="selectRoutine"
            @open-routine-context-menu="openRoutineContextMenu" @disconnect="disconnect"
            @update:table-search="tableSearch = $event" @update:view-search="viewSearch = $event"
            @update:stored-procedure-search="storedProcedureSearch = $event"
            @update:function-search="functionSearch = $event" />

        <!-- Main Content -->
        <div class="flex-1 flex flex-col overflow-hidden bg-background">
            <!-- Tab Bar -->
            <div class="flex items-center border-b border-border bg-muted/20 px-1 pt-1 gap-1 overflow-x-auto">
                <TooltipProvider :delay-duration="300">
                    <TooltipRoot v-for="tab in tabs" :key="tab.id">
                        <TooltipTrigger as-child>
                            <div @click="activeTabId = tab.id"
                                class="group relative flex items-center justify-between gap-2 px-4 py-2 text-sm font-medium cursor-pointer rounded-t-lg transition-all select-none min-w-[140px] max-w-[240px] border-l border-r border-t border-transparent hover:bg-background/50"
                                :class="{ 'bg-background text-foreground border-border shadow-sm mb-[-1px]': activeTabId === tab.id, 'text-muted-foreground hover:text-foreground': activeTabId !== tab.id }">
                                <div class="flex items-center gap-2 truncate">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round" class="lucide lucide-terminal-square">
                                        <path d="m7 11 2-2-2-2" />
                                        <path d="M11 13h4" />
                                        <rect width="18" height="18" x="3" y="3" rx="2" ry="2" />
                                    </svg>
                                    <span class="truncate">{{ tab.name }}</span>
                                </div>
                                <button @click.stop="closeTab(tab.id)"
                                    class="rounded-sm p-0.5 hover:bg-muted text-muted-foreground/50 hover:text-foreground transition-all opacity-0 group-hover:opacity-100">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round" class="lucide lucide-x">
                                        <path d="M18 6 6 18" />
                                        <path d="m6 6 12 12" />
                                    </svg>
                                </button>
                                <!-- Active Indicator Line -->
                                <div v-if="activeTabId === tab.id"
                                    class="absolute top-0 left-0 right-0 h-0.5 bg-primary rounded-t-full"></div>
                            </div>
                        </TooltipTrigger>
                        <TooltipContent side="bottom" :side-offset="4"
                            class="z-50 overflow-hidden rounded-md border bg-popover px-3 py-1.5 text-sm text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95">
                            {{ tab.name }}
                        </TooltipContent>
                    </TooltipRoot>
                </TooltipProvider>

                <button @click="addTab"
                    class="flex items-center justify-center h-8 w-8 ml-1 rounded-md text-muted-foreground hover:text-foreground hover:bg-muted/50 transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-plus">
                        <path d="M5 12h14" />
                        <path d="M12 5v14" />
                    </svg>
                </button>
            </div>

            <DbActivityMonitor v-if="isActivityMonitorOpen" :activity-task-count="activityTaskCount"
                :monitor-refresh-rate="monitorRefreshRate" :latest-monitor-sample="latestMonitorSample"
                :connection-chart-points="connectionChartPoints" :read-qps-chart-points="readQpsChartPoints"
                :write-qps-chart-points="writeQpsChartPoints" :long-running-total="longRunningTotal"
                :activity-tasks-list="activityTasksList" :format-activity-time="formatActivityTime"
                @close="isActivityMonitorOpen = false" @kill-all="killAllActivityTasks" @focus-task="focusActivityTask"
                @kill-task="killActivityTask" @update:monitor-refresh-rate="monitorRefreshRate = $event" />

            <!-- Query Area -->
            <div v-else-if="activeTab" class="flex flex-col h-full overflow-hidden query-area-container">
                <DbQueryWorkspacePane ref="workspaceRef" :active-tab="activeTab" :tables="tables"
                    :get-columns="fetchTableColumns" :editor-settings="editorSettings" :is-read-only="isReadOnly"
                    @beautify-query="beautifyQuery" @explain-with-ai="explainWithAI" @save-routine="handleSaveRoutine"
                    @run-query="runQuery" @stop-query="stopQuery" @start-resizing="startResizing" />
                <DbQueryResultsPane ref="resultsPaneRef" v-if="activeTab" :activeTab="activeTab"
                    :isReadOnly="isReadOnly" v-model:selectedRowIndex="selectedRowIndex"
                    v-model:selectedColumn="selectedColumn" v-model:selectedRowData="selectedRowData"
                    :openAiCopilot="openAiCopilot" :getEditorType="getEditorType" :saveCellEdit="saveCellEdit"
                    :toggleSort="toggleSort" :startColumnResize="startColumnResize" :handleCellClick="handleCellClick"
                    :handleRowContextMenu="handleRowContextMenu" :isImageValue="isImageValue"
                    :openImagePreview="openImagePreview"
                    :openMockDataModal="() => { if (mockDataModal) mockDataModal.isOpen = true }"
                    :openInsertRowModal="() => { if (insertRowModal) insertRowModal.isOpen = true }"
                    :startResultSetResize="startResultSetResize" :getResultSetCardStyle="getResultSetCardStyle" />

                <!-- ER Diagram View -->
                <div v-if="activeTab.isERView" class="flex-1 overflow-hidden bg-background">
                    <ERDiagram :tableName="activeTab.tableName || ''"
                        :columns="activeTab.tablesData && activeTab.tableName ? activeTab.tablesData[activeTab.tableName] : []"
                        :relationships="activeTab.relationships || []" :tablesData="activeTab.tablesData || {}"
                        :isDark="isDarkTheme" />
                </div>

                <!-- Table Designer View -->
                <div v-if="activeTab.isDesignView" class="flex-1 overflow-hidden bg-background">
                    <TableStructureDesigner :key="activeTab.id" :table-name="activeTab.tableName || ''"
                        :connection-id="props.connectionId" :db-type="props.dbType" @close="closeTab(activeTab.id)"
                        @refresh="loadTables" @success="handleDesignerSuccess" />
                </div>

            </div>

            <div v-else
                class="flex flex-col items-center justify-center h-full text-muted-foreground animate-in fade-in zoom-in-95 duration-300">
                <div class="h-16 w-16 rounded-2xl bg-muted/50 flex items-center justify-center mb-4 shadow-sm">
                    <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-layers text-primary">
                        <path
                            d="m12.83 2.18a2 2 0 0 0-1.66 0L2.6 6.08a1 1 0 0 0 0 1.83l8.58 3.91a2 2 0 0 0 1.66 0l8.58-3.9a1 1 0 0 0 0-1.83Z" />
                        <path d="m22 17.65-9.17 4.16a2 2 0 0 1-1.66 0L2 17.65" />
                        <path d="m22 12.65-9.17 4.16a2 2 0 0 1-1.66 0L2 12.65" />
                    </svg>
                </div>
                <h3 class="text-lg font-semibold text-foreground">No Open Queries</h3>
                <p class="text-sm max-w-sm text-center mt-2 mb-6">Open a new tab to start writing queries.</p>
                <button @click="addTab"
                    class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-9 px-4 py-2 shadow-sm">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-plus mr-2">
                        <path d="M5 12h14" />
                        <path d="M12 5v14" />
                    </svg>
                    New Query
                </button>
            </div>
        </div>
        <DbDashboardContextMenus :context-menu="contextMenu" :disconnect="disconnect"
            :refresh-database="refreshDatabase" :add-tab="addTab" :handle-backup-export="handleBackupExport"
            :handle-generate-database-e-r-diagram="handleGenerateDatabaseERDiagram"
            :handle-database-info="handleDatabaseInfo" :handle-new-table="handleNewTable"
            :handle-new-view="handleNewView" :handle-new-routine="handleNewRoutine"
            :handle-drop-database="handleDropDatabase" :handle-script-routine="handleScriptRoutine"
            :handle-execute-routine="handleExecuteRoutine" :handle-duplicate-routine="handleDuplicateRoutine"
            :handle-delete-routine="handleDeleteRoutine" :handle-folder-refresh="handleFolderRefresh"
            :handle-folder-collapse="handleFolderCollapse" :handle-copy-row="handleCopyRow"
            :handle-copy-row-with-header="handleCopyRowWithHeader" :handle-copy-cell-value="handleCopyCellValue"
            :handle-copy-cell-value-with-header="handleCopyCellValueWithHeader"
            :handle-add-where-to-condition="handleAddWhereToCondition" :handle-set-null="handleSetNull"
            :handle-set-empty="handleSetEmpty" :handle-set-default="handleSetDefault"
            :handle-select-top100="handleSelectTop100" :handle-view-e-r-diagram="handleViewERDiagram"
            :handle-view-design="handleViewDesign" :handle-script-table-as="handleScriptTableAs"
            :handle-generate-create-statement="handleGenerateCreateStatement"
            :handle-copy-table-name="handleCopyTableName" :handle-truncate-table="handleTruncateTable"
            :handle-drop-table="handleDropTable" :handle-export="handleExport" :handle-import="handleImport"
            :handle-select-top100-view="handleSelectTop100View" />

        <DbSafeModeModal :is-open="safeModeConfirmation && safeModeConfirmation.isOpen"
            :query-to-run="safeModeConfirmation.queryToRun" @cancel="cancelSafeModeQuery"
            @confirm="confirmSafeModeQuery" />

        <DbUpdateConfirmationModal :is-open="!!(updateConfirmation && updateConfirmation.isOpen)"
            :table-name="updateConfirmation?.tableName || ''" :column="updateConfirmation?.column || ''"
            :original-value="updateConfirmation?.originalValue"
            :new-value-display="formatValueForDisplay(updateConfirmation?.newValue)" @close="cancelUpdate"
            @confirm="confirmUpdate" />

        <DbInsertRowModal :is-open="!!(insertRowModal && insertRowModal.isOpen)"
            :table-name="insertRowModal?.tableName || ''" :columns="insertRowModal?.columns || []"
            :values="insertRowModal?.values || {}" :null-columns="insertRowModal?.nullColumns || {}"
            :error="insertRowModal?.error || ''" :is-inserting="!!insertRowModal?.isInserting"
            :column-defs="insertRowModal?.columnDefs || {}" :get-input-type="getInputType"
            :get-number-step="getNumberStep" @close="cancelInsertRow" @confirm="confirmInsertRow"
            @toggle-null="toggleInsertNull" @update:value="updateInsertRowValue" />
        <DbMockDataModals :generator-open="mockDataModal.isOpen" :generator-table-name="mockDataModal.tableName"
            :generator-row-count="mockDataModal.rowCount" :confirm-open="mockDataConfirm.isOpen"
            :confirm-table-name="mockDataConfirm.tableName" :confirm-row-count="mockDataConfirm.rowCount"
            :confirm-loading="mockDataConfirm.isLoading" @close-generator="mockDataModal.isOpen = false"
            @update:generator-row-count="mockDataModal.rowCount = $event" @open-confirm="openMockDataConfirm"
            @close-confirm="mockDataConfirm.isOpen = false" @confirm-insert="confirmMockDataInsert" />



        <DbImportOptionsModal :is-open="showImportOptions" :target-table="importOptions.tableName"
            :file-path="importOptions.filePath" :is-mssql="props.dbType === 'mssql'"
            :enable-identity-insert="importOptions.enableIdentityInsert" @close="showImportOptions = false"
            @confirm="confirmImport" @update:enable-identity-insert="importOptions.enableIdentityInsert = $event" />
        <DbDatabaseInfoModal :is-open="dbInfoModal.isOpen" :is-loading="dbInfoModal.isLoading" :info="dbInfoModal.info"
            @close="dbInfoModal.isOpen = false" />
        <DbDropDatabaseModal :is-open="dropDbConfirmation.isOpen" :db-name="dropDbConfirmation.dbName"
            :is-loading="dropDbConfirmation.isLoading" @confirm="confirmDropDatabase"
            @close="dropDbConfirmation.isOpen = false" />
        <DbTableActionModal :is-open="tableActionConfirmation.isOpen" :action="tableActionConfirmation.action"
            :table-name="tableActionConfirmation.tableName" :is-loading="tableActionConfirmation.isLoading"
            @confirm="confirmTableAction" @close="tableActionConfirmation.isOpen = false" />
        <DbExportDatabaseModal :is-open="exportDbModal.isOpen" :folder-path="exportDbModal.folderPath"
            :format="exportDbModal.format as 'SQL' | 'CSV' | 'JSON' | 'Excel'" :is-loading="exportDbModal.isLoading"
            @close="exportDbModal.isOpen = false" @confirm="confirmExportDb"
            @update:format="exportDbModal.format = $event" />

        <Toast ref="toastRef" />
        <SettingsDialog :is-open="isSettingsOpen" @close="isSettingsOpen = false" @save="handleSettingsSave" />
        <QueryHistory :is-open="isHistoryOpen" :connection-name="connectionName" @close="isHistoryOpen = false"
            @run-query="handleRunHistoryQuery" />

        <DbImagePreviewModal :image-url="imagePreviewUrl" @close="imagePreviewUrl = null" />

        <DbAICopilotOverlay :is-open="aiCopilot.isOpen" :mode="aiCopilot.mode" :mode-options="aiCopilotModeOptions"
            :prompt="aiCopilot.prompt" :backend-language="aiCopilot.backendLanguage" :is-loading="aiCopilot.isLoading"
            :result="aiCopilot.result" :error="aiCopilot.error" :latency-ms="aiCopilot.latencyMs"
            :has-suggested-sql="!!aiCopilot.suggestedSQL" @close="aiCopilot.isOpen = false" @run="runAiCopilot"
            @apply-sql="applyAiSqlToEditor" @update:mode="setAiCopilotMode" @update:prompt="aiCopilot.prompt = $event"
            @update:backend-language="aiCopilot.backendLanguage = $event" />
    </div>
</template>

<script lang="ts" setup>
import { defineAsyncComponent, ref, reactive, watch, onMounted, computed, shallowRef, nextTick, markRaw, onUnmounted } from 'vue';
import { ExecuteQuery, DisconnectDB, GetPrimaryKeys, GetForeignKeys, GetRoutineDefinition, ExportTable, ImportTable, SelectExportFile, SelectImportFile, CancelQuery, ExecuteTransientQuery, GetTableDefinition, LoadSetting, GetQueryHistory } from '../../wailsjs/go/main/App';
import { format } from 'sql-formatter';

// Components
import TableStructureDesigner from './TableStructureDesigner.vue';
import Toast from './Toast.vue';
import SettingsDialog from './SettingsDialog.vue';
import QueryHistory from './QueryHistory.vue';
import DbSidebar from './dashboard/DbSidebar.vue';
import DbActivityMonitor from './dashboard/DbActivityMonitor.vue';
import DbDashboardContextMenus from './dashboard/DbDashboardContextMenus.vue';
import DbQueryWorkspacePane from './dashboard/DbQueryWorkspacePane.vue';
import DbQueryResultsPane from './dashboard/DbQueryResultsPane.vue';
import DbMockDataModals from './dashboard/DbMockDataModals.vue';
import DbTableActionModal from './dashboard/DbTableActionModal.vue';
import DbDatabaseInfoModal from './dashboard/DbDatabaseInfoModal.vue';
import DbDropDatabaseModal from './dashboard/DbDropDatabaseModal.vue';
import DbExportDatabaseModal from './dashboard/DbExportDatabaseModal.vue';
import DbSafeModeModal from './dashboard/DbSafeModeModal.vue';
import DbImportOptionsModal from './dashboard/DbImportOptionsModal.vue';
import DbInsertRowModal from './dashboard/DbInsertRowModal.vue';
import DbUpdateConfirmationModal from './dashboard/DbUpdateConfirmationModal.vue';
import DbImagePreviewModal from './dashboard/DbImagePreviewModal.vue';
import DbAICopilotOverlay from './dashboard/DbAICopilotOverlay.vue';
import { TooltipProvider, TooltipRoot, TooltipTrigger, TooltipContent } from 'radix-vue';
import { completeWithSavedProvider } from '../composables/useAiProvider';

// Composables
import { isDarkTheme } from '../composables/useTheme';
import { useTabs } from '../composables/useTabs';
import { useSidebar } from '../composables/useSidebar';
import { useRecordOperations } from '../composables/useRecordOperations';
import { useQueryResultsView } from '../composables/useQueryResultsView';
import { useActivityMonitor } from '../composables/useActivityMonitor';
import { useQueryExecution } from '../composables/useQueryExecution';
import { useQueryAnalysis } from '../composables/useQueryAnalysis';
import { useResultSetLayout } from '../composables/useResultSetLayout';
import { useDashboardContextMenus } from '../composables/useDashboardContextMenus';
import { useDatabaseAdminModals } from '../composables/useDatabaseAdminModals';
import { useMockDataModal } from '../composables/useMockDataModal';
import { useTableActions } from '../composables/useTableActions';

// Types
import { QueryTab } from '../types/dashboard';
import { ResultSet } from '../types/database';

const ERDiagram = defineAsyncComponent(() => import('./ERDiagram.vue'));

const resultsPaneRef = ref<InstanceType<typeof DbQueryResultsPane> | null>(null);
const toastRef = ref<InstanceType<typeof Toast> | null>(null);
const isSettingsOpen = ref(false);
const isHistoryOpen = ref(false);
const isActivityMonitorOpen = ref(false);
const imagePreviewUrl = ref<string | null>(null);

const globalSettings = ref<any>({});
const safeModeEnabled = computed(() => {
    return globalSettings.value?.general?.enableSafeMode !== false;
});
const perfLoggingEnabled = computed(() => {
    return globalSettings.value?.general?.enablePerfLogs === true;
});

const openImagePreview = (url: any) => {
    if (url) imagePreviewUrl.value = String(url);
};

const isImageValue = (val: any, col: string) => {
    if (val === null || val === undefined) return false;
    const str = String(val).trim();
    if (str.length < 5) return false;

    // Check for common image extensions in URL/Path
    const isUrl = str.startsWith('http://') || str.startsWith('https://');
    const isBase64 = str.startsWith('data:image/');

    // Check if it's a URL or if it looks like a path with image extension
    const extensionMatch = str.split('?')[0].split('#')[0].toLowerCase().match(/\.(jpg|jpeg|png|gif|webp|svg|bmp|ico)$/i);
    const hasImageExt = !!extensionMatch;

    // Also check column names if they contain 'image', 'picture', 'avatar', 'logo', 'blob'
    const colLower = (col || '').toLowerCase();
    const isImageColumn = colLower.includes('image') ||
        colLower.includes('picture') ||
        colLower.includes('avatar') ||
        colLower.includes('logo') ||
        colLower.includes('photo') ||
        colLower.includes('thumb') ||
        colLower.includes('blob');

    // If it's a URL and has image extension, or it's an image column and is a URL
    if (isBase64) return true;
    if (isUrl && (hasImageExt || isImageColumn)) return true;

    // For local paths, we check if it has an image extension
    if (hasImageExt && (str.includes('/') || str.includes('\\'))) return true;

    return false;
};

const handleRunHistoryQuery = (query: string) => {
    isHistoryOpen.value = false;
    if (activeTab.value && !activeTab.value.isERView && !activeTab.value.isDesignView) {
        activeTab.value.query = query;
        setTimeout(() => {
            runQuery();
        }, 50);
    } else {
        addTab();
        setTimeout(() => {
            if (activeTab.value) {
                activeTab.value.query = query;
                runQuery();
            }
        }, 100);
    }
};

// Global settings state
const editorSettings = ref({
    fontFamily: "'JetBrains Mono', monospace",
    fontSize: 14
});

const handleSettingsSave = (newSettings: any) => {
    globalSettings.value = newSettings || {};

    // Update local editor settings
    if (newSettings && newSettings.editor) {
        editorSettings.value.fontFamily = newSettings.editor.fontFamily;
        editorSettings.value.fontSize = newSettings.editor.fontSize;
    }
};

// Emits/Props setup
const emit = defineEmits(['disconnect']);

const props = defineProps<{
    connectionId: string;
    connectionName?: string;
    dbType: string;
    isReadOnly?: boolean;
}>();
const connectionIdRef = computed(() => props.connectionId);
const connectionNameRef = computed(() => props.connectionName);
const dbTypeRef = computed(() => props.dbType);
const isReadOnlyRef = computed(() => props.isReadOnly || false);

// --- Composable: Tabs ---
const {
    tabs,
    activeTabId,
    activeTab,
    addTab,
    closeTab,
    generateId
} = useTabs();

const {
    activityTasksList,
    activityTaskCount,
    monitorRefreshRate,
    latestMonitorSample,
    connectionChartPoints,
    readQpsChartPoints,
    writeQpsChartPoints,
    longRunningTotal,
    formatActivityTime,
    focusActivityTask,
    killActivityTask,
    killAllActivityTasks,
    startMonitorTimer,
    stopMonitorTimer,
} = useActivityMonitor({
    tabs,
    activeTabId,
    isActivityMonitorOpen,
    connectionId: connectionIdRef,
    onCancelError: (message) => {
        toastRef.value?.error(message);
    },
    onCancelSuccess: (message) => {
        toastRef.value?.success(message);
    }
});

// --- Composable: Sidebar ---
const {
    tableSearch,
    viewSearch,
    storedProcedureSearch,
    functionSearch,
    tables,
    views,
    storedProcedures,
    functions,
    openFolders,
    toggleFolder,
    filteredTables,
    filteredViews,
    filteredStoredProcedures,
    filteredFunctions,
    loadTables,
    refreshTables,
    refreshViews,
    refreshStoredProcedures,
    refreshFunctions
} = useSidebar(connectionIdRef);

// Other local state
const workspaceRef = ref<InstanceType<typeof DbQueryWorkspacePane> | null>(null);
const selectedRowIndex = ref<Array<number | string> | number | string | null>(null);
const selectedColumn = ref<string | null>(null);
const selectedRowData = ref<any>(null);
const tableSchemas = ref<Record<string, string[]>>({});

const {
    safeModeConfirmation,
    runQuery,
    stopQuery,
    confirmSafeModeQuery,
    cancelSafeModeQuery,
    cleanupAllStreams,
} = useQueryExecution({
    connectionId: connectionIdRef,
    connectionName: connectionNameRef,
    dbType: dbTypeRef,
    activeTab,
    safeModeEnabled,
    perfLoggingEnabled,
    generateId,
    getSelectedQuery: () => workspaceRef.value?.getSelection() || '',
});

// --- Composable: Record Operations ---
const {
    updateConfirmation,
    insertRowModal,
    isEditable,
    initiateQuickUpdate,
    confirmUpdate,
    openInsertRowModal,
    toggleInsertNull,
    confirmInsertRow,
    getInputType,
    getNumberStep,
    cancelInsertRow,
    cancelUpdate
} = useRecordOperations(
    connectionIdRef,
    props.isReadOnly || false,
    activeTab,
    () => runQuery(),
    toastRef
);

const updateInsertRowValue = (payload: { col: string; value: any }) => {
    if (!insertRowModal.value) return;
    insertRowModal.value.values[payload.col] = payload.value;
};

const {
    startResizing,
    startColumnResize,
    startResultSetResize,
    getResultSetCardStyle,
    stopAllResizing,
} = useResultSetLayout(activeTab);


// Sidebar filtered computed items are now in useSidebar

const {
    contextMenu,
    openDbContextMenu,
    openFolderContextMenu,
    openContextMenu,
    handleRowContextMenu,
    openViewContextMenu,
    openRoutineContextMenu,
    closeContextMenu,
    handleGlobalClick,
} = useDashboardContextMenus({
    isInsideResults: (target) => !!resultsPaneRef.value?.containsTarget(target),
    onOutsideSelection: () => {
        selectedRowIndex.value = null;
        selectedColumn.value = null;
    },
});

const {
    dbInfoModal,
    dropDbConfirmation,
    exportDbModal,
    handleBackupExport,
    confirmExportDb,
    handleDatabaseInfo,
    handleDropDatabase,
    confirmDropDatabase,
} = useDatabaseAdminModals({
    connectionId: connectionIdRef,
    closeContextMenu,
    onSuccess: (message) => {
        toastRef.value?.success(message);
    },
    onError: (message) => {
        toastRef.value?.error(message);
    },
    onDatabaseDropped: () => {
        emit('disconnect');
    },
});

const handleGenerateDatabaseERDiagram = () => {
    closeContextMenu();
    addTab();
    if (activeTab.value) {
        activeTab.value.name = 'ER Diagram';
        activeTab.value.isERView = true;
    }
};

// Redundant handleNewView and handleNewRoutine removed as they are defined later in the file.

const refreshDatabase = async () => {
    closeContextMenu();
    await loadTables();
    if (toastRef.value) {
        toastRef.value.success('Database refreshed successfully.');
    }
};

const handleDesignerSuccess = (message: string) => {
    if (toastRef.value) {
        toastRef.value.success(message);
    }
};

const handleNewTable = () => {
    closeContextMenu();
    addTab();
    if (activeTab.value) {
        activeTab.value.name = 'New Table';
        activeTab.value.tableName = '';
        activeTab.value.isDesignView = true;
    }
};

const handleFolderRefresh = async () => {
    closeContextMenu();
    try {
        if (contextMenu.targetFolder === 'Tables') {
            await refreshTables();
        } else if (contextMenu.targetFolder === 'Views') {
            await refreshViews();
        } else if (contextMenu.targetFolder === 'Programmability') {
            await Promise.all([refreshStoredProcedures(), refreshFunctions()]);
        }
        if (toastRef.value) {
            toastRef.value.success(`${contextMenu.targetFolder} refreshed successfully.`);
        }
    } catch (error) {
        if (toastRef.value) {
            toastRef.value.error(`Failed to refresh ${contextMenu.targetFolder}.`);
        }
    }
};

const handleFolderCollapse = () => {
    closeContextMenu();
    if (openFolders.value.includes(contextMenu.targetFolder)) {
        openFolders.value = openFolders.value.filter(f => f !== contextMenu.targetFolder);
    }
};

const handleCopyRow = () => {
    // If the right-clicked row is part of our multi-selection, we'll just copy the entire selection
    const isMultiSelected = Array.isArray(selectedRowIndex.value)
        ? (contextMenu.targetRowIndex !== null && selectedRowIndex.value.includes(contextMenu.targetRowIndex))
        : false;

    if (isMultiSelected) {
        copySelectedRow(false);
        closeContextMenu();
        return;
    }

    if (contextMenu.targetRow) {
        const values = Object.values(contextMenu.targetRow).map(v => v === null ? 'NULL' : String(v)).join('\t');
        navigator.clipboard.writeText(values);
        closeContextMenu();
    }
};

const handleCopyCellValue = () => {
    // If we right-clicked a cell that is currently part of a selection (especially multi-cell), copy the whole selection
    if (contextMenu.targetRowIndex !== null && contextMenu.targetColumn) {
        if (resultsPaneRef.value?.isCellSelected(contextMenu.targetRowIndex as string | number, contextMenu.targetColumn)) {
            resultsPaneRef.value?.copyCurrentSelection();
            closeContextMenu();
            if (toastRef.value) toastRef.value.success('Selection copied to clipboard');
            return;
        }
    }

    if (contextMenu.targetRow && contextMenu.targetColumn) {
        const val = contextMenu.targetRow[contextMenu.targetColumn];
        const str = val === null ? 'NULL' : String(val);
        navigator.clipboard.writeText(str);
        if (toastRef.value) toastRef.value.success('Cell value copied to clipboard');
        closeContextMenu();
    }
};

const handleCopyRowWithHeader = () => {
    const isMultiSelected = Array.isArray(selectedRowIndex.value)
        ? (contextMenu.targetRowIndex !== null && selectedRowIndex.value.includes(contextMenu.targetRowIndex))
        : false;

    if (isMultiSelected) {
        copySelectedRow(true);
        closeContextMenu();
        return;
    }

    if (contextMenu.targetRow && activeTab.value && activeTab.value.resultSets && activeTab.value.resultSets.length > 0) {
        const rs = activeTab.value.resultSets[0];
        const columns = rs.columns;
        const row = contextMenu.targetRow;

        const headerLine = columns.join('\t');
        const valueLine = columns.map(col => {
            const val = row[col];
            return val === null ? 'NULL' : String(val);
        }).join('\t');

        navigator.clipboard.writeText(`${headerLine}\n${valueLine}`);
        closeContextMenu();
    }
};

const handleCopyCellValueWithHeader = () => {
    if (contextMenu.targetRow && contextMenu.targetColumn) {
        const col = contextMenu.targetColumn;
        const val = contextMenu.targetRow[col];
        const str = val === null ? 'NULL' : String(val);
        navigator.clipboard.writeText(`${col}: ${str}`);
        closeContextMenu();
    }
};

const handleAddWhereToCondition = () => {
    if (contextMenu.targetRow && contextMenu.targetColumn && activeTab.value) {
        const col = contextMenu.targetColumn;
        const val = contextMenu.targetRow[col];
        const tab = activeTab.value;

        const type = (props.dbType || '').toLowerCase();
        let escapedCol = col;

        if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroachdb') || type.includes('sqlite') || type.includes('duckdb')) {
            escapedCol = `"${col}"`;
        } else if (type.includes('mysql') || type.includes('mariadb') || type.includes('databend')) {
            escapedCol = `\`${col}\``;
        } else if (type.includes('mssql') || type.includes('sqlserver')) {
            escapedCol = `[${col}]`;
        }

        let formattedVal = val;
        if (val === null) {
            formattedVal = 'IS NULL';
        } else if (typeof val === 'number') {
            formattedVal = `= ${val}`;
        } else {
            // Escape single quotes for SQL
            const escapedStr = String(val).replace(/'/g, "''");
            formattedVal = `= '${escapedStr}'`;
        }

        const conditionKeyword = /\bWHERE\b/i.test(tab.query) ? 'AND' : 'WHERE';
        const condition = `\n${conditionKeyword} ${escapedCol} ${formattedVal}`;
        tab.query += condition;

        if (toastRef.value) {
            toastRef.value.success(`Added ${conditionKeyword} condition for ${col}`);
        }
        closeContextMenu();
    }
};

const copySelectedCell = () => {
    const isSelected = Array.isArray(selectedRowIndex.value) ? selectedRowIndex.value.length > 0 : selectedRowIndex.value !== null;
    if (isSelected && selectedColumn.value && activeTab.value) {
        const selVal = Array.isArray(selectedRowIndex.value) ? selectedRowIndex.value[0] : selectedRowIndex.value;
        let rowData: any = null;
        if (typeof selVal === 'number') {
            rowData = filteredResults.value[selVal];
        } else if (typeof selVal === 'string' && selVal.startsWith('sub-')) {
            const parts = selVal.split('-');
            const rsIdx = parseInt(parts[1]) + 1;
            const rIdx = parseInt(parts[2]);
            if (activeTab.value.resultSets && activeTab.value.resultSets[rsIdx]) {
                rowData = activeTab.value.resultSets[rsIdx].rows[rIdx];
            }
        }

        if (rowData) {
            const val = rowData[selectedColumn.value];
            const str = val === null ? 'NULL' : String(val);
            navigator.clipboard.writeText(str);
            if (toastRef.value) toastRef.value.success('Cell value copied to clipboard');
        }
    }
};

const copySelectedRow = (withHeader: boolean = false) => {
    const isSelected = Array.isArray(selectedRowIndex.value) ? selectedRowIndex.value.length > 0 : selectedRowIndex.value !== null;
    if (isSelected && activeTab.value) {
        const indices = Array.isArray(selectedRowIndex.value)
            ? selectedRowIndex.value.slice().sort((a, b) => {
                if (typeof a === 'number' && typeof b === 'number') return a - b;
                return String(a).localeCompare(String(b));
            })
            : [selectedRowIndex.value];

        let headersStr = '';
        let rowsStrs: string[] = [];

        for (const selVal of indices) {
            let rowData: any = null;
            let columns: string[] = [];

            if (typeof selVal === 'number') {
                rowData = filteredResults.value[selVal];
                if (activeTab.value.resultSets && activeTab.value.resultSets[0]) {
                    columns = activeTab.value.resultSets[0].columns;
                }
            } else if (typeof selVal === 'string' && selVal.startsWith('sub-')) {
                const parts = selVal.split('-');
                const rsIdx = parseInt(parts[1]) + 1;
                const rIdx = parseInt(parts[2]);
                if (activeTab.value.resultSets && activeTab.value.resultSets[rsIdx]) {
                    rowData = activeTab.value.resultSets[rsIdx].rows[rIdx];
                    columns = activeTab.value.resultSets[rsIdx].columns;
                }
            }

            if (rowData && columns.length > 0) {
                if (!headersStr) headersStr = columns.join('\t');
                const valueLine = columns.map(col => {
                    const val = rowData[col];
                    return val === null ? 'NULL' : String(val);
                }).join('\t');
                rowsStrs.push(valueLine);
            }
        }

        if (rowsStrs.length > 0) {
            if (withHeader) {
                navigator.clipboard.writeText(`${headersStr}\n${rowsStrs.join('\n')}`);
                if (toastRef.value) toastRef.value.success(`${rowsStrs.length} row(s) with header copied to clipboard`);
            } else {
                navigator.clipboard.writeText(rowsStrs.join('\n'));
                if (toastRef.value) toastRef.value.success(`${rowsStrs.length} row(s) copied to clipboard`);
            }
        }
    }
};

// Import Options State
const showImportOptions = ref(false);
const importOptions = ref({
    filePath: '',
    format: '',
    tableName: '',
    enableIdentityInsert: false
});

const handleExport = async () => {
    if (!contextMenu.targetTable) return;
    const tableName = contextMenu.targetTable;

    const result = await SelectExportFile(`${tableName}_export.json`);

    if (result) {
        let format = "json";
        if (result.endsWith(".csv")) format = "csv";
        else if (result.endsWith(".sql")) format = "sql";
        else if (result.endsWith(".xlsx")) format = "excel";

        try {
            const resp = await ExportTable(props.connectionId, tableName, format, result);
            if (resp !== "Success") {
                toastRef.value?.error(resp);
            } else {
                toastRef.value?.success("Export successful!");
            }
        } catch (e) {
            toastRef.value?.error("Error exporting: " + e);
        }
    }
    closeContextMenu();
};

const handleImport = async () => {
    if (!contextMenu.targetTable) return;
    const tableName = contextMenu.targetTable;

    const result = await SelectImportFile();

    if (result) {
        let format = "json";
        if (result.endsWith(".csv")) format = "csv";
        else if (result.endsWith(".sql")) format = "sql";
        else if (result.endsWith(".xlsx")) format = "excel";

        importOptions.value = {
            filePath: result,
            format: format,
            tableName: tableName,
            enableIdentityInsert: false
        };
        showImportOptions.value = true;
    }
    closeContextMenu();
};

const confirmImport = async () => {
    showImportOptions.value = false;
    try {
        const resp = await ImportTable(
            props.connectionId,
            importOptions.value.tableName,
            importOptions.value.format,
            importOptions.value.filePath,
            importOptions.value.enableIdentityInsert
        );
        if (resp !== "Success") {
            toastRef.value?.error(resp);
        } else {
            toastRef.value?.success("Import Successful!");
            // Optionally refresh if the table is open or just notify
        }
    } catch (e) {
        toastRef.value?.error("Error importing: " + e);
    }
};


const {
    activeResultSet,
    filteredResults,
    getColumns
} = useQueryResultsView(
    computed(() => tabs.value.find(t => t.id === activeTabId.value))
);

const toggleSort = (col: string) => {
    const tab = tabs.value.find(t => t.id === activeTabId.value);
    if (!tab) return;

    if (tab.sortColumn === col) {
        if (tab.sortDirection === 'asc') {
            tab.sortDirection = 'desc';
        } else if (tab.sortDirection === 'desc') {
            tab.sortDirection = null;
            tab.sortColumn = undefined;
        } else {
            tab.sortDirection = 'asc';
        }
    } else {
        tab.sortColumn = col;
        tab.sortDirection = 'asc';
    }
};

// closeTab moved to useTabs

// loadTables moved to useSidebar

const selectTable = async (tableName: string) => {
    // Check if table is already open in a tab
    const existingTab = tabs.value.find(t => t.tableName === tableName);

    if (existingTab) {
        activeTabId.value = existingTab.id;
        return;
    }

    // Check if current active tab is empty/pristine
    const currentTab = activeTab.value;
    const isPristine = currentTab &&
        !currentTab.tableName &&
        !currentTab.query &&
        !currentTab.isDesignView &&
        !currentTab.isERView &&
        !currentTab.queryExecuted;

    if (!isPristine) {
        addTab();
    }

    if (activeTab.value) {
        const type = (props.dbType || '').toLowerCase();
        activeTab.value.tableName = tableName;
        activeTab.value.name = tableName; // Update tab name to table name

        let escapedTableName = tableName;
        if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroachdb') || type.includes('sqlite') || type.includes('duckdb')) {
            escapedTableName = `"${tableName}"`;
        } else if (type.includes('mysql') || type.includes('mariadb') || type.includes('databend')) {
            escapedTableName = `\`${tableName}\``;
        } else if (type.includes('mssql') || type.includes('sqlserver')) {
            if (tableName.includes('.') && !tableName.includes('[')) {
                escapedTableName = tableName.split('.').map(p => `[${p}]`).join('.');
            } else if (!tableName.startsWith('[')) {
                escapedTableName = `[${tableName}]`;
            }
        }

        if (type.includes('mssql') || type.includes('sqlserver')) {
            activeTab.value.query = `SELECT TOP 100 * FROM ${escapedTableName}`;
        } else {
            activeTab.value.query = `SELECT * FROM ${escapedTableName} LIMIT 100`;
        }

        // Fetch Primary Keys
        try {
            activeTab.value.primaryKeys = (await GetPrimaryKeys(props.connectionId, tableName)) || [];
        } catch (e) {
            console.error("Failed to fetch primary keys", e);
            activeTab.value.primaryKeys = [];
        }

        runQuery();
        checkRowCount(tableName);
    }
};

const selectRoutine = async (name: string, routineType: 'PROCEDURE' | 'FUNCTION') => {
    // Check if routine is already open in a tab
    const existingTab = tabs.value.find(t => t.name === name && !t.isDesignView && !t.isERView);

    if (existingTab) {
        activeTabId.value = existingTab.id;
        return;
    }

    // Check if current active tab is empty/pristine
    const currentTab = activeTab.value;
    const isPristine = currentTab &&
        !currentTab.tableName &&
        !currentTab.query &&
        !currentTab.isDesignView &&
        !currentTab.isERView &&
        !currentTab.queryExecuted;

    if (!isPristine) {
        addTab();
    }

    if (activeTab.value) {
        activeTab.value.name = name;
        activeTab.value.isLoading = true;
        activeTab.value.tableName = ""; // It's not a table
        activeTab.value.isRoutine = true;
        activeTab.value.routineName = name;
        activeTab.value.routineType = routineType;

        try {
            const definition = await GetRoutineDefinition(props.connectionId, name, routineType);
            if (definition.startsWith("Error: ")) {
                activeTab.value.error = definition;
                activeTab.value.query = `-- Failed to fetch definition for ${name}`;
            } else {
                activeTab.value.query = definition;
            }
        } catch (e) {
            console.error("Failed to fetch routine definition", e);
            activeTab.value.error = "Failed to fetch routine definition";
        } finally {
            activeTab.value.isLoading = false;
        }
    }
};

const checkRowCount = async (tableName: string) => {
    if (!activeTab.value) return;

    // Reset previous count
    activeTab.value.totalRowCount = undefined;

    const type = (props.dbType || '').toLowerCase();

    let escapedTableName = tableName;
    if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroachdb') || type.includes('sqlite') || type.includes('duckdb')) {
        escapedTableName = `"${tableName}"`;
    } else if (type.includes('mysql') || type.includes('mariadb') || type.includes('databend')) {
        escapedTableName = `\`${tableName}\``;
    } else if (type.includes('mssql') || type.includes('sqlserver')) {
        if (tableName.includes('.') && !tableName.includes('[')) {
            escapedTableName = tableName.split('.').map(p => `[${p}]`).join('.');
        } else if (!tableName.startsWith('[')) {
            escapedTableName = `[${tableName}]`;
        }
    }

    let countQuery = `SELECT COUNT(*) FROM ${escapedTableName}`;

    try {
        const reqId = generateId();
        const res = await ExecuteTransientQuery(props.connectionId, countQuery);
        if (res.error) {
            console.warn("Failed to get row count", res.error);
        } else if (res.resultSets && res.resultSets.length > 0 && res.resultSets[0].rows && res.resultSets[0].rows.length > 0) {
            const row = res.resultSets[0].rows[0];
            const val = Object.values(row)[0];
            activeTab.value.totalRowCount = Number(val);
        }
    } catch (e) {
        console.warn("Failed to get row count", e);
    }
};

const openDesignView = (tableName: string) => {
    // Check if already open
    const existingTab = tabs.value.find(t => t.tableName === tableName && t.isDesignView);
    if (existingTab) {
        activeTabId.value = existingTab.id;
        return;
    }

    const newId = generateId();
    tabs.value.push({
        id: newId,
        name: `Design: ${tableName}`,
        tableName: tableName,
        query: '',
        resultSets: [],
        primaryKeys: [],
        filters: {},
        sortColumn: undefined,
        sortDirection: null,
        error: '',
        isLoading: false,
        queryExecuted: false,
        activeQueryIds: [],
        resultViewTab: 'data',
        isDesignView: true,
        isExplaining: false,
        isERView: false,
        isPartialStats: false,
        editorHeight: 300,
        columnWidths: {}
    });
    activeTabId.value = newId;
};

const handleViewDesign = () => {
    if (contextMenu.targetTable) {
        openDesignView(contextMenu.targetTable);
        closeContextMenu();
    }
};

// ... View Logic ...

const selectView = (viewName: string) => {
    const existingTab = tabs.value.find(t => t.tableName === viewName && !t.isDesignView && !t.isERView);
    if (existingTab) {
        activeTabId.value = existingTab.id;
        return;
    }

    const currentTab = activeTab.value;
    const isPristine = currentTab &&
        !currentTab.tableName &&
        !currentTab.query &&
        !currentTab.isDesignView &&
        !currentTab.isERView &&
        !currentTab.queryExecuted;

    if (!isPristine) {
        addTab();
    }

    if (activeTab.value) {
        const type = (props.dbType || '').toLowerCase();
        activeTab.value.tableName = viewName;
        activeTab.value.name = viewName;

        if (type.includes('mssql') || type.includes('sqlserver')) {
            activeTab.value.query = `SELECT TOP 100 * FROM ${viewName}`;
        } else {
            activeTab.value.query = `SELECT * FROM ${viewName} LIMIT 100`;
        }

        // Views usually don't have PKs in the same way, or at least we typically don't edit them directly via grid easily without triggers/configuration.
        // So we might skip GetPrimaryKeys or just let it fail gracefully.
        activeTab.value.primaryKeys = [];

        runQuery();
        checkRowCount(viewName);
    }
};

const handleSelectTop100View = () => {
    if (contextMenu.targetView) {
        selectView(contextMenu.targetView);
        closeContextMenu();
    }
};

const handleNewView = () => {
    addTab();
    if (activeTab.value) {
        activeTab.value.name = "New View";
        activeTab.value.query = `-- Create a new view
CREATE VIEW NewViewName AS
SELECT *
FROM ExistingTable;`;
    }
    closeContextMenu();
};

const handleDuplicateRoutine = async () => {
    const routine = contextMenu.targetRoutine;
    const type = contextMenu.targetRoutineType;
    if (!routine) return;

    closeContextMenu();

    try {
        const definition = await GetRoutineDefinition(props.connectionId, routine, type);
        if (definition.startsWith("Error: ")) {
            if (toastRef.value) toastRef.value.error("Failed to fetch routine definition for duplication");
            return;
        }

        addTab();
        if (activeTab.value) {
            const copyName = `${routine}_copy`;
            // Attempt to replace the name in the definition (very simple replacement)
            let newDefinition = definition.replace(new RegExp(routine, 'g'), copyName);

            activeTab.value.name = `New ${copyName}`;
            activeTab.value.query = newDefinition;
            activeTab.value.isRoutine = true;
            activeTab.value.routineName = copyName;
            activeTab.value.routineType = type;
        }
    } catch (e) {
        console.error("Failed to duplicate routine", e);
    }
};

const handleDeleteRoutine = async () => {
    const routine = contextMenu.targetRoutine;
    const type = contextMenu.targetRoutineType;
    if (!routine) return;

    if (!confirm(`Are you sure you want to delete ${type.toLowerCase()} '${routine}'?`)) {
        closeContextMenu();
        return;
    }

    closeContextMenu();

    const dropSql = `DROP ${type} ${routine}`;
    try {
        const res = await ExecuteQuery(props.connectionId, dropSql, generateId());
        if (res.error) {
            if (toastRef.value) toastRef.value.error(`Failed to delete routine: ${res.error}`);
        } else {
            if (toastRef.value) toastRef.value.success(`${type} deleted successfully`);
            // Refresh sidebar
            if (type === 'PROCEDURE') {
                refreshStoredProcedures();
            } else {
                refreshFunctions();
            }
        }
    } catch (e) {
        console.error("Error deleting routine", e);
    }
};

const handleExecuteRoutine = () => {
    const routine = contextMenu.targetRoutine;
    const type = contextMenu.targetRoutineType;
    if (!routine) return;

    closeContextMenu();
    addTab();

    if (activeTab.value) {
        activeTab.value.name = `Exec: ${routine}`;
        const dbType = (props.dbType || '').toLowerCase();

        let template = '';
        if (dbType.includes('postgres') || dbType.includes('greenplum')) {
            if (type === 'FUNCTION') {
                template = `SELECT ${routine}(/* parameters */);`;
            } else {
                template = `CALL ${routine}(/* parameters */);`;
            }
        } else if (dbType.includes('mysql') || dbType.includes('mariadb')) {
            if (type === 'FUNCTION') {
                template = `SELECT ${routine}(/* parameters */);`;
            } else {
                template = `CALL ${routine}(/* parameters */);`;
            }
        } else if (dbType.includes('mssql') || dbType.includes('sqlserver')) {
            template = `EXEC ${routine} /* parameters */;`;
        } else {
            template = `-- Execution template for ${routine}
-- ${type === 'PROCEDURE' ? 'CALL/EXEC' : 'SELECT'} ${routine}(...);`;
        }

        activeTab.value.query = template;
    }
};

const handleSaveRoutine = async () => {
    if (!activeTab.value || !activeTab.value.isRoutine) return;

    activeTab.value.isLoading = true;
    try {
        const res = await ExecuteQuery(props.connectionId, activeTab.value.query, generateId());
        if (res.error) {
            activeTab.value.error = res.error;
            if (toastRef.value) toastRef.value.error("Failed to save routine");
        } else {
            activeTab.value.queryExecuted = true;
            activeTab.value.error = "";
            if (toastRef.value) toastRef.value.success("Routine saved and updated successfully");

            // Refresh sidebar list
            if (activeTab.value.routineType === 'PROCEDURE') {
                refreshStoredProcedures();
            } else {
                refreshFunctions();
            }
        }
    } catch (e) {
        console.error("Error saving routine", e);
        activeTab.value.error = "Failed to save routine";
    } finally {
        activeTab.value.isLoading = false;
    }
};

const handleNewRoutine = (type: 'PROCEDURE' | 'FUNCTION') => {
    closeContextMenu();
    addTab();

    if (activeTab.value) {
        const dbType = (props.dbType || '').toLowerCase();
        const placeholderName = type === 'PROCEDURE' ? 'new_procedure' : 'new_function';
        activeTab.value.name = `New ${type.toLowerCase()}`;
        activeTab.value.isRoutine = true;
        activeTab.value.routineType = type;
        activeTab.value.routineName = placeholderName;

        let template = '';
        if (dbType.includes('postgres') || dbType.includes('greenplum')) {
            if (type === 'PROCEDURE') {
                template = `CREATE OR REPLACE PROCEDURE ${placeholderName}(/* parameters */)
LANGUAGE plpgsql
AS $$
BEGIN
    -- procedure body
END;
$$;`;
            } else {
                template = `CREATE OR REPLACE FUNCTION ${placeholderName}(/* parameters */)
RETURNS /* type */
LANGUAGE plpgsql
AS $$
BEGIN
    -- function body
    RETURN /* value */;
END;
$$;`;
            }
        } else if (dbType.includes('mysql') || dbType.includes('mariadb')) {
            if (type === 'PROCEDURE') {
                template = `DELIMITER //
CREATE PROCEDURE ${placeholderName}(/* parameters */)
BEGIN
    -- procedure body
END //
DELIMITER ;`;
            } else {
                template = `CREATE FUNCTION ${placeholderName}(/* parameters */)
RETURNS /* type */
DETERMINISTIC
BEGIN
    -- function body
    RETURN /* value */;
END;`;
            }
        } else if (dbType.includes('mssql') || dbType.includes('sqlserver')) {
            if (type === 'PROCEDURE') {
                template = `CREATE PROCEDURE ${placeholderName}
    /* @param1 type, @param2 type */
AS
BEGIN
    SET NOCOUNT ON;
    -- procedure body
END;`;
            } else {
                template = `CREATE FUNCTION ${placeholderName}
(
    /* @param1 type, @param2 type */
)
RETURNS /* type */
AS
BEGIN
    -- function body
    RETURN /* value */
END;`;
            }
        } else {
            template = `CREATE ${type} ${placeholderName} ...`;
        }

        activeTab.value.query = template;
    }
};

const handleScriptRoutine = () => {
    const routine = contextMenu.targetRoutine;
    if (!routine) return;

    // We don't have a GetRoutineDefinition method yet, so placeholders for now,
    // or we can try to guess/query standard schemas.
    // Ideally we should add GetRoutineDefinition backend method.
    // For now, let's open a query that *would* show the definition, or just a stub.

    // Actually, scripting requires fetching the definition which is DB specific.
    // Let's create a tab that *queries* the definition using helper SQL.

    addTab();
    if (activeTab.value) {
        activeTab.value.name = `Script: ${routine}`;
        activeTab.value.query = `-- Scripting for ${routine} (${contextMenu.targetRoutineType})
-- Note: Provide a backend method 'GetRoutineDefinition' for better support.

-- Postgres:
-- SELECT pg_get_functiondef('${routine}'::regproc);

-- MySQL:
-- SHOW CREATE PROCEDURE ${routine};

-- MSSQL:
-- sp_helptext '${routine}';
`;

        // Try to be smart for MSSQL at least
        const type = (props.dbType || '').toLowerCase();
        if (type.includes('mssql')) {
            activeTab.value.query = `EXEC sp_helptext '${routine}'`;
            setTimeout(() => runQuery(), 50);
        } else if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroach')) {
            activeTab.value.query = `SELECT pg_get_functiondef('${routine}'::regproc)`;
            // This might fail if schema is needed or not in search path, but good attempt
        } else if (type.includes('mysql') || type.includes('maria') || type.includes('databend')) {
            activeTab.value.query = `SHOW CREATE ${contextMenu.targetRoutineType} ${routine}`;
            setTimeout(() => runQuery(), 50);
        } else if (type.includes('sqlite') || type.includes('libsql')) {
            activeTab.value.query = `SELECT sql FROM sqlite_master WHERE name = '${routine}'`;
            setTimeout(() => runQuery(), 50);
        }

    }
    closeContextMenu();
};

const handleSelectTop100 = () => {
    if (contextMenu.targetTable) {
        selectTable(contextMenu.targetTable);
        closeContextMenu();
    }
};

const getEscapedIdentifier = (identifier: string) => {
    const type = (props.dbType || '').toLowerCase();
    if (type.includes('mysql') || type.includes('mariadb') || type.includes('databend')) {
        return `\`${identifier.replace(/`/g, '``')}\``;
    }
    if (type.includes('mssql') || type.includes('sqlserver')) {
        return `[${identifier.replace(/]/g, ']]')}]`;
    }
    return `"${identifier.replace(/"/g, '""')}"`;
};

const getEscapedTableName = (tableName: string) => {
    const type = (props.dbType || '').toLowerCase();
    if (type.includes('mssql') || type.includes('sqlserver')) {
        if (tableName.includes('.') && !tableName.includes('[')) {
            return tableName.split('.').map(part => getEscapedIdentifier(part)).join('.');
        }
        if (!tableName.startsWith('[')) {
            return getEscapedIdentifier(tableName);
        }
        return tableName;
    }

    if (tableName.includes('.') && !tableName.includes('"') && !tableName.includes('`')) {
        return tableName.split('.').map(part => getEscapedIdentifier(part)).join('.');
    }
    return getEscapedIdentifier(tableName);
};

const openSqlTemplateTab = (tabName: string, sql: string) => {
    const currentTab = activeTab.value;
    const isPristine = currentTab &&
        !currentTab.tableName &&
        !currentTab.query &&
        !currentTab.isDesignView &&
        !currentTab.isERView &&
        !currentTab.queryExecuted;

    if (!isPristine) {
        addTab();
    }

    if (activeTab.value) {
        activeTab.value.name = tabName;
        activeTab.value.tableName = '';
        activeTab.value.query = sql;
        activeTab.value.error = '';
        activeTab.value.queryExecuted = false;
        activeTab.value.resultSets = markRaw([]);
        activeTab.value.resultViewTab = 'data';
    }
};

const handleScriptTableAs = async (action: 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE') => {
    const tableName = contextMenu.targetTable;
    if (!tableName) return;

    const escapedTableName = getEscapedTableName(tableName);
    const type = (props.dbType || '').toLowerCase();

    try {
        if (action === 'SELECT') {
            const sql = (type.includes('mssql') || type.includes('sqlserver'))
                ? `SELECT TOP 100 *\nFROM ${escapedTableName};`
                : `SELECT *\nFROM ${escapedTableName}\nLIMIT 100;`;
            openSqlTemplateTab(`Script ${action}: ${tableName}`, sql);
            closeContextMenu();
            return;
        }

        const columns = await GetTableDefinition(props.connectionId, tableName);
        if (!columns || columns.length === 0) {
            toastRef.value?.error('Could not read table columns for scripting.');
            closeContextMenu();
            return;
        }

        const pkColumns = columns.filter(col => col.primaryKey).map(col => col.name);
        const firstColumn = columns[0]?.name;
        const fallbackWhereColumns = firstColumn ? [firstColumn] : [];

        if (action === 'INSERT') {
            const insertColumns = columns.filter(col => !col.autoIncrement);
            const targetColumns = insertColumns.length > 0 ? insertColumns : columns;
            const columnLines = targetColumns.map(col => `  ${getEscapedIdentifier(col.name)}`).join(',\n');
            const valueLines = targetColumns.map(col => `  /* ${col.name} */`).join(',\n');
            const sql = `INSERT INTO ${escapedTableName} (\n${columnLines}\n)\nVALUES (\n${valueLines}\n);`;
            openSqlTemplateTab(`Script ${action}: ${tableName}`, sql);
            closeContextMenu();
            return;
        }

        if (action === 'UPDATE') {
            const setColumns = columns.filter(col => !col.primaryKey);
            const targetSetColumns = setColumns.length > 0 ? setColumns : columns;
            const whereColumns = pkColumns.length > 0 ? pkColumns : fallbackWhereColumns;
            const setLines = targetSetColumns.map(col => `  ${getEscapedIdentifier(col.name)} = /* ${col.name} */`).join(',\n');
            const whereLines = whereColumns.map(col => `  ${getEscapedIdentifier(col)} = /* ${col} */`).join('\n  AND ');
            const sql = `UPDATE ${escapedTableName}\nSET\n${setLines}\nWHERE\n${whereLines || '  /* condition */'};`;
            openSqlTemplateTab(`Script ${action}: ${tableName}`, sql);
            closeContextMenu();
            return;
        }

        const deleteWhereColumns = pkColumns.length > 0 ? pkColumns : fallbackWhereColumns;
        const deleteWhereLines = deleteWhereColumns.map(col => `  ${getEscapedIdentifier(col)} = /* ${col} */`).join('\n  AND ');
        const deleteSql = `DELETE FROM ${escapedTableName}\nWHERE\n${deleteWhereLines || '  /* condition */'};`;
        openSqlTemplateTab(`Script ${action}: ${tableName}`, deleteSql);
    } catch (e) {
        console.error('Failed to script table statement', e);
        toastRef.value?.error('Failed to generate SQL script.');
    } finally {
        closeContextMenu();
    }
};

const handleGenerateCreateStatement = async () => {
    const tableName = contextMenu.targetTable;
    if (!tableName) return;

    try {
        const columns = await GetTableDefinition(props.connectionId, tableName);
        if (!columns || columns.length === 0) {
            toastRef.value?.error('Could not read table columns.');
            closeContextMenu();
            return;
        }

        const pkColumns = columns.filter(col => col.primaryKey).map(col => col.name);
        const columnDefs = columns.map(col => {
            let line = `${getEscapedIdentifier(col.name)} ${col.type || 'TEXT'}`;
            if (!col.nullable) line += ' NOT NULL';
            if (col.defaultValue !== null && col.defaultValue !== undefined && String(col.defaultValue) !== '') {
                line += ` DEFAULT ${String(col.defaultValue)}`;
            }
            return line;
        });

        if (pkColumns.length > 0) {
            columnDefs.push(`PRIMARY KEY (${pkColumns.map(col => getEscapedIdentifier(col)).join(', ')})`);
        }

        const createSql = `CREATE TABLE ${getEscapedTableName(tableName)} (\n${columnDefs.map(line => `  ${line}`).join(',\n')}\n);`;
        await navigator.clipboard.writeText(createSql);
        toastRef.value?.success('CREATE TABLE statement copied to clipboard.');
    } catch (e) {
        console.error('Failed to generate CREATE statement', e);
        toastRef.value?.error('Failed to generate CREATE TABLE statement.');
    } finally {
        closeContextMenu();
    }
};

const {
    tableActionConfirmation,
    handleTruncateTable,
    handleDropTable,
    confirmTableAction,
} = useTableActions({
    connectionId: connectionIdRef,
    dbType: dbTypeRef,
    isReadOnly: isReadOnlyRef,
    getTargetTable: () => contextMenu.targetTable,
    closeContextMenu,
    generateId,
    executeQuery: (connectionId, query, requestId) => ExecuteQuery(connectionId, query, requestId),
    getEscapedTableName,
    loadTables,
    checkRowCount,
    onSuccess: (message) => {
        toastRef.value?.success(message);
    },
    onError: (message) => {
        toastRef.value?.error(message);
    },
});

const randomInt = (min: number, max: number) => Math.floor(Math.random() * (max - min + 1)) + min;
const randomPick = <T>(arr: T[]): T => arr[randomInt(0, arr.length - 1)];
const randomString = (len: number) => {
    const chars = 'abcdefghijklmnopqrstuvwxyz';
    let out = '';
    for (let i = 0; i < len; i++) out += chars[randomInt(0, chars.length - 1)];
    return out;
};

const escapeSqlString = (value: string) => value.replace(/'/g, "''");

const generateMockSqlValue = (col: any, rowIndex: number) => {
    const type = String(col.type || '').toLowerCase();
    const name = String(col.name || '').toLowerCase();
    const dbType = (props.dbType || '').toLowerCase();
    const nullable = !!col.nullable;

    if (nullable && Math.random() < 0.08) return 'NULL';

    if (type.includes('bool') || type.includes('bit')) {
        if (dbType.includes('mssql')) return Math.random() > 0.5 ? '1' : '0';
        return Math.random() > 0.5 ? 'TRUE' : 'FALSE';
    }
    if (type.includes('int') || type.includes('serial')) {
        return String(randomInt(1, 100000));
    }
    if (type.includes('decimal') || type.includes('numeric') || type.includes('float') || type.includes('double') || type.includes('real')) {
        return (Math.random() * 10000).toFixed(2);
    }
    if (type.includes('date') && !type.includes('time')) {
        const d = new Date(Date.now() - randomInt(0, 365) * 86400000);
        const iso = d.toISOString().slice(0, 10);
        return `'${iso}'`;
    }
    if (type.includes('time') || type.includes('timestamp') || type.includes('datetime')) {
        const d = new Date(Date.now() - randomInt(0, 365) * 86400000);
        const iso = d.toISOString().slice(0, 19).replace('T', ' ');
        return `'${iso}'`;
    }
    if (type.includes('uuid') || name.includes('uuid')) {
        return `'${crypto.randomUUID()}'`;
    }

    if (name.includes('email')) {
        return `'${escapeSqlString(`${randomString(6)}${rowIndex}@example.com`)}'`;
    }
    if (name.includes('name')) {
        return `'${escapeSqlString(`${randomPick(['Alex', 'Nina', 'John', 'May', 'Leo'])} ${randomPick(['Kim', 'Park', 'Smith', 'Tan'])}`)}'`;
    }
    if (name.includes('phone')) {
        return `'${escapeSqlString(`08${randomInt(10000000, 99999999)}`)}'`;
    }
    if (type.includes('json')) {
        return `'${escapeSqlString(JSON.stringify({ seed: rowIndex, active: Math.random() > 0.5 }))}'`;
    }

    return `'${escapeSqlString(`${randomPick(['sample', 'mock', 'demo'])}_${randomString(5)}_${rowIndex}`)}'`;
};

const {
    mockDataModal,
    mockDataConfirm,
    openMockDataModal,
    openMockDataConfirm,
    confirmMockDataInsert,
} = useMockDataModal({
    activeTab,
    connectionId: connectionIdRef,
    generateId,
    getTableDefinition: (connectionId, tableName) => GetTableDefinition(connectionId, tableName),
    executeQuery: (connectionId, query, requestId) => ExecuteQuery(connectionId, query, requestId),
    getEscapedTableName,
    getEscapedIdentifier,
    generateMockSqlValue,
    onRefreshCurrentTable: () => runQuery(true),
    onRefreshRowCount: (tableName) => checkRowCount(tableName),
    onError: (message) => {
        toastRef.value?.error(message);
    },
    onSuccess: (message) => {
        toastRef.value?.success(message);
    },
});

const handleCopyTableName = async () => {
    if (!contextMenu.targetTable) return;
    try {
        await navigator.clipboard.writeText(contextMenu.targetTable);
        toastRef.value?.success(`Copied table name: ${contextMenu.targetTable}`);
    } catch (e) {
        console.error('Failed to copy table name', e);
        toastRef.value?.error('Failed to copy table name.');
    } finally {
        closeContextMenu();
    }
};


const openERDiagramTab = async (tableName: string) => {
    const existingTab = tabs.value.find(t => t.name === `ER: ${tableName}`);
    if (existingTab) {
        activeTabId.value = existingTab.id;
        return;
    }

    const newId = generateId();

    // Create tab
    const newTab: QueryTab = {
        id: newId,
        name: `ER: ${tableName}`,
        tableName: tableName,
        query: '',
        resultSets: [],
        primaryKeys: [],
        filters: {},
        sortColumn: undefined,
        sortDirection: null,
        error: '',
        isLoading: true,
        queryExecuted: true,
        isERView: true,
        relationships: [],
        activeQueryIds: [],
        resultViewTab: 'data',
        editorHeight: 300,
        columnWidths: {}
    };

    tabs.value.push(newTab);
    activeTabId.value = newId;

    try {
        // 1. Get Foreign Keys First (Bidirectional)
        const fks = (await GetForeignKeys(props.connectionId, tableName)) || [];
        newTab.relationships = fks;

        // 2. Identify all tables involved (Main table + any referenced/referencing tables)
        const relatedTables = new Set<string>();
        relatedTables.add(tableName);
        fks.forEach((fk: any) => {
            relatedTables.add(fk.table);
            relatedTables.add(fk.refTable);
            // fk.table is the child (Referencing), fk.refTable is the parent (Referenced)
        });

        // 3. Fetch Schema for EACH table
        const tablesData: Record<string, { name: string, type: string }[]> = {};
        const type = (props.dbType || '').toLowerCase();

        // Helper to get query for a table
        const getSchemaQuery = (tbl: string) => {
            if (type.includes('mssql') || type.includes('sqlserver')) {
                return `SELECT COLUMN_NAME, DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '${tbl}'`;
            } else if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroach')) {
                return `SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '${tbl}'`;
            } else if (type.includes('mysql') || type.includes('maria') || type.includes('databend')) {
                return `DESCRIBE ${tbl}`;
            } else if (type.includes('sqlite') || type.includes('libsql')) {
                return `PRAGMA table_info(${tbl})`;
            } else {
                return `SELECT * FROM ${tbl} LIMIT 1`;
            }
        };

        // Execute queries in parallel
        const promises = Array.from(relatedTables).map(async (tbl) => {
            const query = getSchemaQuery(tbl);
            const reqId = generateId();
            newTab.activeQueryIds.push(reqId);
            try {
                const res = await ExecuteQuery(props.connectionId, query, reqId);
                // Need to handle resultSets here too
                if (!res.error && res.resultSets && res.resultSets.length > 0) {
                    const rs = res.resultSets[0];
                    // Convert array rows to object rows
                    const rows = (rs.rows || []).map((row: any[]) =>
                        Object.fromEntries((rs.columns || []).map((col: string, i: number) => [col, row[i]]))
                    );
                    tablesData[tbl] = rows.map((col: any) => {
                        const name = col.COLUMN_NAME || col.column_name || col.Field || col.name || col.Name || 'unknown';
                        const type = col.DATA_TYPE || col.data_type || col.Type || col.type || 'string';
                        return { name, type };
                    });
                }
            } catch (e) {
                console.warn(`Failed to fetch schema for ${tbl}`, e);
            } finally {
                newTab.activeQueryIds = newTab.activeQueryIds.filter(id => id !== reqId);
            }
        });

        await Promise.all(promises);
        newTab.tablesData = tablesData;

        // Keep main table columns in results for legacy/other uses if needed
        // IF we want to show it in grid? ER view handles its own rendering.
        // But for consistency:
        if (tablesData[tableName]) {
            // We'd need to mock a result set if we want to populate resultSets
            // But ER view reads tablesData.
        }

    } catch (e: any) {
        newTab.error = e.toString();
    } finally {
        newTab.isLoading = false;
    }
};

const handleViewERDiagram = () => {
    if (contextMenu.targetTable) {
        openERDiagramTab(contextMenu.targetTable);
        closeContextMenu();
    }
};

const fetchTableColumns = async (tableName: string): Promise<string[]> => {
    if (tableSchemas.value[tableName]) {
        return tableSchemas.value[tableName];
    }

    const type = (props.dbType || '').toLowerCase();
    let query = '';

    if (type.includes('mssql') || type.includes('sqlserver')) {
        query = `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '${tableName}'`;
    } else if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroach')) {
        query = `SELECT column_name FROM information_schema.columns WHERE table_name = '${tableName}'`;
    } else if (type.includes('mysql') || type.includes('maria') || type.includes('databend')) {
        query = `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '${tableName}'`;
    } else if (type.includes('sqlite') || type.includes('libsql')) {
        // SQLite PRAGMA returns a result set we need to parse differently if we use generic ExecuteQuery
        // But let's try standard schema table if available or just PRAGMA
        query = `SELECT name FROM pragma_table_info('${tableName}')`;
    } else {
        return [];
    }

    try {
        const reqId = generateId();
        // unique ID for schema fetch to avoid collision
        const res = await ExecuteQuery(props.connectionId, query, reqId);

        if (!res.error && res.resultSets && res.resultSets.length > 0) {
            const rs = res.resultSets[0];
            const cols = rs.rows.map(row => {
                // Row is array or map depending on driver?
                // ExecuteQuery returns ResultSet which has Rows [][]interface{} and Columns []string
                // Wait, existing code says:
                // const mappedRows = batchRows.map((row: any[]) => ...

                // However, ExecuteQuery (non-stream) return generic result.
                // Let's check App.go ExecuteQuery return type. structure is QueryResult { ResultSets []ResultSet }
                // ResultSet struct in Go is { Columns []string, Rows [][]interface{} }
                // So row is []interface{}.
                if (Array.isArray(row)) {
                    return String(row[0]);
                }
                return String(row);
            });
            tableSchemas.value[tableName] = cols;
            return cols;
        }
    } catch (e) {
        console.warn(`Failed to fetch columns for ${tableName}`, e);
    }
    return [];
};

const {
    aiCopilot,
    aiCopilotModeOptions,
    openAiCopilot,
    setAiCopilotMode,
    runAiCopilot,
    applyAiSqlToEditor,
    analyzeQuery,
    explainWithAI,
} = useQueryAnalysis({
    activeTab,
    connectionId: connectionIdRef,
    connectionName: connectionNameRef,
    dbType: dbTypeRef,
    tables,
    getSelectedQuery: () => workspaceRef.value?.getSelection() || '',
    fetchTableColumns,
    loadHistory: (scope) => GetQueryHistory(scope),
    complete: (messages, options) => completeWithSavedProvider(messages, options),
    onToastError: (message) => {
        toastRef.value?.error(message);
    },
    onToastSuccess: (message) => {
        toastRef.value?.success(message);
    }
});

const beautifyQuery = () => {
    if (!activeTab.value || !activeTab.value.query) return;

    try {
        const type = (props.dbType || '').toLowerCase();
        let language = 'sql';

        if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroach')) language = 'postgresql';
        else if (type.includes('mysql') || type.includes('maria') || type.includes('databend')) language = 'mysql';
        else if (type.includes('mssql') || type.includes('sqlserver')) language = 'transactsql';
        else if (type.includes('sqlite') || type.includes('libsql')) language = 'sqlite';

        activeTab.value.query = format(activeTab.value.query, {
            language: language as any,
            tabWidth: 4,
            keywordCase: 'upper',
            linesBetweenQueries: 2
        });
    } catch (e) {
        console.error("Failed to format query", e);
    }
};

const disconnect = async () => {
    try {
        await DisconnectDB(props.connectionId);
        emit('disconnect', props.connectionId);
    } catch (e) {
        console.error("Failed to disconnect", e);
        emit('disconnect', props.connectionId);
    }
};

// Editing helpers moved to useRecordOperations

const getRowId = (row: any, index: number) => {
    return index;
};

const handleCellClick = (item: any, col: string, rsIndex: number = 0) => {
    if (!isEditable(col)) return;

    let value = item.data[col];
    const editorType = getEditorType(col);

    // Enhanced parsing for dates and datetimes
    if ((editorType === 'datetime-local' || editorType === 'date') && value && typeof value === 'string') {
        try {
            const d = new Date(value);
            if (!isNaN(d.getTime())) {
                const pad = (n: number) => String(n).padStart(2, '0');
                if (editorType === 'datetime-local') {
                    // Local format: YYYY-MM-DDTHH:mm:ss
                    value = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}:${pad(d.getSeconds())}`;
                } else {
                    // Date only format: YYYY-MM-DD
                    value = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
                }
            }
        } catch (e) {
            console.warn("Failed to parse date for editor", value);
        }
    }

    activeTab.value!.editingCell = {
        rowId: item.index,
        col: col,
        value: value,
        resultSetIndex: rsIndex
    };

    nextTick(() => {
        const input = document.getElementById(`edit-input-${rsIndex}-${item.index}-${col}`);
        if (input) (input as HTMLInputElement).focus();
    });
};

const getEditorType = (col: string): string => {
    if (!activeTab.value || !activeTab.value.resultSets || activeTab.value.resultSets.length === 0) return 'text';

    const rs = activeTab.value.resultSets[0];
    if (!rs.columnTypes || rs.columnTypes.length === 0) return 'text';

    const colIndex = rs.columns.indexOf(col);
    if (colIndex === -1) return 'text';

    const meta = rs.columnTypes[colIndex];
    if (!meta) return 'text';

    const type = (meta.type || '').toUpperCase();

    if (type.includes('DATE') && !type.includes('TIME')) return 'date';
    if (type.includes('TIME') || type.includes('TIMESTAMP')) return 'datetime-local';
    if (type.includes('INT') || type.includes('DECIMAL') || type.includes('NUMERIC') || type.includes('FLOAT') || type.includes('REAL') || type.includes('DOUBLE') || type.includes('BIT')) {
        return 'number';
    }

    return 'text';
};

const saveCellEdit = async (item: any, col: string) => {
    if (!activeTab.value || !activeTab.value.editingCell || !activeTab.value.tableName) return;

    const newValue = activeTab.value.editingCell.value;
    const originalValue = item.data[col];

    if (newValue === originalValue) {
        activeTab.value.editingCell = null;
        return;
    }

    // Open Confirmation Modal
    updateConfirmation.value = {
        isOpen: true,
        tableName: activeTab.value.tableName,
        column: col,
        originalValue: originalValue,
        newValue: newValue,
        rowIndex: item.index,
        item: item
    };
};


// Update functions moved to useRecordOperations


const handleKeydown = (e: KeyboardEvent) => {
    // Check if user is typing in an input or textarea
    const target = e.target as HTMLElement;
    const isInput = target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable;

    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        runQuery();
    }
    if ((e.ctrlKey || e.metaKey) && (e.key === 'd' || e.key === 'D')) {
        e.preventDefault();
        if (activeTab.value && activeTab.value.tableName) {
            openDesignView(activeTab.value.tableName);
        }
    }
    if (e.shiftKey && e.altKey && (e.key === 'f' || e.key === 'F')) {
        e.preventDefault();
        beautifyQuery();
    }

    // Datatable Copy Keybindings are now handled natively inside DbQueryResultsPane.vue 
    // to support complex multi-cell drag selections seamlessly.
};

onMounted(async () => {
    try {
        const savedSettingsJson = await LoadSetting('user_settings');
        if (savedSettingsJson) {
            globalSettings.value = JSON.parse(savedSettingsJson);
        }
    } catch (e) {
        console.error("Failed to load global settings in dashboard", e);
    }

    if (props.connectionId) {
        loadTables();
        addTab();
    }
    window.addEventListener('keydown', handleKeydown, true);
    window.addEventListener('click', handleGlobalClick);
    window.addEventListener('open-sql-file', handleOpenSqlFile as EventListener);
    startMonitorTimer();
});

onUnmounted(() => {
    cleanupAllStreams();
    window.removeEventListener('keydown', handleKeydown, true);
    window.removeEventListener('click', handleGlobalClick);
    window.removeEventListener('open-sql-file', handleOpenSqlFile as EventListener);
    stopAllResizing();
    stopMonitorTimer();
});

const handleOpenSqlFile = (e: CustomEvent) => {
    const detail = e.detail;
    if (detail && detail.connectionId === props.connectionId) {
        let targetTab = activeTab.value;
        const isPristine = targetTab &&
            !targetTab.tableName &&
            !targetTab.query &&
            !targetTab.isDesignView &&
            !targetTab.isERView &&
            !targetTab.queryExecuted;

        if (!isPristine) {
            addTab();
            targetTab = activeTab.value;
        }

        if (targetTab) {
            targetTab.name = `File: ${detail.fileName || '.sql'}`;
            targetTab.query = detail.content;

            // Optionally run the query immediately (or let user do it)
            // setTimeout(() => runQuery(), 50); 
        }
    }
};

// Handlers for record operations (exposed for template)
const handleSetNull = () => initiateQuickUpdate(null, contextMenu.targetRow, contextMenu.targetColumn);
const handleSetEmpty = () => initiateQuickUpdate('', contextMenu.targetRow, contextMenu.targetColumn);
const handleSetDefault = () => initiateQuickUpdate({ _quramate_sql_default: true }, contextMenu.targetRow, contextMenu.targetColumn);

const formatValueForDisplay = (val: any) => {
    if (val === null) return 'NULL';
    if (typeof val === 'object' && val._quramate_sql_default) return '<DEFAULT>';
    return String(val);
};

const getColDef = (col: string) => {
    return insertRowModal.value?.columnDefs?.[col] || null;
};

// Watcher for connection changes

watch(() => props.connectionId, (newId) => {
    if (newId) {
        loadTables();
        tabs.value = [];
        addTab();
    }
});

watch(activeTabId, () => {
    // Clear selection when switching tabs to prevent runQuery from picking up previous tab's selection
    // if the editor is shared and selection isn't automatically reset by setValue.
    // We don't have a direct clearSelection on SqlEditor yet, but we could add it.
    // For now, let's just reset the selectedRowIndex to be safe.
    selectedRowIndex.value = null;
    selectedColumn.value = null;
});
</script>

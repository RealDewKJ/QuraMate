<template>
    <div class="flex h-full bg-background text-foreground font-sans">
        <DbSidebar :connection-name="connectionName" :db-type="dbType" :activity-task-count="activityTaskCount"
            :table-search="tableSearch" :view-search="viewSearch" :stored-procedure-search="storedProcedureSearch"
            :function-search="functionSearch" :filtered-tables="filteredTables" :filtered-views="filteredViews"
            :filtered-stored-procedures="filteredStoredProcedures" :filtered-functions="filteredFunctions"
            :open-folders="openFolders" @open-db-context-menu="openDbContextMenu" @open-history="openHistoryPanel"
            @open-activity-monitor="openActivityMonitorTab" @open-ai-copilot="openAiCopilot"
            @open-settings="isSettingsOpen = true" @open-database-info="handleDatabaseInfo" @toggle-folder="toggleFolder"
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

            <DbActivityMonitor v-if="activeTab?.isActivityMonitorView" :activity-task-count="activityTaskCount"
                :monitor-refresh-rate="monitorRefreshRate" :latest-monitor-sample="latestMonitorSample"
                :connection-chart-points="connectionChartPoints" :read-qps-chart-points="readQpsChartPoints"
                :write-qps-chart-points="writeQpsChartPoints" :long-running-total="longRunningTotal"
                :activity-tasks-list="activityTasksList" :format-activity-time="formatActivityTime"
                @close="closeActivityMonitorTab" @kill-all="killAllActivityTasks" @focus-task="openActivityTaskInNewTab"
                @kill-task="killActivityTask" @update:monitor-refresh-rate="monitorRefreshRate = $event" />

            <!-- Query Area -->
            <div v-else-if="activeTab" class="flex flex-col h-full overflow-hidden query-area-container">
                <DbQueryWorkspacePane ref="workspaceRef" :active-tab="activeTab" :tables="tables"
                    :get-columns="fetchTableColumns" :editor-settings="editorSettings" :is-read-only="isReadOnly"
                    @beautify-query="beautifyQuery" @explain-with-ai="explainWithAI" @explain-plan="openExplainPlanTab"
                    @save-plan-baseline="saveExecutionPlanBaseline" @compare-plan-baseline="compareExecutionPlanBaseline"
                    @open-snippets="snippetLibrary.isOpen = true" @save-routine="handleSaveRoutine"
                    @run-query="runQuery" @stop-query="stopQuery" @start-resizing="startResizing" />
                <DbQueryResultsPane ref="resultsPaneRef" v-if="activeTab" :activeTab="activeTab"
                    :isReadOnly="isReadOnly" v-model:selectedRowIndex="selectedRowIndex"
                    v-model:selectedColumn="selectedColumn" v-model:selectedRowData="selectedRowData"
                    :openAiCopilot="openAiCopilot" :getEditorType="getEditorType" :saveCellEdit="saveCellEdit"
                    :toggleSort="toggleSort" :startColumnResize="startColumnResize" :handleCellClick="handleCellClick"
                    :handleRowContextMenu="handleRowContextMenu" :isImageValue="isImageValue"
                    :screenshotShortcutLabel="screenshotShortcutLabel"
                    :showScreenshotShortcutHint="showScreenshotShortcutHint"
                    :exportGridImage="exportQueryResultGridImage"
                    :openImagePreview="openImagePreview"
                    :openMockDataModal="openMockDataModal"
                    :openInsertRowModal="openInsertRowModal"
                    :pasteRowsFromClipboard="pasteRowsFromClipboard"
                    :openHeaderContextMenu="openHeaderContextMenu"
                    :startResultSetResize="startResultSetResize" :getResultSetCardStyle="getResultSetCardStyle" />

                <!-- ER Diagram View -->
                <div v-if="activeTab.isERView" class="flex-1 overflow-hidden bg-background">
                    <ERDiagram :tableName="activeTab.tableName || ''"
                        :columns="(activeTab.tablesData && activeTab.tableName ? activeTab.tablesData[activeTab.tableName] : []) || []"
                        :relationships="activeTab.relationships || []" :tablesData="activeTab.tablesData || {}"
                        :isDark="isDarkTheme" />
                </div>

                <!-- Table Designer View -->
                <KeepAlive>
                    <TableStructureDesigner v-if="activeTab.isDesignView" :key="activeTab.id"
                        :table-name="activeTab.tableName || ''" :connection-id="props.connectionId"
                        :db-type="props.dbType" @close="closeTab(activeTab.id)" @refresh="loadTables"
                        @success="handleDesignerSuccess" />
                </KeepAlive>
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
        <DbDashboardContextMenus :context-menu="contextMenu" :db-type="dbType" :is-read-only="isReadOnly"
            :close-context-menu="closeContextMenu" :open-history="openHistoryPanel"
            :open-activity-monitor="openActivityMonitorTab" :disconnect="disconnect"
            :open-folders="openFolders" :handle-folder-toggle="handleFolderToggle"
            :handle-folder-expand-all="handleFolderExpandAll" :handle-folder-collapse-all="handleFolderCollapseAll"
            :open-transaction-sandbox="openTransactionSandbox"
            :open-schema-compare="openSchemaCompareMigrationPreview"
            :refresh-database="refreshDatabase" :add-tab="addTab" :handle-backup-export="handleBackupExport"
            :handle-generate-database-e-r-diagram="handleGenerateDatabaseERDiagram"
            :handle-database-info="handleDatabaseInfo" :handle-new-table="handleNewTable"
            :handle-new-view="handleNewView" :handle-new-routine="handleNewRoutine"
            :handle-drop-database="handleDropDatabase" :handle-script-routine="handleScriptRoutine"
            :handle-execute-routine="handleExecuteRoutine" :handle-duplicate-routine="handleDuplicateRoutine"
            :handle-delete-routine="handleDeleteRoutine" :handle-folder-refresh="handleFolderRefresh"
            :handle-copy-row="handleCopyRow"
            :handle-copy-row-with-header="handleCopyRowWithHeader" :handle-copy-cell-value="handleCopyCellValue"
            :handle-copy-cell-value-with-header="handleCopyCellValueWithHeader"
            :handle-copy-header-name="handleCopyHeaderName" :handle-copy-header-row="handleCopyHeaderRow"
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
        <DbPastePreviewModal :is-open="!!(pastePreviewModal && pastePreviewModal.isOpen)"
            :table-name="pastePreviewModal?.tableName || ''" :columns="pastePreviewModal?.columns || []"
            :rows="pastePreviewModal?.rows || []" :is-inserting="!!pastePreviewModal?.isInserting"
            :error="pastePreviewModal?.error || ''" :column-defs="pastePreviewModal?.columnDefs || {}"
            :get-input-type="getInputType" :get-number-step="getNumberStep" @close="cancelPastePreview"
            @confirm="confirmPastePreviewInsert" @auto-fix="autoFixPastePreviewRows" @toggle-row="togglePastePreviewRow"
            @update:value="updatePastePreviewValue" />

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
        <DbSchemaCompareWizardModal
            :is-open="schemaCompareWizard.isOpen"
            :db-type="props.dbType"
            :source-name="schemaCompareWizard.sourceName"
            :target-name="schemaCompareWizard.targetName"
            @close="schemaCompareWizard.isOpen = false"
            @confirm="confirmSchemaCompareWizard"
            @update:source-name="schemaCompareWizard.sourceName = $event"
            @update:target-name="schemaCompareWizard.targetName = $event"
        />
        <DbSnippetLibraryModal
            :is-open="snippetLibrary.isOpen"
            :snippets="allSnippetItems"
            :db-type="props.dbType"
            @close="snippetLibrary.isOpen = false"
            @apply="applySnippetToEditor"
            @save-custom="saveCustomSnippet"
            @delete-custom="deleteCustomSnippet"
        />
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
        <SettingsDialog
            :is-open="isSettingsOpen"
            :show-sql-server-settings="props.dbType === 'mssql'"
            @close="isSettingsOpen = false"
            @save="handleSettingsSave"
        />
        <QueryHistory
            :is-open="isHistoryOpen"
            :connection-name="connectionName"
            :history-enabled="queryHistoryEnabled"
            @close="isHistoryOpen = false"
            @run-query="handleRunHistoryQuery"
            @notify-success="toastRef?.success($event)"
            @notify-error="toastRef?.error($event)"
        />

        <DbImagePreviewModal :image-url="imagePreviewUrl" @close="imagePreviewUrl = null" />
        <DbResultImageDialog :is-open="resultImageDialog.isOpen" :image-url="resultImageDialog.imageUrl"
            :file-name="resultImageDialog.fileName" :table-name="resultImageDialog.tableName"
            :timestamp-label="resultImageDialog.timestampLabel" :rendered-rows="resultImageDialog.renderedRows"
            :total-rows="resultImageDialog.totalRows" @close="closeResultImageDialog" @copy="copyResultImage"
            @save="saveResultImage" @share="shareResultImage" @open-new-tab="openResultImageFull"
            @copy-file-name="copyResultImageFileName" />

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
import { ExecuteQuery, DisconnectDB, GetPrimaryKeys, GetForeignKeys, GetRoutineDefinition, ExportTable, GetTables, ImportTable, SelectExportFile, SelectImportFile, CancelQuery, ExecuteTransientQuery, GetTableDefinition, LoadSetting, SaveSetting, GetQueryHistory, WriteTextFile } from '../../wailsjs/go/app/App';
import type { database } from '../../wailsjs/go/models';
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
import DbPastePreviewModal from './dashboard/DbPastePreviewModal.vue';
import DbUpdateConfirmationModal from './dashboard/DbUpdateConfirmationModal.vue';
import DbImagePreviewModal from './dashboard/DbImagePreviewModal.vue';
import DbResultImageDialog from './dashboard/DbResultImageDialog.vue';
import DbAICopilotOverlay from './dashboard/DbAICopilotOverlay.vue';
import DbSchemaCompareWizardModal from './dashboard/DbSchemaCompareWizardModal.vue';
import DbSnippetLibraryModal, { type DashboardSnippetItem } from './dashboard/DbSnippetLibraryModal.vue';
import { TooltipProvider, TooltipRoot, TooltipTrigger, TooltipContent } from 'radix-vue';
import { completeWithSavedProvider } from '../composables/useAiProvider';

// Composables
import { isDarkTheme } from '../composables/useTheme';
import { useTabs } from '../composables/useTabs';
import { useSidebar } from '../composables/useSidebar';
import { useRecordOperations } from '../composables/useRecordOperations';
import { useQueryResultsView } from '../composables/useQueryResultsView';
import { useActivityMonitor } from '../composables/useActivityMonitor';
import type { ActivityTask } from '../composables/useActivityMonitor';
import { useQueryExecution } from '../composables/useQueryExecution';
import { useQueryAnalysis } from '../composables/useQueryAnalysis';
import { useResultSetLayout } from '../composables/useResultSetLayout';
import { useDashboardContextMenus } from '../composables/useDashboardContextMenus';
import { useDatabaseAdminModals } from '../composables/useDatabaseAdminModals';
import { useMockDataModal } from '../composables/useMockDataModal';
import {
    DEFAULT_GRID_SCREENSHOT_SHORTCUT,
    buildResultGridImage,
    copyImageBlobToClipboard,
    downloadBlobAsFile,
    shortcutMatchesEvent
} from '../composables/useResultGridScreenshot';
import { useTableActions } from '../composables/useTableActions';
import { useSchemaVisualizer } from '../composables/useSchemaVisualizer';

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
const schemaCompareWizard = reactive({
    isOpen: false,
    sourceName: '',
    targetName: ''
});
const snippetLibrary = reactive({
    isOpen: false,
    customItems: [] as DashboardSnippetItem[],
});

const globalSettings = ref<any>({});
const safeModeEnabled = computed(() => {
    return globalSettings.value?.general?.enableSafeMode !== false;
});
const perfLoggingEnabled = computed(() => {
    return globalSettings.value?.general?.enablePerfLogs === true;
});
const queryHistoryEnabled = computed(() => {
    return globalSettings.value?.general?.enableQueryHistory !== false;
});
const workspacePersistenceEnabled = computed(() => {
    return globalSettings.value?.general?.persistWorkspaceState === true;
});
const queryHistoryRetentionDays = computed(() => {
    const rawValue = Number(globalSettings.value?.general?.queryHistoryRetentionDays);
    if (!Number.isFinite(rawValue) || rawValue <= 0) {
        return 30;
    }
    return Math.min(3650, Math.max(1, Math.trunc(rawValue)));
});
const showScreenshotShortcutHint = computed(() => true);
const screenshotShortcutLabel = computed(() => DEFAULT_GRID_SCREENSHOT_SHORTCUT);
const builtInSnippetItems = computed<DashboardSnippetItem[]>(() => {
    const normalized = (props.dbType || '').toLowerCase();
    const common: DashboardSnippetItem[] = [
        {
            id: 'ddl-create-index-review',
            title: 'Index Review Checklist',
            description: 'Review missing and duplicate indexes before applying changes.',
            category: 'Maintenance',
            sql: `-- Index review checklist
-- 1) Inspect current indexes
-- 2) Validate predicate / sort coverage
-- 3) Estimate write amplification before rollout`,
            isBuiltIn: true,
        },
        {
            id: 'dml-safe-bulk-update',
            title: 'Safe Bulk Update',
            description: 'Transaction-first DML template with preview and rollback.',
            category: 'DML',
            sql: `BEGIN;

SELECT COUNT(*) AS would_affect_rows
FROM {{table_name}}
WHERE {{predicate}};

UPDATE {{table_name}}
SET {{set_clause}}
WHERE {{predicate}};

ROLLBACK;
-- COMMIT;`,
            isBuiltIn: true,
        },
    ];

    if (normalized.includes('mssql') || normalized.includes('sqlserver')) {
        return common.concat([
            {
                id: 'mssql-lock-scan',
                title: 'Blocking Sessions Snapshot',
                description: 'Inspect waiting sessions and blockers on SQL Server.',
                category: 'Maintenance',
                dbTypes: ['mssql', 'sqlserver'],
                sql: `SELECT
    r.session_id,
    r.blocking_session_id,
    r.status,
    r.wait_type,
    DB_NAME(r.database_id) AS database_name,
    SUBSTRING(t.text, (r.statement_start_offset / 2) + 1,
        ((CASE r.statement_end_offset WHEN -1 THEN DATALENGTH(t.text) ELSE r.statement_end_offset END - r.statement_start_offset) / 2) + 1) AS statement_text
FROM sys.dm_exec_requests r
CROSS APPLY sys.dm_exec_sql_text(r.sql_handle) t
WHERE r.session_id <> @@SPID
  AND (DB_NAME(r.database_id) = '{{database_name}}' OR '{{database_name}}' = '')
ORDER BY r.blocking_session_id DESC, r.session_id;`,
                isBuiltIn: true,
            },
            {
                id: 'mssql-index-fragmentation',
                title: 'Index Fragmentation Review',
                description: 'Inspect fragmentation and page counts before rebuild/reorganize.',
                category: 'DDL',
                dbTypes: ['mssql', 'sqlserver'],
                sql: `SELECT
    OBJECT_NAME(ips.object_id) AS table_name,
    i.name AS index_name,
    ips.avg_fragmentation_in_percent,
    ips.page_count
FROM sys.dm_db_index_physical_stats(DB_ID(), NULL, NULL, NULL, 'LIMITED') ips
JOIN sys.indexes i
  ON ips.object_id = i.object_id AND ips.index_id = i.index_id
WHERE ips.page_count > 100
ORDER BY ips.avg_fragmentation_in_percent DESC;`,
                isBuiltIn: true,
            },
        ]);
    }

    if (normalized.includes('postgres') || normalized.includes('greenplum') || normalized.includes('redshift') || normalized.includes('cockroach')) {
        return common.concat([
            {
                id: 'pg-lock-scan',
                title: 'Blocking Sessions Snapshot',
                description: 'Inspect blockers and blocked sessions on PostgreSQL-family databases.',
                category: 'Maintenance',
                dbTypes: ['postgres', 'greenplum', 'redshift', 'cockroach'],
                sql: `SELECT
    blocked.pid AS blocked_pid,
    blocker.pid AS blocker_pid,
    blocked.query AS blocked_query,
    blocker.query AS blocker_query
FROM pg_stat_activity blocked
JOIN pg_locks blocked_locks ON blocked.pid = blocked_locks.pid
JOIN pg_locks blocker_locks
  ON blocked_locks.locktype = blocker_locks.locktype
 AND blocked_locks.database IS NOT DISTINCT FROM blocker_locks.database
 AND blocked_locks.relation IS NOT DISTINCT FROM blocker_locks.relation
 AND blocked_locks.page IS NOT DISTINCT FROM blocker_locks.page
 AND blocked_locks.tuple IS NOT DISTINCT FROM blocker_locks.tuple
 AND blocked_locks.classid IS NOT DISTINCT FROM blocker_locks.classid
 AND blocked_locks.objid IS NOT DISTINCT FROM blocker_locks.objid
 AND blocked_locks.objsubid IS NOT DISTINCT FROM blocker_locks.objsubid
 AND blocked_locks.pid <> blocker_locks.pid
JOIN pg_stat_activity blocker ON blocker.pid = blocker_locks.pid
WHERE NOT blocked_locks.granted
  AND blocker_locks.granted;`,
                isBuiltIn: true,
            },
        ]);
    }

    return common.concat([
        {
            id: 'mysql-processlist-review',
            title: 'Processlist Review',
            description: 'Inspect active sessions and long-running commands.',
            category: 'Maintenance',
            dbTypes: ['mysql', 'mariadb', 'databend'],
            sql: `SHOW FULL PROCESSLIST;`,
            isBuiltIn: true,
        },
    ]);
});
const allSnippetItems = computed(() => builtInSnippetItems.value.concat(snippetLibrary.customItems));

const resultImageDialog = ref<{
    isOpen: boolean;
    imageUrl: string;
    blob: Blob | null;
    fileName: string;
    tableName: string;
    timestampLabel: string;
    renderedRows: number;
    totalRows: number;
}>({
    isOpen: false,
    imageUrl: '',
    blob: null,
    fileName: '',
    tableName: '',
    timestampLabel: '',
    renderedRows: 0,
    totalRows: 0
});

const loadSnippetLibrary = async () => {
    try {
        const savedSnippetsJson = await LoadSetting(snippetLibraryStorageKey.value);
        if (!savedSnippetsJson) {
            snippetLibrary.customItems = [];
            return;
        }
        const parsed = JSON.parse(savedSnippetsJson);
        snippetLibrary.customItems = Array.isArray(parsed) ? parsed : [];
    } catch (e) {
        console.error('Failed to load snippet library', e);
        snippetLibrary.customItems = [];
    }
};

const openImagePreview = (url: any) => {
    if (url) imagePreviewUrl.value = String(url);
};

const copyTextToClipboard = async (text: string, successMessage: string, errorMessage: string): Promise<boolean> => {
    try {
        await navigator.clipboard.writeText(text);
        toastRef.value?.success(successMessage);
        return true;
    } catch (e) {
        console.error('Clipboard write failed', e);
        toastRef.value?.error(errorMessage);
        return false;
    }
};

const createStringHash = (input: string): string => {
    let hash = 0;
    for (let i = 0; i < input.length; i += 1) {
        hash = ((hash << 5) - hash) + input.charCodeAt(i);
        hash |= 0;
    }
    return Math.abs(hash).toString(16);
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

const clearResultImageDialogPreview = () => {
    if (resultImageDialog.value.imageUrl) {
        URL.revokeObjectURL(resultImageDialog.value.imageUrl);
    }
    resultImageDialog.value.imageUrl = '';
    resultImageDialog.value.blob = null;
};

const closeResultImageDialog = () => {
    resultImageDialog.value.isOpen = false;
    clearResultImageDialogPreview();
};

const generateQueryResultGridImage = async () => {
    const tab = activeTab.value;
    if (!tab || tab.resultViewTab !== 'data' || tab.isERView || tab.isDesignView) {
        throw new Error('Switch to a data result grid before taking a screenshot.');
    }

    const primaryResultSet = tab.resultSets?.[0];
    if (!primaryResultSet || !primaryResultSet.columns || primaryResultSet.columns.length === 0) {
        throw new Error('No query result grid available to export.');
    }

    const tableName = String(tab.tableName || tab.name || 'query_result');
    const image = await buildResultGridImage({
        resultSet: primaryResultSet,
        tableName,
        filters: tab.filters || {},
        sortColumn: tab.sortColumn,
        sortDirection: tab.sortDirection,
        maxRows: 16
    });

    return { image, tableName };
};

const exportQueryResultGridImage = async () => {
    try {
        const { image, tableName } = await generateQueryResultGridImage();

        clearResultImageDialogPreview();
        resultImageDialog.value.isOpen = true;
        resultImageDialog.value.blob = image.blob;
        resultImageDialog.value.fileName = image.fileName;
        resultImageDialog.value.tableName = tableName;
        resultImageDialog.value.timestampLabel = image.timestampLabel;
        resultImageDialog.value.renderedRows = image.renderedRows;
        resultImageDialog.value.totalRows = image.totalRows;
        resultImageDialog.value.imageUrl = URL.createObjectURL(image.blob);
    } catch (e) {
        console.error('Failed to export query result screenshot', e);
        toastRef.value?.error(`Failed to export screenshot: ${e}`);
    }
};

const copyResultImage = async () => {
    if (!resultImageDialog.value.blob) return;
    const copied = await copyImageBlobToClipboard(resultImageDialog.value.blob);
    if (copied) {
        toastRef.value?.success('Image copied to clipboard.');
    } else {
        toastRef.value?.error('Clipboard image copy is not supported on this system.');
    }
};

const saveResultImage = () => {
    if (!resultImageDialog.value.blob) return;
    downloadBlobAsFile(resultImageDialog.value.blob, resultImageDialog.value.fileName || 'query_result.png');
    toastRef.value?.success('Image saved.');
};

const shareResultImage = async () => {
    if (!resultImageDialog.value.blob) return;

    const nav = navigator as Navigator & {
        share?: (data: ShareData) => Promise<void>;
        canShare?: (data?: ShareData) => boolean;
    };
    const file = new File([resultImageDialog.value.blob], resultImageDialog.value.fileName || 'query_result.png', { type: 'image/png' });

    try {
        if (nav.share && nav.canShare && nav.canShare({ files: [file] })) {
            await nav.share({
                title: resultImageDialog.value.tableName || 'Query Result',
                text: `${resultImageDialog.value.tableName || 'Query Result'} • ${resultImageDialog.value.timestampLabel}`,
                files: [file]
            });
            toastRef.value?.success('Image shared.');
            return;
        }
    } catch (e) {
        console.error('Share failed', e);
    }

    const copied = await copyImageBlobToClipboard(resultImageDialog.value.blob);
    if (copied) {
        toastRef.value?.success('Share is unavailable here. Image copied to clipboard instead.');
    } else {
        toastRef.value?.error('Share is unavailable on this system.');
    }
};

const openResultImageFull = () => {
    if (!resultImageDialog.value.imageUrl) return;
    window.open(resultImageDialog.value.imageUrl, '_blank', 'noopener,noreferrer');
};

const copyResultImageFileName = async () => {
    if (!resultImageDialog.value.fileName) return;
    await copyTextToClipboard(resultImageDialog.value.fileName, 'File name copied.', 'Failed to copy file name');
};

// Emits/Props setup
const emit = defineEmits(['disconnect']);

const props = defineProps<{
    connectionId: string;
    connectionName?: string;
    sessionKey?: string;
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
    tabCounter,
    activeTab,
    addTab,
    closeTab,
    generateId
} = useTabs();

const DASHBOARD_TAB_SESSION_VERSION = 1;

interface PersistedDashboardTab {
    id: string;
    name: string;
    tableName?: string;
    query: string;
    resultViewTab: 'data' | 'messages' | 'analysis';
    editorHeight: number;
    columnWidths: Record<string, number>;
    isDesignView?: boolean;
    isERView?: boolean;
    relationships?: any[];
    tablesData?: Record<string, { name: string, type: string }[]>;
    isRoutine?: boolean;
    routineName?: string;
    routineType?: 'PROCEDURE' | 'FUNCTION';
    sqlFilePath?: string;
    isActivityMonitorView?: boolean;
}

interface PersistedDashboardSession {
    version: number;
    activeTabId: string | null;
    tabCounter: number;
    tabs: PersistedDashboardTab[];
}

const sanitizeSessionSegment = (value: string | undefined): string => {
    const sanitized = String(value || 'default').trim().toLowerCase().replace(/[^a-z0-9._-]+/g, '_');
    return sanitized || 'default';
};

const tabSessionStorageKey = computed(() => {
    if (props.sessionKey) {
        return `dashboard_session:${sanitizeSessionSegment(props.sessionKey)}`;
    }

    const dbTypeSegment = sanitizeSessionSegment(props.dbType);
    const connectionSegment = sanitizeSessionSegment(props.connectionName || props.connectionId);
    return `dashboard_session:${dbTypeSegment}:${connectionSegment}`;
});

const planBaselineStorageKey = computed(() => {
    const dbTypeSegment = sanitizeSessionSegment(props.dbType);
    const connectionSegment = sanitizeSessionSegment(props.connectionName || props.connectionId);
    return `dashboard_plan_baselines:${dbTypeSegment}:${connectionSegment}`;
});

const snippetLibraryStorageKey = computed(() => {
    const dbTypeSegment = sanitizeSessionSegment(props.dbType);
    return `dashboard_snippets:${dbTypeSegment}`;
});

interface SavedPlanBaseline {
    fingerprint: string;
    queryHash: string;
    tabName: string;
    capturedAt: string;
    summary: {
        resultSetCount: number;
        totalRows: number;
        columnCounts: number[];
        previewHash: string;
    };
}

const isRestoringTabSession = ref(false);
let tabSessionSaveTimer: ReturnType<typeof setTimeout> | null = null;

const toPersistedDashboardTab = (tab: QueryTab): PersistedDashboardTab => ({
    id: tab.id,
    name: tab.name,
    tableName: tab.tableName,
    query: tab.query || '',
    resultViewTab: tab.resultViewTab || 'data',
    editorHeight: tab.editorHeight || 300,
    columnWidths: tab.columnWidths || {},
    isDesignView: !!tab.isDesignView,
    isERView: !!tab.isERView,
    relationships: tab.relationships,
    tablesData: tab.tablesData,
    isRoutine: !!tab.isRoutine,
    routineName: tab.routineName,
    routineType: tab.routineType,
    sqlFilePath: tab.sqlFilePath,
    isActivityMonitorView: !!tab.isActivityMonitorView,
});

const fromPersistedDashboardTab = (saved: PersistedDashboardTab): QueryTab => {
    const normalizedResultViewTab = saved.resultViewTab === 'messages' || saved.resultViewTab === 'analysis'
        ? saved.resultViewTab
        : 'data';
    const normalizedEditorHeight = typeof saved.editorHeight === 'number' && saved.editorHeight > 120 && saved.editorHeight < 1200
        ? saved.editorHeight
        : 300;

    return {
        id: saved.id || generateId(),
        name: saved.name || 'Query',
        tableName: saved.tableName,
        query: saved.query || '',
        resultSets: markRaw([]),
        primaryKeys: [],
        filters: {},
        sortColumn: undefined,
        sortDirection: null,
        error: '',
        isLoading: false,
        isExplaining: false,
        explanation: undefined,
        isAiExplaining: false,
        aiExplanation: undefined,
        queryExecuted: false,
        activeQueryIds: [],
        resultViewTab: normalizedResultViewTab,
        totalRowCount: undefined,
        isPartialStats: false,
        editorHeight: normalizedEditorHeight,
        columnWidths: saved.columnWidths || {},
        isDesignView: !!saved.isDesignView,
        isERView: !!saved.isERView,
        relationships: saved.relationships || [],
        tablesData: saved.tablesData,
        isRoutine: !!saved.isRoutine,
        routineName: saved.routineName,
        routineType: saved.routineType,
        sqlFilePath: saved.sqlFilePath,
        isActivityMonitorView: !!saved.isActivityMonitorView,
    };
};

const persistTabSession = async () => {
    if (!workspacePersistenceEnabled.value) {
        return;
    }

    const payload: PersistedDashboardSession = {
        version: DASHBOARD_TAB_SESSION_VERSION,
        activeTabId: activeTabId.value,
        tabCounter: tabCounter.value,
        tabs: tabs.value.map(toPersistedDashboardTab),
    };

    try {
        await SaveSetting(tabSessionStorageKey.value, JSON.stringify(payload));
    } catch (e) {
        console.error('Failed to persist tab session', e);
    }
};

const restoreTabSession = async (): Promise<boolean> => {
    if (!workspacePersistenceEnabled.value) {
        return false;
    }

    try {
        const raw = await LoadSetting(tabSessionStorageKey.value);
        if (!raw) {
            return false;
        }

        const parsed = JSON.parse(raw) as Partial<PersistedDashboardSession>;
        if (parsed.version !== DASHBOARD_TAB_SESSION_VERSION || !Array.isArray(parsed.tabs) || parsed.tabs.length === 0) {
            return false;
        }

        isRestoringTabSession.value = true;

        const restoredTabs = parsed.tabs
            .filter((tab): tab is PersistedDashboardTab => !!tab && typeof tab.id === 'string')
            .filter((tab) => !tab.isActivityMonitorView)
            .map((tab) => fromPersistedDashboardTab(tab));

        if (restoredTabs.length === 0) {
            return false;
        }

        tabs.value = restoredTabs;
        const maxQueryNumber = restoredTabs.reduce((maxValue, tab) => {
            const match = tab.name.match(/^Query\s+(\d+)$/i);
            if (!match) return maxValue;
            const current = Number(match[1]);
            return Number.isNaN(current) ? maxValue : Math.max(maxValue, current);
        }, 0);

        const parsedCounter = typeof parsed.tabCounter === 'number' ? parsed.tabCounter : 0;
        tabCounter.value = Math.max(parsedCounter, maxQueryNumber, restoredTabs.length);

        if (parsed.activeTabId && restoredTabs.some(tab => tab.id === parsed.activeTabId)) {
            activeTabId.value = parsed.activeTabId;
        } else {
            activeTabId.value = restoredTabs[restoredTabs.length - 1].id;
        }

        return true;
    } catch (e) {
        console.error('Failed to restore tab session', e);
        return false;
    } finally {
        isRestoringTabSession.value = false;
    }
};

const schedulePersistTabSession = () => {
    if (!workspacePersistenceEnabled.value) return;
    if (isRestoringTabSession.value) return;
    if (tabSessionSaveTimer) {
        clearTimeout(tabSessionSaveTimer);
    }

    tabSessionSaveTimer = setTimeout(() => {
        tabSessionSaveTimer = null;
        void persistTabSession();
    }, 450);
};

watch([tabs, activeTabId, tabCounter], schedulePersistTabSession, { deep: true });
watch(workspacePersistenceEnabled, (enabled) => {
    if (!enabled && tabSessionSaveTimer) {
        clearTimeout(tabSessionSaveTimer);
        tabSessionSaveTimer = null;
    }
});

const initializeDashboardState = async () => {
    if (!props.connectionId) return;
    loadTables();

    const restored = await restoreTabSession();
    if (!restored) {
        tabs.value = [];
        addTab();
    }
};
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

const openActivityMonitorTab = () => {
    const existingTab = tabs.value.find((tab) => tab.isActivityMonitorView);
    if (existingTab) {
        activeTabId.value = existingTab.id;
        return;
    }

    const newTabId = addTab();
    const monitorTab = tabs.value.find((tab) => tab.id === newTabId);
    if (!monitorTab) {
        return;
    }

    monitorTab.name = 'Activity Monitor';
    monitorTab.query = '';
    monitorTab.error = '';
    monitorTab.resultSets = markRaw([]);
    monitorTab.queryExecuted = false;
    monitorTab.isActivityMonitorView = true;
};

const openHistoryPanel = () => {
    isHistoryOpen.value = true;
};

const closeActivityMonitorTab = () => {
    const currentTab = activeTab.value;
    if (currentTab?.isActivityMonitorView) {
        closeTab(currentTab.id);
        return;
    }

    const monitorTab = tabs.value.find((tab) => tab.isActivityMonitorView);
    if (monitorTab) {
        closeTab(monitorTab.id);
    }
};

const openActivityTaskInNewTab = (task: ActivityTask) => {
    const newTabId = addTab();
    const targetTab = tabs.value.find((tab) => tab.id === newTabId);
    if (!targetTab) {
        return;
    }

    const shortTaskId = task.id?.slice(0, 8) || 'task';
    targetTab.name = `Activity ${shortTaskId}`;
    targetTab.query = task.query || '';
    targetTab.queryExecuted = false;
    targetTab.error = '';
    targetTab.resultSets = markRaw([]);
    targetTab.totalRowCount = undefined;
    targetTab.isPartialStats = false;
    targetTab.resultViewTab = 'data';
};

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
    queryHistoryEnabled,
    queryHistoryRetentionDays,
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
    pasteRowsFromClipboard,
    pastePreviewModal,
    updatePastePreviewValue,
    togglePastePreviewRow,
    autoFixPastePreviewRows,
    confirmPastePreviewInsert,
    cancelPastePreview,
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
    openHeaderContextMenu,
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

const { generateDatabaseERDiagram } = useSchemaVisualizer({
    connectionId: props.connectionId,
    dbType: props.dbType,
    executeQuery: ExecuteQuery,
    getTables: GetTables,
    getForeignKeys: GetForeignKeys,
    generateId,
});

const handleGenerateDatabaseERDiagram = async () => {
    closeContextMenu();
    addTab();
    if (!activeTab.value) {
        return;
    }

    await generateDatabaseERDiagram(activeTab.value);
};

// Redundant handleNewView and handleNewRoutine removed as they are defined later in the file.

const getFileNameFromPath = (filePath: string): string => {
    if (!filePath) return 'query.sql';
    const parts = filePath.split(/[/\\]/);
    return parts[parts.length - 1] || 'query.sql';
};

const getDefaultSqlFileName = (tab: QueryTab): string => {
    const baseName = (tab.name || 'query')
        .replace(/^File:\s*/i, '')
        .replace(/[^a-zA-Z0-9._-]+/g, '_')
        .replace(/^_+|_+$/g, '');
    const normalized = baseName || 'query';
    return normalized.toLowerCase().endsWith('.sql') ? normalized : `${normalized}.sql`;
};

const saveCurrentQueryToFile = async (saveAs: boolean = false) => {
    if (!activeTab.value || activeTab.value.isERView || activeTab.value.isDesignView) {
        return;
    }

    if (activeTab.value.isRoutine) {
        await handleSaveRoutine();
        return;
    }

    let targetPath = saveAs ? '' : (activeTab.value.sqlFilePath || '');
    if (!targetPath) {
        targetPath = await SelectExportFile(getDefaultSqlFileName(activeTab.value));
        if (!targetPath) {
            return;
        }
    }

    const result = await WriteTextFile(targetPath, activeTab.value.query || '');
    if (result) {
        toastRef.value?.error(`Failed to save SQL file: ${result}`);
        return;
    }

    activeTab.value.sqlFilePath = targetPath;
    const fileName = getFileNameFromPath(targetPath);
    activeTab.value.name = `File: ${fileName}`;
    toastRef.value?.success(`Saved SQL file: ${fileName}`);
};
const refreshDatabase = async (showToast: boolean = true) => {
    closeContextMenu();
    await loadTables();
    if (showToast && toastRef.value) {
        toastRef.value.success('Database refreshed successfully.');
    }
};

const refreshCurrentContext = async () => {
    if (!activeTab.value || activeTab.value.isERView || activeTab.value.isDesignView) {
        await refreshDatabase();
        return;
    }

    if (activeTab.value.isLoading) {
        return;
    }

    if (activeTab.value.query && activeTab.value.query.trim().length > 0) {
        await runQuery(true);
        return;
    }

    await refreshDatabase();
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

const handleFolderToggle = () => {
    closeContextMenu();
    const folder = contextMenu.targetFolder;
    if (!folder) return;
    if (openFolders.value.includes(folder)) {
        openFolders.value = openFolders.value.filter(f => f !== folder);
        return;
    }
    openFolders.value = [...openFolders.value, folder];
};

const handleFolderExpandAll = () => {
    closeContextMenu();
    const folder = contextMenu.targetFolder;
    if (!folder) return;

    const toOpen = new Set(openFolders.value);
    toOpen.add(folder);
    if (folder === 'Programmability') {
        toOpen.add('Stored Procedures');
        toOpen.add('Functions');
    }
    openFolders.value = Array.from(toOpen);
};

const handleFolderCollapseAll = () => {
    closeContextMenu();
    const folder = contextMenu.targetFolder;
    if (!folder) return;

    if (folder === 'Programmability') {
        openFolders.value = openFolders.value.filter(
            f => f !== 'Programmability' && f !== 'Stored Procedures' && f !== 'Functions'
        );
        return;
    }
    openFolders.value = openFolders.value.filter(f => f !== folder);
};

const handleCopyRow = async () => {
    // If the right-clicked row is part of our multi-selection, we'll just copy the entire selection
    const isMultiSelected = Array.isArray(selectedRowIndex.value)
        ? (contextMenu.targetRowIndex !== null && selectedRowIndex.value.includes(contextMenu.targetRowIndex))
        : false;

    if (isMultiSelected) {
        await copySelectedRow(false);
        closeContextMenu();
        return;
    }

    if (contextMenu.targetRow) {
        const values = Object.values(contextMenu.targetRow).map(v => v === null ? 'NULL' : String(v)).join('\t');
        await copyTextToClipboard(values, 'Row copied to clipboard', 'Failed to copy row to clipboard');
        closeContextMenu();
    }
};

const handleCopyCellValue = async () => {
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
        await copyTextToClipboard(str, 'Cell value copied to clipboard', 'Failed to copy cell value');
        closeContextMenu();
    }
};

const handleCopyRowWithHeader = async () => {
    const isMultiSelected = Array.isArray(selectedRowIndex.value)
        ? (contextMenu.targetRowIndex !== null && selectedRowIndex.value.includes(contextMenu.targetRowIndex))
        : false;

    if (isMultiSelected) {
        await copySelectedRow(true);
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

        await copyTextToClipboard(`${headerLine}\n${valueLine}`, 'Row with header copied to clipboard', 'Failed to copy row with header');
        closeContextMenu();
    }
};

const handleCopyCellValueWithHeader = async () => {
    if (contextMenu.targetRow && contextMenu.targetColumn) {
        const col = contextMenu.targetColumn;
        const val = contextMenu.targetRow[col];
        const str = val === null ? 'NULL' : String(val);
        await copyTextToClipboard(`${col}: ${str}`, 'Cell value with header copied to clipboard', 'Failed to copy cell value with header');
        closeContextMenu();
    }
};

const handleCopyHeaderName = async () => {
    if (!contextMenu.targetColumn) {
        return;
    }

    await copyTextToClipboard(
        contextMenu.targetColumn,
        'Column header copied to clipboard',
        'Failed to copy column header'
    );
    closeContextMenu();
};

const handleCopyHeaderRow = async () => {
    if (!activeTab.value || contextMenu.targetResultSetIndex === null) {
        return;
    }

    const resultSet = activeTab.value.resultSets?.[contextMenu.targetResultSetIndex];
    const columns = resultSet?.columns;
    if (!Array.isArray(columns) || columns.length === 0) {
        return;
    }

    await copyTextToClipboard(
        columns.join('\t'),
        'Header row copied to clipboard',
        'Failed to copy header row'
    );
    closeContextMenu();
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

const copySelectedCell = async () => {
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
            await copyTextToClipboard(str, 'Cell value copied to clipboard', 'Failed to copy cell value');
        }
    }
};

const copySelectedRow = async (withHeader: boolean = false) => {
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
                await copyTextToClipboard(`${headersStr}\n${rowsStrs.join('\n')}`, `${rowsStrs.length} row(s) with header copied to clipboard`, 'Failed to copy rows with header');
            } else {
                await copyTextToClipboard(rowsStrs.join('\n'), `${rowsStrs.length} row(s) copied to clipboard`, 'Failed to copy rows');
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

        const escapedTableName = getEscapedTableName(tableName);

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

    const escapedTableName = getEscapedTableName(tableName);
    const countQuery = `SELECT COUNT(*) FROM ${escapedTableName}`;

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
        const escapedViewName = getEscapedTableName(viewName);
        activeTab.value.tableName = viewName;
        activeTab.value.name = viewName;

        if (type.includes('mssql') || type.includes('sqlserver')) {
            activeTab.value.query = `SELECT TOP 100 * FROM ${escapedViewName}`;
        } else {
            activeTab.value.query = `SELECT * FROM ${escapedViewName} LIMIT 100`;
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
            const newDefinition = definition.split(routine).join(copyName);

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
    if (props.isReadOnly) {
        toastRef.value?.error('This connection is read-only.');
        closeContextMenu();
        return;
    }
    if (type !== 'PROCEDURE' && type !== 'FUNCTION') {
        toastRef.value?.error('Unsupported routine type.');
        closeContextMenu();
        return;
    }

    const confirmationInput = window.prompt(`Type "${routine}" to confirm dropping this ${type.toLowerCase()}.`);
    if (confirmationInput === null) {
        closeContextMenu();
        return;
    }
    if (confirmationInput.trim() !== routine) {
        toastRef.value?.error('Routine name does not match. Drop cancelled.');
        closeContextMenu();
        return;
    }

    closeContextMenu();

    const dropSql = `DROP ${type} ${getEscapedTableName(routine)}`;
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
        const escapedRoutineIdentifier = getEscapedTableName(routine);
        const escapedRoutineLiteral = escapeSqlLiteral(routine);
        activeTab.value.name = `Script: ${routine}`;
        activeTab.value.query = `-- Scripting for ${routine} (${contextMenu.targetRoutineType})
-- Note: Provide a backend method 'GetRoutineDefinition' for better support.

-- Postgres:
-- SELECT pg_get_functiondef('${escapedRoutineLiteral}'::regproc);

-- MySQL:
-- SHOW CREATE PROCEDURE ${escapedRoutineIdentifier};

-- MSSQL:
-- sp_helptext '${escapedRoutineLiteral}';
`;

        // Try to be smart for MSSQL at least
        const type = (props.dbType || '').toLowerCase();
        if (type.includes('mssql')) {
            activeTab.value.query = `EXEC sp_helptext '${escapedRoutineLiteral}'`;
            setTimeout(() => runQuery(), 50);
        } else if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroach')) {
            activeTab.value.query = `SELECT pg_get_functiondef('${escapedRoutineLiteral}'::regproc)`;
            // This might fail if schema is needed or not in search path, but good attempt
        } else if (type.includes('mysql') || type.includes('maria') || type.includes('databend')) {
            activeTab.value.query = `SHOW CREATE ${contextMenu.targetRoutineType} ${escapedRoutineIdentifier}`;
            setTimeout(() => runQuery(), 50);
        } else if (type.includes('sqlite') || type.includes('libsql')) {
            activeTab.value.query = `SELECT sql FROM sqlite_master WHERE name = '${escapedRoutineLiteral}'`;
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

const escapeSqlLiteral = (value: string) => value.replace(/'/g, "''");

const normalizeNameSegment = (segment: string): string => {
    const trimmed = segment.trim();
    if (trimmed.startsWith('[') && trimmed.endsWith(']')) {
        return trimmed.slice(1, -1).replace(/]]/g, ']');
    }
    if ((trimmed.startsWith('"') && trimmed.endsWith('"')) || (trimmed.startsWith('`') && trimmed.endsWith('`'))) {
        return trimmed.slice(1, -1);
    }
    return trimmed;
};

const splitQualifiedName = (rawName: string): { schema?: string; name: string } => {
    const input = String(rawName || '').trim();
    if (!input.includes('.')) {
        return { name: normalizeNameSegment(input) };
    }

    const parts = input.split('.');
    const name = normalizeNameSegment(parts.pop() || input);
    const schema = normalizeNameSegment(parts.join('.'));
    return schema ? { schema, name } : { name };
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

const buildExplainPlanSql = (sql: string) => {
    const trimmedSql = sql.trim().replace(/;+\s*$/, '');
    const type = (props.dbType || '').toLowerCase();
    if (type.includes('mssql') || type.includes('sqlserver')) {
        return `SET SHOWPLAN_TEXT ON;
${trimmedSql};
SET SHOWPLAN_TEXT OFF;`;
    }
    if (type.includes('mysql') || type.includes('mariadb') || type.includes('databend')) {
        return `EXPLAIN FORMAT=JSON
${trimmedSql};`;
    }
    return `EXPLAIN (FORMAT JSON, ANALYZE FALSE, VERBOSE TRUE)
${trimmedSql};`;
};

const openExplainPlanTab = () => {
    const sourceSql = (workspaceRef.value?.getSelection?.() || activeTab.value?.query || '').trim();
    if (!sourceSql) {
        toastRef.value?.error('Write or select a query first.');
        return;
    }

    openSqlTemplateTab('Execution Plan', buildExplainPlanSql(sourceSql));
};

const summarizeCurrentPlanResult = (): SavedPlanBaseline | null => {
    const tab = activeTab.value;
    if (!tab || !tab.resultSets || tab.resultSets.length === 0) {
        return null;
    }

    const resultSets = tab.resultSets.map((resultSet) => ({
        columns: resultSet.columns,
        rows: (resultSet.rows || []).slice(0, 10),
    }));
    const totalRows = tab.resultSets.reduce((sum, resultSet) => sum + (resultSet.rows?.length || 0), 0);
    const serialized = JSON.stringify(resultSets);

    return {
        fingerprint: `${sanitizeSessionSegment(props.dbType)}:${createStringHash((tab.query || '').trim())}`,
        queryHash: createStringHash((tab.query || '').trim()),
        tabName: tab.name,
        capturedAt: new Date().toISOString(),
        summary: {
            resultSetCount: tab.resultSets.length,
            totalRows,
            columnCounts: tab.resultSets.map((resultSet) => resultSet.columns?.length || 0),
            previewHash: createStringHash(serialized),
        },
    };
};

const saveExecutionPlanBaseline = async () => {
    const baseline = summarizeCurrentPlanResult();
    if (!baseline) {
        toastRef.value?.error('Run an execution plan query first.');
        return;
    }

    try {
        await SaveSetting(planBaselineStorageKey.value, JSON.stringify(baseline));
        toastRef.value?.success('Plan baseline saved.');
    } catch (e) {
        console.error('Failed to save plan baseline', e);
        toastRef.value?.error('Failed to save plan baseline.');
    }
};

const compareExecutionPlanBaseline = async () => {
    const current = summarizeCurrentPlanResult();
    if (!current || !activeTab.value) {
        toastRef.value?.error('Run an execution plan query first.');
        return;
    }

    try {
        const raw = await LoadSetting(planBaselineStorageKey.value);
        if (!raw) {
            toastRef.value?.error('No saved baseline found for this connection.');
            return;
        }

        const saved = JSON.parse(raw) as SavedPlanBaseline;
        const lines = [
            `Baseline captured: ${saved.capturedAt}`,
            `Baseline tab: ${saved.tabName}`,
            `Current tab: ${current.tabName}`,
            '',
            `Result sets: ${saved.summary.resultSetCount} -> ${current.summary.resultSetCount}`,
            `Total rows (preview): ${saved.summary.totalRows} -> ${current.summary.totalRows}`,
            `Column counts: ${saved.summary.columnCounts.join(', ')} -> ${current.summary.columnCounts.join(', ')}`,
            `Preview hash changed: ${saved.summary.previewHash === current.summary.previewHash ? 'No' : 'Yes'}`,
            `Query hash changed: ${saved.queryHash === current.queryHash ? 'No' : 'Yes'}`,
        ];

        activeTab.value.explanation = lines.join('\n');
        activeTab.value.resultViewTab = 'analysis';
        toastRef.value?.success('Baseline comparison ready in Analysis tab.');
    } catch (e) {
        console.error('Failed to compare plan baseline', e);
        toastRef.value?.error('Failed to compare plan baseline.');
    }
};

const applySnippetToEditor = (payload: { snippet: DashboardSnippetItem; resolvedSql: string }) => {
    snippetLibrary.isOpen = false;
    if (!activeTab.value || activeTab.value.isERView || activeTab.value.isDesignView) {
        addTab();
    }
    if (!activeTab.value) {
        return;
    }
    activeTab.value.query = payload.resolvedSql;
    activeTab.value.name = payload.snippet.title;
    activeTab.value.queryExecuted = false;
    activeTab.value.resultSets = markRaw([]);
    activeTab.value.error = '';
};

const saveCustomSnippet = async (payload: { title: string; description: string; sql: string; category: string }) => {
    const item: DashboardSnippetItem = {
        id: `custom-${Date.now()}`,
        title: payload.title,
        description: payload.description,
        sql: payload.sql,
        category: payload.category,
        isBuiltIn: false,
    };
    snippetLibrary.customItems = [...snippetLibrary.customItems, item];
    try {
        await SaveSetting(snippetLibraryStorageKey.value, JSON.stringify(snippetLibrary.customItems));
        toastRef.value?.success('Runbook saved.');
    } catch (e) {
        console.error('Failed to save snippet', e);
        toastRef.value?.error('Failed to save runbook.');
    }
};

const deleteCustomSnippet = async (snippetId: string) => {
    snippetLibrary.customItems = snippetLibrary.customItems.filter((item) => item.id !== snippetId);
    try {
        await SaveSetting(snippetLibraryStorageKey.value, JSON.stringify(snippetLibrary.customItems));
        toastRef.value?.success('Runbook deleted.');
    } catch (e) {
        console.error('Failed to delete snippet', e);
        toastRef.value?.error('Failed to delete runbook.');
    }
};

const splitSimpleSqlStatements = (sql: string): string[] => {
    return (sql || '')
        .split(';')
        .map(part => part.trim())
        .filter(part => part.length > 0);
};

const inferAffectedRowsPreview = (statement: string): string | null => {
    const trimmed = statement.trim().replace(/;$/, '');

    const updateMatch = trimmed.match(/^update\s+([^\s]+)\s+set[\s\S]*?\bwhere\b\s+([\s\S]+)$/i);
    if (updateMatch) {
        return `SELECT COUNT(*) AS would_affect_rows FROM ${updateMatch[1]} WHERE ${updateMatch[2]};`;
    }

    const deleteMatch = trimmed.match(/^delete\s+from\s+([^\s]+)\s+where\s+([\s\S]+)$/i);
    if (deleteMatch) {
        return `SELECT COUNT(*) AS would_affect_rows FROM ${deleteMatch[1]} WHERE ${deleteMatch[2]};`;
    }

    return null;
};

const getTransactionKeywords = () => {
    const type = (props.dbType || '').toLowerCase();
    if (type.includes('mssql') || type.includes('sqlserver')) {
        return {
            begin: 'BEGIN TRANSACTION;',
            commit: 'COMMIT TRANSACTION;',
            rollback: 'ROLLBACK TRANSACTION;'
        };
    }
    return {
        begin: 'BEGIN;',
        commit: 'COMMIT;',
        rollback: 'ROLLBACK;'
    };
};

const openTransactionSandbox = () => {
    if (props.isReadOnly) {
        toastRef.value?.error('This connection is read-only.');
        return;
    }

    const selectedSql = workspaceRef.value?.getSelection?.() || '';
    const baseSql = (selectedSql || activeTab.value?.query || '').trim();
    const statements = splitSimpleSqlStatements(baseSql);
    const previewQueries = statements
        .map((statement) => inferAffectedRowsPreview(statement))
        .filter((sql): sql is string => !!sql);
    const tx = getTransactionKeywords();

    const previewBlock = previewQueries.length > 0
        ? previewQueries.join('\n')
        : '-- Add UPDATE/DELETE statements with WHERE clauses to see affected-row preview queries.';

    const mutationBlock = statements.length > 0
        ? statements.map(statement => `${statement.replace(/;$/, '')};`).join('\n')
        : '-- Example:\n-- UPDATE your_table SET column_name = value WHERE id = 1;';

    const sql = `-- Transaction Sandbox
-- Steps:
-- 1) Run the preview block to inspect impact.
-- 2) Review the mutation block.
-- 3) Keep ${tx.rollback} for safe dry-run, or switch to ${tx.commit} after verification.
${tx.begin}

-- Preview affected rows
${previewBlock}

-- Mutation block
${mutationBlock}

-- Finalize
${tx.rollback}
-- ${tx.commit}`;

    openSqlTemplateTab('Transaction Sandbox', sql);
};

const getSchemaCompareDefaults = () => {
    const type = (props.dbType || '').toLowerCase();
    if (type.includes('mssql') || type.includes('sqlserver')) {
        return { sourceName: 'dbo', targetName: 'stg' };
    }
    if (type.includes('mysql') || type.includes('mariadb') || type.includes('databend')) {
        return { sourceName: 'app_source', targetName: 'app_target' };
    }
    return { sourceName: 'public', targetName: 'staging' };
};

const buildSchemaCompareMigrationSql = (sourceName: string, targetName: string) => {
    const type = (props.dbType || '').toLowerCase();
    const sourceLiteral = escapeSqlLiteral(sourceName);
    const targetLiteral = escapeSqlLiteral(targetName);
    let sql = '';

    if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroach')) {
        sql = `-- Schema Compare + Migration Preview (PostgreSQL family)
-- Set schemas before running:
-- source_schema: desired schema
-- target_schema: schema to be updated
WITH params AS (
  SELECT '${sourceLiteral}'::text AS source_schema, '${targetLiteral}'::text AS target_schema
),
source_cols AS (
  SELECT
    c.table_name,
    c.column_name,
    c.data_type,
    c.is_nullable,
    c.character_maximum_length,
    c.numeric_precision,
    c.numeric_scale,
    c.datetime_precision,
    c.column_default,
    CASE
      WHEN c.data_type IN ('character varying', 'character') AND c.character_maximum_length IS NOT NULL
        THEN c.data_type || '(' || c.character_maximum_length || ')'
      WHEN c.data_type IN ('numeric', 'decimal') AND c.numeric_precision IS NOT NULL
        THEN c.data_type || '(' || c.numeric_precision || COALESCE(',' || c.numeric_scale, '') || ')'
      WHEN c.data_type IN ('time without time zone', 'time with time zone', 'timestamp without time zone', 'timestamp with time zone') AND c.datetime_precision IS NOT NULL
        THEN c.data_type || '(' || c.datetime_precision || ')'
      ELSE c.data_type
    END AS type_declaration
  FROM information_schema.columns c, params p
  WHERE c.table_schema = p.source_schema
),
target_cols AS (
  SELECT
    c.table_name,
    c.column_name,
    c.data_type,
    c.is_nullable,
    c.character_maximum_length,
    c.numeric_precision,
    c.numeric_scale,
    c.datetime_precision,
    c.column_default,
    CASE
      WHEN c.data_type IN ('character varying', 'character') AND c.character_maximum_length IS NOT NULL
        THEN c.data_type || '(' || c.character_maximum_length || ')'
      WHEN c.data_type IN ('numeric', 'decimal') AND c.numeric_precision IS NOT NULL
        THEN c.data_type || '(' || c.numeric_precision || COALESCE(',' || c.numeric_scale, '') || ')'
      WHEN c.data_type IN ('time without time zone', 'time with time zone', 'timestamp without time zone', 'timestamp with time zone') AND c.datetime_precision IS NOT NULL
        THEN c.data_type || '(' || c.datetime_precision || ')'
      ELSE c.data_type
    END AS type_declaration
  FROM information_schema.columns c, params p
  WHERE c.table_schema = p.target_schema
)
SELECT
  COALESCE(s.table_name, t.table_name) AS table_name,
  COALESCE(s.column_name, t.column_name) AS column_name,
  CASE
    WHEN s.column_name IS NULL THEN 'ONLY_IN_TARGET'
    WHEN t.column_name IS NULL THEN 'MISSING_IN_TARGET'
    WHEN s.type_declaration <> t.type_declaration
      OR s.is_nullable <> t.is_nullable
      OR COALESCE(s.column_default, '') <> COALESCE(t.column_default, '')
      OR COALESCE(s.character_maximum_length, -1) <> COALESCE(t.character_maximum_length, -1)
      OR COALESCE(s.numeric_precision, -1) <> COALESCE(t.numeric_precision, -1)
      OR COALESCE(s.numeric_scale, -1) <> COALESCE(t.numeric_scale, -1)
      OR COALESCE(s.datetime_precision, -1) <> COALESCE(t.datetime_precision, -1)
    THEN 'DIFF'
    ELSE 'MATCH'
  END AS status,
  s.type_declaration AS source_type,
  t.type_declaration AS target_type,
  s.is_nullable AS source_nullable,
  t.is_nullable AS target_nullable,
  s.column_default AS source_default,
  t.column_default AS target_default
FROM source_cols s
FULL OUTER JOIN target_cols t
  ON s.table_name = t.table_name AND s.column_name = t.column_name
WHERE s.column_name IS NULL
   OR t.column_name IS NULL
   OR s.type_declaration <> t.type_declaration
   OR s.is_nullable <> t.is_nullable
   OR COALESCE(s.column_default, '') <> COALESCE(t.column_default, '')
   OR COALESCE(s.character_maximum_length, -1) <> COALESCE(t.character_maximum_length, -1)
   OR COALESCE(s.numeric_precision, -1) <> COALESCE(t.numeric_precision, -1)
   OR COALESCE(s.numeric_scale, -1) <> COALESCE(t.numeric_scale, -1)
   OR COALESCE(s.datetime_precision, -1) <> COALESCE(t.datetime_precision, -1)
ORDER BY 1, 2;

-- Migration preview: add missing columns to target schema
WITH params AS (
  SELECT '${sourceLiteral}'::text AS source_schema, '${targetLiteral}'::text AS target_schema
)
SELECT
  'ALTER TABLE "' || p.target_schema || '"."' || s.table_name || '" ADD COLUMN "' || s.column_name || '" ' ||
  CASE
    WHEN s.data_type IN ('character varying', 'character') AND s.character_maximum_length IS NOT NULL
      THEN s.data_type || '(' || s.character_maximum_length || ')'
    WHEN s.data_type IN ('numeric', 'decimal') AND s.numeric_precision IS NOT NULL
      THEN s.data_type || '(' || s.numeric_precision || COALESCE(',' || s.numeric_scale, '') || ')'
    WHEN s.data_type IN ('time without time zone', 'time with time zone', 'timestamp without time zone', 'timestamp with time zone') AND s.datetime_precision IS NOT NULL
      THEN s.data_type || '(' || s.datetime_precision || ')'
    ELSE s.data_type
  END ||
  CASE WHEN s.column_default IS NOT NULL THEN ' DEFAULT ' || s.column_default ELSE '' END ||
  CASE WHEN s.is_nullable = 'NO' THEN ' NOT NULL;' ELSE ';' END AS migration_sql
FROM information_schema.columns s, params p
LEFT JOIN information_schema.columns t
  ON t.table_schema = p.target_schema
 AND t.table_name = s.table_name
 AND t.column_name = s.column_name
WHERE s.table_schema = p.source_schema
  AND t.column_name IS NULL
ORDER BY s.table_name, s.ordinal_position;`;
    } else if (type.includes('mssql') || type.includes('sqlserver')) {
        sql = `-- Schema Compare + Migration Preview (SQL Server)
-- Compare source and target schemas
WITH params AS (
    SELECT N'${sourceLiteral}' AS source_schema, N'${targetLiteral}' AS target_schema
),
source_cols AS (
    SELECT
      TABLE_NAME,
      COLUMN_NAME,
      DATA_TYPE,
      IS_NULLABLE,
      CHARACTER_MAXIMUM_LENGTH,
      NUMERIC_PRECISION,
      NUMERIC_SCALE,
      DATETIME_PRECISION,
      COLUMN_DEFAULT,
      CASE
        WHEN DATA_TYPE IN ('varchar','nvarchar','char','nchar','binary','varbinary')
          THEN DATA_TYPE + '(' + CASE WHEN CHARACTER_MAXIMUM_LENGTH = -1 THEN 'max' ELSE CAST(CHARACTER_MAXIMUM_LENGTH AS VARCHAR(20)) END + ')'
        WHEN DATA_TYPE IN ('decimal','numeric')
          THEN DATA_TYPE + '(' + CAST(NUMERIC_PRECISION AS VARCHAR(20)) + ',' + CAST(NUMERIC_SCALE AS VARCHAR(20)) + ')'
        WHEN DATA_TYPE IN ('datetime2','datetimeoffset','time')
          THEN DATA_TYPE + '(' + CAST(COALESCE(DATETIME_PRECISION, 7) AS VARCHAR(20)) + ')'
        ELSE DATA_TYPE
      END AS TYPE_DECLARATION
    FROM INFORMATION_SCHEMA.COLUMNS
    CROSS JOIN params p
    WHERE TABLE_SCHEMA = p.source_schema
),
target_cols AS (
    SELECT
      TABLE_NAME,
      COLUMN_NAME,
      DATA_TYPE,
      IS_NULLABLE,
      CHARACTER_MAXIMUM_LENGTH,
      NUMERIC_PRECISION,
      NUMERIC_SCALE,
      DATETIME_PRECISION,
      COLUMN_DEFAULT,
      CASE
        WHEN DATA_TYPE IN ('varchar','nvarchar','char','nchar','binary','varbinary')
          THEN DATA_TYPE + '(' + CASE WHEN CHARACTER_MAXIMUM_LENGTH = -1 THEN 'max' ELSE CAST(CHARACTER_MAXIMUM_LENGTH AS VARCHAR(20)) END + ')'
        WHEN DATA_TYPE IN ('decimal','numeric')
          THEN DATA_TYPE + '(' + CAST(NUMERIC_PRECISION AS VARCHAR(20)) + ',' + CAST(NUMERIC_SCALE AS VARCHAR(20)) + ')'
        WHEN DATA_TYPE IN ('datetime2','datetimeoffset','time')
          THEN DATA_TYPE + '(' + CAST(COALESCE(DATETIME_PRECISION, 7) AS VARCHAR(20)) + ')'
        ELSE DATA_TYPE
      END AS TYPE_DECLARATION
    FROM INFORMATION_SCHEMA.COLUMNS
    CROSS JOIN params p
    WHERE TABLE_SCHEMA = p.target_schema
)
SELECT
    COALESCE(s.TABLE_NAME, t.TABLE_NAME) AS table_name,
    COALESCE(s.COLUMN_NAME, t.COLUMN_NAME) AS column_name,
    CASE
      WHEN s.COLUMN_NAME IS NULL THEN 'ONLY_IN_TARGET'
      WHEN t.COLUMN_NAME IS NULL THEN 'MISSING_IN_TARGET'
      WHEN s.TYPE_DECLARATION <> t.TYPE_DECLARATION
        OR s.IS_NULLABLE <> t.IS_NULLABLE
        OR ISNULL(s.COLUMN_DEFAULT, '') <> ISNULL(t.COLUMN_DEFAULT, '')
        OR ISNULL(s.CHARACTER_MAXIMUM_LENGTH, -1) <> ISNULL(t.CHARACTER_MAXIMUM_LENGTH, -1)
        OR ISNULL(s.NUMERIC_PRECISION, -1) <> ISNULL(t.NUMERIC_PRECISION, -1)
        OR ISNULL(s.NUMERIC_SCALE, -1) <> ISNULL(t.NUMERIC_SCALE, -1)
      THEN 'DIFF'
      ELSE 'MATCH'
    END AS status,
    s.TYPE_DECLARATION AS source_type,
    t.TYPE_DECLARATION AS target_type,
    s.IS_NULLABLE AS source_nullable,
    t.IS_NULLABLE AS target_nullable,
    s.COLUMN_DEFAULT AS source_default,
    t.COLUMN_DEFAULT AS target_default
FROM source_cols s
FULL OUTER JOIN target_cols t
  ON s.TABLE_NAME = t.TABLE_NAME AND s.COLUMN_NAME = t.COLUMN_NAME
WHERE s.COLUMN_NAME IS NULL
   OR t.COLUMN_NAME IS NULL
   OR s.TYPE_DECLARATION <> t.TYPE_DECLARATION
   OR s.IS_NULLABLE <> t.IS_NULLABLE
   OR ISNULL(s.COLUMN_DEFAULT, '') <> ISNULL(t.COLUMN_DEFAULT, '')
   OR ISNULL(s.CHARACTER_MAXIMUM_LENGTH, -1) <> ISNULL(t.CHARACTER_MAXIMUM_LENGTH, -1)
   OR ISNULL(s.NUMERIC_PRECISION, -1) <> ISNULL(t.NUMERIC_PRECISION, -1)
   OR ISNULL(s.NUMERIC_SCALE, -1) <> ISNULL(t.NUMERIC_SCALE, -1)
ORDER BY 1, 2;

-- Migration preview: add missing columns in target schema
WITH params AS (
    SELECT N'${sourceLiteral}' AS source_schema, N'${targetLiteral}' AS target_schema
)
SELECT
  'ALTER TABLE [' + p.target_schema + '].[' + s.TABLE_NAME + '] ADD [' + s.COLUMN_NAME + '] ' +
  CASE
    WHEN s.DATA_TYPE IN ('varchar','nvarchar','char','nchar','binary','varbinary')
      THEN s.DATA_TYPE + '(' + CASE WHEN s.CHARACTER_MAXIMUM_LENGTH = -1 THEN 'max' ELSE CAST(s.CHARACTER_MAXIMUM_LENGTH AS VARCHAR(20)) END + ')'
    WHEN s.DATA_TYPE IN ('decimal','numeric')
      THEN s.DATA_TYPE + '(' + CAST(s.NUMERIC_PRECISION AS VARCHAR(20)) + ',' + CAST(s.NUMERIC_SCALE AS VARCHAR(20)) + ')'
    WHEN s.DATA_TYPE IN ('datetime2','datetimeoffset','time')
      THEN s.DATA_TYPE + '(' + CAST(COALESCE(s.DATETIME_PRECISION, 7) AS VARCHAR(20)) + ')'
    ELSE s.DATA_TYPE
  END +
  CASE WHEN s.COLUMN_DEFAULT IS NOT NULL THEN ' DEFAULT ' + s.COLUMN_DEFAULT ELSE '' END +
  CASE WHEN s.IS_NULLABLE = 'NO' THEN ' NOT NULL;' ELSE ';' END AS migration_sql
FROM INFORMATION_SCHEMA.COLUMNS s
CROSS JOIN params p
LEFT JOIN INFORMATION_SCHEMA.COLUMNS t
  ON t.TABLE_SCHEMA = p.target_schema
 AND t.TABLE_NAME = s.TABLE_NAME
 AND t.COLUMN_NAME = s.COLUMN_NAME
WHERE s.TABLE_SCHEMA = p.source_schema
  AND t.COLUMN_NAME IS NULL
ORDER BY s.TABLE_NAME, s.ORDINAL_POSITION;`;
    } else if (type.includes('mysql') || type.includes('mariadb') || type.includes('databend')) {
        sql = `-- Schema Compare + Migration Preview (MySQL family)
-- Compare source and target databases
SELECT
  COALESCE(s.TABLE_NAME, t.TABLE_NAME) AS table_name,
  COALESCE(s.COLUMN_NAME, t.COLUMN_NAME) AS column_name,
  CASE
    WHEN s.COLUMN_NAME IS NULL THEN 'ONLY_IN_TARGET'
    WHEN t.COLUMN_NAME IS NULL THEN 'MISSING_IN_TARGET'
    WHEN s.COLUMN_TYPE <> t.COLUMN_TYPE
      OR s.IS_NULLABLE <> t.IS_NULLABLE
      OR COALESCE(s.COLUMN_DEFAULT, '') <> COALESCE(t.COLUMN_DEFAULT, '')
      OR COALESCE(s.EXTRA, '') <> COALESCE(t.EXTRA, '')
    THEN 'DIFF'
    ELSE 'MATCH'
  END AS status,
  s.COLUMN_TYPE AS source_type,
  t.COLUMN_TYPE AS target_type,
  s.IS_NULLABLE AS source_nullable,
  t.IS_NULLABLE AS target_nullable,
  s.COLUMN_DEFAULT AS source_default,
  t.COLUMN_DEFAULT AS target_default,
  s.EXTRA AS source_extra,
  t.EXTRA AS target_extra
FROM (
    SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = '${sourceLiteral}'
) s
LEFT JOIN (
    SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = '${targetLiteral}'
) t ON s.TABLE_NAME = t.TABLE_NAME AND s.COLUMN_NAME = t.COLUMN_NAME
UNION ALL
SELECT
  t.TABLE_NAME, t.COLUMN_NAME, 'ONLY_IN_TARGET', NULL, t.COLUMN_TYPE, NULL, t.IS_NULLABLE, NULL, t.COLUMN_DEFAULT, NULL, t.EXTRA
FROM (
    SELECT TABLE_NAME, COLUMN_NAME, COLUMN_TYPE, IS_NULLABLE, COLUMN_DEFAULT, EXTRA
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = '${targetLiteral}'
) t
LEFT JOIN (
    SELECT TABLE_NAME, COLUMN_NAME
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = '${sourceLiteral}'
) s ON s.TABLE_NAME = t.TABLE_NAME AND s.COLUMN_NAME = t.COLUMN_NAME
WHERE s.COLUMN_NAME IS NULL
ORDER BY 1, 2;

-- Migration preview: add missing columns to target schema
SELECT
  CONCAT('ALTER TABLE \`', '${targetLiteral}', '\`.\`', s.TABLE_NAME, '\` ADD COLUMN \`', s.COLUMN_NAME, '\` ', s.COLUMN_TYPE,
    IF(s.COLUMN_DEFAULT IS NOT NULL, CONCAT(' DEFAULT ', QUOTE(s.COLUMN_DEFAULT)), ''),
    IF(s.IS_NULLABLE = 'NO', ' NOT NULL;', ';')) AS migration_sql
FROM INFORMATION_SCHEMA.COLUMNS s
LEFT JOIN INFORMATION_SCHEMA.COLUMNS t
  ON t.TABLE_SCHEMA = '${targetLiteral}'
 AND t.TABLE_NAME = s.TABLE_NAME
 AND t.COLUMN_NAME = s.COLUMN_NAME
WHERE s.TABLE_SCHEMA = '${sourceLiteral}'
  AND t.COLUMN_NAME IS NULL
ORDER BY s.TABLE_NAME, s.ORDINAL_POSITION;`;
    } else {
        sql = `-- Schema Compare + Migration Preview
-- This helper currently supports PostgreSQL, MySQL/MariaDB, and SQL Server.
-- For this database type, run compare manually by querying metadata catalogs.
-- Suggested approach:
-- 1) Export table/column metadata from source and target
-- 2) Join on (table, column)
-- 3) Generate ALTER statements for missing/different columns`;
    }

    return sql;
};

const openSchemaCompareMigrationPreview = () => {
    const defaults = getSchemaCompareDefaults();
    schemaCompareWizard.sourceName = defaults.sourceName;
    schemaCompareWizard.targetName = defaults.targetName;
    schemaCompareWizard.isOpen = true;
};

const confirmSchemaCompareWizard = () => {
    const sourceName = schemaCompareWizard.sourceName.trim();
    const targetName = schemaCompareWizard.targetName.trim();
    if (!sourceName || !targetName) {
        toastRef.value?.error('Please provide both source and target names.');
        return;
    }

    const sql = buildSchemaCompareMigrationSql(sourceName, targetName);
    schemaCompareWizard.isOpen = false;
    openSqlTemplateTab('Schema Compare + Migration Preview', sql);
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

        const columns: database.ColumnDefinition[] = await GetTableDefinition(props.connectionId, tableName);
        if (!columns || columns.length === 0) {
            toastRef.value?.error('Could not read table columns for scripting.');
            closeContextMenu();
            return;
        }

        const pkColumns = columns.filter((col: database.ColumnDefinition) => col.primaryKey).map((col: database.ColumnDefinition) => col.name);
        const firstColumn = columns[0]?.name;
        const fallbackWhereColumns = firstColumn ? [firstColumn] : [];

        if (action === 'INSERT') {
            const insertColumns = columns.filter((col: database.ColumnDefinition) => !col.autoIncrement);
            const targetColumns = insertColumns.length > 0 ? insertColumns : columns;
            const columnLines = targetColumns.map((col: database.ColumnDefinition) => `  ${getEscapedIdentifier(col.name)}`).join(',\n');
            const valueLines = targetColumns.map((col: database.ColumnDefinition) => `  /* ${col.name} */`).join(',\n');
            const sql = `INSERT INTO ${escapedTableName} (\n${columnLines}\n)\nVALUES (\n${valueLines}\n);`;
            openSqlTemplateTab(`Script ${action}: ${tableName}`, sql);
            closeContextMenu();
            return;
        }

        if (action === 'UPDATE') {
            const setColumns = columns.filter((col: database.ColumnDefinition) => !col.primaryKey);
            const targetSetColumns = setColumns.length > 0 ? setColumns : columns;
            const whereColumns = pkColumns.length > 0 ? pkColumns : fallbackWhereColumns;
            const setLines = targetSetColumns.map((col: database.ColumnDefinition) => `  ${getEscapedIdentifier(col.name)} = /* ${col.name} */`).join(',\n');
            const whereLines = whereColumns.map((col: string) => `  ${getEscapedIdentifier(col)} = /* ${col} */`).join('\n  AND ');
            const sql = `UPDATE ${escapedTableName}\nSET\n${setLines}\nWHERE\n${whereLines || '  /* condition */'};`;
            openSqlTemplateTab(`Script ${action}: ${tableName}`, sql);
            closeContextMenu();
            return;
        }

        const deleteWhereColumns = pkColumns.length > 0 ? pkColumns : fallbackWhereColumns;
        const deleteWhereLines = deleteWhereColumns.map((col: string) => `  ${getEscapedIdentifier(col)} = /* ${col} */`).join('\n  AND ');
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
        const columns: database.ColumnDefinition[] = await GetTableDefinition(props.connectionId, tableName);
        if (!columns || columns.length === 0) {
            toastRef.value?.error('Could not read table columns.');
            closeContextMenu();
            return;
        }

        const pkColumns = columns.filter((col: database.ColumnDefinition) => col.primaryKey).map((col: database.ColumnDefinition) => col.name);
        const columnDefs = columns.map((col: database.ColumnDefinition) => {
            let line = `${getEscapedIdentifier(col.name)} ${col.type || 'TEXT'}`;
            if (!col.nullable) line += ' NOT NULL';
            if (col.defaultValue !== null && col.defaultValue !== undefined && String(col.defaultValue) !== '') {
                line += ` DEFAULT ${String(col.defaultValue)}`;
            }
            return line;
        });

        if (pkColumns.length > 0) {
            columnDefs.push(`PRIMARY KEY (${pkColumns.map((col: string) => getEscapedIdentifier(col)).join(', ')})`);
        }

        const createSql = `CREATE TABLE ${getEscapedTableName(tableName)} (\n${columnDefs.map((line: string) => `  ${line}`).join(',\n')}\n);`;
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
    const existingTab = tabs.value.find(t => t.name === `Schema: ${tableName}` || t.name === `ER: ${tableName}`);
    if (existingTab) {
        activeTabId.value = existingTab.id;
        return;
    }

    const newId = generateId();

    // Create tab
    const newTab: QueryTab = {
        id: newId,
        name: `Schema: ${tableName}`,
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

    // Important: use the reactive tab instance after push so async updates trigger UI immediately.
    const erTab = tabs.value.find(t => t.id === newId);
    if (!erTab) {
        return;
    }

    try {
        // 1. Get Foreign Keys First (Bidirectional)
        const fks = (await GetForeignKeys(props.connectionId, tableName)) || [];
        erTab.relationships = fks;

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
            const { schema, name } = splitQualifiedName(tbl);
            const escapedNameLiteral = escapeSqlLiteral(name);
            const escapedSchemaLiteral = schema ? escapeSqlLiteral(schema) : '';
            if (type.includes('mssql') || type.includes('sqlserver')) {
                const dot = tbl.indexOf('.');
                if (dot > 0) {
                    const schemaName = tbl.slice(0, dot);
                    const tableName = tbl.slice(dot + 1);
                    const escapedSchema = schemaName.replace(/]/g, ']]');
                    const escapedTable = tableName.replace(/]/g, ']]');
                    return `SELECT c.name AS COLUMN_NAME, typ.name AS DATA_TYPE
FROM sys.columns c
INNER JOIN sys.types typ ON c.user_type_id = typ.user_type_id
WHERE c.object_id = OBJECT_ID(N'[${escapedSchema}].[${escapedTable}]')
ORDER BY c.column_id`;
                }
                const schemaFilter = schema ? ` AND TABLE_SCHEMA = '${escapedSchemaLiteral}'` : '';
                return `SELECT COLUMN_NAME, DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '${escapedNameLiteral}'${schemaFilter}`;
            } else if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroach')) {
                const schemaFilter = schema ? ` AND table_schema = '${escapedSchemaLiteral}'` : '';
                return `SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '${escapedNameLiteral}'${schemaFilter}`;
            } else if (type.includes('mysql') || type.includes('maria') || type.includes('databend')) {
                return `DESCRIBE ${getEscapedTableName(tbl)}`;
            } else if (type.includes('sqlite') || type.includes('libsql')) {
                return `SELECT name AS column_name, type AS data_type FROM pragma_table_info('${escapeSqlLiteral(tbl)}')`;
            }
            return `SELECT * FROM ${getEscapedTableName(tbl)} LIMIT 1`;
        };

        // Execute queries in parallel
        const promises = Array.from(relatedTables).map(async (tbl) => {
            const query = getSchemaQuery(tbl);
            const reqId = generateId();
            erTab.activeQueryIds.push(reqId);
            try {
                const res = await ExecuteQuery(props.connectionId, query, reqId);
                // Need to handle resultSets here too
                if (!res.error && res.resultSets && res.resultSets.length > 0) {
                    const rs = res.resultSets[0];
                    const rows = (rs.rows || []).map((row: any) => {
                        if (Array.isArray(row)) {
                            return Object.fromEntries((rs.columns || []).map((col: string, i: number) => [col, row[i]]));
                        }
                        if (row && typeof row === 'object') {
                            return row;
                        }
                        return {};
                    });
                    tablesData[tbl] = rows.map((col: any) => {
                        const name = col.COLUMN_NAME || col.column_name || col.Field || col.field || col.name || col.Name || col.column || 'unknown';
                        const type = col.DATA_TYPE || col.data_type || col.Type || col.type || col.dataType || 'string';
                        return { name, type };
                    });
                }
            } catch (e) {
                console.warn(`Failed to fetch schema for ${tbl}`, e);
            } finally {
                erTab.activeQueryIds = erTab.activeQueryIds.filter(id => id !== reqId);
            }
        });

        await Promise.all(promises);
        erTab.tablesData = tablesData;

        // Keep main table columns in results for legacy/other uses if needed
        // IF we want to show it in grid? ER view handles its own rendering.
        // But for consistency:
        if (tablesData[tableName]) {
            // We'd need to mock a result set if we want to populate resultSets
            // But ER view reads tablesData.
        }

    } catch (e: any) {
        erTab.error = e.toString();
    } finally {
        erTab.isLoading = false;
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
    const { schema, name } = splitQualifiedName(tableName);
    const escapedNameLiteral = escapeSqlLiteral(name);
    const escapedSchemaLiteral = schema ? escapeSqlLiteral(schema) : '';
    let query = '';

    if (type.includes('mssql') || type.includes('sqlserver')) {
        const schemaFilter = schema ? ` AND TABLE_SCHEMA = '${escapedSchemaLiteral}'` : '';
        query = `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '${escapedNameLiteral}'${schemaFilter}`;
    } else if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroach')) {
        const schemaFilter = schema ? ` AND table_schema = '${escapedSchemaLiteral}'` : '';
        query = `SELECT column_name FROM information_schema.columns WHERE table_name = '${escapedNameLiteral}'${schemaFilter}`;
    } else if (type.includes('mysql') || type.includes('maria') || type.includes('databend')) {
        query = `SELECT COLUMN_NAME FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_SCHEMA = DATABASE() AND TABLE_NAME = '${escapedNameLiteral}'`;
    } else if (type.includes('sqlite') || type.includes('libsql')) {
        // SQLite PRAGMA returns a result set we need to parse differently if we use generic ExecuteQuery
        // But let's try standard schema table if available or just PRAGMA
        query = `SELECT name FROM pragma_table_info('${escapeSqlLiteral(tableName)}')`;
    } else {
        return [];
    }

    try {
        const reqId = generateId();
        // unique ID for schema fetch to avoid collision
        const res = await ExecuteQuery(props.connectionId, query, reqId);

        if (!res.error && res.resultSets && res.resultSets.length > 0) {
            const rs = res.resultSets[0];
            const cols = rs.rows.map((row: unknown) => {
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


const isEditableFormElement = (target: HTMLElement | null): boolean => {
    if (!target) return false;

    const tag = target.tagName;
    if (tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || target.isContentEditable) {
        const insideMonaco = !!target.closest('.monaco-editor') || target.classList.contains('inputarea');
        return !insideMonaco;
    }

    return false;
};

const handleKeydown = (e: KeyboardEvent) => {
    const target = e.target as HTMLElement | null;
    if (isEditableFormElement(target)) {
        return;
    }

    const withModifier = e.ctrlKey || e.metaKey;
    const key = e.key.toLowerCase();

    if (shortcutMatchesEvent(e, DEFAULT_GRID_SCREENSHOT_SHORTCUT)) {
        e.preventDefault();
        void exportQueryResultGridImage();
        return;
    }

    if (withModifier && key === 'enter') {
        e.preventDefault();
        void runQuery();
        return;
    }

    if (withModifier && key === 's') {
        e.preventDefault();
        void saveCurrentQueryToFile(e.shiftKey);
        return;
    }

    if (e.key === 'F5') {
        e.preventDefault();
        void refreshCurrentContext();
        return;
    }

    if (withModifier && key === 'r') {
        e.preventDefault();
        if (e.shiftKey) {
            void refreshDatabase();
        } else {
            void refreshCurrentContext();
        }
        return;
    }

    if (withModifier && key === 'n') {
        e.preventDefault();
        addTab();
        return;
    }

    if (withModifier && key === 'w') {
        e.preventDefault();
        if (activeTab.value) {
            closeTab(activeTab.value.id);
        }
        return;
    }

    if (withModifier && key === 'd') {
        e.preventDefault();
        if (activeTab.value && activeTab.value.tableName) {
            openDesignView(activeTab.value.tableName);
        }
        return;
    }

    if (e.shiftKey && e.altKey && key === 'f') {
        e.preventDefault();
        beautifyQuery();
    }
};

onMounted(async () => {
    try {
        const savedSettingsJson = await LoadSetting('user_settings');
        if (savedSettingsJson) {
            globalSettings.value = JSON.parse(savedSettingsJson);
        }
        await loadSnippetLibrary();
    } catch (e) {
        console.error("Failed to load dashboard settings", e);
    }

    await initializeDashboardState();

    window.addEventListener('keydown', handleKeydown, true);
    window.addEventListener('click', handleGlobalClick);
    window.addEventListener('open-sql-file', handleOpenSqlFile as EventListener);
    startMonitorTimer();
});

onUnmounted(() => {
    if (tabSessionSaveTimer) {
        clearTimeout(tabSessionSaveTimer);
        tabSessionSaveTimer = null;
    }
    void persistTabSession();

    cleanupAllStreams();
    window.removeEventListener('keydown', handleKeydown, true);
    window.removeEventListener('click', handleGlobalClick);
    window.removeEventListener('open-sql-file', handleOpenSqlFile as EventListener);
    stopAllResizing();
    stopMonitorTimer();
    clearResultImageDialogPreview();
});

interface OpenSqlFileDetail {
    content: string;
    fileName?: string;
    filePath?: string;
    connectionId: string;
}

const handleOpenSqlFile = (e: CustomEvent<OpenSqlFileDetail>) => {
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
            targetTab.query = detail.content || '';
            targetTab.sqlFilePath = detail.filePath || '';

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
        void initializeDashboardState();
        void loadSnippetLibrary();
    }
});

watch(() => props.dbType, () => {
    void loadSnippetLibrary();
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






























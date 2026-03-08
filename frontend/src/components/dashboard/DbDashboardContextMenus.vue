<script lang="ts" setup>
import { Kbd } from '../ui/kbd';

interface Props {
    contextMenu: any;
    disconnect: () => void | Promise<void>;
    refreshDatabase: () => void | Promise<void>;
    addTab: () => void;
    handleBackupExport: () => void | Promise<void>;
    handleGenerateDatabaseERDiagram: () => void;
    handleDatabaseInfo: () => void | Promise<void>;
    handleNewTable: () => void;
    handleNewView: () => void;
    handleNewRoutine: (type: 'PROCEDURE' | 'FUNCTION') => void;
    handleDropDatabase: () => void | Promise<void>;
    handleScriptRoutine: () => void;
    handleExecuteRoutine: () => void;
    handleDuplicateRoutine: () => void | Promise<void>;
    handleDeleteRoutine: () => void | Promise<void>;
    handleFolderRefresh: () => void | Promise<void>;
    handleFolderCollapse: () => void;
    handleCopyRow: () => void;
    handleCopyRowWithHeader: () => void;
    handleCopyCellValue: () => void;
    handleCopyCellValueWithHeader: () => void;
    handleAddWhereToCondition: () => void;
    handleSetNull: () => void;
    handleSetEmpty: () => void;
    handleSetDefault: () => void;
    handleSelectTop100: () => void;
    handleViewERDiagram: () => void;
    handleViewDesign: () => void;
    handleScriptTableAs: (action: 'SELECT' | 'INSERT' | 'UPDATE' | 'DELETE') => void | Promise<void>;
    handleGenerateCreateStatement: () => void | Promise<void>;
    handleCopyTableName: () => void | Promise<void>;
    handleTruncateTable: () => void | Promise<void>;
    handleDropTable: () => void | Promise<void>;
    handleExport: () => void | Promise<void>;
    handleImport: () => void | Promise<void>;
    handleSelectTop100View: () => void;
}

const {
    contextMenu,
    disconnect,
    refreshDatabase,
    addTab,
    handleBackupExport,
    handleGenerateDatabaseERDiagram,
    handleDatabaseInfo,
    handleNewTable,
    handleNewView,
    handleNewRoutine,
    handleDropDatabase,
    handleScriptRoutine,
    handleExecuteRoutine,
    handleDuplicateRoutine,
    handleDeleteRoutine,
    handleFolderRefresh,
    handleFolderCollapse,
    handleCopyRow,
    handleCopyRowWithHeader,
    handleCopyCellValue,
    handleCopyCellValueWithHeader,
    handleAddWhereToCondition,
    handleSetNull,
    handleSetEmpty,
    handleSetDefault,
    handleSelectTop100,
    handleViewERDiagram,
    handleViewDesign,
    handleScriptTableAs,
    handleGenerateCreateStatement,
    handleCopyTableName,
    handleTruncateTable,
    handleDropTable,
    handleExport,
    handleImport,
    handleSelectTop100View,
} = defineProps<Props>();
</script>

<template>        <!-- Context Menu for Tables -->
        <div v-if="contextMenu.show"
            class="fixed z-50 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[160px] context-menu-fixed"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleSelectTop100"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-list-start">
                    <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
                    <rect x="8" y="2" width="8" height="4" rx="1" ry="1" />
                </svg>
                Select Top 100
            </button>
            <!-- Original View Design button, now replaced by Design Table -->
            <!-- <button @click="handleViewDesign"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-pen-tool">
                    <path d="m12 19 7-7 3 3-7 7-3-3z" />
                    <path d="m18 13-1.5-7.5L2 2l3.5 14.5L13 18l5-5z" />
                    <path d="m2 2 7.586 7.586" />
                    <circle cx="11" cy="11" r="2" />
                </svg>
                View Design
            </button> -->
            <button @click="handleViewERDiagram"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-network">
                    <rect x="16" y="16" width="6" height="6" rx="1" />
                    <rect x="2" y="16" width="6" height="6" rx="1" />
                    <rect x="9" y="2" width="6" height="6" rx="1" />
                    <path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3" />
                    <path d="M12 12V8" />
                </svg>
                Schema Visualizer
            </button>
            <button @click="handleViewDesign"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-pencil-ruler">
                    <path d="M2 22h20" />
                    <path d="M12 6 2 16v6h6l10-10" />
                    <path d="m9 9 5 5" />
                </svg>
                Design Table
            </button>
            <div class="border-t border-border my-1"></div>
            <div class="relative group">
                <button
                    class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center justify-between gap-2">
                    <div class="flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                            fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                            stroke-linejoin="round" class="lucide lucide-file-code-2">
                            <path d="M16 22H4a2 2 0 0 1-2-2V8.5L8.5 2H16a2 2 0 0 1 2 2v4.5" />
                            <polyline points="14 2 14 8 8 8" />
                            <path d="m20 14-2 2 2 2" />
                            <path d="m24 14 2 2-2 2" />
                            <path d="m22 12-2 8" />
                        </svg>
                        Script Table as...
                    </div>
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-chevron-right">
                        <path d="m9 18 6-6-6-6" />
                    </svg>
                </button>
                <div
                    class="absolute left-full top-0 ml-1 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[150px] hidden group-hover:block">
                    <button @click="handleScriptTableAs('SELECT')"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground">
                        SELECT
                    </button>
                    <button @click="handleScriptTableAs('INSERT')"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground">
                        INSERT
                    </button>
                    <button @click="handleScriptTableAs('UPDATE')"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground">
                        UPDATE
                    </button>
                    <button @click="handleScriptTableAs('DELETE')"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground">
                        DELETE
                    </button>
                </div>
            </div>
            <button @click="handleGenerateCreateStatement"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-file-code-2">
                    <path d="M16 22H4a2 2 0 0 1-2-2V8.5L8.5 2H16a2 2 0 0 1 2 2v4.5" />
                    <polyline points="14 2 14 8 8 8" />
                    <path d="m20 14-2 2 2 2" />
                    <path d="m24 14 2 2-2 2" />
                    <path d="m22 12-2 8" />
                </svg>
                Generate Create Statement
            </button>
            <button @click="handleCopyTableName"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-copy">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </svg>
                Copy Name
            </button>
            <div class="border-t border-border my-1"></div>
            <button @click="handleTruncateTable"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-eraser">
                    <path d="m7 21-4.3-4.3a2 2 0 0 1 0-2.8l9.6-9.6a2 2 0 0 1 2.8 0l6.2 6.2a2 2 0 0 1 0 2.8L14 21" />
                    <path d="M22 21H7" />
                    <path d="m5 11 9 9" />
                </svg>
                Truncate Table
            </button>
            <button @click="handleDropTable"
                class="w-full text-left px-3 py-1.5 text-sm text-destructive hover:bg-destructive/10 hover:text-destructive flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-trash-2">
                    <path d="M3 6h18" />
                    <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                    <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                    <line x1="10" x2="10" y1="11" y2="17" />
                    <line x1="14" x2="14" y1="11" y2="17" />
                </svg>
                Drop Table
            </button>
            <div class="border-t border-border my-1"></div>
            <button @click="handleExport"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-upload">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                    <polyline points="17 8 12 3 7 8" />
                    <line x1="12" x2="12" y1="3" y2="15" />
                </svg>
                Export Data
            </button>
            <button @click="handleImport"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-download">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                    <polyline points="7 10 12 15 17 10" />
                    <line x1="12" x2="12" y1="15" y2="3" />
                </svg>
                Import Data
            </button>
        </div>

        <!-- View Context Menu -->
        <div v-if="contextMenu.showView"
            class="fixed z-50 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[160px] context-menu-fixed"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleSelectTop100View"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-list-start">
                    <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
                    <rect x="8" y="2" width="8" height="4" rx="1" ry="1" />
                </svg>
                Select Top 100
            </button>
            <button @click="handleNewView"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-plus-circle">
                    <circle cx="12" cy="12" r="10" />
                    <path d="M8 12h8" />
                    <path d="M12 8v8" />
                </svg>
                New View
            </button>
        </div>

        <!-- Database Context Menu -->
        <div v-if="contextMenu.showDb"
            class="fixed z-50 min-w-[160px] bg-popover text-popover-foreground rounded-md border border-border shadow-md py-1 animate-in fade-in zoom-in-95 duration-100 context-menu-fixed"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="refreshDatabase"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-refresh-cw">
                    <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                    <path d="M3 3v5h5" />
                </svg>
                Refresh
            </button>
            <button @click="addTab"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-terminal-square">
                    <path d="m7 11 2-2-2-2" />
                    <path d="M11 13h4" />
                    <rect width="18" height="18" x="3" y="3" rx="2" ry="2" />
                </svg>
                New Query
            </button>
            <div class="h-px bg-border my-1"></div>
            <button @click="handleBackupExport"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-archive">
                    <rect width="20" height="5" x="2" y="3" rx="1" />
                    <path d="M4 8v11a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8" />
                    <path d="M10 12h4" />
                </svg>
                Backup / Export
            </button>
            <button @click="handleGenerateDatabaseERDiagram"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-network">
                    <rect x="16" y="16" width="6" height="6" rx="1" />
                    <rect x="2" y="16" width="6" height="6" rx="1" />
                    <rect x="9" y="2" width="6" height="6" rx="1" />
                    <path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3" />
                    <path d="M12 12V8" />
                </svg>
                Schema Visualizer
            </button>
            <button @click="handleDatabaseInfo"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-info">
                    <circle cx="12" cy="12" r="10" />
                    <path d="M12 16v-4" />
                    <path d="M12 8h.01" />
                </svg>
                Database Info
            </button>
            <div class="h-px bg-border my-1"></div>
            <div class="relative group">
                <button
                    class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center justify-between gap-2 transition-colors">
                    <div class="flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-plus-circle">
                            <circle cx="12" cy="12" r="10" />
                            <path d="M8 12h8" />
                            <path d="M12 8v8" />
                        </svg>
                        Create New...
                    </div>
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-chevron-right">
                        <path d="m9 18 6-6-6-6" />
                    </svg>
                </button>
                <div
                    class="absolute left-full top-0 ml-1 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[140px] hidden group-hover:block animate-in fade-in slide-in-from-left-1 duration-150">
                    <button @click="handleNewTable"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-table">
                            <path d="M12 3v18" />
                            <rect width="18" height="18" x="3" y="3" rx="2" />
                            <path d="M3 9h18" />
                            <path d="M3 15h18" />
                        </svg>
                        Table
                    </button>
                    <button @click="handleNewView"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-eye text-green-500">
                            <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
                            <circle cx="12" cy="12" r="3" />
                        </svg>
                        View
                    </button>
                    <button @click="handleNewRoutine('PROCEDURE')"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-scroll text-blue-400">
                            <path d="M8 17a5 5 0 0 1 5-5c1.1 0 2 .9 2 2v6a2 2 0 0 1-4 0v-6.5" />
                            <path d="M12 2H8a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V9a5 5 0 0 0-5-5Z" />
                        </svg>
                        Procedure
                    </button>
                    <button @click="handleNewRoutine('FUNCTION')"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-function-square text-purple-400">
                            <rect width="18" height="18" x="3" y="3" rx="2" ry="2" />
                            <path d="M9 17c2 0 2.8-1 2.8-2.8V10c0-2 1-3.3 3.2-3" />
                            <path d="M9 11.2h5.7" />
                        </svg>
                        Function
                    </button>
                </div>
            </div>
            <div class="h-px bg-border my-1"></div>
            <button @click="handleDropDatabase"
                class="w-full text-left px-3 py-1.5 text-sm text-destructive hover:bg-destructive/10 hover:text-destructive focus:bg-destructive/10 focus:text-destructive flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-trash-2">
                    <path d="M3 6h18" />
                    <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                    <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                </svg>
                Drop Database
            </button>
            <div class="h-px bg-border my-1"></div>
            <button @click="disconnect"
                class="w-full text-left px-3 py-1.5 text-sm text-destructive hover:bg-destructive/10 hover:text-destructive focus:bg-destructive/10 focus:text-destructive flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-log-out">
                    <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
                    <polyline points="16 17 21 12 16 7" />
                    <line x1="21" x2="9" y1="12" y2="12" />
                </svg>
                Disconnect
            </button>
        </div>

        <!-- Routine Context Menu -->
        <div v-if="contextMenu.showRoutine"
            class="fixed z-50 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[160px] context-menu-fixed"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleScriptRoutine"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-file-code">
                    <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
                    <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
                </svg>
                Script as Create
            </button>
            <button @click="handleExecuteRoutine"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-play">
                    <polygon points="5 3 19 12 5 21 5 3" />
                </svg>
                Execute
            </button>
            <button @click="handleDuplicateRoutine"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-copy">
                    <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
                    <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
                </svg>
                Duplicate
            </button>
            <div class="h-px bg-border my-1"></div>
            <button @click="handleDeleteRoutine"
                class="w-full text-left px-3 py-1.5 text-sm hover:text-destructive hover:bg-destructive/10 flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-trash-2">
                    <path d="M3 6h18" />
                    <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                    <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                    <line x1="10" x2="10" y1="11" y2="17" />
                    <line x1="14" x2="14" y1="11" y2="17" />
                </svg>
                Delete
            </button>
        </div>

        <!-- Folder Context Menu -->
        <div v-if="contextMenu.showFolder"
            class="fixed z-50 min-w-[160px] bg-popover text-popover-foreground rounded-md border border-border shadow-md py-1 animate-in fade-in zoom-in-95 duration-100 context-menu-fixed"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleFolderRefresh"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-refresh-cw">
                    <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                    <path d="M3 3v5h5" />
                </svg>
                Refresh {{ contextMenu.targetFolder }}
            </button>
            <div v-if="contextMenu.targetFolder === 'Tables'" class="h-px bg-border my-1"></div>
            <button v-if="contextMenu.targetFolder === 'Tables'" @click="handleNewTable"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-plus">
                    <path d="M5 12h14" />
                    <path d="M12 5v14" />
                </svg>
                Create Table
            </button>
            <div v-if="contextMenu.targetFolder === 'Stored Procedures' || contextMenu.targetFolder === 'Programmability'"
                class="h-px bg-border my-1"></div>
            <button
                v-if="contextMenu.targetFolder === 'Stored Procedures' || contextMenu.targetFolder === 'Programmability'"
                @click="handleNewRoutine('PROCEDURE')"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-plus">
                    <path d="M5 12h14" />
                    <path d="M12 5v14" />
                </svg>
                New Procedure
            </button>
            <div v-if="contextMenu.targetFolder === 'Functions' || contextMenu.targetFolder === 'Programmability'"
                class="h-px bg-border my-1"></div>
            <button v-if="contextMenu.targetFolder === 'Functions' || contextMenu.targetFolder === 'Programmability'"
                @click="handleNewRoutine('FUNCTION')"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-plus">
                    <path d="M5 12h14" />
                    <path d="M12 5v14" />
                </svg>
                New Function
            </button>
            <div class="h-px bg-border my-1"></div>
            <button @click="handleFolderCollapse"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2 transition-colors">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-folder-closed">
                    <path
                        d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z" />
                </svg>
                Collapse
            </button>
        </div>

        <!-- Row Context Menu -->
        <div v-if="contextMenu.showRow"
            class="fixed z-50 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[160px] context-menu-fixed"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleCopyRow"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-copy">
                        <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
                        <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
                    </svg>
                    Copy Row
                </div>
                <Kbd class="text-[10px] pointer-events-none h-4 px-1">Ctrl + Shift + C</Kbd>
            </button>
            <button @click="handleCopyRowWithHeader"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-copy-plus">
                        <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
                        <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
                        <path d="M15 15h6" />
                        <path d="M18 12v6" />
                    </svg>
                    Copy Row With Header
                </div>
            </button>
            <button @click="handleCopyCellValue"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center justify-between gap-2">
                <div class="flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-clipboard-copy">
                        <rect width="8" height="4" x="8" y="2" rx="1" ry="1" />
                        <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
                    </svg>
                    Copy Cell Value
                </div>
                <Kbd class="text-[10px] pointer-events-none h-4 px-1">Ctrl + C</Kbd>
            </button>
            <button @click="handleCopyCellValueWithHeader"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-clipboard-type">
                    <rect width="8" height="4" x="8" y="2" rx="1" ry="1" />
                    <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
                    <path d="M9 12v2h6v-2" />
                    <path d="M12 12v7" />
                </svg>
                Copy Cell Value With Header
            </button>
            <div class="border-t border-border my-1"></div>
            <button @click="handleAddWhereToCondition"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-filter">
                    <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
                </svg>
                Add Where To Condition
            </button>
            <div class="border-t border-border my-1"></div>
            <div class="relative group">
                <button
                    class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center justify-between gap-2">
                    <div class="flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-pencil">
                            <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
                            <path d="m15 5 4 4" />
                        </svg>
                        Set Value
                    </div>
                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-chevron-right">
                        <path d="m9 18 6-6-6-6" />
                    </svg>
                </button>
                <!-- Submenu -->
                <div
                    class="absolute left-full top-0 ml-1 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[120px] hidden group-hover:block">
                    <button @click="handleSetNull"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground">
                        Set to NULL
                    </button>
                    <button @click="handleSetEmpty"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground">
                        Set to Empty
                    </button>
                    <button @click="handleSetDefault"
                        class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground">
                        Set to Default
                    </button>
                </div>
            </div>
        </div>

</template>

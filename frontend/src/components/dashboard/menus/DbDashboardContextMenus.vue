<script lang="ts" setup>
import { computed, nextTick, ref, watch } from 'vue';

import type { DashboardContextMenuState } from '../../../composables/useDashboardContextMenus';

import { Kbd } from '../../ui/kbd';

interface Props {
    contextMenu: DashboardContextMenuState;
    dbType: string;
    isReadOnly: boolean;
    openFolders: string[];
    closeContextMenu: () => void;
    openHistory: () => void;
    openActivityMonitor: () => void;
    disconnect: () => void | Promise<void>;
    refreshDatabase: () => void | Promise<void>;
    addTab: () => void;
    openTransactionSandbox: () => void | Promise<void>;
    openSchemaCompare: () => void | Promise<void>;
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
    handleFolderToggle: () => void;
    handleFolderExpandAll: () => void;
    handleFolderCollapseAll: () => void;
    handleCopyRow: () => void | Promise<void>;
    handleCopyRowWithHeader: () => void | Promise<void>;
    handleViewCellDetails: () => void | Promise<void>;
    handleCopyCellValue: () => void | Promise<void>;
    handleCopyCellValueWithHeader: () => void | Promise<void>;
    handleCopyHeaderName: () => void | Promise<void>;
    handleCopyHeaderRow: () => void | Promise<void>;
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
    dbType,
    isReadOnly,
    openFolders,
    closeContextMenu,
    openHistory,
    openActivityMonitor,
    disconnect,
    refreshDatabase,
    addTab,
    openTransactionSandbox,
    openSchemaCompare,
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
    handleFolderToggle,
    handleFolderExpandAll,
    handleFolderCollapseAll,
    handleCopyRow,
    handleCopyRowWithHeader,
    handleViewCellDetails,
    handleCopyCellValue,
    handleCopyCellValueWithHeader,
    handleCopyHeaderName,
    handleCopyHeaderRow,
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

type MenuAction = () => void | Promise<void>;

const dbMenuRef = ref<HTMLElement | null>(null);
const createNewTriggerRef = ref<HTMLButtonElement | null>(null);
const createNewSubmenuRef = ref<HTMLElement | null>(null);
const isCreateSubmenuOpen = ref(false);
const createSubmenuAlignLeft = ref(false);
const scriptTableAsSubmenuRef = ref<HTMLElement | null>(null);
const isScriptTableAsSubmenuOpen = ref(false);
const scriptTableAsSubmenuAlignLeft = ref(false);
const setValueSubmenuRef = ref<HTMLElement | null>(null);
const isSetValueSubmenuOpen = ref(false);
const setValueSubmenuAlignLeft = ref(false);

const normalizedDbType = computed(() => (dbType || '').toLowerCase());
const canMutateSchema = computed(() => !isReadOnly);
const menuPanelClass = 'fixed z-50 min-w-[160px] rounded-2xl border border-border/80 bg-popover/95 py-1 text-popover-foreground shadow-xl ring-1 ring-black/5 backdrop-blur animate-in fade-in zoom-in-95 duration-100 context-menu-fixed';
const dbMenuPanelClass = 'fixed z-50 min-w-[180px] rounded-2xl border border-border/80 bg-popover/95 py-1 text-popover-foreground shadow-xl ring-1 ring-black/5 backdrop-blur animate-in fade-in zoom-in-95 duration-100 context-menu-fixed';
const submenuPanelClass = 'absolute left-full top-0 ml-1 hidden rounded-2xl border border-border/80 bg-popover/95 py-1 text-popover-foreground shadow-xl ring-1 ring-black/5 backdrop-blur group-hover:block';
const createSubmenuPanelClass = 'absolute top-0 z-10 min-w-[140px] rounded-2xl border border-border/80 bg-popover/95 py-1 text-popover-foreground shadow-xl ring-1 ring-black/5 backdrop-blur animate-in fade-in duration-100';
const menuItemClass = 'flex w-full items-center gap-2 px-3 py-2 text-left text-sm transition-colors hover:bg-accent hover:text-accent-foreground';
const menuItemBetweenClass = `${menuItemClass} justify-between`;
const plainMenuItemClass = `${menuItemClass} justify-start`;
const destructiveMenuItemClass = 'flex w-full items-center gap-2 px-3 py-2 text-left text-sm text-destructive transition-colors hover:bg-destructive/10 hover:text-destructive focus:bg-destructive/10 focus:text-destructive';
const menuSeparatorClass = 'h-px bg-border my-1';
const supportsRoutines = computed(() => {
    const type = normalizedDbType.value;
    return type.includes('postgres')
        || type.includes('greenplum')
        || type.includes('mysql')
        || type.includes('mariadb')
        || type.includes('mssql')
        || type.includes('sqlserver');
});

const currentFolder = computed(() => contextMenu.targetFolder || '');
const isTablesFolder = computed(() => currentFolder.value === 'Tables');
const isViewsFolder = computed(() => currentFolder.value === 'Views');
const isStoredProceduresFolder = computed(() => currentFolder.value === 'Stored Procedures');
const isFunctionsFolder = computed(() => currentFolder.value === 'Functions');
const isProgrammabilityFolder = computed(() => currentFolder.value === 'Programmability');
const canCreateProcedure = computed(() => supportsRoutines.value && (isStoredProceduresFolder.value || isProgrammabilityFolder.value));
const canCreateFunction = computed(() => supportsRoutines.value && (isFunctionsFolder.value || isProgrammabilityFolder.value));
const isFolderOpen = computed(() => openFolders.includes(currentFolder.value));
const folderToggleLabel = computed(() => (isFolderOpen.value ? 'Collapse Folder' : 'Expand Folder'));

const getDbMenuItems = (): HTMLElement[] => {
    if (!dbMenuRef.value) {
        return [];
    }
    return Array.from(
        dbMenuRef.value.querySelectorAll<HTMLElement>('[data-db-menu-item]:not([disabled])')
    ).filter((el) => el.offsetParent !== null);
};

const moveDbMenuFocus = (direction: 1 | -1) => {
    const items = getDbMenuItems();
    if (items.length === 0) {
        return;
    }
    const current = document.activeElement as HTMLElement | null;
    const index = items.findIndex((item) => item === current);
    const next = index >= 0 ? (index + direction + items.length) % items.length : 0;
    items[next]?.focus();
};

const runDbAction = async (action: MenuAction) => {
    await action();
    isCreateSubmenuOpen.value = false;
    closeContextMenu();
};

const openCreateSubmenu = async () => {
    if (!canMutateSchema.value) {
        return;
    }
    isCreateSubmenuOpen.value = true;
    await nextTick();
    if (createNewSubmenuRef.value) {
        const submenuRect = createNewSubmenuRef.value.getBoundingClientRect();
        createSubmenuAlignLeft.value = submenuRect.right > window.innerWidth;
    }
};

const closeCreateSubmenu = () => {
    isCreateSubmenuOpen.value = false;
    createSubmenuAlignLeft.value = false;
};

const toggleCreateSubmenu = async () => {
    if (isCreateSubmenuOpen.value) {
        closeCreateSubmenu();
        return;
    }
    await openCreateSubmenu();
};

const openScriptTableAsSubmenu = async () => {
    isScriptTableAsSubmenuOpen.value = true;
    await nextTick();
    if (scriptTableAsSubmenuRef.value) {
        const submenuRect = scriptTableAsSubmenuRef.value.getBoundingClientRect();
        scriptTableAsSubmenuAlignLeft.value = submenuRect.right > window.innerWidth;
    }
};

const closeScriptTableAsSubmenu = () => {
    isScriptTableAsSubmenuOpen.value = false;
    scriptTableAsSubmenuAlignLeft.value = false;
};

const toggleScriptTableAsSubmenu = async () => {
    if (isScriptTableAsSubmenuOpen.value) {
        closeScriptTableAsSubmenu();
        return;
    }
    await openScriptTableAsSubmenu();
};

const openSetValueSubmenu = async () => {
    isSetValueSubmenuOpen.value = true;
    await nextTick();
    if (setValueSubmenuRef.value) {
        const submenuRect = setValueSubmenuRef.value.getBoundingClientRect();
        setValueSubmenuAlignLeft.value = submenuRect.right > window.innerWidth;
    }
};

const closeSetValueSubmenu = () => {
    isSetValueSubmenuOpen.value = false;
    setValueSubmenuAlignLeft.value = false;
};

const toggleSetValueSubmenu = async () => {
    if (isSetValueSubmenuOpen.value) {
        closeSetValueSubmenu();
        return;
    }
    await openSetValueSubmenu();
};

const focusFirstCreateSubmenuItem = () => {
    if (!createNewSubmenuRef.value) {
        return;
    }
    const firstItem = createNewSubmenuRef.value.querySelector<HTMLElement>('[data-db-menu-item]');
    firstItem?.focus();
};

const handleDbMenuKeydown = async (event: KeyboardEvent) => {
    const active = document.activeElement as HTMLElement | null;
    const activeInSubmenu = !!(active && createNewSubmenuRef.value?.contains(active));

    if (event.key === 'Escape') {
        event.preventDefault();
        closeContextMenu();
        return;
    }

    if (event.key === 'ArrowDown') {
        event.preventDefault();
        moveDbMenuFocus(1);
        return;
    }

    if (event.key === 'ArrowUp') {
        event.preventDefault();
        moveDbMenuFocus(-1);
        return;
    }

    if (event.key === 'Home') {
        event.preventDefault();
        getDbMenuItems()[0]?.focus();
        return;
    }

    if (event.key === 'End') {
        event.preventDefault();
        const items = getDbMenuItems();
        items[items.length - 1]?.focus();
        return;
    }

    if (event.key === 'ArrowRight' && active === createNewTriggerRef.value) {
        event.preventDefault();
        await openCreateSubmenu();
        focusFirstCreateSubmenuItem();
        return;
    }

    if (event.key === 'ArrowLeft' && activeInSubmenu) {
        event.preventDefault();
        closeCreateSubmenu();
        createNewTriggerRef.value?.focus();
        return;
    }

    if ((event.key === 'Enter' || event.key === ' ') && active?.hasAttribute('data-db-menu-item')) {
        event.preventDefault();
        active.click();
    }
};

const handleOpenHistory = async () => {
    await runDbAction(openHistory);
};

const handleOpenActivityMonitor = async () => {
    await runDbAction(openActivityMonitor);
};

const handleNewQuery = async () => {
    await runDbAction(addTab);
};

watch(
    () => contextMenu.showDb,
    async (show) => {
        if (!show) {
            closeCreateSubmenu();
            return;
        }

        closeCreateSubmenu();
        await nextTick();
        dbMenuRef.value?.focus();
        getDbMenuItems()[0]?.focus();
    }
);

watch(() => contextMenu.show, (show) => {
    if (!show) {
        closeScriptTableAsSubmenu();
    }
});

watch(() => contextMenu.showRow, (show) => {
    if (!show) {
        closeSetValueSubmenu();
    }
});
</script>

<template>        <!-- Context Menu for Tables -->
        <div v-if="contextMenu.show"
            :class="menuPanelClass"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleSelectTop100"
                :class="menuItemClass">
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
                :class="plainMenuItemClass">
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
                :class="menuItemClass">
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
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-pencil-ruler">
                    <path d="M2 22h20" />
                    <path d="M12 6 2 16v6h6l10-10" />
                    <path d="m9 9 5 5" />
                </svg>
                Design Table
            </button>
            <div :class="menuSeparatorClass"></div>
            <div class="relative" @mouseenter="openScriptTableAsSubmenu" @mouseleave="closeScriptTableAsSubmenu">
                <button
                    :aria-expanded="isScriptTableAsSubmenuOpen ? 'true' : 'false'" aria-haspopup="menu"
                    @click="toggleScriptTableAsSubmenu" @focus="openScriptTableAsSubmenu"
                    :class="menuItemBetweenClass">
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
                <div v-if="isScriptTableAsSubmenuOpen" ref="scriptTableAsSubmenuRef"
                    :class="[
                        createSubmenuPanelClass,
                        'min-w-[150px]',
                        scriptTableAsSubmenuAlignLeft ? 'right-full mr-1' : 'left-full ml-1'
                    ]">
                    <button @click="handleScriptTableAs('SELECT')"
                        :class="menuItemClass">
                        SELECT
                    </button>
                    <button @click="handleScriptTableAs('INSERT')"
                        :class="menuItemClass">
                        INSERT
                    </button>
                    <button @click="handleScriptTableAs('UPDATE')"
                        :class="menuItemClass">
                        UPDATE
                    </button>
                    <button @click="handleScriptTableAs('DELETE')"
                        :class="menuItemClass">
                        DELETE
                    </button>
                </div>
            </div>
            <button @click="handleGenerateCreateStatement"
                :class="menuItemClass">
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
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-copy">
                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                </svg>
                Copy Name
            </button>
            <div :class="menuSeparatorClass"></div>
            <button @click="handleTruncateTable"
                :class="menuItemClass">
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
                :class="destructiveMenuItemClass">
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
            <div :class="menuSeparatorClass"></div>
            <button @click="handleExport"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-upload">
                    <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4" />
                    <polyline points="17 8 12 3 7 8" />
                    <line x1="12" x2="12" y1="3" y2="15" />
                </svg>
                Export Table...
            </button>
            <button @click="handleImport"
                :class="menuItemClass">
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
            :class="menuPanelClass"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleSelectTop100View"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-list-start">
                    <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
                    <rect x="8" y="2" width="8" height="4" rx="1" ry="1" />
                </svg>
                Select Top 100
            </button>
            <button @click="handleNewView"
                :class="menuItemClass">
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
        <div v-if="contextMenu.showDb" ref="dbMenuRef" role="menu" tabindex="0"
            aria-label="Database context menu" @keydown="handleDbMenuKeydown"
            :class="dbMenuPanelClass"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button data-db-menu-item role="menuitem" @click="runDbAction(refreshDatabase)"
                :class="menuItemBetweenClass">
                <div class="flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-refresh-cw">
                        <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                        <path d="M3 3v5h5" />
                    </svg>
                    Refresh
                </div>
                <Kbd class="text-[10px] pointer-events-none h-4 px-1">Ctrl + R</Kbd>
            </button>
            <button data-db-menu-item role="menuitem" @click="handleNewQuery"
                :class="menuItemBetweenClass">
                <div class="flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-terminal-square">
                        <path d="m7 11 2-2-2-2" />
                        <path d="M11 13h4" />
                        <rect width="18" height="18" x="3" y="3" rx="2" ry="2" />
                    </svg>
                    New Query
                </div>
                <Kbd class="text-[10px] pointer-events-none h-4 px-1">Ctrl + N</Kbd>
            </button>
            <button data-db-menu-item role="menuitem" @click="runDbAction(openTransactionSandbox)"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-shield-check">
                    <path d="M20 13c0 5-3.5 7.5-8 9-4.5-1.5-8-4-8-9V6l8-4 8 4z" />
                    <path d="m9 12 2 2 4-4" />
                </svg>
                Transaction Sandbox
            </button>
            <button data-db-menu-item role="menuitem" @click="runDbAction(openSchemaCompare)"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-git-compare-arrows">
                    <circle cx="5" cy="6" r="3" />
                    <path d="M12 6h7" />
                    <path d="m16 10 3-4-3-4" />
                    <circle cx="19" cy="18" r="3" />
                    <path d="M12 18H5" />
                    <path d="m8 14-3 4 3 4" />
                </svg>
                Schema Compare
            </button>
            <button data-db-menu-item role="menuitem" @click="handleOpenHistory"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-history">
                    <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                    <path d="M3 3v5h5" />
                    <path d="M12 7v5l4 2" />
                </svg>
                Query History
            </button>
            <button data-db-menu-item role="menuitem" @click="handleOpenActivityMonitor"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-activity">
                    <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
                </svg>
                Activity Monitor
            </button>
            <div class="h-px bg-border my-1"></div>
            <button data-db-menu-item role="menuitem" @click="runDbAction(handleBackupExport)"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-archive">
                    <rect width="20" height="5" x="2" y="3" rx="1" />
                    <path d="M4 8v11a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8" />
                    <path d="M10 12h4" />
                </svg>
                Backup / Export
            </button>
            <button data-db-menu-item role="menuitem" @click="runDbAction(handleGenerateDatabaseERDiagram)"
                :class="menuItemClass">
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
            <button data-db-menu-item role="menuitem" @click="runDbAction(handleDatabaseInfo)"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-info">
                    <circle cx="12" cy="12" r="10" />
                    <path d="M12 16v-4" />
                    <path d="M12 8h.01" />
                </svg>
                Database Info
            </button>
            <div v-if="canMutateSchema" class="h-px bg-border my-1"></div>
            <div v-if="canMutateSchema" class="relative"
                @mouseenter="openCreateSubmenu" @mouseleave="closeCreateSubmenu">
                <button ref="createNewTriggerRef" data-db-menu-item role="menuitem"
                    :aria-expanded="isCreateSubmenuOpen ? 'true' : 'false'" aria-haspopup="menu"
                    @click="toggleCreateSubmenu" @focus="openCreateSubmenu"
                    :class="menuItemBetweenClass">
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
                <div v-if="isCreateSubmenuOpen" ref="createNewSubmenuRef" role="menu"
                    :class="[
                        createSubmenuPanelClass,
                        createSubmenuAlignLeft ? 'right-full mr-1' : 'left-full ml-1'
                    ]">
                    <button data-db-menu-item role="menuitem" @click="runDbAction(handleNewTable)"
                        :class="menuItemClass">
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
                    <button data-db-menu-item role="menuitem" @click="runDbAction(handleNewView)"
                        :class="menuItemClass">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-eye text-green-500">
                            <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
                            <circle cx="12" cy="12" r="3" />
                        </svg>
                        View
                    </button>
                    <button v-if="supportsRoutines" data-db-menu-item role="menuitem"
                        @click="runDbAction(() => handleNewRoutine('PROCEDURE'))"
                        :class="menuItemClass">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-scroll text-blue-400">
                            <path d="M8 17a5 5 0 0 1 5-5c1.1 0 2 .9 2 2v6a2 2 0 0 1-4 0v-6.5" />
                            <path d="M12 2H8a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V9a5 5 0 0 0-5-5Z" />
                        </svg>
                        Procedure
                    </button>
                    <button v-if="supportsRoutines" data-db-menu-item role="menuitem"
                        @click="runDbAction(() => handleNewRoutine('FUNCTION'))"
                        :class="menuItemClass">
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
            <div v-if="canMutateSchema" class="h-px bg-border my-1"></div>
            <button v-if="canMutateSchema" data-db-menu-item role="menuitem" @click="runDbAction(handleDropDatabase)"
                :class="destructiveMenuItemClass">
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
            <button data-db-menu-item role="menuitem" @click="runDbAction(disconnect)"
                :class="destructiveMenuItemClass">
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
            :class="menuPanelClass"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleScriptRoutine"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-file-code">
                    <path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71" />
                    <path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71" />
                </svg>
                Script as Create
            </button>
            <button @click="handleExecuteRoutine"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-play">
                    <polygon points="5 3 19 12 5 21 5 3" />
                </svg>
                Execute
            </button>
            <button @click="handleDuplicateRoutine"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-copy">
                    <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
                    <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
                </svg>
                Duplicate
            </button>
            <div :class="menuSeparatorClass"></div>
            <button @click="handleDeleteRoutine"
                :class="destructiveMenuItemClass">
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
            :class="menuPanelClass"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleFolderRefresh"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-refresh-cw">
                    <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                    <path d="M3 3v5h5" />
                </svg>
                Refresh {{ contextMenu.targetFolder }}
            </button>
            <button @click="refreshDatabase"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-database-zap">
                    <ellipse cx="12" cy="5" rx="9" ry="3" />
                    <path d="M3 5v14c0 1.7 4 3 9 3s9-1.3 9-3V5" />
                    <path d="m7 12 3 3 7-7" />
                </svg>
                Refresh Database
            </button>
            <div v-if="canMutateSchema && (isTablesFolder || isViewsFolder || canCreateProcedure || canCreateFunction)"
                :class="menuSeparatorClass"></div>
            <button v-if="canMutateSchema && isTablesFolder" @click="handleNewTable"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-plus">
                    <path d="M5 12h14" />
                    <path d="M12 5v14" />
                </svg>
                Create Table
            </button>
            <button v-if="canMutateSchema && isViewsFolder" @click="handleNewView"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-plus">
                    <path d="M5 12h14" />
                    <path d="M12 5v14" />
                </svg>
                New View
            </button>
            <button v-if="canMutateSchema && canCreateProcedure" @click="handleNewRoutine('PROCEDURE')"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-plus">
                    <path d="M5 12h14" />
                    <path d="M12 5v14" />
                </svg>
                New Procedure
            </button>
            <button v-if="canMutateSchema && canCreateFunction" @click="handleNewRoutine('FUNCTION')"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-plus">
                    <path d="M5 12h14" />
                    <path d="M12 5v14" />
                </svg>
                New Function
            </button>
            <div v-if="isProgrammabilityFolder" :class="menuSeparatorClass"></div>
            <button v-if="isProgrammabilityFolder" @click="handleFolderExpandAll"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-folders">
                    <path d="M20 7h-8l-2-2H4a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h16a2 2 0 0 0 2-2V9a2 2 0 0 0-2-2Z" />
                    <path d="M8 12h8" />
                    <path d="M12 8v8" />
                </svg>
                Expand All
            </button>
            <button v-if="isProgrammabilityFolder" @click="handleFolderCollapseAll"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-folder-x">
                    <path d="M3 5a2 2 0 0 1 2-2h3l2 2h9a2 2 0 0 1 2 2v1H3V5Z" />
                    <path d="M3 8h18v9a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8Z" />
                    <path d="m9 12 6 6" />
                    <path d="m15 12-6 6" />
                </svg>
                Collapse All
            </button>
            <div :class="menuSeparatorClass"></div>
            <button @click="handleFolderToggle"
                :class="plainMenuItemClass">
                <svg v-if="isFolderOpen" xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-folder-closed">
                    <path
                        d="M20 20a2 2 0 0 0 2-2V8a2 2 0 0 0-2-2h-7.9a2 2 0 0 1-1.69-.9L9.6 3.9A2 2 0 0 0 7.93 3H4a2 2 0 0 0-2 2v13a2 2 0 0 0 2 2Z" />
                </svg>
                <svg v-else xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-folder-open">
                    <path
                        d="M6 5H3a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h15a2 2 0 0 0 1.94-1.5l2-8A2 2 0 0 0 20 7h-8l-2-2Z" />
                </svg>
                {{ folderToggleLabel }}
            </button>
        </div>

        <!-- Row Context Menu -->
        <div v-if="contextMenu.showRow"
            :class="menuPanelClass"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleCopyRow"
                :class="menuItemBetweenClass">
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
                :class="menuItemBetweenClass">
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
                :class="menuItemBetweenClass">
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
            <button @click="handleViewCellDetails"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-expand">
                    <path d="m15 3 6 6" />
                    <path d="M21 3h-6" />
                    <path d="M21 9V3" />
                    <path d="m9 21-6-6" />
                    <path d="M3 21h6" />
                    <path d="M3 15v6" />
                </svg>
                View Full Value
            </button>
            <button @click="handleCopyCellValueWithHeader"
                :class="plainMenuItemClass">
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
            <div :class="menuSeparatorClass"></div>
            <button @click="handleAddWhereToCondition"
                :class="plainMenuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-filter">
                    <polygon points="22 3 2 3 10 12.46 10 19 14 21 14 12.46 22 3" />
                </svg>
                Add Where To Condition
            </button>
            <div :class="menuSeparatorClass"></div>
            <div class="relative" @mouseenter="openSetValueSubmenu" @mouseleave="closeSetValueSubmenu">
                <button
                    :aria-expanded="isSetValueSubmenuOpen ? 'true' : 'false'" aria-haspopup="menu"
                    @click="toggleSetValueSubmenu" @focus="openSetValueSubmenu"
                    :class="menuItemBetweenClass">
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
                <div v-if="isSetValueSubmenuOpen" ref="setValueSubmenuRef"
                    :class="[
                        createSubmenuPanelClass,
                        'min-w-[120px]',
                        setValueSubmenuAlignLeft ? 'right-full mr-1' : 'left-full ml-1'
                    ]">
                    <button @click="handleSetNull"
                        :class="menuItemClass">
                        Set to NULL
                    </button>
                    <button @click="handleSetEmpty"
                        :class="menuItemClass">
                        Set to Empty
                    </button>
                    <button @click="handleSetDefault"
                        :class="menuItemClass">
                        Set to Default
                    </button>
                </div>
            </div>
        </div>

        <!-- Header Context Menu -->
        <div v-if="contextMenu.showHeader"
            :class="menuPanelClass"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleCopyHeaderName"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-tag">
                    <path d="M20 12V5a2 2 0 0 0-2-2h-7l-9 9 9 9 9-9Z" />
                    <line x1="7" x2="7.01" y1="7" y2="7" />
                </svg>
                Copy Column Header
            </button>
            <button @click="handleCopyHeaderRow"
                :class="menuItemClass">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-rows-3">
                    <path d="M21 6H3" />
                    <path d="M21 12H3" />
                    <path d="M21 18H3" />
                </svg>
                Copy Header Row
            </button>
        </div>

</template>

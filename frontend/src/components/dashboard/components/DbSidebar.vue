<script lang="ts" setup>
import { useI18n } from 'vue-i18n';
type RoutineType = 'PROCEDURE' | 'FUNCTION';

interface Props {
    connectionName?: string;
    dbType: string;
    activityTaskCount: number;
    isSqlNotebooksActive?: boolean;
    isAiCopilotActive?: boolean;
    tableSearch: string;
    viewSearch: string;
    storedProcedureSearch: string;
    functionSearch: string;
    filteredTables: string[];
    filteredViews: string[];
    filteredStoredProcedures: string[];
    filteredFunctions: string[];
    openFolders: string[];
}

const props = defineProps<Props>();

const emit = defineEmits<{
    'open-db-context-menu': [event: MouseEvent];
    'open-history': [];
    'open-activity-monitor': [];
    'open-sql-notebooks': [];
    'open-ai-copilot': [];
    'open-database-info': [];
    'open-settings': [];
    'toggle-folder': [folder: string];
    'open-folder-context-menu': [event: MouseEvent, folder: string];
    'select-table': [table: string];
    'open-table-context-menu': [event: MouseEvent, table: string];
    'select-view': [view: string];
    'open-view-context-menu': [event: MouseEvent, view: string];
    'select-routine': [name: string, type: RoutineType];
    'open-routine-context-menu': [event: MouseEvent, name: string, type: RoutineType];
    'disconnect': [];
    'update:table-search': [value: string];
    'update:view-search': [value: string];
    'update:stored-procedure-search': [value: string];
    'update:function-search': [value: string];
}>();
const { t } = useI18n({ useScope: 'global' });
</script>

<template>
    <div class="w-64 border-r border-border bg-card flex flex-col transition-all duration-300">
        <div class="p-4 border-b border-border flex items-center gap-2 cursor-pointer hover:bg-accent/50 transition-colors"
            @contextmenu.prevent="emit('open-db-context-menu', $event)">
                    <img src="../../../assets/images/new-icon.png" width="24" height="24" alt="Logo" class="object-contain" />
            <span class="font-semibold tracking-tight truncate flex-1" :title="connectionName || dbType || 'Database'">
                {{ connectionName || dbType || 'Database' }}
            </span>
        </div>

        <div class="px-3 py-2 border-b border-border/70 bg-muted/20">
            <div class="flex items-center justify-end gap-1">
                <button @click="emit('open-history')" :title="t('common.queryHistory.title')"
                    class="h-8 w-8 inline-flex items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-history">
                        <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                        <path d="M3 3v5h5" />
                        <path d="M12 7v5l4 2" />
                    </svg>
                </button>
                <button @click="emit('open-activity-monitor')" title="Activity Monitor"
                    class="relative h-8 w-8 inline-flex items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-activity">
                        <path d="M22 12h-4l-3 9L9 3l-3 9H2" />
                    </svg>
                    <span v-if="activityTaskCount > 0"
                        class="absolute -top-1 -right-1 inline-flex items-center justify-center min-w-4 h-4 px-1 rounded-full text-[10px] font-bold bg-primary text-primary-foreground">
                        {{ activityTaskCount }}
                    </span>
                </button>
                <button
                    @click="emit('open-sql-notebooks')"
                    title="SQL Notebooks"
                    class="h-8 w-8 inline-flex items-center justify-center rounded-md transition-colors"
                    :class="props.isSqlNotebooksActive
                        ? 'bg-accent text-foreground'
                        : 'text-muted-foreground hover:bg-accent hover:text-foreground'"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-book-open-text">
                        <path d="M12 7v14" />
                        <path d="M16 12h2" />
                        <path d="M16 8h2" />
                        <path
                            d="M3 18a2 2 0 0 1 2-2h7a4 4 0 0 1 4 4V6a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2z" />
                        <path
                            d="M21 18a2 2 0 0 0-2-2h-7a4 4 0 0 0-4 4V6a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2z" />
                    </svg>
                </button>
                <button @click="emit('open-ai-copilot')" :title="t('common.aiCopilot.title')"
                    class="h-8 w-8 inline-flex items-center justify-center rounded-md transition-colors"
                    :class="props.isAiCopilotActive
                        ? 'bg-accent text-foreground'
                        : 'text-muted-foreground hover:bg-accent hover:text-foreground'">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-sparkles">
                        <path d="M12 3l1.912 5.813L20 10.5l-6.088 1.688L12 18l-1.912-5.813L4 10.5l6.088-1.687z" />
                        <path d="M5 3v4" />
                        <path d="M19 17v4" />
                        <path d="M3 5h4" />
                        <path d="M17 19h4" />
                    </svg>
                </button>
                <button @click="emit('open-settings')" :title="t('common.settings.title')"
                    class="h-8 w-8 inline-flex items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-settings">
                        <path
                            d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
                        <circle cx="12" cy="12" r="3" />
                    </svg>
                </button>
                <button @click="emit('open-database-info')" :title="t('common.databaseInfo.title')"
                    class="h-8 w-8 inline-flex items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-info">
                        <circle cx="12" cy="12" r="10" />
                        <path d="M12 16v-4" />
                        <path d="M12 8h.01" />
                    </svg>
                </button>
            </div>
        </div>

        <div class="flex-1 overflow-y-auto px-2 py-2">
            <div class="space-y-1">
                <div>
                    <div @click="emit('toggle-folder', 'Tables')"
                        @contextmenu.prevent="emit('open-folder-context-menu', $event, 'Tables')"
                        class="flex items-center gap-2 px-2 py-1.5 text-sm font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-chevron-right transition-transform duration-200"
                            :class="{ 'rotate-90': openFolders.includes('Tables') }">
                            <path d="m9 18 6-6-6-6" />
                        </svg>
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-table text-blue-500">
                            <path d="M12 3v18" />
                            <rect width="18" height="18" x="3" y="3" rx="2" />
                            <path d="M3 9h18" />
                            <path d="M3 15h18" />
                        </svg>
                        <span>Tables</span>
                    </div>

                    <div v-show="openFolders.includes('Tables')" class="ml-4 mt-1 border-l border-border pl-2">
                        <div class="mb-2 pr-2">
                            <div class="relative">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round"
                                    class="lucide lucide-search absolute left-2 top-2 text-muted-foreground">
                                    <circle cx="11" cy="11" r="8" />
                                    <path d="m21 21-4.3-4.3" />
                                </svg>
                                <input :value="tableSearch" type="text" placeholder="Filter tables..."
                                    @input="emit('update:table-search', ($event.target as HTMLInputElement).value)"
                                    class="w-full h-8 pl-7 pr-2 rounded-md border border-input bg-background px-3 py-1 text-xs shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50">
                            </div>
                        </div>
                        <ul class="space-y-0.5">
                            <li v-for="table in filteredTables" :key="table"
                                class="flex items-center gap-2 px-2 py-1.5 text-xs rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors truncate"
                                @click="emit('select-table', table)"
                                @contextmenu.prevent="emit('open-table-context-menu', $event, table)">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-table-2 text-muted-foreground">
                                    <path
                                        d="M9 3H5a2 2 0 0 0-2 2v4m6-6h10a2 2 0 0 1 2 2v4M9 3v18m0 0h10a2 2 0 0 0 2-2V9M9 21H5a2 2 0 0 1-2-2V9m0 0h18" />
                                </svg>
                                <span class="truncate">{{ table }}</span>
                            </li>
                            <li v-if="filteredTables.length === 0"
                                class="text-xs text-muted-foreground py-2 italic ml-2">
                                No tables found.
                            </li>
                        </ul>
                    </div>
                </div>

                <div>
                    <div @click="emit('toggle-folder', 'Views')"
                        @contextmenu.prevent="emit('open-folder-context-menu', $event, 'Views')"
                        class="flex items-center gap-2 px-2 py-1.5 text-sm font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-chevron-right transition-transform duration-200"
                            :class="{ 'rotate-90': openFolders.includes('Views') }">
                            <path d="m9 18 6-6-6-6" />
                        </svg>
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-eye text-green-500">
                            <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
                            <circle cx="12" cy="12" r="3" />
                        </svg>
                        <span>Views</span>
                    </div>
                    <div v-show="openFolders.includes('Views')" class="ml-4 mt-1 border-l border-border pl-2">
                        <div class="mb-2 pr-2">
                            <div class="relative">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round"
                                    class="lucide lucide-search absolute left-2 top-2 text-muted-foreground">
                                    <circle cx="11" cy="11" r="8" />
                                    <path d="m21 21-4.3-4.3" />
                                </svg>
                                <input :value="viewSearch" type="text" placeholder="Filter views..."
                                    @input="emit('update:view-search', ($event.target as HTMLInputElement).value)"
                                    class="w-full h-8 pl-7 pr-2 rounded-md border border-input bg-background px-3 py-1 text-xs shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50">
                            </div>
                        </div>
                        <ul class="space-y-0.5">
                            <li v-for="view in filteredViews" :key="view"
                                class="flex items-center gap-2 px-2 py-1.5 text-xs rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors truncate"
                                @click="emit('select-view', view)"
                                @contextmenu.prevent="emit('open-view-context-menu', $event, view)">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-eye text-green-500">
                                    <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
                                    <circle cx="12" cy="12" r="3" />
                                </svg>
                                <span class="truncate">{{ view }}</span>
                            </li>
                            <li v-if="filteredViews.length === 0"
                                class="text-xs text-muted-foreground py-2 italic ml-2">
                                No views found.
                            </li>
                        </ul>
                    </div>
                </div>

                <div>
                    <div @click="emit('toggle-folder', 'Programmability')"
                        @contextmenu.prevent="emit('open-folder-context-menu', $event, 'Programmability')"
                        class="flex items-center gap-2 px-2 py-1.5 text-sm font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-chevron-right transition-transform duration-200"
                            :class="{ 'rotate-90': openFolders.includes('Programmability') }">
                            <path d="m9 18 6-6-6-6" />
                        </svg>
                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-code-2 text-purple-500">
                            <path d="m18 16 4-4-4-4" />
                            <path d="m6 8-4 4 4 4" />
                            <path d="m14.5 4-5 16" />
                        </svg>
                        <span>Programmability</span>
                    </div>
                    <div v-show="openFolders.includes('Programmability')" class="ml-4 mt-1 border-l border-border pl-2">
                        <div>
                            <div @click="emit('toggle-folder', 'Stored Procedures')"
                                @contextmenu.prevent="emit('open-folder-context-menu', $event, 'Stored Procedures')"
                                class="flex items-center gap-2 px-2 py-1.5 text-xs font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round"
                                    class="lucide lucide-chevron-right transition-transform duration-200"
                                    :class="{ 'rotate-90': openFolders.includes('Stored Procedures') }">
                                    <path d="m9 18 6-6-6-6" />
                                </svg>
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-scroll text-blue-400">
                                    <path d="M8 17a5 5 0 0 1 5-5c1.1 0 2 .9 2 2v6a2 2 0 0 1-4 0v-6.5" />
                                    <path d="M12 2H8a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V9a5 5 0 0 0-5-5Z" />
                                </svg>
                                <span>Stored Procedures</span>
                            </div>
                            <div v-show="openFolders.includes('Stored Procedures')"
                                class="ml-4 border-l border-border pl-2">
                                <div class="mb-2 pr-2 mt-1">
                                    <div class="relative">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-search absolute left-2 top-2 text-muted-foreground">
                                            <circle cx="11" cy="11" r="8" />
                                            <path d="m21 21-4.3-4.3" />
                                        </svg>
                                        <input :value="storedProcedureSearch" type="text"
                                            placeholder="Filter procedures..."
                                            @input="emit('update:stored-procedure-search', ($event.target as HTMLInputElement).value)"
                                            class="w-full h-8 pl-7 pr-2 rounded-md border border-input bg-background px-3 py-1 text-xs shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50">
                                    </div>
                                </div>
                                <ul class="space-y-0.5">
                                    <li v-for="sp in filteredStoredProcedures" :key="sp"
                                        class="flex items-center gap-2 px-2 py-1.5 text-xs rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors truncate"
                                        @click="emit('select-routine', sp, 'PROCEDURE')"
                                        @contextmenu.prevent="emit('open-routine-context-menu', $event, sp, 'PROCEDURE')">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-scroll-text text-muted-foreground">
                                            <path
                                                d="M8 21h12a2 2 0 0 0 2-2v-2a10 10 0 0 0-10-10H4a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h4" />
                                            <path d="M19 17V5a2 2 0 0 0-2-2H4" />
                                        </svg>
                                        <span class="truncate">{{ sp }}</span>
                                    </li>
                                    <li v-if="filteredStoredProcedures.length === 0"
                                        class="text-xs text-muted-foreground py-1 italic ml-2">
                                        No stored procedures.
                                    </li>
                                </ul>
                            </div>
                        </div>

                        <div>
                            <div @click="emit('toggle-folder', 'Functions')"
                                @contextmenu.prevent="emit('open-folder-context-menu', $event, 'Functions')"
                                class="flex items-center gap-2 px-2 py-1.5 text-xs font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round"
                                    class="lucide lucide-chevron-right transition-transform duration-200"
                                    :class="{ 'rotate-90': openFolders.includes('Functions') }">
                                    <path d="m9 18 6-6-6-6" />
                                </svg>
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-function-square text-purple-400">
                                    <rect width="18" height="18" x="3" y="3" rx="2" ry="2" />
                                    <path d="M9 17c2 0 2.8-1 2.8-2.8V10c0-2 1-3.3 3.2-3" />
                                    <path d="M9 11.2h5.7" />
                                </svg>
                                <span>Functions</span>
                            </div>
                            <div v-show="openFolders.includes('Functions')" class="ml-4 border-l border-border pl-2">
                                <div class="mb-2 pr-2 mt-1">
                                    <div class="relative">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-search absolute left-2 top-2 text-muted-foreground">
                                            <circle cx="11" cy="11" r="8" />
                                            <path d="m21 21-4.3-4.3" />
                                        </svg>
                                        <input :value="functionSearch" type="text" placeholder="Filter functions..."
                                            @input="emit('update:function-search', ($event.target as HTMLInputElement).value)"
                                            class="w-full h-8 pl-7 pr-2 rounded-md border border-input bg-background px-3 py-1 text-xs shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50">
                                    </div>
                                </div>
                                <ul class="space-y-0.5">
                                    <li v-for="fn in filteredFunctions" :key="fn"
                                        class="flex items-center gap-2 px-2 py-1.5 text-xs rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors truncate"
                                        @click="emit('select-routine', fn, 'FUNCTION')"
                                        @contextmenu.prevent="emit('open-routine-context-menu', $event, fn, 'FUNCTION')">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round"
                                            class="lucide lucide-braces text-muted-foreground">
                                            <path
                                                d="M8 3H7a2 2 0 0 0-2 2v5a2 2 0 0 1-2 2 2 2 0 0 1 2 2v5c0 1.1.9 2 2 2h1" />
                                            <path
                                                d="M16 21h1a2 2 0 0 0 2-2v-5c0-1.1.9-2 2-2a2 2 0 0 1-2-2V5a2 2 0 0 0-2-2h-1" />
                                        </svg>
                                        <span class="truncate">{{ fn }}</span>
                                    </li>
                                    <li v-if="filteredFunctions.length === 0"
                                        class="text-xs text-muted-foreground py-1 italic ml-2">
                                        No functions.
                                    </li>
                                </ul>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div class="p-4 border-t border-border">
            <button @click="emit('disconnect')"
                class="w-full flex items-center justify-center gap-2 text-destructive hover:bg-destructive/10 hover:text-destructive text-sm font-medium py-2 px-4 rounded-md transition-colors">
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
    </div>
</template>

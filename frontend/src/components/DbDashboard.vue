<template>
    <div class="flex h-full bg-background text-foreground font-sans">
        <!-- Sidebar -->
        <div class="w-64 border-r border-border bg-card flex flex-col transition-all duration-300">
            <div class="p-4 border-b border-border flex items-center gap-2 cursor-pointer hover:bg-accent/50 transition-colors"
                @contextmenu.prevent="openDbContextMenu($event)">
                <div class="h-6 w-6 rounded bg-primary flex items-center justify-center text-primary-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-database">
                        <ellipse cx="12" cy="5" rx="9" ry="3" />
                        <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                        <path d="M3 12A9 3 0 0 0 21 12" />
                    </svg>
                </div>
                <span class="font-semibold tracking-tight truncate flex-1"
                    :title="connectionName || dbType || 'Database'">
                    {{ connectionName || dbType || 'Database' }}
                </span>
            </div>

            <!-- Removed top search bar, moved inside Tables folder -->

            <div class="flex-1 overflow-y-auto px-2 py-2">
                <div class="space-y-1">
                    <!-- Tables Folder -->
                    <div>
                        <div @click="toggleFolder('Tables')"
                            @contextmenu.prevent="openFolderContextMenu($event, 'Tables')"
                            class="flex items-center gap-2 px-2 py-1.5 text-sm font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round"
                                class="lucide lucide-chevron-right transition-transform duration-200"
                                :class="{ 'rotate-90': openFolders.includes('Tables') }">
                                <path d="m9 18 6-6-6-6" />
                            </svg>
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-table text-blue-500">
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
                                    <input v-model="tableSearch" type="text" placeholder="Filter tables..."
                                        class="w-full h-8 pl-7 pr-2 rounded-md border border-input bg-background px-3 py-1 text-xs shadow-sm transition-colors file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50">
                                </div>
                            </div>

                            <ul class="space-y-0.5">
                                <li v-for="table in filteredTables" :key="table"
                                    class="flex items-center gap-2 px-2 py-1.5 text-xs rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors truncate"
                                    @click="selectTable(table)" @contextmenu.prevent="openContextMenu($event, table)">
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

                    <!-- Views Folder -->
                    <div>
                        <div @click="toggleFolder('Views')"
                            @contextmenu.prevent="openFolderContextMenu($event, 'Views')"
                            class="flex items-center gap-2 px-2 py-1.5 text-sm font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round"
                                class="lucide lucide-chevron-right transition-transform duration-200"
                                :class="{ 'rotate-90': openFolders.includes('Views') }">
                                <path d="m9 18 6-6-6-6" />
                            </svg>
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-eye text-green-500">
                                <path d="M2 12s3-7 10-7 10 7 10 7-3 7-10 7-10-7-10-7Z" />
                                <circle cx="12" cy="12" r="3" />
                            </svg>
                            <span>Views</span>
                        </div>
                        <div v-show="openFolders.includes('Views')" class="ml-4 mt-1 border-l border-border pl-2">
                            <ul class="space-y-0.5">
                                <li v-for="view in filteredViews" :key="view"
                                    class="flex items-center gap-2 px-2 py-1.5 text-xs rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors truncate"
                                    @click="selectView(view)" @contextmenu.prevent="openViewContextMenu($event, view)">
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

                    <!-- Programmability Folder -->
                    <div>
                        <div @click="toggleFolder('Programmability')"
                            @contextmenu.prevent="openFolderContextMenu($event, 'Programmability')"
                            class="flex items-center gap-2 px-2 py-1.5 text-sm font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round"
                                class="lucide lucide-chevron-right transition-transform duration-200"
                                :class="{ 'rotate-90': openFolders.includes('Programmability') }">
                                <path d="m9 18 6-6-6-6" />
                            </svg>
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-code-2 text-purple-500">
                                <path d="m18 16 4-4-4-4" />
                                <path d="m6 8-4 4 4 4" />
                                <path d="m14.5 4-5 16" />
                            </svg>
                            <span>Programmability</span>
                        </div>
                        <div v-show="openFolders.includes('Programmability')"
                            class="ml-4 mt-1 border-l border-border pl-2">
                            <!-- Stored Procedures -->
                            <div>
                                <div @click="toggleFolder('Stored Procedures')"
                                    @contextmenu.prevent="openFolderContextMenu($event, 'Stored Procedures')"
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
                                        <path
                                            d="M12 2H8a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V9a5 5 0 0 0-5-5Z" />
                                    </svg>
                                    <span>Stored Procedures</span>
                                </div>
                                <div v-show="openFolders.includes('Stored Procedures')"
                                    class="ml-4 border-l border-border pl-2">
                                    <ul class="space-y-0.5">
                                        <li v-for="sp in filteredStoredProcedures" :key="sp"
                                            class="flex items-center gap-2 px-2 py-1.5 text-xs rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors truncate"
                                            @click="selectRoutine(sp, 'PROCEDURE')"
                                            @contextmenu.prevent="openRoutineContextMenu($event, sp, 'PROCEDURE')">
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

                            <!-- Functions -->
                            <div>
                                <div @click="toggleFolder('Functions')"
                                    @contextmenu.prevent="openFolderContextMenu($event, 'Functions')"
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
                                <div v-show="openFolders.includes('Functions')"
                                    class="ml-4 border-l border-border pl-2">
                                    <ul class="space-y-0.5">
                                        <li v-for="fn in filteredFunctions" :key="fn"
                                            class="flex items-center gap-2 px-2 py-1.5 text-xs rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors truncate"
                                            @click="selectRoutine(fn, 'FUNCTION')"
                                            @contextmenu.prevent="openRoutineContextMenu($event, fn, 'FUNCTION')">
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

            <!-- Settings and Disconnect Buttons -->
            <div class="p-4 border-t border-border flex flex-col gap-2">
                <button @click="isHistoryOpen = true"
                    class="w-full flex items-center justify-center gap-2 text-muted-foreground hover:bg-muted/50 hover:text-foreground text-sm font-medium py-2 px-4 rounded-md transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-history">
                        <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                        <path d="M3 3v5h5" />
                        <path d="M12 7v5l4 2" />
                    </svg>
                    History
                </button>
                <button @click="isSettingsOpen = true"
                    class="w-full flex items-center justify-center gap-2 text-muted-foreground hover:bg-muted/50 hover:text-foreground text-sm font-medium py-2 px-4 rounded-md transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-settings">
                        <path
                            d="M12.22 2h-.44a2 2 0 0 0-2 2v.18a2 2 0 0 1-1 1.73l-.43.25a2 2 0 0 1-2 0l-.15-.08a2 2 0 0 0-2.73.73l-.22.38a2 2 0 0 0 .73 2.73l.15.1a2 2 0 0 1 1 1.72v.51a2 2 0 0 1-1 1.74l-.15.09a2 2 0 0 0-.73 2.73l.22.38a2 2 0 0 0 2.73.73l.15-.08a2 2 0 0 1 2 0l.43.25a2 2 0 0 1 1 1.73V20a2 2 0 0 0 2 2h.44a2 2 0 0 0 2-2v-.18a2 2 0 0 1 1-1.73l.43-.25a2 2 0 0 1 2 0l.15.08a2 2 0 0 0 2.73-.73l.22-.39a2 2 0 0 0-.73-2.73l-.15-.08a2 2 0 0 1-1-1.74v-.5a2 2 0 0 1 1-1.74l.15-.09a2 2 0 0 0 .73-2.73l-.22-.38a2 2 0 0 0-2.73-.73l-.15.08a2 2 0 0 1-2 0l-.43-.25a2 2 0 0 1-1-1.73V4a2 2 0 0 0-2-2z" />
                        <circle cx="12" cy="12" r="3" />
                    </svg>
                    Settings
                </button>
                <button @click="disconnect"
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

            <!-- Query Area -->
            <div v-if="activeTab" class="flex flex-col h-full overflow-hidden query-area-container">
                <div v-if="!activeTab.isERView && !activeTab.isDesignView"
                    class="flex flex-col border-b border-border bg-card p-4 gap-3 relative shrink-0 min-h-[0px]"
                    :style="{ height: activeTab.editorHeight + 'px' }">
                    <div class="relative w-full flex-1 min-h-0">
                        <SqlEditor ref="sqlEditorRef" v-model="activeTab.query" :tables="tables"
                            :get-columns="fetchTableColumns" :font-family="editorSettings.fontFamily"
                            :font-size="editorSettings.fontSize" />

                        <!-- Char count overlay -->
                        <div class="absolute bottom-1 right-3 z-10 flex items-center gap-2 pointer-events-none">
                            <div
                                class="text-xs text-muted-foreground bg-background/80 px-2 py-1 rounded backdrop-blur-sm border border-border pointer-events-auto">
                                {{ activeTab.query.length }} chars
                            </div>
                        </div>
                    </div>

                    <div class="flex items-center justify-between">
                        <div class="flex items-center gap-4 text-xs text-muted-foreground">
                            <div v-if="activeTab.isLoading && !activeTab.executionTime"
                                class="flex items-center gap-2 text-primary">
                                <svg class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none"
                                    viewBox="0 0 24 24">
                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                        stroke-width="4"></circle>
                                    <path class="opacity-75" fill="currentColor"
                                        d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z">
                                    </path>
                                </svg>
                                Executing...
                            </div>
                            <div v-else-if="activeTab.executionTime !== undefined" class="flex items-center gap-1.5 ">
                                <span class="flex items-center gap-1" title="Execution Time (Database)">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                        fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                        stroke-linejoin="round" class="lucide lucide-database">
                                        <ellipse cx="12" cy="5" rx="9" ry="3" />
                                        <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
                                        <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
                                    </svg>
                                    <span>Exec: {{ activeTab.executionTime }}ms</span>
                                </span>
                                <span class="text-border mx-1">|</span>
                                <span class="flex items-center gap-1" title="Fetch/Transfer Time">
                                    <div v-if="activeTab.isLoading" class="flex items-center gap-1">
                                        <svg class="animate-spin h-3 w-3 text-primary"
                                            xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                                stroke-width="4"></circle>
                                            <path class="opacity-75" fill="currentColor"
                                                d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z">
                                            </path>
                                        </svg>
                                    </div>
                                    <svg v-else xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-arrow-down">
                                        <line x1="12" x2="12" y1="5" y2="19" />
                                        <polyline points="19 12 12 19 5 12" />
                                    </svg>
                                    <span>Fetch: {{ activeTab.fetchTime !== undefined ? activeTab.fetchTime : '...'
                                        }}{{ activeTab.isLoading ? '...' : 'ms' }}</span>
                                </span>
                            </div>
                        </div>

                    </div>

                    <div class="flex items-center gap-2">
                        <div v-if="isReadOnly"
                            class="px-2 py-1 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-500 text-xs rounded border border-yellow-200 dark:border-yellow-900/50 mr-2 flex items-center gap-1 select-none cursor-help"
                            title="Database is in Read-Only mode. Modifications are disabled.">
                            <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-lock">
                                <rect width="18" height="11" x="3" y="11" rx="2" ry="2" />
                                <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                            </svg>
                            Read Only
                        </div>

                        <!-- <button @click="analyzeQuery" :disabled="activeTab.isLoading || !activeTab.query"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-4 py-2 shadow-sm"
                            title="Analyze Query (Explain)">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-microscope mr-2">
                                <path d="M6 18h8" />
                                <path d="M3 22h18" />
                                <path d="M14 22a7 7 0 1 0 0-14h-1" />
                                <path d="M9 14h2" />
                                <path d="M9 12a2 2 0 0 1-2-2V6h6v4a2 2 0 0 1-2 2Z" />
                                <path d="M12 6V3a1 1 0 0 0-1-1H9a1 1 0 0 0-1 1v3" />
                            </svg>
                            Analyze
                        </button> -->

                        <button @click="beautifyQuery" :disabled="activeTab.isLoading || !activeTab.query"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-4 py-2 shadow-sm"
                            title="Format SQL (Shift + Alt + F)">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-wrap-text mr-2">
                                <line x1="3" x2="21" y1="6" y2="6" />
                                <path d="M3 12h15a3 3 0 1 1 0 6h-4" />
                                <polyline points="16 16 14 18 16 20" />
                            </svg>
                            Beautify
                        </button>

                        <button v-if="activeTab.isRoutine" @click="handleSaveRoutine" :disabled="activeTab.isLoading"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-primary/50 bg-primary/10 text-primary hover:bg-primary/20 h-9 px-4 py-2 shadow-sm"
                            title="Save / Update Routine">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-save mr-2">
                                <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
                                <polyline points="17 21 17 13 7 13 7 21" />
                                <polyline points="7 3 7 8 15 8" />
                            </svg>
                            Save
                        </button>

                        <button v-if="activeTab.isLoading" @click="stopQuery"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 bg-destructive text-destructive-foreground hover:bg-destructive/90 h-9 px-4 py-2 shadow-sm">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-square mr-2 fill-current">
                                <rect width="18" height="18" x="3" y="3" rx="2" />
                            </svg>
                            Stop
                        </button>

                        <button v-else @click="runQuery"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-9 px-4 py-2 shadow-sm">
                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-play mr-2">
                                <polygon points="5 3 19 12 5 21 5 3" />
                            </svg>
                            Run Query
                        </button>
                    </div>
                </div>

                <!-- Resizer Handle -->
                <div v-if="!activeTab.isERView && !activeTab.isDesignView"
                    class="h-1.5 hover:bg-primary/30 cursor-row-resize flex items-center justify-center transition-colors group z-20 shrink-0"
                    @mousedown="startResizing">
                    <div class="w-8 h-1 bg-border rounded-full group-hover:bg-primary/50"></div>
                </div>

                <!-- Results Area -->
                <div v-if="!activeTab.isERView && !activeTab.isDesignView"
                    class="flex-1 overflow-hidden bg-muted/10 flex flex-col">

                    <!-- Data/Messages Sub-Tabs -->
                    <div v-if="activeTab.queryExecuted || activeTab.error"
                        class="flex items-center border-b border-border bg-muted/20 px-2 pt-1 gap-0.5 shrink-0">
                        <button @click="activeTab.resultViewTab = 'data'"
                            class="relative px-4 py-1.5 text-xs font-medium rounded-t-md transition-all select-none border-l border-r border-t border-transparent"
                            :class="activeTab.resultViewTab === 'data' ? 'bg-background text-foreground border-border shadow-sm mb-[-1px]' : 'text-muted-foreground hover:text-foreground hover:bg-background/50'">
                            <div class="flex items-center gap-1.5">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-table-2">
                                    <path
                                        d="M9 3H5a2 2 0 0 0-2 2v4m6-6h10a2 2 0 0 1 2 2v4M9 3v18m0 0h10a2 2 0 0 0 2-2V9M9 21H5a2 2 0 0 1-2-2V9m0 0h18" />
                                </svg>
                                Results
                            </div>
                            <div v-if="activeTab.resultViewTab === 'data'"
                                class="absolute top-0 left-0 right-0 h-0.5 bg-primary rounded-t-full"></div>
                        </button>
                        <button @click="activeTab.resultViewTab = 'messages'"
                            class="relative px-4 py-1.5 text-xs font-medium rounded-t-md transition-all select-none border-l border-r border-t border-transparent"
                            :class="activeTab.resultViewTab === 'messages' ? 'bg-background text-foreground border-border shadow-sm mb-[-1px]' : 'text-muted-foreground hover:text-foreground hover:bg-background/50'">
                            <div class="flex items-center gap-1.5">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-message-square-text">
                                    <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                                    <path d="M13 8H7" />
                                    <path d="M17 12H7" />
                                </svg>
                                Messages
                            </div>
                            <div v-if="activeTab.resultViewTab === 'messages'"
                                class="absolute top-0 left-0 right-0 h-0.5 bg-primary rounded-t-full"></div>
                        </button>
                        <button @click="activeTab.resultViewTab = 'analysis'" v-if="activeTab.explanation"
                            class="relative px-4 py-1.5 text-xs font-medium rounded-t-md transition-all select-none border-l border-r border-t border-transparent"
                            :class="activeTab.resultViewTab === 'analysis' ? 'bg-background text-foreground border-border shadow-sm mb-[-1px]' : 'text-muted-foreground hover:text-foreground hover:bg-background/50'">
                            <div class="flex items-center gap-1.5">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-microscope">
                                    <path d="M6 18h8" />
                                    <path d="M3 22h18" />
                                    <path d="M14 22a7 7 0 1 0 0-14h-1" />
                                    <path d="M9 14h2" />
                                    <path d="M9 12a2 2 0 0 1-2-2V6h6v4a2 2 0 0 1-2 2Z" />
                                    <path d="M12 6V3a1 1 0 0 0-1-1H9a1 1 0 0 0-1 1v3" />
                                </svg>
                                Analysis
                            </div>
                            <div v-if="activeTab.resultViewTab === 'analysis'"
                                class="absolute top-0 left-0 right-0 h-0.5 bg-primary rounded-t-full"></div>
                        </button>
                    </div>

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
                            <div class="flex-1 text-sm font-medium break-all font-mono">{{ activeTab.error }}</div>
                        </div>

                        <!-- Results List (Multiple Sets) -->
                        <div v-else-if="activeTab.resultSets && activeTab.resultSets.length > 0"
                            class="flex flex-col gap-4 h-full">

                            <!-- Primary Result Set (Virtual List) -->
                            <div v-if="activeTab.resultSets[0]"
                                class="flex-1 border border-border rounded-lg shadow-sm bg-card flex flex-col min-h-[0px] overflow-hidden">

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
                                                    :class="{ 'bg-accent/50': activeTab.editingCell && activeTab.editingCell.rowId === item.index && activeTab.editingCell.col === col }"
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
                                        <button
                                            v-if="activeTab.tableName && !props.isReadOnly && activeTab.primaryKeys.length > 0"
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
                            </div>

                            <!-- Subsequent Result Sets (Standard Tables) -->
                            <div v-for="(resultSet, rsIndex) in activeTab.resultSets.slice(1)" :key="rsIndex + 1"
                                class="flex-1 border border-border rounded-lg shadow-sm bg-card flex flex-col min-h-[0px] overflow-hidden">

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
                                                    class="px-4 py-2 whitespace-nowrap text-foreground font-mono text-xs border-r border-transparent hover:border-border overflow-hidden"
                                                    :style="{ width: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', minWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px', maxWidth: activeTab.columnWidths[col] ? activeTab.columnWidths[col] + 'px' : '150px' }"
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
                    <div v-if="activeTab.resultViewTab === 'analysis' && activeTab.explanation"
                        class="flex-1 overflow-auto p-4">
                        <div
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
                    </div>
                </div>

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
                        :connection-id="props.connectionId" @close="closeTab(activeTab.id)" @refresh="loadTables" />
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
        <!-- Context Menu for Tables -->
        <div v-if="contextMenu.show"
            class="fixed z-50 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[160px]"
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
                View ER Diagram
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
            class="fixed z-50 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[160px]"
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
            class="fixed z-50 min-w-[160px] bg-popover text-popover-foreground rounded-md border border-border shadow-md py-1 animate-in fade-in zoom-in-95 duration-100"
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
            class="fixed z-50 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[160px]"
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
            class="fixed z-50 min-w-[160px] bg-popover text-popover-foreground rounded-md border border-border shadow-md py-1 animate-in fade-in zoom-in-95 duration-100"
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
            class="fixed z-50 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[160px]"
            :style="{ top: `${contextMenu.position.y}px`, left: `${contextMenu.position.x}px` }">
            <button @click="handleCopyRow"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-copy">
                    <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
                    <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
                </svg>
                Copy Row
            </button>
            <button @click="handleCopyCellValue"
                class="w-full text-left px-3 py-1.5 text-sm hover:bg-accent hover:text-accent-foreground flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-clipboard-copy">
                    <rect width="8" height="4" x="8" y="2" rx="1" ry="1" />
                    <path d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2" />
                </svg>
                Copy Cell Value
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

        <!-- Update Confirmation Modal -->
        <div v-if="updateConfirmation && updateConfirmation.isOpen"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div
                class="bg-card w-full max-w-md rounded-lg shadow-lg border border-border p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200">
                <div class="flex items-center justify-between">
                    <h3 class="text-lg font-semibold text-foreground">Confirm Update</h3>
                    <button @click="cancelUpdate" class="text-muted-foreground hover:text-foreground">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-x">
                            <path d="M18 6 6 18" />
                            <path d="m6 6 12 12" />
                        </svg>
                    </button>
                </div>

                <p class="text-sm text-muted-foreground">
                    Are you sure you want to update data in table <span class="font-medium text-foreground">{{
                        updateConfirmation.tableName }}</span>?
                </p>

                <div class="bg-muted/50 p-3 rounded-md space-y-2 text-sm">
                    <div class="flex flex-col gap-1">
                        <span class="text-xs font-semibold text-muted-foreground uppercase">Column</span>
                        <span class="font-mono text-foreground">{{ updateConfirmation.column }}</span>
                    </div>
                    <div class="grid grid-cols-2 gap-4">
                        <div class="flex flex-col gap-1">
                            <span class="text-xs font-semibold text-muted-foreground uppercase">Old Value</span>
                            <div
                                class="font-mono text-destructive text-xs max-h-48 overflow-y-auto whitespace-pre-wrap break-words border border-destructive/20 bg-destructive/5 p-2 rounded">
                                {{ updateConfirmation.originalValue === null ? 'NULL' : updateConfirmation.originalValue
                                }}
                            </div>
                        </div>
                        <div class="flex flex-col gap-1">
                            <span class="text-xs font-semibold text-muted-foreground uppercase">New Value</span>
                            <div
                                class="font-mono text-green-600 dark:text-green-500 text-xs max-h-48 overflow-y-auto whitespace-pre-wrap break-words border border-green-500/20 bg-green-500/5 p-2 rounded">
                                {{ formatValueForDisplay(updateConfirmation.newValue) }}
                            </div>
                        </div>
                    </div>
                </div>

                <div class="flex justify-end gap-3 pt-2">
                    <button @click="cancelUpdate"
                        class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                        Cancel
                    </button>
                    <button @click="confirmUpdate"
                        class="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm">
                        Confirm Update
                    </button>
                </div>
            </div>
        </div>

        <!-- Insert Row Modal -->
        <div v-if="insertRowModal && insertRowModal.isOpen"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm"
            @mousedown.self="cancelInsertRow">
            <div @keydown.stop @mousedown.stop
                class="bg-card w-full max-w-lg rounded-lg shadow-lg border border-border p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200 max-h-[80vh] flex flex-col">
                <div class="flex items-center justify-between">
                    <h3 class="text-lg font-semibold text-foreground flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-plus-circle text-primary">
                            <circle cx="12" cy="12" r="10" />
                            <path d="M8 12h8" />
                            <path d="M12 8v8" />
                        </svg>
                        Insert Row
                    </h3>
                    <button @click="cancelInsertRow" class="text-muted-foreground hover:text-foreground">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-x">
                            <path d="M18 6 6 18" />
                            <path d="m6 6 12 12" />
                        </svg>
                    </button>
                </div>

                <p class="text-sm text-muted-foreground">
                    Insert a new row into <span class="font-medium text-foreground">{{ insertRowModal.tableName
                    }}</span>
                </p>

                <div class="flex-1 overflow-y-auto space-y-3 pr-1">
                    <div v-for="col in insertRowModal.columns" :key="col" class="flex flex-col gap-1">
                        <div class="flex items-center justify-between">
                            <label
                                class="text-xs font-semibold text-muted-foreground uppercase flex items-center gap-1">
                                {{ col }}
                                <span v-if="getColDef(col)?.primaryKey"
                                    class="text-[9px] px-1 py-0.5 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-500 rounded font-normal normal-case">PK</span>
                                <span v-if="getColDef(col)?.autoIncrement"
                                    class="text-[9px] px-1 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded font-normal normal-case">Auto</span>
                                <span v-if="getColDef(col)"
                                    class="text-[9px] px-1 py-0.5 bg-muted text-muted-foreground rounded font-normal normal-case">{{
                                        getColDef(col)?.type }}</span>
                            </label>
                            <button @click="toggleInsertNull(col)"
                                class="text-[10px] px-1.5 py-0.5 rounded transition-colors"
                                :class="insertRowModal.nullColumns[col] ? 'bg-muted text-foreground' : 'text-muted-foreground hover:bg-muted/50'">
                                NULL
                            </button>
                        </div>
                        <!-- Type-aware inputs -->
                        <template v-if="!insertRowModal.nullColumns[col]">
                            <!-- Boolean/bit -->
                            <div v-if="getInputType(col) === 'checkbox'" class="flex items-center gap-2 h-8">
                                <input type="checkbox" v-model="insertRowModal.values[col]" :true-value="'1'"
                                    :false-value="'0'"
                                    class="insert-row-input w-4 h-4 rounded border-input bg-background text-primary focus:ring-ring" />
                                <span class="text-sm text-muted-foreground">{{ insertRowModal.values[col] === '1' ?
                                    'True' : 'False' }}</span>
                            </div>
                            <!-- Date -->
                            <input v-else-if="getInputType(col) === 'date'" v-model="insertRowModal.values[col]"
                                type="date"
                                class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono" />
                            <!-- DateTime -->
                            <input v-else-if="getInputType(col) === 'datetime-local'"
                                v-model="insertRowModal.values[col]" type="datetime-local" step="1"
                                class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono" />
                            <!-- Time -->
                            <input v-else-if="getInputType(col) === 'time'" v-model="insertRowModal.values[col]"
                                type="time" step="1"
                                class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono" />
                            <!-- Number -->
                            <input v-else-if="getInputType(col) === 'number'" v-model="insertRowModal.values[col]"
                                type="number" :step="getNumberStep(col)" :placeholder="`Enter ${col}`"
                                class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono" />
                            <!-- Textarea for text/blob -->
                            <textarea v-else-if="getInputType(col) === 'textarea'" v-model="insertRowModal.values[col]"
                                :placeholder="`Enter ${col}`" rows="3"
                                class="insert-row-input w-full px-3 py-2 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono resize-y" />
                            <!-- Default text input -->
                            <input v-else v-model="insertRowModal.values[col]" type="text" :placeholder="`Enter ${col}`"
                                class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono" />
                        </template>
                        <div v-else
                            class="w-full h-8 px-3 rounded-md border border-input bg-muted/50 text-sm text-muted-foreground flex items-center italic font-mono">
                            NULL
                        </div>
                    </div>
                </div>

                <div v-if="insertRowModal.error"
                    class="text-sm text-destructive bg-destructive/10 p-2 rounded-md break-all">
                    {{ insertRowModal.error }}
                </div>

                <div class="flex justify-end gap-3 pt-2">
                    <button @click="cancelInsertRow"
                        class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                        Cancel
                    </button>
                    <button @click="confirmInsertRow" :disabled="insertRowModal.isInserting"
                        class="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm disabled:opacity-50 inline-flex items-center gap-2">
                        <svg v-if="insertRowModal.isInserting" class="animate-spin h-3 w-3"
                            xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                            <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4">
                            </circle>
                            <path class="opacity-75" fill="currentColor" d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z">
                            </path>
                        </svg>
                        Insert
                    </button>
                </div>
            </div>
        </div>



        <!-- Import Options Modal -->
        <div v-if="showImportOptions"
            class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm">
            <div
                class="bg-card w-full max-w-md rounded-lg shadow-lg border border-border p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200">
                <div class="flex items-center justify-between">
                    <h3 class="text-lg font-semibold text-foreground">Import Options</h3>
                    <button @click="showImportOptions = false" class="text-muted-foreground hover:text-foreground">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-x">
                            <path d="M18 6 6 18" />
                            <path d="m6 6 12 12" />
                        </svg>
                    </button>
                </div>

                <div class="space-y-4">
                    <div class="flex flex-col gap-1 text-sm">
                        <span class="font-semibold text-muted-foreground">Target Table:</span>
                        <span>{{ importOptions.tableName }}</span>
                    </div>
                    <div class="flex flex-col gap-1 text-sm">
                        <span class="font-semibold text-muted-foreground">File:</span>
                        <span class="truncate" :title="importOptions.filePath">{{ importOptions.filePath }}</span>
                    </div>

                    <div v-if="props.dbType === 'mssql'" class="flex items-center space-x-2 pt-2">
                        <input type="checkbox" id="identityInsert" v-model="importOptions.enableIdentityInsert"
                            class="h-4 w-4 rounded border-input bg-background text-primary focus:ring-primary">
                        <label for="identityInsert"
                            class="text-sm font-medium leading-none peer-disabled:cursor-not-allowed peer-disabled:opacity-70">
                            Enable Identity Insert (SET IDENTITY_INSERT ON)
                        </label>
                    </div>
                </div>

                <div class="flex justify-end gap-3 pt-4">
                    <button @click="showImportOptions = false"
                        class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                        Cancel
                    </button>
                    <button @click="confirmImport"
                        class="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm">
                        Import
                    </button>
                </div>
            </div>
        </div>
        <Toast ref="toastRef" />
        <SettingsDialog :is-open="isSettingsOpen" @close="isSettingsOpen = false" @save="handleSettingsSave" />
        <QueryHistory :is-open="isHistoryOpen" :connection-name="connectionName" @close="isHistoryOpen = false"
            @run-query="handleRunHistoryQuery" />

        <!-- Image Preview Modal -->
        <div v-if="imagePreviewUrl"
            class="fixed inset-0 z-[100] flex items-center justify-center bg-black/80 backdrop-blur-sm p-10"
            @click="imagePreviewUrl = null">
            <div class="relative max-w-full max-h-full flex items-center justify-center animate-in zoom-in-95 duration-200"
                @click.stop>
                <img :src="imagePreviewUrl"
                    class="max-w-full max-h-[85vh] rounded-lg shadow-2xl object-contain border-4 border-background bg-card" />
                <button @click="imagePreviewUrl = null"
                    class="absolute -top-4 -right-4 h-10 w-10 rounded-full bg-background border border-border shadow-lg flex items-center justify-center hover:bg-muted transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
                <div class="absolute -bottom-10 left-0 right-0 text-center">
                    <a :href="imagePreviewUrl" target="_blank"
                        class="text-white hover:text-primary transition-colors text-sm font-medium underline underline-offset-4">Open
                        in browser</a>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref, reactive, watch, onMounted, computed, shallowRef, nextTick, markRaw, onUnmounted } from 'vue';
import { GetTables, GetViews, GetStoredProcedures, GetFunctions, ExecuteQuery, DisconnectDB, GetPrimaryKeys, GetForeignKeys, GetRoutineDefinition, UpdateRecord, ExportTable, ImportTable, SelectExportFile, SelectImportFile, CancelQuery, ExecuteQueryStream, ExplainQuery, ExecuteTransientQuery, GetTableDefinition, SaveQueryHistory } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';
import { format } from 'sql-formatter';
import { useVirtualList } from '@vueuse/core';

// Components
import SqlEditor from './SqlEditor.vue';
import ERDiagram from './ERDiagram.vue';
import TableStructureDesigner from './TableStructureDesigner.vue';
import Toast from './Toast.vue';
import SettingsDialog from './SettingsDialog.vue';
import QueryHistory from './QueryHistory.vue';
import DatePicker from './DatePicker.vue';
import { TooltipProvider, TooltipRoot, TooltipTrigger, TooltipContent } from 'radix-vue';

// Composables
import { isDarkTheme } from '../composables/useTheme';
import { useTabs } from '../composables/useTabs';
import { useSidebar } from '../composables/useSidebar';
import { useRecordOperations } from '../composables/useRecordOperations';

// Types
import { QueryTab, ContextMenuState } from '../types/dashboard';
import { ColumnMetadata, ResultSet } from '../types/database';

const toastRef = ref<InstanceType<typeof Toast> | null>(null);
const isSettingsOpen = ref(false);
const isHistoryOpen = ref(false);
const imagePreviewUrl = ref<string | null>(null);

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

// --- Composable: Tabs ---
const {
    tabs,
    activeTabId,
    activeTab,
    addTab,
    closeTab,
    generateId
} = useTabs();

// --- Composable: Sidebar ---
const {
    tableSearch,
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
} = useSidebar(props.connectionId);

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
    props.connectionId,
    props.isReadOnly || false,
    activeTab,
    () => runQuery(), // Wrapped in arrow because runQuery is hoisted or defined later
    toastRef
);

// Other local state
const sqlEditorRef = ref<any>(null);
const selectedRowIndex = ref<number | string | null>(null);
const tableSchemas = ref<Record<string, string[]>>({});


// Resizing State
const isResizing = ref(false);
const isColumnResizing = ref(false);
const resizingColumn = ref<string | null>(null);
const startX = ref(0);
const startWidth = ref(0);

const startResizing = (e: MouseEvent) => {
    isResizing.value = true;
    document.addEventListener('mousemove', doResize);
    document.addEventListener('mouseup', stopResizing);
    document.body.style.cursor = 'row-resize';
};

const doResize = (e: MouseEvent) => {
    if (!isResizing.value || !activeTab.value) return;

    const queryArea = document.querySelector('.query-area-container');
    if (queryArea) {
        const rect = queryArea.getBoundingClientRect();
        const newHeight = e.clientY - rect.top;

        // Min 120px to keep buttons visible, Max leave 150px for results
        if (newHeight >= 120 && newHeight <= window.innerHeight - 150) {
            activeTab.value.editorHeight = newHeight;
        }
    }
};

const stopResizing = () => {
    isResizing.value = false;
    document.removeEventListener('mousemove', doResize);
    document.removeEventListener('mouseup', stopResizing);
    document.body.style.cursor = '';
};

// Column Resizing logic
const startColumnResize = (e: MouseEvent, col: string) => {
    if (!activeTab.value) return;

    isColumnResizing.value = true;
    resizingColumn.value = col;
    startX.value = e.clientX;

    // Default width if not set
    if (!activeTab.value.columnWidths[col]) {
        // Find the th element to get its current width
        const th = (e.target as HTMLElement).closest('th');
        startWidth.value = th ? th.offsetWidth : 150;
    } else {
        startWidth.value = activeTab.value.columnWidths[col];
    }

    document.addEventListener('mousemove', doColumnResize);
    document.addEventListener('mouseup', stopColumnResize);
    document.body.style.cursor = 'col-resize';
    e.preventDefault();
    e.stopPropagation();
};

const doColumnResize = (e: MouseEvent) => {
    if (!isColumnResizing.value || !activeTab.value || !resizingColumn.value) return;

    const deltaX = e.clientX - startX.value;
    const newWidth = Math.max(50, startWidth.value + deltaX);

    activeTab.value.columnWidths[resizingColumn.value] = newWidth;
};

const stopColumnResize = () => {
    isColumnResizing.value = false;
    resizingColumn.value = null;
    document.removeEventListener('mousemove', doColumnResize);
    document.removeEventListener('mouseup', stopColumnResize);
    document.body.style.cursor = '';
};


// Sidebar filtered computed items are now in useSidebar

// Context Menu State
const contextMenu = reactive({
    showDb: false,
    showFolder: false,
    show: false,
    showRow: false,
    showView: false,
    showRoutine: false,
    position: { x: 0, y: 0 },
    targetTable: '',
    targetRow: null as any,
    targetColumn: '',
    targetFolder: '',
    targetView: '',
    targetRoutine: '',
    targetRoutineType: 'PROCEDURE' as 'PROCEDURE' | 'FUNCTION'
});

const openDbContextMenu = (event: MouseEvent) => {
    closeContextMenu();
    contextMenu.position = { x: event.clientX, y: event.clientY };
    contextMenu.showDb = true;
};

const refreshDatabase = async () => {
    closeContextMenu();
    await loadTables();
    if (toastRef.value) {
        toastRef.value.success('Database refreshed successfully.');
    }
};

const openFolderContextMenu = (event: MouseEvent, folderName: string) => {
    closeContextMenu();
    contextMenu.targetFolder = folderName;
    contextMenu.position = { x: event.clientX, y: event.clientY };
    contextMenu.showFolder = true;
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

const openContextMenu = (event: MouseEvent, table: string) => {
    closeContextMenu();
    contextMenu.targetTable = table;
    contextMenu.position = { x: event.clientX, y: event.clientY };
    contextMenu.show = true;
};

const handleRowContextMenu = (event: MouseEvent, row: any, col: string) => {
    closeContextMenu();
    contextMenu.targetRow = row;
    contextMenu.targetColumn = col;
    contextMenu.position = { x: event.clientX, y: event.clientY };
    contextMenu.showRow = true;
};

const openViewContextMenu = (event: MouseEvent, view: string) => {
    closeContextMenu();
    contextMenu.targetView = view;
    contextMenu.position = { x: event.clientX, y: event.clientY };
    contextMenu.showView = true;
};

const openRoutineContextMenu = (event: MouseEvent, routine: string, type: 'PROCEDURE' | 'FUNCTION') => {
    closeContextMenu();
    contextMenu.targetRoutine = routine;
    contextMenu.targetRoutineType = type;
    contextMenu.position = { x: event.clientX, y: event.clientY };
    contextMenu.showRoutine = true;
};

const closeContextMenu = () => {
    contextMenu.showDb = false;
    contextMenu.showFolder = false;
    contextMenu.show = false;
    contextMenu.showRow = false;
    contextMenu.showView = false;
    contextMenu.showRoutine = false;
};

const handleCopyRow = () => {
    if (contextMenu.targetRow) {
        const values = Object.values(contextMenu.targetRow).map(v => v === null ? 'NULL' : String(v)).join('\t');
        navigator.clipboard.writeText(values);
        closeContextMenu();
    }
};

const handleCopyCellValue = () => {
    if (contextMenu.targetRow && contextMenu.targetColumn) {
        const val = contextMenu.targetRow[contextMenu.targetColumn];
        const str = val === null ? 'NULL' : String(val);
        navigator.clipboard.writeText(str);
        closeContextMenu();
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


// Virtual List Logic - Adapted for the first result set for now, or we need multiple virtual lists
// For simplicity in this iteration, let's make the virtual list only apply to the first result set if it exists.
// Or we render multiple tables, but maybe only the first one is virtualized if it's huge?
// Actually, `useVirtualList` takes a source array.
// If we have multiple results, we might need a component for each result table.
// For now, let's keep `filteredResults` mapped to `resultSets[0]` to handle the main case,
// and if there are multiple, we might render them simply or focus on the first one.
// Wait, the user wants "select 1; select 2" to show BOTH.
// `useVirtualList` works on a single list.
// If I change the UI to list multiple tables, I need multiple virtual lists or just standard tables for small results.
// Let's stick to virtualizing the first result set for now to minimize risk,
// and render subsequent result sets as standard tables (assuming they are smaller summaries or we accept performance hit for secondary results).
// Better yet, let's try to map `activeTab.results` to `activeTab.resultSets[0].rows` dynamically.

const currentResultSetIndex = ref(0); // Track which result set is "active" for the main view?
// Or better: Render ALL result sets stack vertically.
// But `useVirtualList` is global per component instance here.
// I will create a `ResultSetTable` component? No, I should stick to single file edits.
// Let's try: Main view uses virtualization. If multiple results, maybe we only virtualize the largest/first?
// Or, we simplifiy and just say: The `virtualList` is backing the `filteredResults`.
// `filteredResults` can be a computed that flattens headers? No.
// Let's try to support just ONE virtual list for the "Primary" result (usually the last one or the one with most rows?).
// Actually, standard behavior: Show "Grid 1", "Grid 2".
// I will modify `filteredResults` to target `resultSets[0]` (backward compat)
// AND add a section to show *other* result sets below, maybe without virtualization (limit 100?) or just standard rendering.

const activeResultSet = computed(() => {
    if (!activeTab.value || !activeTab.value.resultSets || activeTab.value.resultSets.length === 0) return null;
    return activeTab.value.resultSets[0];
});

const filteredResults = computed(() => {
    if (!activeResultSet.value) return [];
    let data = activeResultSet.value.rows || [];

    // Safety check for activeTab
    const tab = activeTab.value;
    if (!tab) return data;

    const filters = tab.filters;

    if (filters && Object.keys(filters).length > 0) {
        data = data.filter(row => {
            for (const [col, filterText] of Object.entries(filters)) {
                if (!filterText) continue;
                const val = row[col];
                const strVal = val === null ? 'NULL' : String(val).toLowerCase();
                if (!strVal.includes(filterText.toLowerCase())) return false;
            }
            return true;
        });
    }

    // Sort logic
    if (tab.sortColumn && tab.sortDirection) {
        const col = tab.sortColumn;
        const dir = tab.sortDirection;

        data = [...data].sort((a, b) => {
            const valA = a[col];
            const valB = b[col];

            if (valA === valB) return 0;
            if (valA === null) return 1; // Nulls last
            if (valB === null) return -1;

            if (valA < valB) return dir === 'asc' ? -1 : 1;
            if (valA > valB) return dir === 'asc' ? 1 : -1;
            return 0;
        });
    }

    return data;
});


const { list: virtualList, containerProps, wrapperProps } = useVirtualList(filteredResults, {
    itemHeight: 37,
    overscan: 10,
});


const padTop = computed(() => {
    if (virtualList.value.length === 0) return 0;
    const start = virtualList.value[0].index;
    return start * 37; // itemHeight
});

const padBottom = computed(() => {
    if (virtualList.value.length === 0) return 0;
    const end = virtualList.value[virtualList.value.length - 1].index;
    const total = filteredResults.value.length;
    return (total - end - 1) * 37; // itemHeight
});


// ... (useVirtualList stays same) ...

// ... (getColumns needs update) ...
const getColumns = (tab: QueryTab) => {
    // This is used for the virtual table (first result set)
    if (tab.resultSets && tab.resultSets.length > 0 && tab.resultSets[0].columns) return tab.resultSets[0].columns;
    return [];
};

// Tab management moved to useTabs

const toggleSort = (col: string) => {
    if (!activeTab.value) return;

    if (activeTab.value.sortColumn === col) {
        if (activeTab.value.sortDirection === 'asc') {
            activeTab.value.sortDirection = 'desc';
        } else if (activeTab.value.sortDirection === 'desc') {
            activeTab.value.sortDirection = null;
            activeTab.value.sortColumn = undefined;
        } else {
            activeTab.value.sortDirection = 'asc';
        }
    } else {
        activeTab.value.sortColumn = col;
        activeTab.value.sortDirection = 'asc';
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



const runQuery = async () => {

    if (!activeTab.value) return;

    activeTab.value.error = '';
    activeTab.value.resultSets = [];
    activeTab.value.filters = {};
    activeTab.value.queryExecuted = false;
    activeTab.value.isLoading = true;
    activeTab.value.isExplaining = false; // Reset explain state
    activeTab.value.explanation = undefined; // Reset explanation
    activeTab.value.executionTime = undefined;
    activeTab.value.fetchTime = undefined;
    activeTab.value.editingCell = null;
    activeTab.value.totalRowCount = undefined;
    activeTab.value.isPartialStats = false;

    const startTime = performance.now();
    const reqId = generateId();
    const tab = activeTab.value;
    tab.activeQueryIds.push(reqId);

    let queryToRun = tab.query;
    if (sqlEditorRef.value) {
        const selection = sqlEditorRef.value.getSelection();
        if (selection && selection.trim()) {
            queryToRun = selection;
        }
    }

    // Track whether we received first batch (for execution time)
    let firstBatchReceived = false;

    const cleanup = () => {
        EventsOff('query:batch:' + reqId);
        EventsOff('query:done:' + reqId);
        EventsOff('query:error:' + reqId);
        EventsOff('query:stats:' + reqId);
        tab.activeQueryIds = tab.activeQueryIds.filter((id: string) => id !== reqId);
    };

    EventsOn('query:stats:' + reqId, (stats: any) => {
        if (tab.activeQueryIds.includes(reqId)) {
            // Check phase. If execution, only set execution time.
            if (stats.phase === 'execution') {
                tab.executionTime = stats.time;
            } else {
                // Fetch phase or default
                if (stats.rows >= 0) {
                    tab.totalRowCount = stats.rows;
                } else {
                    tab.totalRowCount = undefined;
                }

                // If we get "time" in fetch phase, that is EXECUTION time in our new go logic
                // and "fetchTime" is the separate one.
                if (stats.time !== undefined) tab.executionTime = stats.time;
                if (stats.fetchTime !== undefined) tab.fetchTime = stats.fetchTime;

                tab.isPartialStats = stats.partial;
            }
        }
    });

    // Set up event listeners BEFORE calling ExecuteQueryStream
    EventsOn('query:batch:' + reqId, (batch: any) => {
        if (!firstBatchReceived) {
            firstBatchReceived = true;
            // Record time to first batch as "execution time" IF not already set by stats
            if (tab.executionTime === undefined) {
                tab.executionTime = Math.round(performance.now() - startTime);
            }
        }

        const rsIdx = batch.resultSetIdx;
        const columns = batch.columns || [];
        const columnTypes = batch.columnTypes || [];
        const batchRows = batch.rows || [];

        // Convert array rows to object rows. 
        // Note: If the query has duplicate column names (e.g., SELECT A, *), 
        // Object.fromEntries will only keep the LAST occurrence in the object.
        // However, the template uses columns array + index for keys, so it will still render correctly.
        const mappedRows = batchRows.map((row: any[]) =>
            Object.fromEntries(columns.map((col: string, i: number) => [col, row[i]]))
        );

        // Ensure result set exists at this index
        while (tab.resultSets.length <= rsIdx) {
            tab.resultSets.push({ columns: [], columnTypes: [], rows: [] });
        }

        const rs = tab.resultSets[rsIdx];
        if (columns.length > 0 && rs.columns.length === 0) {
            rs.columns = columns;
            rs.columnTypes = columnTypes;

            // Calculate initial widths if not already set
            columns.forEach((col: string, i: number) => {
                if (!tab.columnWidths[col]) {
                    const meta = columnTypes[i];
                    tab.columnWidths[col] = calculateInitialWidth(meta);
                }
            });
        }
        // Append rows
        rs.rows = markRaw(rs.rows.concat(mappedRows));

        // Show data as soon as first data batch arrives
        if (!tab.queryExecuted) {
            tab.queryExecuted = true;
            // Only switch to 'data' tab if we actually have rows
            tab.resultViewTab = (columns.length > 0 && mappedRows.length > 0) ? 'data' : 'messages';
        }
    });

    const calculateInitialWidth = (meta: ColumnMetadata): number => {
        if (!meta) return 150;

        const type = (meta.type || '').toUpperCase();
        const length = meta.length || 0;

        // User's specific rule: max initial 300 for max, blob, long blob or very long fields
        // In some drivers, -1 or very large numbers represent "MAX"
        if (type.includes('BLOB') || length === -1 || length > 1000 || type === 'TEXT' || type === 'LONGTEXT') {
            return 300;
        }

        if (type.includes('CHAR') || type.includes('TEXT') || type.includes('STRING')) {
            if (length > 0) {
                // Approx 8.5px per character + padding
                return Math.min(300, Math.max(120, length * 8.5 + 32));
            }
            return 200; // Default for text without length
        }

        if (type.includes('INT') || type.includes('DECIMAL') || type.includes('NUMERIC') || type.includes('BIT')) {
            return 120;
        }

        if (type.includes('DATE') || type.includes('TIME') || type.includes('TIMESTAMP')) {
            return 180;
        }

        return 150; // Default
    };

    EventsOn('query:done:' + reqId, () => {
        tab.isLoading = false;
        tab.completionTime = new Date().toLocaleString();
        if (!tab.queryExecuted) {
            tab.queryExecuted = true;
            const hasDataWithRows = tab.resultSets.some((rs: any) => rs.columns && rs.columns.length > 0 && rs.rows && rs.rows.length > 0);
            tab.resultViewTab = hasDataWithRows ? 'data' : 'messages';
        }
        if (!firstBatchReceived) {
            if (tab.executionTime === undefined) {
                tab.executionTime = Math.round(performance.now() - startTime);
            }
        }
        cleanup();
    });

    EventsOn('query:error:' + reqId, (errMsg: string) => {
        tab.error = errMsg;
        tab.isLoading = false;
        tab.queryExecuted = true;
        tab.executionTime = Math.round(performance.now() - startTime);
        tab.completionTime = new Date().toLocaleString();
        tab.resultViewTab = 'messages';
        cleanup();
    });

    // Start the streaming query (returns immediately)
    try {
        if (queryToRun && queryToRun.trim().length > 0) {
            SaveQueryHistory(queryToRun, props.connectionName || props.dbType || '');
        }
        const err = await ExecuteQueryStream(props.connectionId, queryToRun, reqId);
        if (err) {
            tab.error = err;
            tab.isLoading = false;
            tab.queryExecuted = true;
            tab.resultViewTab = 'messages';
            cleanup();
        }
    } catch (e: any) {
        tab.error = e.toString();
        tab.isLoading = false;
        tab.executionTime = Math.round(performance.now() - startTime);
        cleanup();
    }
};

const stopQuery = async () => {
    if (!activeTab.value) return;
    const tab = activeTab.value;

    try {
        const ids = [...tab.activeQueryIds];
        await Promise.all(ids.map(id => CancelQuery(id)));
        // The ExecuteQuery promise in runQuery/ER logic will handle the error (context canceled)
    } catch (e) {
        console.error("Error stopping query:", e);
    }
};

const analyzeQuery = async () => {
    if (!activeTab.value || !activeTab.value.query.trim()) return;

    activeTab.value.isExplaining = true;
    activeTab.value.error = ''; // Clear previous errors
    activeTab.value.explanation = "Analyzing...";
    activeTab.value.resultViewTab = 'analysis';
    // Ensure the tab pane is visible even if queryExecuted is false
    // We treat explanation as a type of execution result
    activeTab.value.queryExecuted = true;

    let queryToAnalyze = activeTab.value.query;
    if (sqlEditorRef.value) {
        const selection = sqlEditorRef.value.getSelection();
        if (selection && selection.trim()) {
            queryToAnalyze = selection;
        }
    }

    try {
        const plan = await ExplainQuery(props.connectionId, queryToAnalyze);
        activeTab.value.explanation = plan;
    } catch (err: any) {
        activeTab.value.error = "Failed to analyze query: " + err.toString();
        activeTab.value.explanation = undefined;
    } finally {
        activeTab.value.isExplaining = false;
    }
};

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

const handleCellClick = (item: any, col: string) => {
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
        value: value
    };

    nextTick(() => {
        const input = document.getElementById(`edit-input-${item.index}-${col}`);
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
};

onMounted(() => {
    if (props.connectionId) {
        loadTables();
        addTab();
    }
    window.addEventListener('keydown', handleKeydown, true);
    window.addEventListener('click', closeContextMenu);
    window.addEventListener('open-sql-file', handleOpenSqlFile as EventListener);
});

onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown, true);
    window.removeEventListener('click', closeContextMenu);
    window.removeEventListener('open-sql-file', handleOpenSqlFile as EventListener);
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
});
</script>

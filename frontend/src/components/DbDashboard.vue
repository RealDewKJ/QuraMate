<template>
    <div class="flex h-full bg-background text-foreground font-sans">
        <!-- Sidebar -->
        <div class="w-64 border-r border-border bg-card flex flex-col transition-all duration-300">
            <div class="p-4 border-b border-border flex items-center gap-2">
                <div class="h-6 w-6 rounded bg-primary flex items-center justify-center text-primary-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-database">
                        <ellipse cx="12" cy="5" rx="9" ry="3" />
                        <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                        <path d="M3 12A9 3 0 0 0 21 12" />
                    </svg>
                </div>
                <span class="font-semibold tracking-tight">Tables</span>
            </div>

            <!-- Removed top search bar, moved inside Tables folder -->

            <div class="flex-1 overflow-y-auto px-2 py-2">
                <div class="space-y-1">
                    <!-- Tables Folder -->
                    <div>
                        <div @click="toggleFolder('Tables')"
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
                            <span class="text-xs text-muted-foreground py-1 block italic ml-2">No views loaded.</span>
                        </div>
                    </div>

                    <!-- Programmability Folder -->
                    <div>
                        <div @click="toggleFolder('Programmability')"
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
                            <span class="text-xs text-muted-foreground py-1 block italic ml-2">Not implemented.</span>
                        </div>
                    </div>

                    <!-- Service Broker Folder -->
                    <div>
                        <div @click="toggleFolder('Service Broker')"
                            class="flex items-center gap-2 px-2 py-1.5 text-sm font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round"
                                class="lucide lucide-chevron-right transition-transform duration-200"
                                :class="{ 'rotate-90': openFolders.includes('Service Broker') }">
                                <path d="m9 18 6-6-6-6" />
                            </svg>
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-server text-orange-500">
                                <rect width="20" height="8" x="2" y="2" rx="2" ry="2" />
                                <rect width="20" height="8" x="2" y="14" rx="2" ry="2" />
                                <line x1="6" x2="6.01" y1="6" y2="6" />
                                <line x1="6" x2="6.01" y1="18" y2="18" />
                            </svg>
                            <span>Service Broker</span>
                        </div>
                        <div v-show="openFolders.includes('Service Broker')"
                            class="ml-4 mt-1 border-l border-border pl-2">
                            <span class="text-xs text-muted-foreground py-1 block italic ml-2">Not implemented.</span>
                        </div>
                    </div>

                    <!-- Storage Folder -->
                    <div>
                        <div @click="toggleFolder('Storage')"
                            class="flex items-center gap-2 px-2 py-1.5 text-sm font-medium rounded-md cursor-pointer hover:bg-accent hover:text-accent-foreground transition-colors select-none">
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round"
                                class="lucide lucide-chevron-right transition-transform duration-200"
                                :class="{ 'rotate-90': openFolders.includes('Storage') }">
                                <path d="m9 18 6-6-6-6" />
                            </svg>
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-database text-yellow-500">
                                <ellipse cx="12" cy="5" rx="9" ry="3" />
                                <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                                <path d="M3 12A9 3 0 0 0 21 12" />
                            </svg>
                            <span>Storage</span>
                        </div>
                        <div v-show="openFolders.includes('Storage')" class="ml-4 mt-1 border-l border-border pl-2">
                            <span class="text-xs text-muted-foreground py-1 block italic ml-2">Not implemented.</span>
                        </div>
                    </div>
                </div>
            </div>

            <!-- Disconnect Button -->
            <div class="p-4 border-t border-border">
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
                <div v-for="tab in tabs" :key="tab.id" @click="activeTabId = tab.id"
                    class="group relative flex items-center justify-between gap-2 px-4 py-2 text-sm font-medium cursor-pointer rounded-t-lg transition-all select-none min-w-[140px] max-w-[240px] border-l border-r border-t border-transparent hover:bg-background/50"
                    :class="{ 'bg-background text-foreground border-border shadow-sm mb-[-1px]': activeTabId === tab.id, 'text-muted-foreground hover:text-foreground': activeTabId !== tab.id }">
                    <div class="flex items-center gap-2 truncate">
                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-terminal-square">
                            <path d="m7 11 2-2-2-2" />
                            <path d="M11 13h4" />
                            <rect width="18" height="18" x="3" y="3" rx="2" ry="2" />
                        </svg>
                        <span class="truncate">{{ tab.name }}</span>
                    </div>
                    <button @click.stop="closeTab(tab.id)"
                        class="rounded-sm p-0.5 hover:bg-muted text-muted-foreground/50 hover:text-foreground transition-all opacity-0 group-hover:opacity-100">
                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-x">
                            <path d="M18 6 6 18" />
                            <path d="m6 6 12 12" />
                        </svg>
                    </button>
                    <!-- Active Indicator Line -->
                    <div v-if="activeTabId === tab.id"
                        class="absolute top-0 left-0 right-0 h-0.5 bg-primary rounded-t-full"></div>
                </div>

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
            <div v-if="activeTab" class="flex flex-col h-full overflow-hidden">
                <div class="flex flex-col border-b border-border bg-card p-4 gap-3 relative">
                    <div class="relative w-full h-64">
                        <SqlEditor v-model="activeTab.query" :tables="tables" />

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
                            <div v-if="activeTab.isLoading" class="flex items-center gap-2 text-primary">
                                <svg class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none"
                                    viewBox="0 0 24 24">
                                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                        stroke-width="4"></circle>
                                    <path class="opacity-75" fill="currentColor"
                                        d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                                    </path>
                                </svg>
                                Executing...
                            </div>
                            <div v-else-if="activeTab.executionTime !== undefined" class="flex items-center gap-1.5">
                                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-timer">
                                    <line x1="10" x2="14" y1="2" y2="2" />
                                    <line x1="12" x2="15" y1="14" y2="11" />
                                    <circle cx="12" cy="14" r="8" />
                                </svg>
                                <span>{{ activeTab.executionTime }}ms</span>
                            </div>
                        </div>

                    </div>

                    <div class="flex items-center gap-2">
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

                        <button @click="runQuery" :disabled="activeTab.isLoading"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-9 px-4 py-2 shadow-sm">
                            <svg v-if="!activeTab.isLoading" xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-play mr-2">
                                <polygon points="5 3 19 12 5 21 5 3" />
                            </svg>
                            Run Query
                        </button>
                    </div>
                </div>

                <!-- Results Area -->
                <div class="flex-1 overflow-auto bg-muted/10 p-4">
                    <!-- Error State -->
                    <div v-if="activeTab.error"
                        class="bg-destructive/10 border border-destructive/20 text-destructive p-4 rounded-lg shadow-sm flex items-start gap-3 animate-in fade-in slide-in-from-top-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-alert-triangle mt-0.5">
                            <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                            <path d="M12 9v4" />
                            <path d="M12 17h.01" />
                        </svg>
                        <div class="flex-1 text-sm font-medium break-all font-mono">{{ activeTab.error }}</div>
                    </div>

                    <!-- Results Table -->
                    <div v-else-if="activeTab.results && activeTab.results.length > 0"
                        class="border border-border rounded-lg shadow-sm bg-card flex flex-col h-full max-h-full overflow-hidden">
                        <div class="flex-1 overflow-auto" v-bind="containerProps">
                            <table class="w-full text-sm text-left relative">
                                <thead
                                    class="text-xs text-muted-foreground uppercase bg-muted sticky top-0 z-10 font-medium">
                                    <tr>
                                        <th v-for="col in getColumns(activeTab)" :key="col" scope="col"
                                            class="px-4 py-3 whitespace-nowrap border-b border-border min-w-[150px] cursor-pointer hover:bg-muted/80 select-none"
                                            @click="toggleSort(col)">
                                            <div class="flex flex-col gap-2">
                                                <div class="flex items-center justify-between gap-2">
                                                    <span>{{ col }}</span>
                                                    <div class="flex flex-col">
                                                        <svg v-if="activeTab.sortColumn === col && activeTab.sortDirection === 'asc'"
                                                            xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                                            viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                                            stroke-width="2" stroke-linecap="round"
                                                            stroke-linejoin="round" class="lucide lucide-chevron-up">
                                                            <path d="m18 15-6-6-6 6" />
                                                        </svg>
                                                        <svg v-if="activeTab.sortColumn === col && activeTab.sortDirection === 'desc'"
                                                            xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                                            viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                                            stroke-width="2" stroke-linecap="round"
                                                            stroke-linejoin="round" class="lucide lucide-chevron-down">
                                                            <path d="m6 9 6 6 6-6" />
                                                        </svg>
                                                        <svg v-if="activeTab.sortColumn !== col"
                                                            xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                                            viewBox="0 0 24 24" fill="none" stroke="currentColor"
                                                            stroke-width="2" stroke-linecap="round"
                                                            stroke-linejoin="round"
                                                            class="lucide lucide-chevrons-up-down text-muted-foreground/30">
                                                            <path d="m7 15 5 5 5-5" />
                                                            <path d="m7 9 5-5 5 5" />
                                                        </svg>
                                                    </div>
                                                </div>
                                                <input v-if="activeTab.primaryKeys.length > 0 || true" type="text"
                                                    v-model="activeTab.filters[col]" placeholder="Filter..."
                                                    class="w-full h-6 px-2 text-[10px] rounded border border-input bg-background focus:outline-none focus:ring-1 focus:ring-ring font-normal normal-case text-foreground cursor-text"
                                                    @click.stop />
                                            </div>
                                        </th>
                                    </tr>
                                </thead>
                                <tbody class="divide-y divide-border">
                                    <tr :style="{ height: `${padTop}px` }"></tr>
                                    <tr v-for="item in virtualList" :key="item.index"
                                        class="bg-card hover:bg-muted/50 transition-colors h-[37px]">
                                        <td v-for="col in getColumns(activeTab)" :key="col"
                                            class="px-4 py-2 whitespace-nowrap text-foreground font-mono text-xs border-r border-transparent hover:border-border cursor-pointer relative"
                                            :class="{ 'bg-accent/50': activeTab.editingCell && activeTab.editingCell.rowId === item.index && activeTab.editingCell.col === col }"
                                            @dblclick="handleCellClick(item, col)">

                                            <div v-if="activeTab.editingCell && activeTab.editingCell.rowId === item.index && activeTab.editingCell.col === col"
                                                class="absolute inset-0 p-0.5">
                                                <input :id="`edit-input-${item.index}-${col}`"
                                                    v-model="activeTab.editingCell.value"
                                                    class="w-full h-full px-2 bg-background text-foreground border border-primary focus:outline-none focus:ring-1 focus:ring-primary rounded-sm shadow-sm"
                                                    @blur="saveCellEdit(item, col)"
                                                    @keydown.enter="saveCellEdit(item, col)"
                                                    @keydown.esc="activeTab.editingCell = null" />
                                            </div>
                                            <span v-else class="truncate block max-w-[300px]"
                                                :title="String(item.data[col])">
                                                {{ item.data[col] === null ? 'NULL' : item.data[col] }}
                                            </span>
                                        </td>
                                    </tr>
                                    <tr :style="{ height: `${padBottom}px` }"></tr>
                                </tbody>
                            </table>
                        </div>
                        <div
                            class="bg-muted/30 px-4 py-2 border-t border-border text-xs text-muted-foreground flex justify-between items-center">
                            <span>{{ filteredResults.length }} rows returned ({{ activeTab.results.length }}
                                total)</span>
                            <span class="font-mono text-[10px] opacity-70">Double-click to edit</span>
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
                        <div class="h-16 w-16 rounded-2xl bg-muted/50 flex items-center justify-center mb-4 shadow-sm">
                            <svg xmlns="http://www.w3.org/2000/svg" width="32" height="32" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-terminal text-primary">
                                <polyline points="4 17 10 11 4 5" />
                                <line x1="12" x2="20" y1="19" y2="19" />
                            </svg>
                        </div>
                        <h3 class="text-lg font-semibold text-foreground">Ready to Query</h3>
                        <p class="text-sm max-w-sm text-center mt-2">Select a table from the sidebar or type a custom
                            SQL query to get started.</p>

                        <div class="flex gap-2 mt-6">
                            <span class="text-xs bg-muted px-2 py-1 rounded border border-border">Ctrl + Enter to
                                Run</span>
                        </div>
                    </div>
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
        <!-- Context Menu -->
        <div v-if="showContextMenu"
            class="fixed z-50 bg-popover text-popover-foreground border border-border shadow-md rounded-md py-1 min-w-[160px]"
            :style="{ top: `${contextMenuPosition.y}px`, left: `${contextMenuPosition.x}px` }">
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
            <button @click="handleViewDesign"
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
            </button>
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
                                {{ updateConfirmation.newValue === null ? 'NULL' : updateConfirmation.newValue }}
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
    </div>
</template>

<script lang="ts" setup>
import { ref, onMounted, computed, watch, nextTick } from 'vue';
import { GetTables, ExecuteQuery, DisconnectDB, GetPrimaryKeys, UpdateRecord } from '../../wailsjs/go/main/App';
import { format } from 'sql-formatter';
import { useVirtualList } from '@vueuse/core';
import SqlEditor from './SqlEditor.vue';

const props = defineProps<{
    connectionId: string;
    dbType: string;
}>();

const emit = defineEmits(['disconnect']);

interface CellEdit {
    rowId: any; // Using row index or joined PK values as ID
    col: string;
    value: any;
}

interface QueryTab {
    id: string;
    name: string;
    tableName?: string; // Store table name if it's a simple SELECT
    query: string;
    results: any[];
    columns: string[];
    primaryKeys: string[];
    filters: Record<string, string>;
    sortColumn?: string;
    sortDirection: 'asc' | 'desc' | null;
    error: string;
    isLoading: boolean;
    queryExecuted: boolean;
    executionTime?: number;
    editingCell?: CellEdit | null;
    isDesignView?: boolean;
}

const tableSearch = ref('');
const tables = ref<string[]>([]);
const tabs = ref<QueryTab[]>([]);
const activeTabId = ref<string | null>(null);

// Sidebar State
const openFolders = ref(['Tables']);

const toggleFolder = (folder: string) => {
    if (openFolders.value.includes(folder)) {
        openFolders.value = openFolders.value.filter(f => f !== folder);
    } else {
        openFolders.value.push(folder);
    }
};

// Context Menu State
const showContextMenu = ref(false);
const contextMenuPosition = ref({ x: 0, y: 0 });
const contextMenuTargetTable = ref('');

// Update Confirmation State
const updateConfirmation = ref<{
    isOpen: boolean;
    tableName: string;
    column: string;
    originalValue: any;
    newValue: any;
    rowIndex: number;
    item: any;
} | null>(null);

// Active Tab Helper
const activeTab = computed(() => tabs.value.find(t => t.id === activeTabId.value));

// Virtual List Logic
const filteredResults = computed(() => {
    if (!activeTab.value) return [];
    let data = activeTab.value.results;
    const filters = activeTab.value.filters;

    // 1. Filter
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

    // 2. Sort
    if (activeTab.value.sortColumn && activeTab.value.sortDirection) {
        const col = activeTab.value.sortColumn;
        const dir = activeTab.value.sortDirection;

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

const { list: virtualList, containerProps, wrapperProps } = useVirtualList(
    filteredResults,
    {
        itemHeight: 37,
        overscan: 10,
    }
);

// Start/End padding for table virtualization
const padTop = computed(() => {
    if (virtualList.value.length === 0) return 0;
    return virtualList.value[0].index * 37;
});

const padBottom = computed(() => {
    if (virtualList.value.length === 0) return 0;
    const lastItem = virtualList.value[virtualList.value.length - 1];
    const total = filteredResults.value.length;
    return (total - 1 - lastItem.index) * 37;
});

const filteredTables = computed(() => {
    if (!tableSearch.value) return tables.value;
    return tables.value.filter(t => t.toLowerCase().includes(tableSearch.value.toLowerCase()));
});

const getColumns = (tab: QueryTab) => {
    if (tab.columns && tab.columns.length > 0) return tab.columns;
    if (!tab.results || tab.results.length === 0) return [];
    return Object.keys(tab.results[0]);
};

const generateId = () => {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
};

const tabCounter = ref(0);

const addTab = () => {
    const newId = generateId();
    tabCounter.value++;
    tabs.value.push({
        id: newId,
        name: `Query ${tabCounter.value}`,
        query: '',
        results: [],
        columns: [],
        primaryKeys: [],
        filters: {},
        sortColumn: undefined,
        sortDirection: null,
        error: '',
        isLoading: false,
        queryExecuted: false
    });
    activeTabId.value = newId;
};

// ... existing code ...

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

const closeTab = (id: string) => {
    const index = tabs.value.findIndex(t => t.id === id);
    if (index !== -1) {
        tabs.value.splice(index, 1);
        if (activeTabId.value === id) {
            if (tabs.value.length > 0) {
                activeTabId.value = tabs.value[tabs.value.length - 1].id;
            } else {
                activeTabId.value = null;
            }
        }
    }
};

const loadTables = async () => {
    try {
        const result = await GetTables(props.connectionId);
        tables.value = result.sort((a, b) => a.localeCompare(b));
    } catch (e) {
        console.error("Failed to load tables", e);
    }
};

const selectTable = async (tableName: string) => {
    if (!activeTab.value) {
        addTab();
    }
    if (activeTab.value) {
        const type = (props.dbType || '').toLowerCase();
        activeTab.value.tableName = tableName;
        activeTab.value.name = tableName; // Update tab name to table name

        if (type.includes('mssql') || type.includes('sqlserver')) {
            activeTab.value.query = `SELECT TOP 100 * FROM ${tableName}`;
        } else {
            activeTab.value.query = `SELECT * FROM ${tableName} LIMIT 100`;
        }

        // Fetch Primary Keys
        try {
            activeTab.value.primaryKeys = await GetPrimaryKeys(props.connectionId, tableName);
        } catch (e) {
            console.error("Failed to fetch primary keys", e);
            activeTab.value.primaryKeys = [];
        }

        runQuery();
    }
};

const openContextMenu = (event: MouseEvent, table: string) => {
    contextMenuTargetTable.value = table;
    const { clientX, clientY } = event;
    contextMenuPosition.value = { x: clientX, y: clientY };
    showContextMenu.value = true;
};

const closeContextMenu = () => {
    showContextMenu.value = false;
};

const handleSelectTop100 = () => {
    if (contextMenuTargetTable.value) {
        selectTable(contextMenuTargetTable.value);
        closeContextMenu();
    }
};

const openDesignTab = (tableName: string) => {
    const type = (props.dbType || '').toLowerCase();
    let query = '';

    if (type.includes('mssql') || type.includes('sqlserver')) {
        query = `SELECT COLUMN_NAME, DATA_TYPE, IS_NULLABLE, CHARACTER_MAXIMUM_LENGTH FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '${tableName}'`;
    } else if (type.includes('postgres')) {
        query = `SELECT column_name, data_type, is_nullable, character_maximum_length FROM information_schema.columns WHERE table_name = '${tableName}'`;
    } else if (type.includes('mysql') || type.includes('maria')) {
        query = `DESCRIBE ${tableName}`;
    } else if (type.includes('sqlite')) {
        query = `PRAGMA table_info(${tableName})`;
    } else {
        query = `-- Could not determine DB type for schema query\nSELECT * FROM ${tableName} LIMIT 1`;
    }

    // Check if design tab already exists
    const existingTab = tabs.value.find(t => t.name === `Design: ${tableName}`);
    if (existingTab) {
        activeTabId.value = existingTab.id;
        return;
    }

    const newId = generateId();
    tabCounter.value++;
    tabs.value.push({
        id: newId,
        name: `Design: ${tableName}`,
        query: query,
        results: [],
        columns: [],
        primaryKeys: [], // Design view is read-only usually
        filters: {},
        sortColumn: undefined,
        sortDirection: null,
        error: '',
        isLoading: false,
        queryExecuted: false,
        isDesignView: true
    });
    activeTabId.value = newId;

    setTimeout(() => {
        runQuery();
    }, 50);
};

const handleViewDesign = () => {
    if (contextMenuTargetTable.value) {
        openDesignTab(contextMenuTargetTable.value);
        closeContextMenu();
    }
};

const runQuery = async () => {
    if (!activeTab.value) return;

    activeTab.value.error = '';
    activeTab.value.results = [];
    activeTab.value.columns = [];
    activeTab.value.filters = {}; // Reset filters on new query run
    activeTab.value.queryExecuted = false;
    activeTab.value.isLoading = true;
    activeTab.value.executionTime = undefined;
    activeTab.value.editingCell = null;

    const startTime = performance.now();

    try {
        const res = await ExecuteQuery(props.connectionId, activeTab.value.query);
        const endTime = performance.now();
        activeTab.value.executionTime = Math.round(endTime - startTime);

        if (res.error) {
            activeTab.value.error = res.error;
        } else {
            activeTab.value.results = res.data || [];
            activeTab.value.columns = res.columns || [];
        }
        activeTab.value.queryExecuted = true;
    } catch (e: any) {
        activeTab.value.error = e.toString();
        const endTime = performance.now();
        activeTab.value.executionTime = Math.round(endTime - startTime);
    } finally {
        activeTab.value.isLoading = false;
    }
};

const beautifyQuery = () => {
    if (!activeTab.value || !activeTab.value.query) return;

    try {
        const type = (props.dbType || '').toLowerCase();
        let language = 'sql';

        if (type.includes('postgres')) language = 'postgresql';
        else if (type.includes('mysql') || type.includes('maria')) language = 'mysql';
        else if (type.includes('mssql') || type.includes('sqlserver')) language = 'transactsql';
        else if (type.includes('sqlite')) language = 'sqlite';

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

// Editing Logic
const isEditable = (col: string) => {
    if (!activeTab.value || !activeTab.value.tableName || activeTab.value.primaryKeys.length === 0) return false;
    if (activeTab.value.isDesignView) return false; // Disable editing in design view
    // Don't edit PKs for now to simplify
    if (activeTab.value.primaryKeys.includes(col)) return false;
    return true;
};

const getRowId = (row: any, index: number) => {
    // Use index as fallback but strictly we need PKs for updates.
    // If we are editing, we must have PKs.
    return index;
};

const handleCellClick = (item: any, col: string) => {
    if (!isEditable(col)) return;

    activeTab.value!.editingCell = {
        rowId: item.index,
        col: col,
        value: item.data[col]
    };

    nextTick(() => {
        const input = document.getElementById(`edit-input-${item.index}-${col}`);
        if (input) (input as HTMLInputElement).focus();
    });
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

const confirmUpdate = async () => {
    if (!updateConfirmation.value || !activeTab.value) return;

    const { tableName, column, newValue, item } = updateConfirmation.value;
    const col = column;

    // Prepare conditions (PKs)
    const conditions: Record<string, any> = {};
    for (const pk of activeTab.value.primaryKeys) {
        conditions[pk] = item.data[pk];
    }

    const updates: Record<string, any> = {};
    updates[col] = newValue;

    try {
        const result = await UpdateRecord(props.connectionId, tableName, updates, conditions);
        if (result === "Success") {
            const currentTab = activeTab.value;
            if (!currentTab) return;

            // Update local state
            item.data[col] = newValue;
            // Also update the original source array
            const realIndex = currentTab.results.findIndex(r => {
                for (const pk of currentTab.primaryKeys) {
                    if (r[pk] !== conditions[pk]) return false;
                }
                return true;
            });
            if (realIndex !== -1) {
                currentTab.results[realIndex][col] = newValue;
            }
        } else {
            console.error("Update failed:", result);
            alert("Update failed: " + result);
        }
    } catch (e) {
        console.error("Update error:", e);
        alert("Update error: " + e);
    } finally {
        if (activeTab.value) {
            activeTab.value.editingCell = null;
        }
        updateConfirmation.value = null; // Close modal
    }
};

const cancelUpdate = () => {
    if (activeTab.value) {
        activeTab.value.editingCell = null;
    }
    updateConfirmation.value = null;
};


const handleKeydown = (e: KeyboardEvent) => {
    if ((e.ctrlKey || e.metaKey) && e.key === 'Enter') {
        runQuery();
    }
    if ((e.ctrlKey || e.metaKey) && (e.key === 'd' || e.key === 'D')) {
        e.preventDefault();
        if (activeTab.value && activeTab.value.tableName) {
            openDesignTab(activeTab.value.tableName);
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
});

import { onUnmounted } from 'vue';
onUnmounted(() => {
    window.removeEventListener('keydown', handleKeydown, true);
    window.removeEventListener('click', closeContextMenu);
});

watch(() => props.connectionId, (newId) => {
    if (newId) {
        loadTables();
        tabs.value = [];
        addTab();
    }
});
</script>

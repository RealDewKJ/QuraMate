<script lang="ts" setup>
import { ref } from 'vue';

import SqlEditor from '../SqlEditor.vue';

const props = defineProps<{
    activeTab: any;
    tables: string[];
    getColumns: (tableName: string) => Promise<string[]>;
    editorSettings: { fontFamily: string; fontSize: number };
    isReadOnly?: boolean;
}>();

const emit = defineEmits<{
    'beautify-query': [];
    'explain-with-ai': [];
    'save-routine': [];
    'run-query': [];
    'stop-query': [];
    'start-resizing': [event: MouseEvent];
}>();

const sqlEditorRef = ref<any>(null);

const getSelection = (): string => {
    return sqlEditorRef.value?.getSelection?.() || '';
};

defineExpose({
    getSelection,
});
</script>

<template>
    <div v-if="activeTab && !activeTab.isERView && !activeTab.isDesignView"
        class="flex flex-col border-b border-border bg-card p-4 gap-3 relative shrink-0 min-h-[0px]"
        :style="{ height: activeTab.editorHeight + 'px' }">
        <div class="relative w-full flex-1 min-h-0">
            <SqlEditor ref="sqlEditorRef" v-model="activeTab.query" :tables="tables" :get-columns="getColumns"
                :font-family="editorSettings.fontFamily" :font-size="editorSettings.fontSize" />

            <div class="absolute bottom-1 right-3 z-10 flex items-center gap-2 pointer-events-none">
                <div
                    class="text-xs text-muted-foreground bg-background/80 px-2 py-1 rounded backdrop-blur-sm border border-border pointer-events-auto">
                    {{ activeTab.query.length }} chars
                </div>
            </div>
        </div>

        <div class="flex items-center justify-between">
            <div class="flex items-center gap-4 text-xs text-muted-foreground">
                <div v-if="activeTab.isLoading && !activeTab.executionTime" class="flex items-center gap-2 text-primary">
                    <svg class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z"></path>
                    </svg>
                    Executing...
                </div>
                <div v-else-if="activeTab.executionTime !== undefined" class="flex items-center gap-1.5 ">
                    <span class="flex items-center gap-1" title="Execution Time (Database)">
                        <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-database">
                            <ellipse cx="12" cy="5" rx="9" ry="3" />
                            <path d="M21 12c0 1.66-4 3-9 3s-9-1.34-9-3" />
                            <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5" />
                        </svg>
                        <span>Exec: {{ activeTab.executionTime }}ms</span>
                    </span>
                    <span class="text-border mx-1">|</span>
                    <span class="flex items-center gap-1" title="Fetch/Transfer Time">
                        <div v-if="activeTab.isLoading" class="flex items-center gap-1">
                            <svg class="animate-spin h-3 w-3 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none"
                                viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor"
                                    d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z"></path>
                            </svg>
                        </div>
                        <svg v-else xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-arrow-down">
                            <line x1="12" x2="12" y1="5" y2="19" />
                            <polyline points="19 12 12 19 5 12" />
                        </svg>
                        <span>Fetch: {{ activeTab.fetchTime !== undefined ? activeTab.fetchTime : '...' }}{{
                            activeTab.isLoading ? '...' : 'ms' }}</span>
                    </span>
                </div>
            </div>
        </div>

        <div class="flex items-center gap-2">
            <div v-if="isReadOnly"
                class="px-2 py-1 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-500 text-xs rounded border border-yellow-200 dark:border-yellow-900/50 mr-2 flex items-center gap-1 select-none cursor-help"
                title="Database is in Read-Only mode. Modifications are disabled.">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-lock">
                    <rect width="18" height="11" x="3" y="11" rx="2" ry="2" />
                    <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                </svg>
                Read Only
            </div>

            <button @click="emit('beautify-query')" :disabled="activeTab.isLoading || !activeTab.query"
                class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-4 py-2 shadow-sm"
                title="Format SQL (Shift + Alt + F)">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-wrap-text mr-2">
                    <line x1="3" x2="21" y1="6" y2="6" />
                    <path d="M3 12h15a3 3 0 1 1 0 6h-4" />
                    <polyline points="16 16 14 18 16 20" />
                </svg>
                Beautify
            </button>

            <button @click="emit('explain-with-ai')" :disabled="activeTab.isLoading || activeTab.isAiExplaining || !activeTab.query"
                class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-9 px-4 py-2 shadow-sm"
                title="Explain SQL with configured AI provider">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-bot mr-2">
                    <path d="M12 8V4H8" />
                    <rect width="16" height="12" x="4" y="8" rx="2" />
                    <path d="M2 14h2" />
                    <path d="M20 14h2" />
                    <path d="M15 13v2" />
                    <path d="M9 13v2" />
                </svg>
                {{ activeTab.isAiExplaining ? 'AI Explaining...' : 'Explain with AI' }}
            </button>

            <button v-if="activeTab.isRoutine" @click="emit('save-routine')" :disabled="activeTab.isLoading"
                class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-primary/50 bg-primary/10 text-primary hover:bg-primary/20 h-9 px-4 py-2 shadow-sm"
                title="Save / Update Routine">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-save mr-2">
                    <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z" />
                    <polyline points="17 21 17 13 7 13 7 21" />
                    <polyline points="7 3 7 8 15 8" />
                </svg>
                Save
            </button>

            <button v-if="activeTab.isLoading" @click="emit('stop-query')"
                class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 bg-destructive text-destructive-foreground hover:bg-destructive/90 h-9 px-4 py-2 shadow-sm">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-square mr-2 fill-current">
                    <rect width="18" height="18" x="3" y="3" rx="2" />
                </svg>
                Stop
            </button>

            <button v-else @click="emit('run-query')"
                class="inline-flex items-center justify-center rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 bg-primary text-primary-foreground hover:bg-primary/90 h-9 px-4 py-2 shadow-sm">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-play mr-2">
                    <polygon points="5 3 19 12 5 21 5 3" />
                </svg>
                Run Query
            </button>
        </div>
    </div>

    <div v-if="activeTab && !activeTab.isERView && !activeTab.isDesignView"
        class="h-1.5 hover:bg-primary/30 cursor-row-resize flex items-center justify-center transition-colors group z-20 shrink-0"
        @mousedown="emit('start-resizing', $event)">
        <div class="w-8 h-1 bg-border rounded-full group-hover:bg-primary/50"></div>
    </div>
</template>

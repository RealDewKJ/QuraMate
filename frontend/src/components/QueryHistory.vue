<template>
    <div v-if="isOpen"
        class="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
        <div
            class="bg-card w-full max-w-4xl max-h-[80vh] h-[80vh] rounded-xl border border-border flex flex-col shadow-lg overflow-hidden relative">
            <div class="p-4 border-b border-border flex items-center justify-between bg-muted/20">
                <h2 class="text-lg font-semibold flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-history text-primary">
                        <path d="M3 12a9 9 0 1 0 9-9 9.75 9.75 0 0 0-6.74 2.74L3 8" />
                        <path d="M3 3v5h5" />
                        <path d="M12 7v5l4 2" />
                    </svg>
                    Query History
                </h2>
                <button @click="$emit('close')"
                    class="h-8 w-8 flex items-center justify-center rounded-md hover:bg-muted text-muted-foreground transition-colors">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div class="flex-1 flex overflow-hidden">
                <!-- Sidebar for DB Type Selection -->
                <div class="w-48 border-r border-border bg-muted/10 p-3 overflow-y-auto">
                    <div class="text-xs font-semibold uppercase text-muted-foreground mb-2 px-2">Filter by Connection
                    </div>
                    <button @click="activeDbType = 'all'"
                        class="w-full text-left px-3 py-2 rounded-md text-sm transition-colors mb-1"
                        :class="activeDbType === 'all' ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted text-foreground'">
                        All Connections
                    </button>
                    <button v-for="type in dbTypes" :key="type" @click="activeDbType = type"
                        class="w-full text-left px-3 py-2 rounded-md text-sm transition-colors mb-1"
                        :class="activeDbType === type ? 'bg-primary/10 text-primary font-medium' : 'hover:bg-muted text-foreground'">
                        {{ type }}
                    </button>
                </div>

                <!-- History List -->
                <div class="flex-1 flex flex-col bg-background relative">
                    <!-- Main Toolbar -->
                    <div class="p-2 border-b border-border flex justify-between items-center bg-muted/5 z-10">
                        <div class="text-sm text-muted-foreground pl-2">{{ filteredHistory.length }} queries</div>
                        <button @click="clearHistory"
                            class="flex items-center gap-1.5 px-3 py-1.5 text-xs text-destructive hover:bg-destructive/10 rounded-md transition-colors">
                            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-trash-2">
                                <path d="M3 6h18" />
                                <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                                <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                                <line x1="10" x2="10" y1="11" y2="17" />
                                <line x1="14" x2="14" y1="11" y2="17" />
                            </svg>
                            Clear Non-Favorites
                        </button>
                    </div>

                    <div class="flex-1 overflow-y-auto p-4 space-y-3">
                        <div v-if="isLoading" class="flex justify-center py-8">
                            <svg class="animate-spin h-6 w-6 text-primary border-primary"
                                xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor"
                                    stroke-width="4"></circle>
                                <path class="opacity-75" fill="currentColor"
                                    d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z"></path>
                            </svg>
                        </div>

                        <div v-else-if="filteredHistory.length === 0"
                            class="flex flex-col items-center justify-center p-12 text-muted-foreground text-center">
                            <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24"
                                fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round"
                                stroke-linejoin="round" class="lucide lucide-search-x mb-4 opacity-50">
                                <path d="m13.5 8.5-5 5" />
                                <path d="m8.5 8.5 5 5" />
                                <circle cx="11" cy="11" r="8" />
                                <path d="m21 21-4.3-4.3" />
                            </svg>
                            <h3 class="text-base font-medium text-foreground mb-1">No history found</h3>
                            <p class="text-sm">You haven't executed any queries yet, or there are none for the selected
                                connection.</p>
                        </div>

                        <div v-for="entry in filteredHistory" :key="entry.id"
                            class="border border-border rounded-lg bg-card overflow-hidden group hover:border-primary/50 transition-colors">
                            <div class="flex items-center justify-between px-3 py-2 bg-muted/30 border-b border-border">
                                <div class="flex items-center gap-3">
                                    <span
                                        class="text-xs font-mono bg-muted px-2 py-0.5 rounded text-muted-foreground border border-border/50">{{
                                            entry.db_type }}</span>
                                    <span class="text-xs text-muted-foreground">{{ formatDate(entry.timestamp) }}</span>
                                </div>
                                <div class="flex items-center gap-1">
                                    <!-- Actions -->
                                    <button @click="copyQuery(entry.query)"
                                        class="p-1.5 rounded-md hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                                        title="Copy code">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-copy">
                                            <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
                                            <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
                                        </svg>
                                    </button>
                                    <button @click="$emit('run-query', entry.query)"
                                        class="p-1.5 rounded-md hover:bg-primary/20 text-primary transition-colors"
                                        title="Load & Run">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-play">
                                            <polygon points="5 3 19 12 5 21 5 3" />
                                        </svg>
                                    </button>
                                    <div class="w-px h-4 bg-border mx-1"></div>
                                    <button @click="toggleFavorite(entry)"
                                        class="p-1.5 rounded-md hover:bg-yellow-500/10 transition-colors"
                                        :class="entry.is_favorite ? 'text-yellow-500' : 'text-muted-foreground hover:text-yellow-600'">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                                            viewBox="0 0 24 24" :fill="entry.is_favorite ? 'currentColor' : 'none'"
                                            stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                            stroke-linejoin="round" class="lucide lucide-star">
                                            <polygon
                                                points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                                        </svg>
                                    </button>
                                    <button @click="deleteEntry(entry.id)"
                                        class="p-1.5 rounded-md hover:bg-destructive/10 text-muted-foreground hover:text-destructive transition-colors ml-1"
                                        title="Delete record">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-trash">
                                            <path d="M3 6h18" />
                                            <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                                            <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                                        </svg>
                                    </button>
                                </div>
                            </div>
                            <div
                                class="p-3 bg-muted/5 font-mono text-sm max-h-32 overflow-y-auto whitespace-pre-wrap rounded-b-lg">
                                {{ entry.query }}
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <!-- Clear Confirmation Modal -->
        <div v-if="showClearConfirm"
            class="fixed inset-0 bg-background/80 backdrop-blur-sm z-[60] flex items-center justify-center p-4">
            <div
                class="bg-card w-full max-w-md rounded-lg border border-border shadow-lg overflow-hidden animate-in zoom-in-95 duration-200">
                <div class="p-4 border-b border-border flex items-center gap-3">
                    <div
                        class="w-10 h-10 rounded-full bg-destructive/10 flex items-center justify-center flex-shrink-0">
                        <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="text-destructive">
                            <path
                                d="m10.29 3.86 8.52 14.73c.47.81-.12 1.83-1.06 1.83H2.25c-.94 0-1.53-1.02-1.06-1.83L9.71 3.86c.47-.81 1.66-.81 2.13 0z" />
                            <line x1="12" x2="12" y1="9" y2="13" />
                            <line x1="12" x2="12.01" y1="17" y2="17" />
                        </svg>
                    </div>
                    <div>
                        <h3 class="font-semibold text-lg">Clear History</h3>
                        <p class="text-sm text-muted-foreground mt-1">Are you sure you want to clear all non-favorite
                            query history? This action cannot be undone.</p>
                    </div>
                </div>
                <div class="p-4 bg-muted/20 flex justify-end gap-3">
                    <button @click="showClearConfirm = false"
                        class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                        Cancel
                    </button>
                    <button @click="executeClearHistory"
                        class="px-4 py-2 text-sm font-medium rounded-md bg-destructive text-destructive-foreground hover:bg-destructive/90 transition-colors shadow-sm">
                        Clear History
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup>
import { ref, onMounted, watch, computed } from 'vue';
import { GetQueryHistory, ToggleFavoriteQuery, DeleteQueryHistory, ClearQueryHistory } from '../../wailsjs/go/main/App';

const props = defineProps({
    isOpen: {
        type: Boolean,
        default: false
    },
    connectionName: {
        type: String,
        default: ''
    }
});

const emit = defineEmits(['close', 'run-query']);

const history = ref([]);
const isLoading = ref(false);
const activeDbType = ref('all');
const showClearConfirm = ref(false);

const dbTypes = computed(() => {
    const types = new Set(history.value.map(h => h.db_type));
    return Array.from(types).sort();
});

const filteredHistory = computed(() => {
    let list = history.value;
    if (activeDbType.value !== 'all') {
        list = list.filter(h => h.db_type === activeDbType.value);
    }
    // Sort favorites to top, then by id descending
    return list.sort((a, b) => {
        if (a.is_favorite && !b.is_favorite) return -1;
        if (!a.is_favorite && b.is_favorite) return 1;
        return b.id - a.id;
    });
});

const fetchHistory = async () => {
    isLoading.value = true;
    try {
        const res = await GetQueryHistory('');
        history.value = res || [];

        // Auto-select current connection if it exists in history
        if (props.connectionName && activeDbType.value === 'all') {
            const hasCurrentConn = history.value.some(h => h.db_type === props.connectionName);
            if (hasCurrentConn) {
                activeDbType.value = props.connectionName;
            }
        }
    } catch (err) {
        console.error("Failed to load history", err);
    } finally {
        isLoading.value = false;
    }
};

const toggleFavorite = async (entry) => {
    try {
        const newVal = !entry.is_favorite;
        await ToggleFavoriteQuery(entry.id, newVal);
        entry.is_favorite = newVal;
    } catch (err) {
        console.error("Failed to toggle favorite", err);
    }
};

const deleteEntry = async (id) => {
    try {
        await DeleteQueryHistory(id);
        history.value = history.value.filter(h => h.id !== id);
    } catch (err) {
        console.error("Failed to delete entry", err);
    }
};

const clearHistory = async () => {
    showClearConfirm.value = true;
};

const executeClearHistory = async () => {
    try {
        await ClearQueryHistory();
        await fetchHistory();
        showClearConfirm.value = false;
    } catch (err) {
        console.error("Failed to clear history", err);
    }
};

const copyQuery = async (text) => {
    try {
        await navigator.clipboard.writeText(text);
        // You could emit a toast here
    } catch (err) {
        console.error("Failed to copy", err);
    }
};

const formatDate = (dateStr) => {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleString();
};

watch(() => props.isOpen, (newVal) => {
    if (newVal) {
        fetchHistory();
    }
});

onMounted(() => {
    if (props.isOpen) {
        fetchHistory();
    }
});
</script>

<template>
    <div v-if="isOpen"
        class="fixed inset-0 bg-background/80 z-50 flex items-center justify-center p-4"
        @mousedown.self="$emit('close')">
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
                    {{ t("common.queryHistory.title") }}
                </h2>
                <button @click="$emit('close')" :aria-label="t('common.queryHistory.closeAriaLabel')"
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
                    <div class="p-3 border-b border-border bg-muted/5 z-10 space-y-2">
                        <div class="flex flex-col md:flex-row gap-2 md:items-center md:justify-between">
                            <div class="relative flex-1 max-w-xl">
                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round"
                                    class="lucide lucide-search absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground">
                                    <circle cx="11" cy="11" r="8" />
                                    <path d="m21 21-4.3-4.3" />
                                </svg>
                                <input v-model="searchText" type="text"
                                    placeholder="Search SQL text, table names, clauses..."
                                    class="w-full h-9 pl-8 pr-10 rounded-md border border-input bg-background text-sm outline-none ring-offset-background focus-visible:ring-2 focus-visible:ring-ring" />
                                <button v-if="searchText" @click="searchText = ''" aria-label="Clear search text"
                                    class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded hover:bg-muted text-muted-foreground"
                                    title="Clear search">
                                    <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12"
                                        viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                        stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-x">
                                        <path d="M18 6 6 18" />
                                        <path d="m6 6 12 12" />
                                    </svg>
                                </button>
                            </div>
                            <button @click="clearHistory"
                                class="flex items-center justify-center gap-1.5 px-3 py-2 text-xs text-destructive hover:bg-destructive/10 rounded-md transition-colors whitespace-nowrap">
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
                        <div class="flex flex-wrap items-center gap-2">
                            <button @click="showFavoritesOnly = !showFavoritesOnly"
                                class="px-2.5 py-1.5 rounded-md text-xs border transition-colors"
                                :class="showFavoritesOnly ? 'border-yellow-500/50 bg-yellow-500/10 text-yellow-600' : 'border-border bg-background text-muted-foreground hover:text-foreground'">
                                Favorites Only
                            </button>
                            <select v-model="dateRange"
                                class="h-8 rounded-md border border-input bg-background px-2 text-xs">
                                <option value="all">All Time</option>
                                <option value="today">Today</option>
                                <option value="7d">Last 7 Days</option>
                                <option value="30d">Last 30 Days</option>
                            </select>
                            <select v-model="sortMode"
                                class="h-8 rounded-md border border-input bg-background px-2 text-xs">
                                <option value="recent">Newest First</option>
                                <option value="oldest">Oldest First</option>
                            </select>
                            <button v-if="hasActiveFilters" @click="resetFilters"
                                class="px-2.5 py-1.5 rounded-md text-xs border border-border bg-background text-muted-foreground hover:text-foreground">
                                Reset Filters
                            </button>
                            <div class="text-xs text-muted-foreground ml-auto">{{ history.length }} / {{ historySummary.total }}
                                queries</div>
                        </div>
                    </div>

                    <div v-if="!historyEnabled" class="mx-4 mt-3 rounded-md border border-blue-500/30 bg-blue-500/10 px-3 py-2 text-xs text-blue-700">
                        {{ t("common.queryHistory.disabledNotice") }}
                    </div>

                    <div v-if="actionError" class="mx-4 mt-3 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-xs text-destructive">
                        {{ actionError }}
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
                            <h3 class="text-base font-medium text-foreground mb-1">{{ t("common.queryHistory.emptyTitle") }}</h3>
                            <p class="text-sm">{{ emptyStateMessage }}</p>
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
                                    <button @click="copyQuery(entry.query)" aria-label="Copy query"
                                        class="p-1.5 rounded-md hover:bg-muted text-muted-foreground hover:text-foreground transition-colors"
                                        title="Copy code">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-copy">
                                            <rect width="14" height="14" x="8" y="8" rx="2" ry="2" />
                                            <path d="M4 16c-1.1 0-2-.9-2-2V4c0-1.1.9-2 2-2h10c1.1 0 2 .9 2 2" />
                                        </svg>
                                    </button>
                                    <button @click="$emit('run-query', entry.query)" aria-label="Load and run query"
                                        class="p-1.5 rounded-md hover:bg-primary/20 text-primary transition-colors"
                                        title="Load & Run">
                                        <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14"
                                            viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                            stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-play">
                                            <polygon points="5 3 19 12 5 21 5 3" />
                                        </svg>
                                    </button>
                                    <div class="w-px h-4 bg-border mx-1"></div>
                                    <button @click="toggleFavorite(entry)" :aria-label="entry.is_favorite ? 'Remove from favorites' : 'Add to favorites'"
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
                                    <button @click="deleteEntry(entry.id)" aria-label="Delete history record"
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
            class="fixed inset-0 bg-background/80 z-[60] flex items-center justify-center p-4">
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

<script lang="ts" setup>
import { computed, onMounted, onUnmounted, ref, shallowRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { ClearQueryHistory, DeleteQueryHistory, GetQueryHistorySummary, SearchQueryHistory, ToggleFavoriteQuery } from '../../wailsjs/go/app/App';
const { t } = useI18n({ useScope: 'global' });

type QueryHistoryEntry = {
    id: number;
    query: string;
    db_type: string;
    timestamp: string;
    is_favorite: boolean;
};

type QueryHistorySummary = {
    total: number;
    db_types: string[];
};

type ActionResult = {
    success: boolean;
    error: string;
};

type DateRange = 'all' | 'today' | '7d' | '30d';
type SortMode = 'recent' | 'oldest';

const props = withDefaults(defineProps<{
    isOpen: boolean;
    connectionName?: string;
    historyEnabled?: boolean;
}>(), {
    isOpen: false,
    connectionName: '',
    historyEnabled: true
});

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'run-query', query: string): void;
    (e: 'notify-success', message: string): void;
    (e: 'notify-error', message: string): void;
}>();

const history = ref<QueryHistoryEntry[]>([]);
const historySummary = ref<QueryHistorySummary>({
    total: 0,
    db_types: []
});
const isLoading = shallowRef(false);
const activeDbType = shallowRef('all');
const showClearConfirm = shallowRef(false);
const searchText = shallowRef('');
const showFavoritesOnly = shallowRef(false);
const dateRange = shallowRef<DateRange>('all');
const sortMode = shallowRef<SortMode>('recent');
const actionError = shallowRef('');

const dbTypes = computed(() => {
    return historySummary.value.db_types;
});

const hasActiveFilters = computed(() =>
    searchText.value.trim().length > 0 ||
    showFavoritesOnly.value ||
    dateRange.value !== 'all' ||
    activeDbType.value !== 'all'
);

const filteredHistory = computed(() => history.value);

const emptyStateMessage = computed(() => {
    if (historySummary.value.total === 0) {
        return "You haven't executed any queries yet.";
    }
    return 'No query history matches your current search filters.';
});

let searchDebounce: ReturnType<typeof setTimeout> | null = null;
let fetchHistoryToken = 0;

const fetchSummary = async () => {
    const summary = await GetQueryHistorySummary();
    historySummary.value = {
        total: summary?.total ?? 0,
        db_types: summary?.db_types ?? []
    };
};

const fetchHistory = async () => {
    const requestToken = ++fetchHistoryToken;
    isLoading.value = true;
    try {
        const res = await SearchQueryHistory(
            searchText.value.trim(),
            activeDbType.value,
            showFavoritesOnly.value,
            dateRange.value,
            sortMode.value,
            2000
        );
        if (requestToken !== fetchHistoryToken) {
            return;
        }
        history.value = res || [];
    } catch (err) {
        if (requestToken !== fetchHistoryToken) {
            return;
        }
        console.error("Failed to load history", err);
        actionError.value = 'Failed to load query history.';
    } finally {
        if (requestToken === fetchHistoryToken) {
            isLoading.value = false;
        }
    }
};

const ensureSuccess = (result: ActionResult, defaultMessage: string) => {
    if (!result?.success) {
        throw new Error(result?.error || defaultMessage);
    }
};

const toggleFavorite = async (entry: QueryHistoryEntry) => {
    try {
        actionError.value = '';
        const newVal = !entry.is_favorite;
        const result = await ToggleFavoriteQuery(entry.id, newVal);
        ensureSuccess(result, 'Failed to update favorite status');
        await Promise.all([fetchSummary(), fetchHistory()]);
        emit('notify-success', newVal ? 'Added to favorites.' : 'Removed from favorites.');
    } catch (err) {
        console.error("Failed to toggle favorite", err);
        actionError.value = err instanceof Error ? err.message : 'Failed to update favorite status.';
        emit('notify-error', actionError.value);
    }
};

const deleteEntry = async (id: number) => {
    try {
        actionError.value = '';
        const result = await DeleteQueryHistory(id);
        ensureSuccess(result, 'Failed to delete history entry');
        await Promise.all([fetchSummary(), fetchHistory()]);
        emit('notify-success', 'History entry deleted.');
    } catch (err) {
        console.error("Failed to delete entry", err);
        actionError.value = err instanceof Error ? err.message : 'Failed to delete history entry.';
        emit('notify-error', actionError.value);
    }
};

const clearHistory = () => {
    showClearConfirm.value = true;
};

const executeClearHistory = async () => {
    try {
        actionError.value = '';
        const result = await ClearQueryHistory();
        ensureSuccess(result, 'Failed to clear non-favorite history');
        await Promise.all([fetchSummary(), fetchHistory()]);
        showClearConfirm.value = false;
        emit('notify-success', 'Non-favorite query history cleared.');
    } catch (err) {
        console.error("Failed to clear history", err);
        actionError.value = err instanceof Error ? err.message : 'Failed to clear non-favorite history.';
        emit('notify-error', actionError.value);
    }
};

const copyQuery = async (text: string) => {
    try {
        await navigator.clipboard.writeText(text);
        emit('notify-success', 'Query copied to clipboard.');
    } catch (err) {
        console.error("Failed to copy", err);
        emit('notify-error', 'Failed to copy query.');
    }
};

const formatDate = (dateStr: string): string => {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    return date.toLocaleString();
};

const resetFilters = () => {
    searchText.value = '';
    showFavoritesOnly.value = false;
    dateRange.value = 'all';
    if (props.connectionName && dbTypes.value.includes(props.connectionName)) {
        activeDbType.value = props.connectionName;
        return;
    }
    activeDbType.value = 'all';
};

const handleEscapeKeydown = (event: KeyboardEvent) => {
    if (event.key !== 'Escape' || !props.isOpen) {
        return;
    }

    if (showClearConfirm.value) {
        showClearConfirm.value = false;
        return;
    }

    emit('close');
};

watch(() => props.connectionName, (newConnectionName) => {
    if (!newConnectionName || !props.isOpen) {
        return;
    }
    if (activeDbType.value === 'all' && dbTypes.value.includes(newConnectionName)) {
        activeDbType.value = newConnectionName;
    }
});

watch(() => props.isOpen, (isOpen) => {
    if (!isOpen) {
        return;
    }
    actionError.value = '';
    fetchSummary()
        .then(() => {
            if (props.connectionName && dbTypes.value.includes(props.connectionName)) {
                activeDbType.value = props.connectionName;
            }
            return fetchHistory();
        })
        .catch((err) => console.error("Failed to load history", err));
}, { immediate: true });

watch([searchText, showFavoritesOnly, dateRange, sortMode, activeDbType], () => {
    if (!props.isOpen) {
        return;
    }

    if (searchDebounce) {
        clearTimeout(searchDebounce);
    }
    searchDebounce = setTimeout(() => {
        fetchHistory().catch((err) => console.error("Failed to search history", err));
    }, 180);
});

onMounted(() => {
    window.addEventListener('keydown', handleEscapeKeydown);
});

onUnmounted(() => {
    window.removeEventListener('keydown', handleEscapeKeydown);
    if (searchDebounce) {
        clearTimeout(searchDebounce);
    }
    fetchHistoryToken += 1;
});
</script>

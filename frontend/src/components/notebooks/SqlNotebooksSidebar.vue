<script lang="ts" setup>
import { computed } from 'vue';

import type { SqlNotebook } from '../../types/sqlNotebook';

const props = defineProps<{
    notebooks: SqlNotebook[];
    activeNotebookId: string | null;
    searchQuery: string;
    isLoading?: boolean;
    activeFilters?: {
        favorites: boolean;
        environment: string | null;
    };
    environmentOptions?: string[];
}>();

const emit = defineEmits<{
    'update:search-query': [value: string];
    'select-notebook': [notebookId: string];
    'create-notebook': [];
    'import-notebook': [];
    'request-delete-notebook': [notebookId: string];
    'toggle-favorite': [notebookId: string];
    'duplicate-notebook': [notebookId: string];
    'update:filters': [value: { favorites: boolean; environment: string | null }];
}>();

const filters = computed(() => props.activeFilters ?? {
    favorites: false,
    environment: null,
});

const toggleFilter = (key: 'favorites') => {
    emit('update:filters', {
        ...filters.value,
        [key]: !filters.value[key],
    });
};

const updateEnvironmentFilter = (value: string | null) => {
    emit('update:filters', {
        ...filters.value,
        environment: value,
    });
};

const formatNotebookDate = (value: string): string => {
    try {
        return new Intl.DateTimeFormat(undefined, {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        }).format(new Date(value));
    } catch (_error) {
        return value;
    }
};
</script>

<template>
    <aside class="flex h-full w-80 flex-col border-r border-border bg-card">
        <div class="space-y-3 border-b border-border p-4">
            <div class="flex items-center justify-between gap-3">
                <div>
                    <h2 class="text-base font-semibold">SQL Notebooks</h2>
                    <p class="text-xs text-muted-foreground">Reusable SQL workflows for this connection.</p>
                </div>
                <div class="flex flex-wrap items-center justify-end gap-2">
                    <button
                        class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-sm font-medium transition-colors hover:bg-accent"
                        @click="emit('import-notebook')"
                    >
                        Import
                    </button>
                    <button
                        class="inline-flex h-9 items-center justify-center rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90"
                        @click="emit('create-notebook')"
                    >
                        New
                    </button>
                </div>
            </div>

            <input
                :value="props.searchQuery"
                type="text"
                placeholder="Search notes, tables, notebooks..."
                class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus:border-primary"
                @input="emit('update:search-query', ($event.target as HTMLInputElement).value)"
            />

            <div class="flex flex-wrap gap-2">
                <button
                    class="inline-flex h-8 items-center justify-center rounded-full border px-3 text-xs font-medium transition-colors"
                    :class="filters.favorites ? 'border-primary bg-primary/10 text-primary' : 'border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground'"
                    @click="toggleFilter('favorites')"
                >
                    Favorites
                </button>
                <button
                    class="inline-flex h-8 items-center justify-center rounded-full border px-3 text-xs font-medium transition-colors"
                    :class="filters.environment === null ? 'border-primary bg-primary/10 text-primary' : 'border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground'"
                    @click="updateEnvironmentFilter(null)"
                >
                    All Environments
                </button>
                <button
                    v-for="environment in props.environmentOptions || []"
                    :key="environment"
                    class="inline-flex h-8 items-center justify-center rounded-full border px-3 text-xs font-medium transition-colors"
                    :class="filters.environment === environment ? 'border-primary bg-primary/10 text-primary' : 'border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground'"
                    @click="updateEnvironmentFilter(environment)"
                >
                    {{ environment }}
                </button>
            </div>
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto p-3">
            <div v-if="props.isLoading" class="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
                Loading notebooks...
            </div>

            <div v-else-if="props.notebooks.length === 0" class="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
                {{ props.searchQuery.trim()
                    ? 'No notebooks matched. Try keywords from saved notes, table names, or notebook titles.'
                    : 'No notebooks yet. Create one when you want to save reusable SQL work.' }}
            </div>

            <div v-else class="space-y-2">
                <button
                    v-for="notebook in props.notebooks"
                    :key="notebook.id"
                    class="block w-full rounded-xl border px-3 py-3 text-left transition-colors"
                    :class="props.activeNotebookId === notebook.id
                        ? 'border-primary bg-primary/5'
                        : 'border-border bg-background hover:bg-accent/50'"
                    @click="emit('select-notebook', notebook.id)"
                >
                    <div class="flex items-start justify-between gap-3">
                        <div class="min-w-0 flex-1">
                            <div class="flex items-start gap-2">
                                <div class="line-clamp-2 flex-1 text-sm font-semibold leading-5">{{ notebook.title }}</div>
                            </div>
                            <div v-if="notebook.metadata.environment || notebook.metadata.owner" class="mt-2 flex flex-wrap items-center gap-2 text-[11px] text-muted-foreground">
                                <span v-if="notebook.metadata.environment" class="rounded-full bg-muted px-2 py-0.5">{{ notebook.metadata.environment }}</span>
                                <span v-if="notebook.metadata.owner" class="rounded-full bg-muted px-2 py-0.5">{{ notebook.metadata.owner }}</span>
                            </div>
                            <div class="mt-2 flex items-center gap-2 text-[11px] text-muted-foreground">
                                <span>{{ notebook.cells.length }} cells</span>
                                <span>&bull;</span>
                                <span>{{ formatNotebookDate(notebook.updatedAt) }}</span>
                            </div>
                        </div>
                        <div class="flex items-center gap-1">
                            <button
                                class="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                                :title="notebook.isFavorite ? 'Remove favorite' : 'Add favorite'"
                                @click.stop="emit('toggle-favorite', notebook.id)"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" :fill="notebook.isFavorite ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                                </svg>
                            </button>
                            <button
                                class="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                                title="Duplicate notebook"
                                @click.stop="emit('duplicate-notebook', notebook.id)"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <rect x="9" y="9" width="13" height="13" rx="2" ry="2" />
                                    <path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1" />
                                </svg>
                            </button>
                            <button
                                class="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-destructive/10 hover:text-destructive"
                                title="Delete notebook"
                                @click.stop="emit('request-delete-notebook', notebook.id)"
                            >
                                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                    <path d="M3 6h18" />
                                    <path d="M8 6V4h8v2" />
                                    <path d="M19 6l-1 14H6L5 6" />
                                    <path d="M10 11v6" />
                                    <path d="M14 11v6" />
                                </svg>
                            </button>
                        </div>
                    </div>
                </button>
            </div>
        </div>
    </aside>
</template>

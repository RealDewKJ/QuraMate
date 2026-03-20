<script lang="ts" setup>
import { computed, ref } from 'vue';

import type { SqlNotebook } from '../../types/sqlNotebook';

const props = defineProps<{
    notebook: SqlNotebook;
    sqlCellCount: number;
    markdownCellCount: number;
    runbookCellCount: number;
    isDirty: boolean;
    isSaving: boolean;
    lastSavedAt: string | null;
    isReadOnly?: boolean;
}>();

const emit = defineEmits<{
    'update-title': [value: string];
    'update-description': [value: string];
    'update-tags': [value: string[]];
    'update-metadata': [value: { environment?: string; purpose?: string; owner?: string }];
    'toggle-favorite': [];
    'add-cell': [type: 'sql' | 'markdown' | 'runbook'];
    share: [];
    save: [];
}>();

const isDetailsOpen = ref(false);

const notebookSummary = computed(() => {
    const parts = [
        props.sqlCellCount > 0 ? `${props.sqlCellCount} SQL` : '',
        props.markdownCellCount > 0 ? `${props.markdownCellCount} Notes` : '',
        props.runbookCellCount > 0 ? `${props.runbookCellCount} Runbook` : '',
    ].filter(Boolean);

    return parts.join(' • ') || 'Empty notebook';
});

const metadataChips = computed(() => {
    return [
        props.notebook.metadata.environment ? `Env: ${props.notebook.metadata.environment}` : '',
        props.notebook.metadata.owner ? `Owner: ${props.notebook.metadata.owner}` : '',
        props.notebook.metadata.purpose ? `Purpose: ${props.notebook.metadata.purpose}` : '',
        ...props.notebook.tags.slice(0, 3).map((tag) => `#${tag}`),
    ].filter(Boolean);
});
</script>

<template>
    <section class="rounded-xl border border-border bg-card p-4 shadow-sm">
        <div class="flex flex-wrap items-start justify-between gap-4">
            <div class="min-w-0 flex-1 space-y-3">
                <input
                    :value="props.notebook.title"
                    type="text"
                    class="w-full rounded-md border border-input bg-background px-3 py-2 text-lg font-semibold outline-none transition-colors focus:border-primary"
                    placeholder="Notebook title"
                    @input="emit('update-title', ($event.target as HTMLInputElement).value)"
                />
                <textarea
                    :value="props.notebook.description"
                    rows="2"
                    class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm text-muted-foreground outline-none transition-colors focus:border-primary"
                    placeholder="Describe the goal, scope, and operating context for this notebook."
                    @input="emit('update-description', ($event.target as HTMLTextAreaElement).value)"
                />

                <div class="flex flex-wrap items-center gap-2">
                    <span class="rounded-full bg-muted px-2.5 py-1 text-[11px] font-medium text-muted-foreground">
                        {{ notebookSummary }}
                    </span>
                    <span
                        v-for="chip in metadataChips"
                        :key="chip"
                        class="rounded-full bg-muted px-2.5 py-1 text-[11px] font-medium text-muted-foreground"
                    >
                        {{ chip }}
                    </span>
                </div>

                <button
                    type="button"
                    class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-sm font-medium transition-colors hover:bg-accent"
                    @click="isDetailsOpen = !isDetailsOpen"
                >
                    {{ isDetailsOpen ? 'Hide Details' : 'Notebook Details' }}
                </button>

                <div v-if="isDetailsOpen" class="grid gap-3 rounded-lg border border-border bg-muted/10 p-3 md:grid-cols-3">
                    <input
                        :value="props.notebook.metadata.environment"
                        type="text"
                        class="h-10 rounded-md border border-input bg-background px-3 text-sm outline-none transition-colors focus:border-primary"
                        placeholder="Environment"
                        @input="emit('update-metadata', { environment: ($event.target as HTMLInputElement).value })"
                    />
                    <input
                        :value="props.notebook.metadata.owner"
                        type="text"
                        class="h-10 rounded-md border border-input bg-background px-3 text-sm outline-none transition-colors focus:border-primary"
                        placeholder="Owner"
                        @input="emit('update-metadata', { owner: ($event.target as HTMLInputElement).value })"
                    />
                    <input
                        :value="props.notebook.metadata.purpose"
                        type="text"
                        class="h-10 rounded-md border border-input bg-background px-3 text-sm outline-none transition-colors focus:border-primary"
                        placeholder="Purpose"
                        @input="emit('update-metadata', { purpose: ($event.target as HTMLInputElement).value })"
                    />
                    <input
                        :value="props.notebook.tags.join(', ')"
                        type="text"
                        class="h-10 rounded-md border border-input bg-background px-3 text-sm outline-none transition-colors focus:border-primary md:col-span-3"
                        placeholder="Tags (comma separated)"
                        @input="emit('update-tags', ($event.target as HTMLInputElement).value.split(',').map((tag) => tag.trim()).filter(Boolean))"
                    />
                </div>
            </div>

            <div class="flex min-w-[220px] flex-col gap-3 rounded-lg border border-border bg-muted/20 p-3">
                <div class="rounded-md border border-dashed border-border bg-background/80 px-3 py-2 text-xs text-muted-foreground">
                    <span v-if="props.isSaving">Saving...</span>
                    <span v-else-if="props.isDirty">Unsaved changes</span>
                    <span v-else-if="props.lastSavedAt">Saved</span>
                    <span v-else>No local saves yet</span>
                    <span class="ml-2 text-[11px]">Ctrl+S</span>
                </div>

                <div class="flex flex-wrap gap-2">
                    <button
                        class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                        :title="props.notebook.isFavorite ? 'Remove favorite' : 'Add favorite'"
                        @click="emit('toggle-favorite')"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" :fill="props.notebook.isFavorite ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                        </svg>
                    </button>
                    <button
                        class="inline-flex h-9 w-9 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                        title="Share notebook"
                        @click="emit('share')"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                            <circle cx="18" cy="5" r="3" />
                            <circle cx="6" cy="12" r="3" />
                            <circle cx="18" cy="19" r="3" />
                            <path d="m8.59 13.51 6.83 3.98" />
                            <path d="m15.41 6.51-6.82 3.98" />
                        </svg>
                    </button>
                </div>

                <div class="flex flex-wrap gap-2">
                    <button
                        class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-sm font-medium transition-colors hover:bg-accent"
                        @click="emit('add-cell', 'sql')"
                    >
                        Add SQL Cell
                    </button>
                    <button
                        class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-sm font-medium transition-colors hover:bg-accent"
                        @click="emit('add-cell', 'markdown')"
                    >
                        Add Note
                    </button>
                    <button
                        class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-sm font-medium transition-colors hover:bg-accent"
                        @click="emit('add-cell', 'runbook')"
                    >
                        Add Runbook Step
                    </button>
                    <button
                        class="inline-flex h-9 items-center justify-center rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50"
                        :disabled="props.isReadOnly || !props.isDirty || props.isSaving"
                        @click="emit('save')"
                    >
                        {{ props.isSaving ? 'Saving...' : 'Save' }}
                    </button>
                </div>
            </div>
        </div>
    </section>
</template>

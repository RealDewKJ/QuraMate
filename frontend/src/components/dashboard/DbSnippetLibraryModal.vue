<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

export interface DashboardSnippetItem {
    id: string;
    title: string;
    description: string;
    sql: string;
    category: string;
    dbTypes?: string[];
    isBuiltIn?: boolean;
}

interface Props {
    isOpen: boolean;
    snippets: DashboardSnippetItem[];
    dbType: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'apply', payload: { snippet: DashboardSnippetItem; resolvedSql: string }): void;
    (e: 'save-custom', payload: { title: string; description: string; sql: string; category: string }): void;
    (e: 'delete-custom', snippetId: string): void;
}>();

const selectedSnippetId = ref<string>('');
const draftTitle = ref('');
const draftDescription = ref('');
const draftCategory = ref('Custom');
const draftSql = ref('');
const variableValues = ref<Record<string, string>>({});

const normalizedDbType = computed(() => (props.dbType || '').toLowerCase());
const visibleSnippets = computed(() =>
    props.snippets.filter((snippet) => !snippet.dbTypes || snippet.dbTypes.some((db) => normalizedDbType.value.includes(db)))
);
const selectedSnippet = computed(() => visibleSnippets.value.find((snippet) => snippet.id === selectedSnippetId.value) || null);
const selectedVariables = computed(() => {
    if (!selectedSnippet.value) {
        return [];
    }
    const matches = selectedSnippet.value.sql.match(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g) || [];
    return Array.from(new Set(matches.map((match) => match.replace(/\{\{|\}\}/g, '').trim())));
});
const resolvedSelectedSql = computed(() => {
    if (!selectedSnippet.value) {
        return '';
    }
    return selectedSnippet.value.sql.replace(/\{\{\s*([a-zA-Z0-9_]+)\s*\}\}/g, (_, variableName: string) => {
        return variableValues.value[variableName] ?? `{{${variableName}}}`;
    });
});

watch(selectedSnippet, (snippet) => {
    if (!snippet) {
        variableValues.value = {};
        return;
    }
    const nextValues: Record<string, string> = {};
    for (const variableName of selectedVariables.value) {
        nextValues[variableName] = variableValues.value[variableName] || '';
    }
    variableValues.value = nextValues;
});

const saveCustomSnippet = () => {
    const title = draftTitle.value.trim();
    const sql = draftSql.value.trim();
    if (!title || !sql) {
        return;
    }

    emit('save-custom', {
        title,
        description: draftDescription.value.trim(),
        sql,
        category: draftCategory.value.trim() || 'Custom',
    });

    draftTitle.value = '';
    draftDescription.value = '';
    draftCategory.value = 'Custom';
    draftSql.value = '';
};

const applySelectedSnippet = () => {
    if (!selectedSnippet.value) {
        return;
    }
    emit('apply', {
        snippet: selectedSnippet.value,
        resolvedSql: resolvedSelectedSql.value,
    });
};
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 p-4 text-left" @mousedown.self="emit('close')">
        <div class="bg-card w-full max-w-6xl rounded-lg shadow-lg border border-border overflow-hidden animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center justify-between px-6 py-4 border-b border-border">
                <div>
                    <h3 class="text-lg font-semibold text-foreground">Saved Snippets & Runbooks</h3>
                    <p class="text-sm text-muted-foreground">Built-in templates by database type plus your saved runbooks.</p>
                </div>
                <button @click="emit('close')" class="text-muted-foreground hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-[300px_minmax(0,1fr)_360px] min-h-[620px]">
                <div class="border-r border-border overflow-auto">
                    <div class="px-4 py-3 border-b border-border text-xs font-semibold uppercase tracking-wide text-muted-foreground">
                        Library
                    </div>
                    <div class="p-2 space-y-1">
                        <button
                            v-for="snippet in visibleSnippets"
                            :key="snippet.id"
                            @click="selectedSnippetId = snippet.id"
                            class="w-full rounded-md border px-3 py-2 text-left transition-colors"
                            :class="selectedSnippetId === snippet.id ? 'border-primary bg-primary/5' : 'border-transparent hover:border-border hover:bg-muted/40'"
                        >
                            <div class="flex items-center justify-between gap-2">
                                <span class="text-sm font-medium text-foreground">{{ snippet.title }}</span>
                                <span class="text-[10px] uppercase tracking-wide text-muted-foreground">{{ snippet.category }}</span>
                            </div>
                            <p class="mt-1 text-xs text-muted-foreground line-clamp-2">{{ snippet.description }}</p>
                        </button>
                    </div>
                </div>

                <div class="flex flex-col min-w-0">
                    <div class="px-5 py-3 border-b border-border flex items-center justify-between gap-4">
                        <div>
                            <h4 class="text-base font-semibold text-foreground">{{ selectedSnippet?.title || 'Select a snippet' }}</h4>
                            <p class="text-xs text-muted-foreground">{{ selectedSnippet?.description || 'Preview SQL templates and apply them into the editor.' }}</p>
                        </div>
                        <div class="flex items-center gap-2">
                            <button
                                v-if="selectedSnippet && !selectedSnippet.isBuiltIn"
                                @click="emit('delete-custom', selectedSnippet.id)"
                                class="px-3 py-1.5 text-xs rounded-md border border-destructive text-destructive hover:bg-destructive/10 transition-colors"
                            >
                                Delete
                            </button>
                            <button
                                v-if="selectedSnippet"
                                @click="applySelectedSnippet"
                                class="px-3 py-1.5 text-xs rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors"
                            >
                                Apply to Editor
                            </button>
                        </div>
                    </div>

                    <div class="flex-1 overflow-auto p-5 bg-muted/10">
                        <div v-if="selectedSnippet" class="space-y-4">
                            <div v-if="selectedVariables.length > 0" class="rounded-md border border-border bg-background p-4">
                                <h5 class="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-3">Template Variables</h5>
                                <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                                    <label v-for="variableName in selectedVariables" :key="variableName" class="block">
                                        <span class="text-xs font-medium text-muted-foreground">{{ variableName }}</span>
                                        <input
                                            v-model="variableValues[variableName]"
                                            class="mt-1 w-full h-9 rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring"
                                        >
                                    </label>
                                </div>
                            </div>
                            <pre class="whitespace-pre-wrap break-words text-[12px] leading-relaxed font-mono text-foreground">{{ resolvedSelectedSql }}</pre>
                        </div>
                        <div v-else class="h-full flex items-center justify-center text-sm text-muted-foreground">
                            Choose a snippet from the left to preview it here.
                        </div>
                    </div>
                </div>

                <div class="border-l border-border bg-muted/10">
                    <div class="px-4 py-3 border-b border-border">
                        <h4 class="text-sm font-semibold text-foreground">Save Custom Runbook</h4>
                        <p class="text-xs text-muted-foreground mt-1">Store repeatable DDL, DML, or maintenance SQL for this connection type.</p>
                    </div>
                    <div class="p-4 space-y-3">
                        <label class="block">
                            <span class="text-xs font-medium text-muted-foreground">Title</span>
                            <input v-model="draftTitle" class="mt-1 w-full h-10 rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring">
                        </label>
                        <label class="block">
                            <span class="text-xs font-medium text-muted-foreground">Category</span>
                            <input v-model="draftCategory" class="mt-1 w-full h-10 rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring">
                        </label>
                        <label class="block">
                            <span class="text-xs font-medium text-muted-foreground">Description</span>
                            <input v-model="draftDescription" class="mt-1 w-full h-10 rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring">
                        </label>
                        <label class="block">
                            <span class="text-xs font-medium text-muted-foreground">SQL Template</span>
                            <textarea v-model="draftSql" rows="12" class="mt-1 w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-mono outline-none focus:ring-2 focus:ring-ring"></textarea>
                        </label>
                        <button
                            @click="saveCustomSnippet"
                            class="w-full h-10 rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors disabled:opacity-50"
                            :disabled="!draftTitle.trim() || !draftSql.trim()"
                        >
                            Save Runbook
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

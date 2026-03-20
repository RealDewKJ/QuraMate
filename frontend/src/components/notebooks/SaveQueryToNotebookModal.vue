<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import type { SqlNotebook } from '../../types/sqlNotebook';

const props = defineProps<{
    isOpen: boolean;
    notebooks: SqlNotebook[];
    preferredNotebookId?: string | null;
    queryTitle: string;
    queryText: string;
}>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'confirm', payload: { notebookId?: string; notebookTitle?: string; sqlTitle: string; description?: string }): void;
}>();

const mode = ref<'existing' | 'new'>('existing');
const selectedNotebookId = ref('');
const sqlTitle = ref('');
const newNotebookTitle = ref('');
const description = ref('');

watch(
    () => props.isOpen,
    (isOpen) => {
        if (!isOpen) {
            return;
        }

        sqlTitle.value = props.queryTitle || 'Saved Query';
        selectedNotebookId.value = props.preferredNotebookId && props.notebooks.some((notebook) => notebook.id === props.preferredNotebookId)
            ? props.preferredNotebookId
            : props.notebooks[0]?.id || '';
        mode.value = props.notebooks.length > 0 ? 'existing' : 'new';
        newNotebookTitle.value = props.queryTitle ? `${props.queryTitle} Notes` : 'New SQL Notebook';
        description.value = '';
    },
    { immediate: true },
);

const canSubmit = computed(() => {
    if (!sqlTitle.value.trim() || !props.queryText.trim()) {
        return false;
    }

    if (mode.value === 'new') {
        return newNotebookTitle.value.trim().length > 0;
    }

    return selectedNotebookId.value.length > 0;
});

const submit = () => {
    if (!canSubmit.value) {
        return;
    }

    emit('confirm', {
        notebookId: mode.value === 'existing' ? selectedNotebookId.value : undefined,
        notebookTitle: mode.value === 'new' ? newNotebookTitle.value.trim() : undefined,
        sqlTitle: sqlTitle.value.trim(),
        description: description.value.trim() || undefined,
    });
};
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 p-4" @mousedown.self="emit('close')">
        <div class="w-full max-w-2xl overflow-hidden rounded-xl border border-border bg-card shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div class="border-b border-border px-6 py-4">
                <h3 class="text-lg font-semibold text-foreground">Save To SQL Notebook</h3>
                <p class="mt-1 text-sm text-muted-foreground">Store the current query in an existing notebook or create a new notebook for it.</p>
            </div>

            <div class="grid gap-5 p-6 lg:grid-cols-[320px_minmax(0,1fr)]">
                <div class="space-y-4">
                    <div class="rounded-xl border border-border bg-muted/20 p-4">
                        <div class="text-xs uppercase tracking-[0.12em] text-muted-foreground">Target</div>
                        <div class="mt-3 grid gap-2">
                            <button
                                class="rounded-md border px-3 py-2 text-left text-sm transition-colors"
                                :class="mode === 'existing' ? 'border-primary bg-primary/5' : 'border-border bg-background hover:bg-accent/50'"
                                @click="mode = 'existing'"
                            >
                                Existing Notebook
                            </button>
                            <button
                                class="rounded-md border px-3 py-2 text-left text-sm transition-colors"
                                :class="mode === 'new' ? 'border-primary bg-primary/5' : 'border-border bg-background hover:bg-accent/50'"
                                @click="mode = 'new'"
                            >
                                New Notebook
                            </button>
                        </div>
                    </div>

                    <div v-if="mode === 'existing'" class="space-y-2">
                        <label class="text-sm font-medium text-foreground">Notebook</label>
                        <select
                            v-model="selectedNotebookId"
                            class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:border-primary"
                        >
                            <option v-for="notebook in props.notebooks" :key="notebook.id" :value="notebook.id">
                                {{ notebook.title }}
                            </option>
                        </select>
                    </div>

                    <div v-else class="space-y-2">
                        <label class="text-sm font-medium text-foreground">New Notebook Title</label>
                        <input
                            v-model="newNotebookTitle"
                            type="text"
                            class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:border-primary"
                        />
                    </div>
                </div>

                <div class="space-y-4">
                    <div class="space-y-2">
                        <label class="text-sm font-medium text-foreground">SQL Cell Title</label>
                        <input
                            v-model="sqlTitle"
                            type="text"
                            class="h-10 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:border-primary"
                        />
                    </div>

                    <div class="space-y-2">
                        <label class="text-sm font-medium text-foreground">Notebook Description</label>
                        <textarea
                            v-model="description"
                            rows="3"
                            placeholder="Optional context for the notebook when creating or enriching it."
                            class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none focus:border-primary"
                        ></textarea>
                    </div>

                    <div class="space-y-2">
                        <div class="text-sm font-medium text-foreground">Query Preview</div>
                        <pre class="max-h-64 overflow-auto rounded-lg border border-border bg-muted/40 p-3 text-xs leading-6 text-foreground"><code>{{ props.queryText }}</code></pre>
                    </div>
                </div>
            </div>

            <div class="flex justify-end gap-3 border-t border-border px-6 py-4">
                <button
                    class="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('close')"
                >
                    Cancel
                </button>
                <button
                    class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:pointer-events-none disabled:opacity-50"
                    :disabled="!canSubmit"
                    @click="submit"
                >
                    Save To Notebook
                </button>
            </div>
        </div>
    </div>
</template>

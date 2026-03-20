<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import type { SqlNotebook, SqlNotebookTemplatePreset } from '../../types/sqlNotebook';

const props = defineProps<{
    isOpen: boolean;
    templates: SqlNotebook[];
    presets: SqlNotebookTemplatePreset[];
    favoritePresetIds: string[];
}>();

const emit = defineEmits<{
    createFromTemplate: [payload: { templateId: string; title: string }];
    createFromPreset: [payload: { presetId: string; title: string }];
    toggleFavoritePreset: [presetId: string];
    close: [];
}>();

const selectedTemplateId = ref('');
const notebookTitle = ref('');
const presetCategoryFilter = ref<'all' | 'operations' | 'release' | 'analysis'>('all');

const selectedTemplate = computed(() => {
    return props.templates.find((template) => template.id === selectedTemplateId.value) ?? null;
});

const visiblePresets = computed(() => {
    const presets = presetCategoryFilter.value === 'all'
        ? props.presets
        : props.presets.filter((preset) => preset.category === presetCategoryFilter.value);

    return [...presets].sort((left, right) => {
        const leftFavorite = props.favoritePresetIds.includes(left.id);
        const rightFavorite = props.favoritePresetIds.includes(right.id);
        if (leftFavorite !== rightFavorite) {
            return leftFavorite ? -1 : 1;
        }
        return left.title.localeCompare(right.title);
    });
});

watch(
    () => props.isOpen,
    (isOpen) => {
        if (!isOpen) {
            return;
        }

        const firstTemplate = props.templates[0] ?? null;
        selectedTemplateId.value = firstTemplate?.id || '';
        notebookTitle.value = firstTemplate ? `${firstTemplate.title} Run` : '';
    },
    { immediate: true },
);

watch(selectedTemplate, (template) => {
    if (!template) {
        return;
    }

    if (!notebookTitle.value.trim() || notebookTitle.value.endsWith(' Run')) {
        notebookTitle.value = `${template.title} Run`;
    }
});

const handleSubmit = () => {
    if (!selectedTemplateId.value) {
        return;
    }

    emit('createFromTemplate', {
        templateId: selectedTemplateId.value,
        title: notebookTitle.value.trim(),
    });
};

const handleCreateFromPreset = (preset: SqlNotebookTemplatePreset) => {
    emit('createFromPreset', {
        presetId: preset.id,
        title: `${preset.title} Run`,
    });
};

const toggleFavoritePreset = (presetId: string) => {
    emit('toggleFavoritePreset', presetId);
};
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[70] flex items-center justify-center bg-black/55 p-4">
        <div class="w-full max-w-2xl rounded-lg border border-border bg-card p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-start justify-between gap-4">
                <div>
                    <h3 class="text-lg font-semibold text-foreground">Create From Template</h3>
                    <p class="mt-1 text-sm text-muted-foreground">Start a new notebook from a reusable template without changing the source template.</p>
                </div>
                <button
                    class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('close')"
                >
                    Close
                </button>
            </div>

            <div v-if="props.templates.length === 0" class="mt-6 rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
                No templates yet. Mark a notebook as template first, then come back to create a run from it.
            </div>

            <div class="mt-6 space-y-4">
                <div class="rounded-xl border border-border bg-background/70 p-4">
                    <div class="flex items-center justify-between gap-3">
                        <div>
                            <div class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Quick Presets</div>
                            <div class="mt-1 text-sm text-muted-foreground">Start from a built-in notebook shape when you do not want to prepare a template first.</div>
                        </div>
                    </div>

                    <div class="mt-4 flex flex-wrap gap-2">
                        <button
                            class="inline-flex h-8 items-center justify-center rounded-full border px-3 text-xs font-medium transition-colors"
                            :class="presetCategoryFilter === 'all' ? 'border-primary bg-primary/10 text-primary' : 'border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground'"
                            @click="presetCategoryFilter = 'all'"
                        >
                            All
                        </button>
                        <button
                            v-for="category in ['operations', 'release', 'analysis']"
                            :key="category"
                            class="inline-flex h-8 items-center justify-center rounded-full border px-3 text-xs font-medium capitalize transition-colors"
                            :class="presetCategoryFilter === category ? 'border-primary bg-primary/10 text-primary' : 'border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground'"
                            @click="presetCategoryFilter = category as 'operations' | 'release' | 'analysis'"
                        >
                            {{ category }}
                        </button>
                    </div>

                    <div class="mt-4 grid gap-3 md:grid-cols-3">
                        <button
                            v-for="preset in visiblePresets"
                            :key="preset.id"
                            class="rounded-xl border border-border bg-card p-4 text-left transition-colors hover:border-primary hover:bg-primary/5"
                            @click="handleCreateFromPreset(preset)"
                        >
                            <div class="flex items-start justify-between gap-2">
                                <div class="text-sm font-semibold">{{ preset.title }}</div>
                                <button
                                    class="inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                                    type="button"
                                    @click.stop="toggleFavoritePreset(preset.id)"
                                >
                                    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" :fill="props.favoritePresetIds.includes(preset.id) ? 'currentColor' : 'none'" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                                        <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2" />
                                    </svg>
                                </button>
                            </div>
                            <p class="mt-1 text-xs text-muted-foreground">{{ preset.description }}</p>
                            <div class="mt-3 flex flex-wrap gap-2 text-[11px] text-muted-foreground">
                                <span v-if="preset.metadata.environment" class="rounded-full bg-muted px-2 py-0.5">{{ preset.metadata.environment }}</span>
                                <span class="rounded-full bg-muted px-2 py-0.5 capitalize">{{ preset.category }}</span>
                                <span>{{ preset.cells.length }} cells</span>
                            </div>
                        </button>
                    </div>
                </div>

                <div v-if="props.templates.length > 0" class="grid gap-4 lg:grid-cols-[minmax(0,0.95fr)_minmax(0,1.05fr)]">
                <div class="space-y-2 rounded-xl border border-border bg-background/80 p-3">
                    <div class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Templates</div>
                    <button
                        v-for="template in props.templates"
                        :key="template.id"
                        class="block w-full rounded-lg border px-3 py-3 text-left transition-colors"
                        :class="selectedTemplateId === template.id ? 'border-primary bg-primary/5' : 'border-border bg-card hover:bg-accent/50'"
                        @click="selectedTemplateId = template.id"
                    >
                        <div class="text-sm font-semibold">{{ template.title }}</div>
                        <div class="mt-1 line-clamp-2 text-xs text-muted-foreground">{{ template.description }}</div>
                        <div class="mt-2 flex flex-wrap gap-2 text-[11px] text-muted-foreground">
                            <span v-if="template.metadata.environment" class="rounded-full bg-muted px-2 py-0.5">{{ template.metadata.environment }}</span>
                            <span>{{ template.cells.length }} cells</span>
                            <span>{{ template.variables.length }} variables</span>
                        </div>
                    </button>
                </div>

                <div class="space-y-4 rounded-xl border border-border bg-background/70 p-4">
                    <div>
                        <label class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">New Notebook Title</label>
                        <input
                            v-model="notebookTitle"
                            type="text"
                            class="mt-2 h-10 w-full rounded-md border border-input bg-background px-3 text-sm outline-none transition-colors focus:border-primary"
                            placeholder="Notebook title"
                        />
                    </div>

                    <div v-if="selectedTemplate" class="space-y-3">
                        <div>
                            <div class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Template Summary</div>
                            <div class="mt-2 text-sm font-medium">{{ selectedTemplate.title }}</div>
                            <p class="mt-1 text-sm text-muted-foreground">{{ selectedTemplate.description }}</p>
                        </div>

                        <div class="grid gap-3 sm:grid-cols-2">
                            <div class="rounded-lg border border-border bg-card px-3 py-2">
                                <div class="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Environment</div>
                                <div class="mt-1 text-sm font-medium">{{ selectedTemplate.metadata.environment || 'Not set' }}</div>
                            </div>
                            <div class="rounded-lg border border-border bg-card px-3 py-2">
                                <div class="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Owner</div>
                                <div class="mt-1 text-sm font-medium">{{ selectedTemplate.metadata.owner || 'Not set' }}</div>
                            </div>
                        </div>

                        <div class="rounded-lg border border-border bg-card px-3 py-3 text-sm text-muted-foreground">
                            The new notebook keeps cells, metadata, tags, and variables from the template, but starts with fresh run results and empty snapshots.
                        </div>
                    </div>

                    <div class="flex justify-end gap-3">
                        <button
                            class="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 text-sm font-medium transition-colors hover:bg-accent"
                            @click="emit('close')"
                        >
                            Cancel
                        </button>
                        <button
                            class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
                            :disabled="!selectedTemplateId"
                            @click="handleSubmit"
                        >
                            Create Notebook
                        </button>
                    </div>
                </div>
                </div>
            </div>
        </div>
    </div>
</template>

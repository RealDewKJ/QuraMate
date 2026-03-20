<script lang="ts" setup>
type ExportFormat = 'json' | 'csv' | 'sql' | 'excel';

interface FormatOption {
    value: ExportFormat;
    label: string;
    description: string;
}

interface Props {
    isOpen: boolean;
    tableName: string;
    format: ExportFormat;
    includeSchema: boolean;
    includeData: boolean;
    dropIfExists: boolean;
    isLoading: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'confirm'): void;
    (e: 'update:format', value: ExportFormat): void;
    (e: 'update:include-schema', value: boolean): void;
    (e: 'update:include-data', value: boolean): void;
    (e: 'update:drop-if-exists', value: boolean): void;
}>();

const formatOptions: FormatOption[] = [
    { value: 'sql', label: 'SQL', description: 'Portable script with INSERT statements and optional CREATE TABLE.' },
    { value: 'excel', label: 'Excel', description: 'Spreadsheet export for sharing and quick review.' },
    { value: 'csv', label: 'CSV', description: 'Flat text export for import into many tools.' },
    { value: 'json', label: 'JSON', description: 'Structured export for APIs, backups, and automation.' },
];
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/55 p-4 text-left">
        <div class="w-full max-w-2xl rounded-2xl border border-border bg-card p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-start justify-between gap-4">
                <div class="space-y-1">
                    <h3 class="text-xl font-semibold text-foreground">Export Table</h3>
                    <p class="text-sm text-muted-foreground">
                        Choose the export format for <span class="font-medium text-foreground">{{ tableName }}</span>.
                    </p>
                </div>
                <button @click="emit('close')" class="rounded-md p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div class="mt-6 grid gap-3 md:grid-cols-2">
                <button
                    v-for="option in formatOptions"
                    :key="option.value"
                    @click="emit('update:format', option.value)"
                    class="rounded-xl border p-4 text-left transition-all"
                    :class="format === option.value
                        ? 'border-primary bg-primary/10 shadow-sm'
                        : 'border-border bg-background/60 hover:border-primary/40 hover:bg-accent/40'"
                >
                    <div class="flex items-center justify-between gap-3">
                        <span class="text-sm font-semibold text-foreground">{{ option.label }}</span>
                        <span
                            class="h-3 w-3 rounded-full border"
                            :class="format === option.value ? 'border-primary bg-primary' : 'border-muted-foreground/40 bg-transparent'"
                        ></span>
                    </div>
                    <p class="mt-2 text-sm leading-relaxed text-muted-foreground">{{ option.description }}</p>
                </button>
            </div>

            <div class="mt-6 grid gap-3 md:grid-cols-2">
                <label class="flex items-start gap-3 rounded-xl border border-border bg-background/60 p-4 transition-colors hover:bg-accent/30">
                    <input
                        type="checkbox"
                        class="mt-1 h-4 w-4 rounded border-input bg-background text-primary focus:ring-primary"
                        :checked="includeData"
                        @change="emit('update:include-data', ($event.target as HTMLInputElement).checked)"
                    >
                    <span class="space-y-1">
                        <span class="block text-sm font-semibold text-foreground">Export table data</span>
                        <span class="block text-sm text-muted-foreground">Writes all rows from the selected table.</span>
                    </span>
                </label>

                <label class="flex items-start gap-3 rounded-xl border border-border bg-background/60 p-4 transition-colors hover:bg-accent/30">
                    <input
                        type="checkbox"
                        class="mt-1 h-4 w-4 rounded border-input bg-background text-primary focus:ring-primary"
                        :checked="includeSchema"
                        @change="emit('update:include-schema', ($event.target as HTMLInputElement).checked)"
                    >
                    <span class="space-y-1">
                        <span class="block text-sm font-semibold text-foreground">Include CREATE TABLE script</span>
                        <span class="block text-sm text-muted-foreground">
                            For CSV, JSON, and Excel, the schema will be saved as a sibling <code>.schema.sql</code> file.
                        </span>
                    </span>
                </label>
            </div>

            <label
                v-if="format === 'sql'"
                class="mt-3 flex items-start gap-3 rounded-xl border border-border bg-background/60 p-4 transition-colors hover:bg-accent/30"
                :class="{ 'opacity-60': !includeSchema }"
            >
                <input
                    type="checkbox"
                    class="mt-1 h-4 w-4 rounded border-input bg-background text-primary focus:ring-primary"
                    :checked="dropIfExists"
                    :disabled="!includeSchema"
                    @change="emit('update:drop-if-exists', ($event.target as HTMLInputElement).checked)"
                >
                <span class="space-y-1">
                    <span class="block text-sm font-semibold text-foreground">Add DROP TABLE IF EXISTS first</span>
                    <span class="block text-sm text-muted-foreground">
                        Prepends a safe drop statement before <code>CREATE TABLE</code> so the SQL script can recreate the table cleanly.
                    </span>
                </span>
            </label>

            <div class="mt-5 rounded-xl border border-amber-500/30 bg-amber-500/10 p-4 text-sm text-amber-950 dark:text-amber-100">
                SQL imports are now restricted to <code class="font-semibold text-amber-900 dark:text-amber-50">CREATE TABLE</code> and <code class="font-semibold text-amber-900 dark:text-amber-50">INSERT INTO</code> statements for the selected table only.
            </div>

            <div class="mt-6 flex items-center justify-end gap-3">
                <button
                    @click="emit('close')"
                    class="rounded-md border border-input bg-background px-4 py-2 text-sm font-medium transition-colors hover:bg-accent"
                >
                    Cancel
                </button>
                <button
                    @click="emit('confirm')"
                    :disabled="isLoading || (!includeData && !includeSchema)"
                    class="flex h-10 items-center gap-2 rounded-md bg-primary px-5 text-sm font-semibold text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
                >
                    <template v-if="isLoading">
                        <div class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent"></div>
                        Exporting...
                    </template>
                    <template v-else>
                        Export Table
                    </template>
                </button>
            </div>
        </div>
    </div>
</template>

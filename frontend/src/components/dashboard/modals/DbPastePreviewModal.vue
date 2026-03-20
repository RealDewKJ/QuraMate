<script lang="ts" setup>
import { computed } from 'vue';

interface PastePreviewRow {
    id: number;
    sourceRowNumber: number;
    values: Record<string, string>;
    include: boolean;
    isValid: boolean;
    error: string;
}

interface Props {
    isOpen: boolean;
    tableName: string;
    columns: string[];
    rows: PastePreviewRow[];
    isInserting: boolean;
    error: string;
    columnDefs: Record<string, any>;
    getInputType: (col: string) => string;
    getNumberStep: (col: string) => string | number;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'confirm'): void;
    (e: 'auto-fix'): void;
    (e: 'toggle-row', rowId: number): void;
    (e: 'update:value', payload: { rowId: number; col: string; value: any }): void;
}>();

const selectedCount = computed(() => props.rows.filter(row => row.include).length);
const validCount = computed(() => props.rows.filter(row => row.isValid).length);
const invalidCount = computed(() => props.rows.length - validCount.value);

const updateValue = (rowId: number, col: string, value: any) => {
    emit('update:value', { rowId, col, value });
};
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @mousedown.self="emit('close')">
        <div @keydown.stop @mousedown.stop
            class="bg-card w-[95vw] max-w-6xl rounded-lg shadow-lg border border-border p-5 space-y-4 animate-in fade-in zoom-in-95 duration-200 max-h-[88vh] flex flex-col">
            <div class="flex items-start justify-between gap-4">
                <div>
                    <h3 class="text-lg font-semibold text-foreground">Paste Preview</h3>
                    <p class="text-sm text-muted-foreground">
                        Review and edit rows before inserting into
                        <span class="font-medium text-foreground">{{ tableName }}</span>
                    </p>
                </div>
                <button @click="emit('close')" class="text-muted-foreground hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div class="grid grid-cols-1 md:grid-cols-3 gap-2 text-xs">
                <div class="rounded-md border border-border bg-muted/40 px-3 py-2">Rows: <span class="font-semibold">{{ rows.length }}</span></div>
                <div class="rounded-md border border-border bg-muted/40 px-3 py-2">Valid: <span class="font-semibold text-emerald-600">{{ validCount }}</span></div>
                <div class="rounded-md border border-border bg-muted/40 px-3 py-2">Selected: <span class="font-semibold">{{ selectedCount }}</span></div>
            </div>

            <div class="flex-1 overflow-auto border border-border rounded-md">
                <table class="w-full text-xs">
                    <thead class="sticky top-0 bg-muted z-10">
                        <tr class="border-b border-border">
                            <th class="px-2 py-2 text-left w-20">Use</th>
                            <th class="px-2 py-2 text-left w-16">Row</th>
                            <th class="px-2 py-2 text-left w-20">Status</th>
                            <th v-for="col in columns" :key="col" class="px-2 py-2 text-left min-w-[140px]">{{ col }}</th>
                            <th class="px-2 py-2 text-left min-w-[220px]">Reason</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="row in rows" :key="row.id" class="border-b border-border/50 align-top"
                            :class="row.include ? 'bg-background' : 'bg-muted/20 opacity-80'">
                            <td class="px-2 py-2">
                                <input type="checkbox" :checked="row.include" @change="emit('toggle-row', row.id)"
                                    class="w-4 h-4 rounded border-input bg-background text-primary focus:ring-ring" />
                            </td>
                            <td class="px-2 py-2 font-mono text-muted-foreground">{{ row.sourceRowNumber }}</td>
                            <td class="px-2 py-2">
                                <span class="inline-flex px-2 py-0.5 rounded border"
                                    :class="row.isValid ? 'border-emerald-500/30 bg-emerald-500/10 text-emerald-600' : 'border-destructive/30 bg-destructive/10 text-destructive'">
                                    {{ row.isValid ? 'Valid' : 'Invalid' }}
                                </span>
                            </td>
                            <td v-for="col in columns" :key="`${row.id}-${col}`" class="px-2 py-2">
                                <template v-if="getInputType(col) === 'checkbox'">
                                    <input type="checkbox" :checked="row.values[col] === '1'"
                                        @change="updateValue(row.id, col, ($event.target as HTMLInputElement).checked ? '1' : '0')"
                                        class="w-4 h-4 rounded border-input bg-background text-primary focus:ring-ring" />
                                </template>
                                <input v-else-if="getInputType(col) === 'number'" :value="row.values[col] || ''"
                                    @input="updateValue(row.id, col, ($event.target as HTMLInputElement).value)" type="number"
                                    :step="getNumberStep(col)"
                                    class="w-full h-8 px-2 rounded-md border border-input bg-background text-foreground font-mono" />
                                <input v-else-if="getInputType(col) === 'datetime-local'" :value="row.values[col] || ''"
                                    @input="updateValue(row.id, col, ($event.target as HTMLInputElement).value)" type="datetime-local" step="1"
                                    class="w-full h-8 px-2 rounded-md border border-input bg-background text-foreground font-mono" />
                                <input v-else :value="row.values[col] || ''"
                                    @input="updateValue(row.id, col, ($event.target as HTMLInputElement).value)" type="text"
                                    class="w-full h-8 px-2 rounded-md border border-input bg-background text-foreground font-mono" />
                            </td>
                            <td class="px-2 py-2 text-destructive break-words">{{ row.error }}</td>
                        </tr>
                    </tbody>
                </table>
            </div>

            <div v-if="invalidCount > 0" class="text-xs text-muted-foreground">
                You can fix invalid rows or uncheck them to skip.
            </div>

            <div v-if="error" class="text-sm text-destructive bg-destructive/10 p-2 rounded-md break-all">
                {{ error }}
            </div>

            <div class="flex justify-end gap-3">
                <button @click="emit('auto-fix')" :disabled="isInserting"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-muted/50 hover:bg-muted transition-colors">
                    Auto-fix Common Issues
                </button>
                <button @click="emit('close')"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                    Cancel
                </button>
                <button @click="emit('confirm')" :disabled="isInserting || selectedCount === 0"
                    class="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm disabled:opacity-50 inline-flex items-center gap-2">
                    <svg v-if="isInserting" class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none"
                        viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z"></path>
                    </svg>
                    Insert Selected
                </button>
            </div>
        </div>
    </div>
</template>

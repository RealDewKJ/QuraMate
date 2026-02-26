<script lang="ts" setup>
interface Props {
    isOpen: boolean;
    tableName: string;
    columns: string[];
    values: Record<string, any>;
    nullColumns: Record<string, boolean>;
    error: string;
    isInserting: boolean;
    columnDefs: Record<string, any>;
    getInputType: (col: string) => string;
    getNumberStep: (col: string) => string | number;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'confirm'): void;
    (e: 'toggle-null', col: string): void;
    (e: 'update:value', payload: { col: string; value: any }): void;
}>();

const getColDef = (col: string) => props.columnDefs?.[col] || null;

const updateValue = (col: string, value: any) => {
    emit('update:value', { col, value });
};
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm" @mousedown.self="emit('close')">
        <div @keydown.stop @mousedown.stop
            class="bg-card w-full max-w-lg rounded-lg shadow-lg border border-border p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200 max-h-[80vh] flex flex-col">
            <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold text-foreground flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-plus-circle text-primary">
                        <circle cx="12" cy="12" r="10" />
                        <path d="M8 12h8" />
                        <path d="M12 8v8" />
                    </svg>
                    Insert Row
                </h3>
                <button @click="emit('close')" class="text-muted-foreground hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <p class="text-sm text-muted-foreground">
                Insert a new row into <span class="font-medium text-foreground">{{ tableName }}</span>
            </p>

            <div class="flex-1 overflow-y-auto space-y-3 pr-1">
                <div v-for="col in columns" :key="col" class="flex flex-col gap-1">
                    <div class="flex items-center justify-between">
                        <label class="text-xs font-semibold text-muted-foreground uppercase flex items-center gap-1">
                            {{ col }}
                            <span v-if="getColDef(col)?.primaryKey"
                                class="text-[9px] px-1 py-0.5 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-500 rounded font-normal normal-case">PK</span>
                            <span v-if="getColDef(col)?.autoIncrement"
                                class="text-[9px] px-1 py-0.5 bg-blue-100 dark:bg-blue-900/30 text-blue-800 dark:text-blue-400 rounded font-normal normal-case">Auto</span>
                            <span v-if="getColDef(col)"
                                class="text-[9px] px-1 py-0.5 bg-muted text-muted-foreground rounded font-normal normal-case">{{ getColDef(col)?.type }}</span>
                        </label>
                        <button @click="emit('toggle-null', col)" class="text-[10px] px-1.5 py-0.5 rounded transition-colors"
                            :class="nullColumns[col] ? 'bg-muted text-foreground' : 'text-muted-foreground hover:bg-muted/50'">
                            NULL
                        </button>
                    </div>

                    <template v-if="!nullColumns[col]">
                        <div v-if="getInputType(col) === 'checkbox'" class="flex items-center gap-2 h-8">
                            <input type="checkbox" :checked="values[col] === '1'" @change="updateValue(col, ($event.target as HTMLInputElement).checked ? '1' : '0')"
                                class="insert-row-input w-4 h-4 rounded border-input bg-background text-primary focus:ring-ring">
                            <span class="text-sm text-muted-foreground">{{ values[col] === '1' ? 'True' : 'False' }}</span>
                        </div>
                        <input v-else-if="getInputType(col) === 'date'" :value="values[col]" @input="updateValue(col, ($event.target as HTMLInputElement).value)"
                            type="date"
                            class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono">
                        <input v-else-if="getInputType(col) === 'datetime-local'" :value="values[col]" @input="updateValue(col, ($event.target as HTMLInputElement).value)"
                            type="datetime-local" step="1"
                            class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono">
                        <input v-else-if="getInputType(col) === 'time'" :value="values[col]" @input="updateValue(col, ($event.target as HTMLInputElement).value)"
                            type="time" step="1"
                            class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono">
                        <input v-else-if="getInputType(col) === 'number'" :value="values[col]" @input="updateValue(col, ($event.target as HTMLInputElement).value)"
                            type="number" :step="getNumberStep(col)" :placeholder="`Enter ${col}`"
                            class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono">
                        <textarea v-else-if="getInputType(col) === 'textarea'" :value="values[col]" @input="updateValue(col, ($event.target as HTMLTextAreaElement).value)"
                            :placeholder="`Enter ${col}`" rows="3"
                            class="insert-row-input w-full px-3 py-2 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono resize-y"></textarea>
                        <input v-else :value="values[col]" @input="updateValue(col, ($event.target as HTMLInputElement).value)"
                            type="text" :placeholder="`Enter ${col}`"
                            class="insert-row-input w-full h-8 px-3 rounded-md border border-input bg-background text-sm text-foreground focus:outline-none focus:ring-1 focus:ring-ring font-mono">
                    </template>
                    <div v-else class="w-full h-8 px-3 rounded-md border border-input bg-muted/50 text-sm text-muted-foreground flex items-center italic font-mono">
                        NULL
                    </div>
                </div>
            </div>

            <div v-if="error" class="text-sm text-destructive bg-destructive/10 p-2 rounded-md break-all">
                {{ error }}
            </div>

            <div class="flex justify-end gap-3 pt-2">
                <button @click="emit('close')"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                    Cancel
                </button>
                <button @click="emit('confirm')" :disabled="isInserting"
                    class="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm disabled:opacity-50 inline-flex items-center gap-2">
                    <svg v-if="isInserting" class="animate-spin h-3 w-3" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                        <path class="opacity-75" fill="currentColor" d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z"></path>
                    </svg>
                    Insert
                </button>
            </div>
        </div>
    </div>
</template>

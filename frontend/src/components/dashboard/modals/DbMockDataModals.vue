<script lang="ts" setup>
interface Props {
    generatorOpen: boolean;
    generatorTableName: string;
    generatorRowCount: number;
    confirmOpen: boolean;
    confirmTableName: string;
    confirmRowCount: number;
    confirmLoading: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
    (e: 'close-generator'): void;
    (e: 'update:generator-row-count', value: number): void;
    (e: 'open-confirm'): void;
    (e: 'close-confirm'): void;
    (e: 'confirm-insert'): void;
}>();
</script>

<template>
    <div v-if="generatorOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        @mousedown.self="emit('close-generator')">
        <div class="bg-card w-full max-w-md rounded-lg shadow-lg border border-border p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold text-foreground flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-flask-conical text-primary">
                        <path d="M10 2v7.31" />
                        <path d="M14 9.3V2" />
                        <path d="M8.5 2h7" />
                        <path
                            d="M14 9.3a5 5 0 0 1 .6 2.4V19a3 3 0 0 1-3 3h-1.2a3 3 0 0 1-3-3v-7.3a5 5 0 0 1 .6-2.4L10 6h4Z" />
                    </svg>
                    Mock Data Generator
                </h3>
                <button @click="emit('close-generator')" class="text-muted-foreground hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <p class="text-sm text-muted-foreground">
                Generate mock rows for table <span class="font-semibold text-foreground">{{ generatorTableName }}</span>
            </p>

            <div class="space-y-2">
                <label class="text-sm font-medium">Rows to generate</label>
                <input :value="generatorRowCount" @input="emit('update:generator-row-count', Number(($event.target as HTMLInputElement).value))" type="number" min="1"
                    max="500"
                    class="w-full h-10 rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"
                    placeholder="e.g. 50" />
                <p class="text-xs text-muted-foreground">Allowed range: 1 - 500 rows</p>
            </div>

            <div class="flex justify-end gap-3 pt-2">
                <button @click="emit('close-generator')"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                    Cancel
                </button>
                <button @click="emit('open-confirm')"
                    class="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm">
                    Continue
                </button>
            </div>
        </div>
    </div>

    <div v-if="confirmOpen" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 p-4 text-left">
        <div class="bg-card w-full max-w-sm rounded-lg shadow-2xl border border-destructive/40 p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center gap-3 text-destructive">
                <svg xmlns="http://www.w3.org/2000/svg" width="22" height="22" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-alert-triangle">
                    <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                    <path d="M12 9v4" />
                    <path d="M12 17h.01" />
                </svg>
                <h3 class="text-lg font-bold">Confirm Mock Insert</h3>
            </div>

            <p class="text-sm text-muted-foreground leading-relaxed">
                Insert <span class="font-bold text-foreground">{{ confirmRowCount }}</span> mock rows
                into table <span class="font-bold text-foreground">"{{ confirmTableName }}"</span>?
            </p>

            <div class="bg-destructive/10 border border-destructive/20 rounded-md p-3 text-[11px] text-destructive/90">
                This action writes real data to the table.
            </div>

            <div class="flex justify-end gap-3 pt-2">
                <button @click="emit('close-confirm')" :disabled="confirmLoading"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent transition-colors disabled:opacity-50">
                    Cancel
                </button>
                <button @click="emit('confirm-insert')" :disabled="confirmLoading"
                    class="px-4 py-2 text-sm font-medium rounded-md bg-destructive text-destructive-foreground hover:bg-destructive/90 transition-colors shadow-sm disabled:opacity-50 inline-flex items-center gap-2">
                    <template v-if="confirmLoading">
                        <div class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
                        Inserting...
                    </template>
                    <template v-else>
                        Confirm Insert
                    </template>
                </button>
            </div>
        </div>
    </div>
</template>

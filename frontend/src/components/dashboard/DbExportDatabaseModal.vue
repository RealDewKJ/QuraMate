<script lang="ts" setup>
type ExportFormat = 'SQL' | 'CSV' | 'JSON' | 'Excel';

interface Props {
    isOpen: boolean;
    folderPath: string;
    format: ExportFormat;
    isLoading: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'confirm'): void;
    (e: 'update:format', value: ExportFormat): void;
}>();

const formats: ExportFormat[] = ['SQL', 'CSV', 'JSON', 'Excel'];
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 text-left">
        <div class="bg-card w-full max-w-sm rounded-lg shadow-lg border border-border p-6 space-y-5 animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold text-foreground tracking-tight">Backup Database</h3>
                <button @click="emit('close')" class="text-muted-foreground hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div class="space-y-4">
                <div class="space-y-2">
                    <label class="text-xs font-bold text-muted-foreground uppercase tracking-wider">Format</label>
                    <div class="grid grid-cols-2 gap-2">
                        <button v-for="fmt in formats" :key="fmt"
                            @click="emit('update:format', fmt)"
                            class="px-3 py-2 text-xs font-bold rounded-md border transition-all"
                            :class="format === fmt ? 'bg-primary border-primary text-primary-foreground shadow-sm' : 'border-input bg-background hover:bg-accent text-muted-foreground'">
                            {{ fmt }}
                        </button>
                    </div>
                </div>

                <div class="space-y-1.5">
                    <label class="text-xs font-bold text-muted-foreground uppercase tracking-wider">Destination</label>
                    <div class="p-2.5 bg-muted/60 rounded-md border border-border">
                        <p class="text-[11px] truncate font-mono text-muted-foreground italic" :title="folderPath">{{ folderPath }}</p>
                    </div>
                    <p class="text-[10px] text-muted-foreground leading-relaxed">
                        Each table will be saved as an individual file.
                    </p>
                </div>
            </div>

            <div class="flex justify-end gap-3 pt-2">
                <button @click="emit('close')"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent transition-colors">
                    Cancel
                </button>
                <button @click="emit('confirm')" :disabled="isLoading"
                    class="flex items-center gap-2 px-6 py-2 text-sm font-bold rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm disabled:opacity-50 h-10">
                    <template v-if="isLoading">
                        <div class="h-4 w-4 animate-spin rounded-full border-2 border-primary-foreground border-t-transparent"></div>
                        Processing...
                    </template>
                    <template v-else>
                        Export Now
                    </template>
                </button>
            </div>
        </div>
    </div>
</template>

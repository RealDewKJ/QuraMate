<script lang="ts" setup>
interface Props {
    isOpen: boolean;
    tableName: string;
    column: string;
    originalValue: any;
    newValueDisplay: string;
}

defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'confirm'): void;
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
        <div class="bg-card w-full max-w-md rounded-lg shadow-lg border border-border p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold text-foreground">Confirm Update</h3>
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
                Are you sure you want to update data in table <span class="font-medium text-foreground">{{ tableName }}</span>?
            </p>

            <div class="bg-muted/50 p-3 rounded-md space-y-2 text-sm">
                <div class="flex flex-col gap-1">
                    <span class="text-xs font-semibold text-muted-foreground uppercase">Column</span>
                    <span class="font-mono text-foreground">{{ column }}</span>
                </div>
                <div class="grid grid-cols-2 gap-4">
                    <div class="flex flex-col gap-1">
                        <span class="text-xs font-semibold text-muted-foreground uppercase">Old Value</span>
                        <div
                            class="font-mono text-destructive text-xs max-h-48 overflow-y-auto whitespace-pre-wrap break-words border border-destructive/20 bg-destructive/5 p-2 rounded">
                            {{ originalValue === null ? 'NULL' : originalValue }}
                        </div>
                    </div>
                    <div class="flex flex-col gap-1">
                        <span class="text-xs font-semibold text-muted-foreground uppercase">New Value</span>
                        <div
                            class="font-mono text-green-600 dark:text-green-500 text-xs max-h-48 overflow-y-auto whitespace-pre-wrap break-words border border-green-500/20 bg-green-500/5 p-2 rounded">
                            {{ newValueDisplay }}
                        </div>
                    </div>
                </div>
            </div>

            <div class="flex justify-end gap-3 pt-2">
                <button @click="emit('close')"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                    Cancel
                </button>
                <button @click="emit('confirm')"
                    class="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm">
                    Confirm Update
                </button>
            </div>
        </div>
    </div>
</template>

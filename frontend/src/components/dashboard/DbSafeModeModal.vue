<script lang="ts" setup>
interface Props {
    isOpen: boolean;
    queryToRun: string;
}

defineProps<Props>();

const emit = defineEmits<{
    (e: 'confirm'): void;
    (e: 'cancel'): void;
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
        <div class="bg-card w-full max-w-md rounded-lg shadow-lg border border-border p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold text-destructive flex items-center gap-2">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-alert-triangle">
                        <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                        <path d="M12 9v4" />
                        <path d="M12 17h.01" />
                    </svg>
                    Safe Mode Warning
                </h3>
                <button @click="emit('cancel')" class="text-muted-foreground hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <p class="text-sm text-foreground">
                You are about to execute an <strong>UPDATE</strong> or <strong>DELETE</strong> query without a WHERE clause.
            </p>
            <div class="bg-destructive/10 border border-destructive/20 rounded p-3 text-xs font-mono text-destructive max-h-32 overflow-y-auto break-all whitespace-pre-wrap">
                {{ queryToRun }}
            </div>
            <p class="text-sm flex font-medium text-destructive">
                This will affect all rows potentially causing data loss. Are you sure you want to proceed?
            </p>

            <div class="flex justify-end gap-3 pt-4">
                <button @click="emit('cancel')"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                    Cancel
                </button>
                <button @click="emit('confirm')"
                    class="px-4 py-2 text-sm font-medium rounded-md bg-destructive text-destructive-foreground hover:bg-destructive/90 transition-colors shadow-sm">
                    Run Query
                </button>
            </div>
        </div>
    </div>
</template>

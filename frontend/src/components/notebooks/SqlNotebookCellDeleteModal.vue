<script lang="ts" setup>
defineProps<{
    isOpen: boolean;
    cellTitle: string;
    cellTypeLabel: string;
}>();

const emit = defineEmits<{
    (e: 'confirm'): void;
    (e: 'close'): void;
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[70] flex items-center justify-center bg-black/55 p-4">
        <div class="w-full max-w-md rounded-lg border border-destructive/30 bg-card p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-start gap-3">
                <div class="mt-0.5 rounded-full bg-destructive/10 p-2 text-destructive">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                        <path d="M12 9v4" />
                        <path d="M12 17h.01" />
                    </svg>
                </div>
                <div class="min-w-0">
                    <h3 class="text-lg font-semibold text-foreground">Delete Cell</h3>
                    <p class="mt-1 text-sm text-muted-foreground">
                        Delete this {{ cellTypeLabel.toLowerCase() }} cell
                        <span class="font-semibold text-foreground">"{{ cellTitle }}"</span>?
                    </p>
                </div>
            </div>

            <div class="mt-4 rounded-md border border-destructive/20 bg-destructive/10 p-3">
                <p class="text-xs font-semibold uppercase tracking-[0.12em] text-destructive">Warning</p>
                <p class="mt-1 text-sm text-destructive/90">
                    This removes the cell content from the notebook canvas. Saved results linked to this cell will also be cleared.
                </p>
            </div>

            <div class="mt-6 flex justify-end gap-3">
                <button
                    class="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('close')"
                >
                    Cancel
                </button>
                <button
                    class="inline-flex h-10 items-center justify-center rounded-md bg-destructive px-4 text-sm font-medium text-destructive-foreground transition-colors hover:bg-destructive/90"
                    @click="emit('confirm')"
                >
                    Delete Cell
                </button>
            </div>
        </div>
    </div>
</template>

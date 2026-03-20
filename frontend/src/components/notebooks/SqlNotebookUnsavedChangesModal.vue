<script lang="ts" setup>
defineProps<{
    isOpen: boolean;
    currentNotebookTitle: string;
    nextNotebookTitle: string;
    isSaving?: boolean;
}>();

const emit = defineEmits<{
    (e: 'save'): void;
    (e: 'discard'): void;
    (e: 'close'): void;
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[70] flex items-center justify-center bg-black/55 p-4">
        <div class="w-full max-w-md rounded-lg border border-amber-500/30 bg-card p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-start gap-3">
                <div class="mt-0.5 rounded-full bg-amber-500/10 p-2 text-amber-600">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                        <path d="M12 9v4" />
                        <path d="M12 17h.01" />
                    </svg>
                </div>
                <div class="min-w-0">
                    <h3 class="text-lg font-semibold text-foreground">Unsaved Changes</h3>
                    <p class="mt-1 text-sm text-muted-foreground">
                        <span class="font-semibold text-foreground">"{{ currentNotebookTitle }}"</span>
                        has unsaved changes. Save before switching to
                        <span class="font-semibold text-foreground">"{{ nextNotebookTitle }}"</span>?
                    </p>
                </div>
            </div>

            <div class="mt-4 rounded-md border border-amber-500/20 bg-amber-500/10 p-3">
                <p class="text-xs font-semibold uppercase tracking-[0.12em] text-amber-700">Warning</p>
                <p class="mt-1 text-sm text-amber-800/90">
                    Choose save to keep your latest notebook edits, discard to switch without saving, or stay here to keep editing.
                </p>
            </div>

            <div class="mt-6 flex flex-wrap justify-end gap-3">
                <button
                    class="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 text-sm font-medium transition-colors hover:bg-accent"
                    :disabled="isSaving"
                    @click="emit('close')"
                >
                    Stay Here
                </button>
                <button
                    class="inline-flex h-10 items-center justify-center rounded-md border border-amber-500/30 bg-amber-500/10 px-4 text-sm font-medium text-amber-700 transition-colors hover:bg-amber-500/15 disabled:cursor-not-allowed disabled:opacity-60"
                    :disabled="isSaving"
                    @click="emit('discard')"
                >
                    Discard Changes
                </button>
                <button
                    class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-60"
                    :disabled="isSaving"
                    @click="emit('save')"
                >
                    {{ isSaving ? 'Saving...' : 'Save & Switch' }}
                </button>
            </div>
        </div>
    </div>
</template>

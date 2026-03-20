<script lang="ts" setup>
defineProps<{
    isOpen: boolean;
    environmentLabel: string;
    notebookTitle: string;
    cellTitle: string;
    queryPreview: string;
}>();

const emit = defineEmits<{
    confirm: [];
    close: [];
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[70] flex items-center justify-center bg-black/55 p-4">
        <div class="w-full max-w-2xl rounded-lg border border-amber-500/30 bg-card p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-start gap-3">
                <div class="mt-0.5 rounded-full bg-amber-500/10 p-2 text-amber-600">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                        <path d="M12 9v4" />
                        <path d="M12 17h.01" />
                    </svg>
                </div>
                <div>
                    <h3 class="text-lg font-semibold text-foreground">Execution Guardrail</h3>
                    <p class="mt-1 text-sm text-muted-foreground">
                        This SQL looks destructive or schema-changing and the notebook is marked as
                        <span class="font-semibold text-foreground">{{ environmentLabel }}</span>.
                    </p>
                </div>
            </div>

            <div class="mt-4 grid gap-3 sm:grid-cols-2">
                <div class="rounded-lg border border-border bg-background/70 p-3">
                    <div class="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Notebook</div>
                    <div class="mt-1 text-sm font-medium">{{ notebookTitle }}</div>
                </div>
                <div class="rounded-lg border border-border bg-background/70 p-3">
                    <div class="text-[11px] uppercase tracking-[0.12em] text-muted-foreground">Cell</div>
                    <div class="mt-1 text-sm font-medium">{{ cellTitle }}</div>
                </div>
            </div>

            <div class="mt-4 rounded-lg border border-amber-500/20 bg-amber-500/10 p-3">
                <div class="text-[11px] font-semibold uppercase tracking-[0.12em] text-amber-700">Query Preview</div>
                <pre class="mt-2 overflow-auto whitespace-pre-wrap text-xs leading-6 text-amber-900"><code>{{ queryPreview }}</code></pre>
            </div>

            <div class="mt-6 flex justify-end gap-3">
                <button
                    class="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('close')"
                >
                    Cancel
                </button>
                <button
                    class="inline-flex h-10 items-center justify-center rounded-md bg-amber-600 px-4 text-sm font-medium text-white transition-colors hover:bg-amber-700"
                    @click="emit('confirm')"
                >
                    Run Anyway
                </button>
            </div>
        </div>
    </div>
</template>

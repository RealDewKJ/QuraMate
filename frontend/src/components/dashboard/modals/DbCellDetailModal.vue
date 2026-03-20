<script lang="ts" setup>
type CellDetailViewMode = 'text' | 'hex' | 'base64';

const props = defineProps<{
    isOpen: boolean;
    column: string;
    value: string;
    formatLabel: string;
    availableViews?: CellDetailViewMode[];
    selectedView?: CellDetailViewMode;
}>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'copy'): void;
    (e: 'update:selectedView', value: CellDetailViewMode): void;
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[95] flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60" @click="emit('close')"></div>

        <div class="relative z-10 flex max-h-[90vh] w-full max-w-4xl flex-col overflow-hidden rounded-xl border border-border bg-card shadow-2xl">
            <div class="flex items-center justify-between gap-3 border-b border-border bg-muted/30 px-5 py-3">
                <div class="min-w-0">
                    <h3 class="truncate text-base font-semibold">Cell Detail</h3>
                    <p class="truncate text-xs text-muted-foreground">
                        {{ column || 'Selected value' }}<span v-if="formatLabel"> · {{ formatLabel }}</span>
                    </p>
                </div>
                <button
                    class="rounded-md p-2 text-muted-foreground transition-colors hover:bg-muted hover:text-foreground"
                    title="Close"
                    @click="emit('close')"
                >
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div class="overflow-auto bg-background/70 p-5">
                <div v-if="(props.availableViews?.length || 0) > 1" class="mb-4 flex items-center gap-2">
                    <button
                        v-for="view in props.availableViews"
                        :key="view"
                        class="inline-flex items-center rounded-md border px-2.5 py-1.5 text-xs font-medium uppercase tracking-[0.12em] transition-colors"
                        :class="props.selectedView === view
                            ? 'border-primary bg-primary/10 text-primary'
                            : 'border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground'"
                        @click="emit('update:selectedView', view)"
                    >
                        {{ view }}
                    </button>
                </div>
                <pre class="overflow-x-auto whitespace-pre-wrap break-words rounded-lg border border-border bg-muted/20 p-4 text-xs text-foreground">{{ value }}</pre>
            </div>

            <div class="flex items-center justify-end gap-2 border-t border-border bg-muted/20 px-4 py-3">
                <button
                    class="inline-flex items-center rounded-md border border-input bg-background px-3 py-2 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('copy')"
                >
                    Copy Value
                </button>
                <button
                    class="inline-flex items-center rounded-md border border-input bg-background px-3 py-2 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('close')"
                >
                    Close
                </button>
            </div>
        </div>
    </div>
</template>

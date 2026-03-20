<script lang="ts" setup>
const props = defineProps<{
    isOpen: boolean;
    code: string;
    isImporting?: boolean;
    errorMessage?: string;
}>();

const emit = defineEmits<{
    close: [];
    import: [];
    'update:code': [value: string];
}>();
</script>

<template>
    <div v-if="props.isOpen" class="fixed inset-0 z-[80] flex items-center justify-center bg-black/55 p-4">
        <div class="w-full max-w-md rounded-lg border border-border bg-card p-6 shadow-2xl">
            <div class="flex items-start justify-between gap-4">
                <div>
                    <h3 class="text-lg font-semibold text-foreground">Join SessionShare</h3>
                    <p class="mt-1 text-sm text-muted-foreground">Paste a SessionShare code from your teammate to pull the notebook into this connection.</p>
                </div>
                <button
                    type="button"
                    class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('close')"
                >
                    Close
                </button>
            </div>

            <label class="mt-5 block">
                <span class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">SessionShare Code</span>
                <input
                    :value="props.code"
                    type="text"
                    placeholder="QN-AB12CD34"
                    class="mt-2 h-11 w-full rounded-md border border-input bg-background px-3 font-mono text-sm uppercase outline-none transition-colors focus:border-primary"
                    @input="emit('update:code', (($event.target as HTMLInputElement).value || '').toUpperCase())"
                />
            </label>

            <div v-if="props.errorMessage" class="mt-3 rounded-lg border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
                {{ props.errorMessage }}
            </div>

            <div class="mt-5 flex justify-end gap-3">
                <button
                    type="button"
                    class="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-4 text-sm font-medium transition-colors hover:bg-accent"
                    @click="emit('close')"
                >
                    Cancel
                </button>
                <button
                    type="button"
                    class="inline-flex h-10 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                    :disabled="props.isImporting || !props.code.trim()"
                    @click="emit('import')"
                >
                    {{ props.isImporting ? 'Joining...' : 'Join SessionShare' }}
                </button>
            </div>
        </div>
    </div>
</template>

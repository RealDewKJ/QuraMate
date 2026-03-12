<script lang="ts" setup>
interface Props {
    isOpen: boolean;
    imageUrl: string;
    fileName: string;
    tableName: string;
    timestampLabel: string;
    renderedRows: number;
    totalRows: number;
}

defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'copy'): void;
    (e: 'save'): void;
    (e: 'share'): void;
    (e: 'open-new-tab'): void;
    (e: 'copy-file-name'): void;
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[80] flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/60" @click="emit('close')"></div>

        <div class="relative z-10 w-full max-w-6xl max-h-[92vh] rounded-xl border border-border bg-card shadow-2xl flex flex-col overflow-hidden">
            <div class="flex items-center justify-between gap-4 border-b border-border px-5 py-3 bg-muted/30">
                <div class="min-w-0">
                    <h3 class="text-base font-semibold truncate">Screenshot Preview: {{ tableName || 'Query Result' }}</h3>
                    <p class="text-xs text-muted-foreground truncate">
                        {{ timestampLabel }} • {{ renderedRows }}/{{ totalRows }} rows
                    </p>
                </div>
                <button @click="emit('close')"
                    class="rounded-md p-2 text-muted-foreground hover:text-foreground hover:bg-muted transition-colors"
                    title="Close">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div class="flex-1 overflow-auto p-4 bg-background/50">
                <img v-if="imageUrl" :src="imageUrl" alt="Result screenshot preview"
                    class="mx-auto max-w-full rounded-lg border border-border bg-white shadow-sm" />
            </div>

            <div class="border-t border-border px-4 py-3 bg-muted/20">
                <div class="flex flex-wrap items-center gap-2">
                    <button @click="emit('copy')"
                        class="inline-flex items-center gap-1.5 rounded-md border border-input bg-background px-3 py-2 text-sm font-medium hover:bg-accent">
                        Copy Image
                    </button>
                    <button @click="emit('save')"
                        class="inline-flex items-center gap-1.5 rounded-md border border-input bg-background px-3 py-2 text-sm font-medium hover:bg-accent">
                        Save Image
                    </button>
                    <button @click="emit('share')"
                        class="inline-flex items-center gap-1.5 rounded-md border border-input bg-background px-3 py-2 text-sm font-medium hover:bg-accent">
                        Share Image
                    </button>
                    <button @click="emit('open-new-tab')"
                        class="inline-flex items-center gap-1.5 rounded-md border border-input bg-background px-3 py-2 text-sm font-medium hover:bg-accent">
                        Open Full Image
                    </button>
                    <button @click="emit('copy-file-name')"
                        class="inline-flex items-center gap-1.5 rounded-md border border-input bg-background px-3 py-2 text-sm font-medium hover:bg-accent">
                        Copy File Name
                    </button>
                </div>
                <p class="mt-2 text-[11px] text-muted-foreground truncate">{{ fileName }}</p>
            </div>
        </div>
    </div>
</template>

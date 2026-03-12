<script lang="ts" setup>
import { computed } from 'vue';

interface Props {
    isOpen: boolean;
    dbType: string;
    sourceName: string;
    targetName: string;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'confirm'): void;
    (e: 'update:source-name', value: string): void;
    (e: 'update:target-name', value: string): void;
}>();

const normalizedDbType = computed(() => (props.dbType || '').toLowerCase());
const sourceLabel = computed(() => {
    if (normalizedDbType.value.includes('mysql') || normalizedDbType.value.includes('mariadb') || normalizedDbType.value.includes('databend')) {
        return 'Source Database';
    }
    return 'Source Schema';
});
const targetLabel = computed(() => {
    if (normalizedDbType.value.includes('mysql') || normalizedDbType.value.includes('mariadb') || normalizedDbType.value.includes('databend')) {
        return 'Target Database';
    }
    return 'Target Schema';
});

const canConfirm = computed(() => props.sourceName.trim().length > 0 && props.targetName.trim().length > 0);
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 p-4 text-left" @mousedown.self="emit('close')">
        <div class="bg-card w-full max-w-lg rounded-lg shadow-lg border border-border p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-start justify-between gap-4">
                <div>
                    <h3 class="text-lg font-semibold text-foreground">Schema Compare Wizard</h3>
                    <p class="text-sm text-muted-foreground mt-1">
                        Configure source and target, then generate compare + migration preview SQL.
                    </p>
                </div>
                <button @click="emit('close')" class="text-muted-foreground hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div class="grid grid-cols-1 gap-4">
                <label class="flex flex-col gap-1">
                    <span class="text-sm font-medium text-foreground">{{ sourceLabel }}</span>
                    <input
                        :value="sourceName"
                        @input="emit('update:source-name', ($event.target as HTMLInputElement).value)"
                        placeholder="public"
                        class="h-10 rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring"
                    >
                </label>

                <label class="flex flex-col gap-1">
                    <span class="text-sm font-medium text-foreground">{{ targetLabel }}</span>
                    <input
                        :value="targetName"
                        @input="emit('update:target-name', ($event.target as HTMLInputElement).value)"
                        placeholder="staging"
                        class="h-10 rounded-md border border-input bg-background px-3 text-sm outline-none focus:ring-2 focus:ring-ring"
                    >
                </label>
            </div>

            <div class="rounded-md border border-border bg-muted/30 px-3 py-2 text-xs text-muted-foreground">
                The generated SQL focuses on column-level diffs and missing-column migration previews.
            </div>

            <div class="flex justify-end gap-3 pt-1">
                <button @click="emit('close')"
                    class="px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent hover:text-accent-foreground transition-colors">
                    Cancel
                </button>
                <button @click="emit('confirm')" :disabled="!canConfirm"
                    class="px-4 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm disabled:opacity-50 disabled:pointer-events-none">
                    Generate SQL
                </button>
            </div>
        </div>
    </div>
</template>

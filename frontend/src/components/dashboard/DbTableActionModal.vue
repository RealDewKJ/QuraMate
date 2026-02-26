<script lang="ts" setup>
interface Props {
    isOpen: boolean;
    action: 'TRUNCATE' | 'DROP';
    tableName: string;
    isLoading: boolean;
}

defineProps<Props>();

const emit = defineEmits<{
    (e: 'confirm'): void;
    (e: 'close'): void;
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-[60] flex items-center justify-center bg-black/60 backdrop-blur-sm p-4 text-left">
        <div class="bg-card w-full max-w-sm rounded-lg shadow-2xl border border-destructive/50 p-6 space-y-4 animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center gap-3 text-destructive">
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-alert-triangle">
                    <path d="m21.73 18-8-14a2 2 0 0 0-3.48 0l-8 14A2 2 0 0 0 4 21h16a2 2 0 0 0 1.73-3Z" />
                    <path d="M12 9v4" />
                    <path d="M12 17h.01" />
                </svg>
                <h3 class="text-lg font-bold uppercase tracking-tight">
                    {{ action === 'TRUNCATE' ? 'Truncate Table' : 'Drop Table' }}
                </h3>
            </div>

            <div class="space-y-3">
                <p class="text-sm text-muted-foreground leading-relaxed">
                    Are you sure you want to
                    <span class="font-bold text-foreground">
                        {{ action === 'TRUNCATE' ? 'truncate' : 'drop' }}
                    </span>
                    table <span class="font-bold text-foreground">"{{ tableName }}"</span>?
                </p>

                <div class="bg-destructive/10 border border-destructive/20 rounded-md p-3">
                    <p class="text-xs text-destructive font-bold uppercase mb-1">Critical Warning</p>
                    <p class="text-[11px] text-destructive/90 leading-normal">
                        <template v-if="action === 'TRUNCATE'">
                            This will remove all rows in this table, but keep the table structure.
                        </template>
                        <template v-else>
                            This will permanently remove both data and table structure.
                        </template>
                    </p>
                </div>
            </div>

            <div class="flex flex-col gap-2 pt-2">
                <button @click="emit('confirm')" :disabled="isLoading"
                    class="w-full h-10 flex items-center justify-center gap-2 px-4 py-2 text-sm font-bold rounded-md bg-destructive text-destructive-foreground hover:bg-destructive/90 transition-colors shadow-sm disabled:opacity-50">
                    <template v-if="isLoading">
                        <div class="h-4 w-4 animate-spin rounded-full border-2 border-white border-t-transparent"></div>
                        Processing...
                    </template>
                    <template v-else>
                        Confirm {{ action === 'TRUNCATE' ? 'Truncate' : 'Drop' }}
                    </template>
                </button>
                <button @click="emit('close')" :disabled="isLoading"
                    class="w-full h-10 px-4 py-2 text-sm font-medium rounded-md border border-input bg-background hover:bg-accent transition-colors disabled:opacity-50">
                    Cancel
                </button>
            </div>
        </div>
    </div>
</template>

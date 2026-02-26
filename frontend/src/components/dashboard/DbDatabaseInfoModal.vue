<script lang="ts" setup>
interface Props {
    isOpen: boolean;
    isLoading: boolean;
    info: any;
}

defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
}>();
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 text-left">
        <div class="bg-card w-full max-w-lg rounded-lg shadow-lg border border-border p-6 space-y-5 animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center justify-between">
                <h3 class="text-lg font-semibold text-foreground tracking-tight">Database Information</h3>
                <button @click="emit('close')" class="text-muted-foreground hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                        class="lucide lucide-x">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>

            <div v-if="isLoading" class="flex flex-col items-center justify-center py-10 gap-3">
                <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
                <p class="text-sm text-muted-foreground">Loading database info...</p>
            </div>

            <div v-else-if="info" class="space-y-4">
                <div class="grid grid-cols-2 gap-4 text-sm">
                    <div class="space-y-1">
                        <label class="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Database Name</label>
                        <p class="text-sm font-semibold truncate">{{ info.dbName }}</p>
                    </div>
                    <div class="space-y-1">
                        <label class="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Version</label>
                        <p class="text-sm font-semibold truncate" :title="info.version">{{ info.version }}</p>
                    </div>
                    <div class="space-y-1">
                        <label class="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Total Size</label>
                        <p class="text-sm font-semibold">{{ info.size }}</p>
                    </div>
                    <div class="space-y-1">
                        <label class="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Tables</label>
                        <p class="text-sm font-semibold">{{ info.tableCount }}</p>
                    </div>
                    <div class="space-y-1">
                        <label class="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Views</label>
                        <p class="text-sm font-semibold">{{ info.viewCount }}</p>
                    </div>
                    <div class="space-y-1">
                        <label class="text-[10px] font-bold text-muted-foreground uppercase tracking-widest">Routines</label>
                        <p class="text-sm font-semibold">{{ info.routineCount }}</p>
                    </div>
                </div>
            </div>

            <div class="flex justify-end pt-2">
                <button @click="emit('close')"
                    class="px-5 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm">
                    Close
                </button>
            </div>
        </div>
    </div>
</template>

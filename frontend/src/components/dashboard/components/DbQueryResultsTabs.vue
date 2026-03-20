<script lang="ts" setup>
defineProps<{
    activeTab: any;
}>();

const emit = defineEmits<{
    'change-tab': [tab: 'data' | 'messages' | 'analysis'];
}>();
</script>

<template>
    <div v-if="activeTab.queryExecuted || activeTab.error"
        class="shrink-0 border-b border-border bg-muted/20 px-3 py-2 dark:border-zinc-700/70 dark:bg-zinc-950/70">
        <div class="inline-flex items-center gap-1 rounded-full border border-border/80 bg-background/80 p-1 shadow-sm backdrop-blur-sm">
        <button @click="emit('change-tab', 'data')"
            class="relative inline-flex items-center rounded-full px-4 py-1.5 text-xs font-medium transition-all select-none"
            :class="activeTab.resultViewTab === 'data' ? 'bg-background text-foreground shadow-sm ring-1 ring-border/70 dark:bg-zinc-950 dark:text-zinc-100 dark:ring-zinc-700/70' : 'text-muted-foreground hover:bg-background/70 hover:text-foreground dark:text-zinc-400 dark:hover:bg-zinc-900/70 dark:hover:text-zinc-100'">
            <div class="flex items-center gap-1.5">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-table-2">
                    <path
                        d="M9 3H5a2 2 0 0 0-2 2v4m6-6h10a2 2 0 0 1 2 2v4M9 3v18m0 0h10a2 2 0 0 0 2-2V9M9 21H5a2 2 0 0 1-2-2V9m0 0h18" />
                </svg>
                Results
            </div>
        </button>
        <button @click="emit('change-tab', 'messages')"
            class="relative inline-flex items-center rounded-full px-4 py-1.5 text-xs font-medium transition-all select-none"
            :class="activeTab.resultViewTab === 'messages' ? 'bg-background text-foreground shadow-sm ring-1 ring-border/70 dark:bg-zinc-950 dark:text-zinc-100 dark:ring-zinc-700/70' : 'text-muted-foreground hover:bg-background/70 hover:text-foreground dark:text-zinc-400 dark:hover:bg-zinc-900/70 dark:hover:text-zinc-100'">
            <div class="flex items-center gap-1.5">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-message-square-text">
                    <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z" />
                    <path d="M13 8H7" />
                    <path d="M17 12H7" />
                </svg>
                Messages
            </div>
        </button>
        <button v-if="activeTab.explanation || activeTab.aiExplanation" @click="emit('change-tab', 'analysis')"
            class="relative inline-flex items-center rounded-full px-4 py-1.5 text-xs font-medium transition-all select-none"
            :class="activeTab.resultViewTab === 'analysis' ? 'bg-background text-foreground shadow-sm ring-1 ring-border/70 dark:bg-zinc-950 dark:text-zinc-100 dark:ring-zinc-700/70' : 'text-muted-foreground hover:bg-background/70 hover:text-foreground dark:text-zinc-400 dark:hover:bg-zinc-900/70 dark:hover:text-zinc-100'">
            <div class="flex items-center gap-1.5">
                <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-microscope">
                    <path d="M6 18h8" />
                    <path d="M3 22h18" />
                    <path d="M14 22a7 7 0 1 0 0-14h-1" />
                    <path d="M9 14h2" />
                    <path d="M9 12a2 2 0 0 1-2-2V6h6v4a2 2 0 0 1-2 2Z" />
                    <path d="M12 6V3a1 1 0 0 0-1-1H9a1 1 0 0 0-1 1v3" />
                </svg>
                Analysis
            </div>
        </button>
        </div>
    </div>
</template>

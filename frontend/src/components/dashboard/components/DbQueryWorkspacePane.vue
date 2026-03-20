<script lang="ts" setup>
import { computed, defineAsyncComponent, ref } from "vue";
import { useEventListener } from "@vueuse/core";

const SqlEditor = defineAsyncComponent(() => import("../../SqlEditor.vue"));

const props = defineProps<{
    activeTab: any;
    tables: string[];
    getColumns: (tableName: string) => Promise<string[]>;
    editorSettings: { fontFamily: string; fontSize: number };
    isReadOnly?: boolean;
}>();

const emit = defineEmits<{
    "beautify-query": [];
    "explain-with-ai": [];
    "explain-plan": [];
    "save-plan-baseline": [];
    "compare-plan-baseline": [];
    "save-to-notebook": [];
    "save-routine": [];
    "run-query": [];
    "stop-query": [];
    "start-resizing": [event: MouseEvent];
}>();

const sqlEditorRef = ref<any>(null);
const toolsMenuOpen = ref(false);
const toolsMenuRef = ref<HTMLElement | null>(null);
const runMenuOpen = ref(false);
const runMenuRef = ref<HTMLElement | null>(null);
const commandPaletteOpen = ref(false);
const commandPaletteQuery = ref("");
const commandPaletteRef = ref<HTMLElement | null>(null);

const getSelection = (): string => {
    return sqlEditorRef.value?.getSelection?.() || "";
};

const hasSelection = computed(() => getSelection().trim().length > 0);
const commandPaletteItems = computed(() => {
    const baseItems = [
        {
            id: "run-query",
            title: "Run Query",
            description: "Execute the current query tab.",
            disabled: !!props.activeTab?.isLoading,
            action: () => emit("run-query"),
        },
        {
            id: "run-selected",
            title: "Run Selected SQL",
            description: "Execute only the selected SQL text.",
            disabled: !!props.activeTab?.isLoading || !hasSelection.value,
            action: () => emit("run-query"),
        },
        {
            id: "beautify",
            title: "Beautify SQL",
            description: "Format the current query text.",
            disabled: !!props.activeTab?.isLoading || !props.activeTab?.query,
            action: () => emit("beautify-query"),
        },
        {
            id: "explain-plan",
            title: "Execution Plan",
            description: "Generate explain / showplan SQL for this query.",
            disabled: !!props.activeTab?.isLoading || !props.activeTab?.query,
            action: () => emit("explain-plan"),
        },
        {
            id: "compare-plan",
            title: "Compare Plan Baseline",
            description: "Compare current result to the saved baseline.",
            disabled:
                !!props.activeTab?.isLoading ||
                !props.activeTab?.resultSets?.length,
            action: () => emit("compare-plan-baseline"),
        },
        {
            id: "save-to-notebook",
            title: "Save To SQL Notebook",
            description:
                "Store the current query in an existing notebook or create a new one.",
            disabled: false,
            action: () => emit("save-to-notebook"),
        },
    ];

    const keyword = commandPaletteQuery.value.trim().toLowerCase();
    if (!keyword) {
        return baseItems;
    }
    return baseItems.filter(
        (item) =>
            item.title.toLowerCase().includes(keyword) ||
            item.description.toLowerCase().includes(keyword),
    );
});

const runToolAction = (action: () => void) => {
    action();
    toolsMenuOpen.value = false;
};

const runPaletteAction = (action: () => void) => {
    action();
    commandPaletteOpen.value = false;
    commandPaletteQuery.value = "";
};

const runQueryAction = (action: () => void) => {
    action();
    runMenuOpen.value = false;
};

const openCommandPalette = () => {
    commandPaletteOpen.value = true;
    commandPaletteQuery.value = "";
    toolsMenuOpen.value = false;
    runMenuOpen.value = false;
};

useEventListener(document, "click", (event: MouseEvent) => {
    const target = event.target as Node | null;
    if (
        toolsMenuOpen.value &&
        (!target || !toolsMenuRef.value?.contains(target))
    ) {
        toolsMenuOpen.value = false;
    }
    if (runMenuOpen.value && (!target || !runMenuRef.value?.contains(target))) {
        runMenuOpen.value = false;
    }
    if (
        commandPaletteOpen.value &&
        (!target || !commandPaletteRef.value?.contains(target))
    ) {
        commandPaletteOpen.value = false;
        commandPaletteQuery.value = "";
    }
});

useEventListener(document, "keydown", (event: KeyboardEvent) => {
    const withModifier = event.ctrlKey || event.metaKey;
    const key = event.key.toLowerCase();

    if (withModifier && key === "k") {
        event.preventDefault();
        if (commandPaletteOpen.value) {
            commandPaletteOpen.value = false;
            commandPaletteQuery.value = "";
            return;
        }
        openCommandPalette();
        return;
    }

    if (event.key === "Escape" && commandPaletteOpen.value) {
        event.preventDefault();
        commandPaletteOpen.value = false;
        commandPaletteQuery.value = "";
    }
});

defineExpose({
    getSelection,
    openCommandPalette,
});
</script>

<template>
    <div
        v-if="activeTab && !activeTab.isERView && !activeTab.isDesignView"
        class="flex flex-col border-b border-border bg-card p-4 gap-3 relative shrink-0 min-h-[0px]"
        :style="{ height: activeTab.editorHeight + 'px' }"
    >
        <div class="relative w-full flex-1 min-h-0">
            <SqlEditor
                ref="sqlEditorRef"
                v-model="activeTab.query"
                :tables="tables"
                :get-columns="getColumns"
                :font-family="editorSettings.fontFamily"
                :font-size="editorSettings.fontSize"
            />

            <div
                class="absolute bottom-1 right-3 z-10 flex items-center gap-2 pointer-events-none"
            >
                <div
                    class="pointer-events-auto rounded-full border border-border/80 bg-background/85 px-2.5 py-1 text-xs text-muted-foreground shadow-sm backdrop-blur-sm"
                >
                    {{ activeTab.query.length }} chars
                </div>
            </div>
        </div>

        <div class="flex items-center justify-between">
            <div class="flex items-center gap-4 text-xs text-muted-foreground">
                <div
                    v-if="activeTab.isLoading && !activeTab.executionTime"
                    class="flex items-center gap-2 text-primary"
                >
                    <svg
                        class="animate-spin h-3 w-3"
                        xmlns="http://www.w3.org/2000/svg"
                        fill="none"
                        viewBox="0 0 24 24"
                    >
                        <circle
                            class="opacity-25"
                            cx="12"
                            cy="12"
                            r="10"
                            stroke="currentColor"
                            stroke-width="4"
                        ></circle>
                        <path
                            class="opacity-75"
                            fill="currentColor"
                            d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z"
                        ></path>
                    </svg>
                    Executing...
                </div>
            </div>
        </div>

        <div
            class="sticky bottom-0 z-20 -mx-4 -mb-4 border-t border-border bg-card/95 px-4 py-3 backdrop-blur supports-[backdrop-filter]:bg-card/85"
        >
            <div
                class="flex flex-wrap items-center justify-between gap-2 overflow-visible"
            >
                <div class="flex min-w-0 flex-wrap items-center gap-2">
                    <div
                        v-if="isReadOnly"
                        class="px-2 py-1 bg-yellow-100 dark:bg-yellow-900/30 text-yellow-800 dark:text-yellow-500 text-xs rounded border border-yellow-200 dark:border-yellow-900/50 flex items-center gap-1 select-none cursor-help"
                        title="Database is in Read-Only mode. Modifications are disabled."
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="12"
                            height="12"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            class="lucide lucide-lock"
                        >
                            <rect
                                width="18"
                                height="11"
                                x="3"
                                y="11"
                                rx="2"
                                ry="2"
                            />
                            <path d="M7 11V7a5 5 0 0 1 10 0v4" />
                        </svg>
                        Read Only
                    </div>

                    <button
                        @click="emit('beautify-query')"
                        :disabled="activeTab.isLoading || !activeTab.query"
                        class="inline-flex h-9 items-center justify-center whitespace-nowrap rounded-full border border-input bg-background px-3 py-2 text-sm font-medium shadow-sm transition-colors ring-offset-background hover:bg-accent hover:text-accent-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 sm:px-4"
                        title="Format SQL (Shift + Alt + F)"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            class="lucide lucide-wrap-text sm:mr-2"
                        >
                            <line x1="3" x2="21" y1="6" y2="6" />
                            <path d="M3 12h15a3 3 0 1 1 0 6h-4" />
                            <polyline points="16 16 14 18 16 20" />
                        </svg>
                        <span class="hidden sm:inline">Beautify</span>
                    </button>
                </div>

                <div class="flex flex-wrap items-center justify-end gap-2">
                    <!-- <div ref="toolsMenuRef" class="relative">
                        <button
                            @click="toolsMenuOpen = !toolsMenuOpen"
                            class="inline-flex h-9 min-w-[44px] items-center justify-center whitespace-nowrap rounded-full border border-input bg-background px-3 py-2 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground sm:px-3.5"
                            title="Open tools and advanced actions"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                class="lucide lucide-wrench sm:mr-2"
                            >
                                <path
                                    d="M14.7 6.3a4 4 0 0 0 5 5L10 21l-7-7 9.7-9.7a4 4 0 0 0 2 2Z"
                                />
                                <path d="M16 4h4v4" />
                            </svg>
                            <span class="hidden sm:inline">Tools</span>
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="14"
                                height="14"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                class="lucide lucide-chevron-down ml-1.5"
                            >
                                <path d="m6 9 6 6 6-6" />
                            </svg>
                        </button>

                        <div
                            v-if="toolsMenuOpen"
                            class="absolute bottom-full left-0 z-30 mb-2 w-64 overflow-hidden rounded-2xl border border-border/80 bg-popover/95 p-2 shadow-xl ring-1 ring-black/5 backdrop-blur animate-in fade-in zoom-in-95 duration-100"
                        >
                            <button
                                @click="
                                    runToolAction(() => emit('explain-with-ai'))
                                "
                                :disabled="
                                    activeTab.isLoading ||
                                    activeTab.isAiExplaining ||
                                    !activeTab.query
                                "
                                class="flex w-full items-start gap-3 rounded-xl px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
                            >
                                <span class="font-medium text-foreground">{{
                                    activeTab.isAiExplaining
                                        ? "AI Explaining..."
                                        : "Explain with AI"
                                }}</span>
                            </button>
                            <button
                                @click="
                                    runToolAction(() => emit('explain-plan'))
                                "
                                :disabled="
                                    activeTab.isLoading || !activeTab.query
                                "
                                class="flex w-full items-start gap-3 rounded-xl px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
                            >
                                <span class="font-medium text-foreground"
                                    >Execution Plan</span
                                >
                            </button>
                            <button
                                @click="
                                    runToolAction(() =>
                                        emit('save-plan-baseline'),
                                    )
                                "
                                :disabled="
                                    activeTab.isLoading ||
                                    !activeTab.resultSets?.length
                                "
                                class="flex w-full items-start gap-3 rounded-xl px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
                            >
                                <span class="font-medium text-foreground"
                                    >Save Plan Baseline</span
                                >
                            </button>
                            <button
                                @click="
                                    runToolAction(() =>
                                        emit('compare-plan-baseline'),
                                    )
                                "
                                :disabled="
                                    activeTab.isLoading ||
                                    !activeTab.resultSets?.length
                                "
                                class="flex w-full items-start gap-3 rounded-xl px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
                            >
                                <span class="font-medium text-foreground"
                                    >Compare to Baseline</span
                                >
                            </button>
                            <button
                                @click="
                                    runToolAction(() => emit('save-to-notebook'))
                                "
                                class="flex w-full items-start gap-3 rounded-xl px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent"
                            >
                                <span class="font-medium text-foreground"
                                    >Save To SQL Notebook</span
                                >
                            </button>
                        </div>
                    </div> -->

                    <button
                        @click="emit('save-to-notebook')"
                        :disabled="!activeTab.query"
                        class="inline-flex h-9 items-center justify-center whitespace-nowrap rounded-full border border-input bg-background px-3 py-2 text-sm font-medium shadow-sm transition-colors hover:bg-accent hover:text-accent-foreground disabled:pointer-events-none disabled:opacity-50 sm:px-4"
                        title="Save current query to SQL Notebook"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            class="lucide lucide-book-plus sm:mr-2"
                        >
                            <path d="M12 7v14" />
                            <path d="M16 8h2" />
                            <path d="M16 12h2" />
                            <path d="M19 15v6" />
                            <path d="M22 18h-6" />
                            <path
                                d="M3 18a2 2 0 0 1 2-2h7a4 4 0 0 1 4 4V6a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2z"
                            />
                            <path
                                d="M21 12V6a2 2 0 0 0-2-2h-9a2 2 0 0 0-2 2v14a4 4 0 0 1 4-4h3"
                            />
                        </svg>
                        <span class="hidden sm:inline">Save To Notebook</span>
                    </button>

                    <button
                        v-if="activeTab.isRoutine"
                        @click="emit('save-routine')"
                        :disabled="activeTab.isLoading"
                        class="inline-flex h-9 items-center justify-center whitespace-nowrap rounded-full border border-primary/50 bg-primary/10 px-3 py-2 text-sm font-medium text-primary shadow-sm transition-colors ring-offset-background hover:bg-primary/20 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 sm:px-4"
                        title="Save or update routine"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            class="lucide lucide-save sm:mr-2"
                        >
                            <path
                                d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"
                            />
                            <polyline points="17 21 17 13 7 13 7 21" />
                            <polyline points="7 3 7 8 15 8" />
                        </svg>
                        <span class="hidden sm:inline">Save</span>
                    </button>

                    <button
                        v-if="activeTab.isLoading"
                        @click="emit('stop-query')"
                        class="inline-flex h-9 items-center justify-center whitespace-nowrap rounded-full bg-destructive px-3 py-2 text-sm font-medium text-destructive-foreground shadow-sm transition-colors ring-offset-background hover:bg-destructive/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 sm:px-4"
                        title="Stop query execution"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            class="lucide lucide-square sm:mr-2 fill-current"
                        >
                            <rect width="18" height="18" x="3" y="3" rx="2" />
                        </svg>
                        <span class="hidden sm:inline">Stop</span>
                    </button>

                    <div v-else ref="runMenuRef" class="relative inline-flex">
                        <button
                            @click="emit('run-query')"
                            class="inline-flex h-9 min-w-[48px] items-center justify-center whitespace-nowrap rounded-l-full bg-primary px-3 py-2 text-sm font-medium text-primary-foreground shadow-sm transition-colors ring-offset-background hover:bg-primary/90 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 sm:min-w-[84px] sm:px-3.5"
                            title="Run query (Ctrl+Enter)"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                                class="lucide lucide-play sm:mr-2"
                            >
                                <polygon points="5 3 19 12 5 21 5 3" />
                            </svg>
                            <span class="hidden sm:inline">Run</span>
                        </button>
                        <button
                            @click="runMenuOpen = !runMenuOpen"
                            class="inline-flex h-9 items-center justify-center rounded-r-full border-l border-white/20 bg-primary px-2.5 text-primary-foreground shadow-sm transition-colors hover:bg-primary/90"
                            title="More run actions"
                        >
                            <svg
                                xmlns="http://www.w3.org/2000/svg"
                                width="14"
                                height="14"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                stroke-width="2"
                                stroke-linecap="round"
                                stroke-linejoin="round"
                            >
                                <path d="m6 9 6 6 6-6" />
                            </svg>
                        </button>

                        <div
                            v-if="runMenuOpen"
                            class="absolute bottom-full right-0 z-30 mb-2 w-56 overflow-hidden rounded-2xl border border-border/80 bg-popover/95 p-2 shadow-xl ring-1 ring-black/5 backdrop-blur animate-in fade-in zoom-in-95 duration-100"
                        >
                            <button
                                @click="runQueryAction(() => emit('run-query'))"
                                class="flex w-full rounded-xl px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent"
                            >
                                Run Query
                            </button>
                            <button
                                @click="runQueryAction(() => emit('run-query'))"
                                :disabled="!hasSelection"
                                class="flex w-full rounded-xl px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
                            >
                                Run Selected SQL
                            </button>
                            <!-- <button
                                @click="
                                    runQueryAction(() => emit('explain-plan'))
                                "
                                class="flex w-full rounded-xl px-3 py-2.5 text-left text-sm transition-colors hover:bg-accent"
                            >
                                Explain Plan
                            </button> -->
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <div
        v-if="commandPaletteOpen"
        class="fixed inset-0 z-[80] flex items-start justify-center bg-black/40 px-4 pt-[12vh]"
        @click.self="commandPaletteOpen = false"
    >
        <div
            ref="commandPaletteRef"
            class="w-full max-w-2xl overflow-hidden rounded-xl border border-border bg-popover shadow-2xl animate-in fade-in zoom-in-95 duration-150"
        >
            <div class="border-b border-border px-4 py-3">
                <input
                    v-model="commandPaletteQuery"
                    autofocus
                    placeholder="Type a command"
                    class="w-full bg-transparent text-sm outline-none placeholder:text-muted-foreground"
                />
                <div class="mt-2 text-[11px] text-muted-foreground">
                    Press `Ctrl+K` to toggle, `Esc` to close.
                </div>
            </div>
            <div class="max-h-[420px] overflow-auto p-2">
                <button
                    v-for="item in commandPaletteItems"
                    :key="item.id"
                    @click="runPaletteAction(item.action)"
                    :disabled="item.disabled"
                    class="flex w-full items-start justify-between gap-4 rounded-lg px-3 py-3 text-left hover:bg-accent disabled:pointer-events-none disabled:opacity-50"
                >
                    <div>
                        <div class="text-sm font-medium text-foreground">
                            {{ item.title }}
                        </div>
                        <div class="mt-1 text-xs text-muted-foreground">
                            {{ item.description }}
                        </div>
                    </div>
                </button>
                <div
                    v-if="commandPaletteItems.length === 0"
                    class="px-3 py-6 text-sm text-muted-foreground"
                >
                    No matching command.
                </div>
            </div>
        </div>
    </div>

    <div
        v-if="activeTab && !activeTab.isERView && !activeTab.isDesignView"
        class="h-1.5 hover:bg-primary/30 cursor-row-resize flex items-center justify-center transition-colors group z-20 shrink-0"
        @mousedown="emit('start-resizing', $event)"
    >
        <div
            class="w-8 h-1 bg-border rounded-full group-hover:bg-primary/50"
        ></div>
    </div>
</template>

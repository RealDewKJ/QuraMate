<template>
    <div v-if="isOpen"
        class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 backdrop-blur-sm transition-all duration-100 animate-in fade-in">
        <div ref="modalRef"
            class="fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 sm:rounded-lg md:w-full animate-in fade-in zoom-in-95 slide-in-from-left-1/2 slide-in-from-top-48">
            <div class="flex flex-col space-y-1.5 text-center sm:text-left">
                <h2 class="text-lg font-semibold leading-none tracking-tight">
                    Saved Connections
                </h2>
                <p class="text-sm text-muted-foreground">
                    Select a connection to load its details.
                </p>
            </div>
            <div class="grid gap-4 py-4 max-h-[60vh] overflow-y-auto">
                <div v-if="connections.length === 0" class="text-center text-muted-foreground text-sm py-8">
                    No saved connections found.
                </div>
                <div v-else class="space-y-2">
                    <div v-for="conn in connections" :key="conn.id"
                        class="flex items-center justify-between p-3 rounded-lg border bg-card hover:bg-accent hover:text-accent-foreground transition-colors cursor-pointer group"
                        @click="emit('select', conn)">
                        <div class="flex items-center gap-3 overflow-hidden">
                            <div
                                class="h-8 w-8 rounded-full bg-primary/10 flex items-center justify-center text-primary">
                                <svg v-if="isPostgresLike(conn.type)" xmlns="http://www.w3.org/2000/svg" width="16"
                                    height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                    stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-database">
                                    <ellipse cx="12" cy="5" rx="9" ry="3" />
                                    <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                                    <path d="M3 12A9 3 0 0 0 21 12" />
                                </svg>
                                <svg v-else-if="isMysqlLike(conn.type)" xmlns="http://www.w3.org/2000/svg" width="16"
                                    height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                    stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-database">
                                    <ellipse cx="12" cy="5" rx="9" ry="3" />
                                    <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                                    <path d="M3 12A9 3 0 0 0 21 12" />
                                </svg>
                                <svg v-else-if="isFileBased(conn.type)" xmlns="http://www.w3.org/2000/svg" width="16"
                                    height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                    stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-file-code">
                                    <path d="M10 12.5 8 15l2 2.5" />
                                    <path d="m14 12.5 2 2.5-2 2.5" />
                                    <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
                                    <path d="M14 2v6h6" />
                                </svg>
                                <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                    viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                    stroke-linecap="round" stroke-linejoin="round" class="lucide lucide-database">
                                    <ellipse cx="12" cy="5" rx="9" ry="3" />
                                    <path d="M3 5V19A9 3 0 0 0 21 19V5" />
                                    <path d="M3 12A9 3 0 0 0 21 12" />
                                </svg>
                            </div>
                            <div class="flex flex-col truncate text-left">
                                <span class="text-sm font-medium truncate">{{
                                    getConnectionLabel(conn)
                                    }}</span>
                                <span class="text-xs text-muted-foreground truncate">{{ conn.host }}:{{ conn.port
                                    }}</span>
                            </div>
                        </div>
                        <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button @click.stop="emit('edit', conn)"
                                class="p-2 rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
                                title="Edit Connection">
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-pencil">
                                    <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
                                    <path d="m15 5 4 4" />
                                </svg>
                            </button>
                            <button @click.stop="emit('remove', conn)"
                                class="p-2 rounded-md hover:bg-destructive hover:text-destructive-foreground transition-colors"
                                title="Delete Connection">
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-trash-2">
                                    <path d="M3 6h18" />
                                    <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                                    <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                                    <line x1="10" x2="10" y1="11" y2="17" />
                                    <line x1="14" x2="14" y1="11" y2="17" />
                                </svg>
                            </button>
                        </div>
                    </div>
                </div>
            </div>
            <div class="flex flex-col-reverse sm:flex-row sm:justify-end sm:space-x-2">
                <button @click="emit('close')"
                    class="inline-flex items-center justify-center whitespace-nowrap rounded-md text-sm font-medium ring-offset-background transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50 border border-input bg-background hover:bg-accent hover:text-accent-foreground h-10 px-4 py-2">
                    Cancel
                </button>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { ref } from "vue";
import { onClickOutside, onKeyStroke } from "@vueuse/core";
import {
    type ConnectionConfig,
    getConnectionLabel,
} from "../../composables/useConnectionForm";

defineProps<{
    isOpen: boolean;
    connections: ConnectionConfig[];
}>();

const emit = defineEmits<{
    close: [];
    select: [conn: ConnectionConfig];
    edit: [conn: ConnectionConfig];
    remove: [conn: ConnectionConfig];
}>();

const modalRef = ref<HTMLElement | null>(null);
onClickOutside(modalRef, () => emit("close"));
onKeyStroke("Escape", () => emit("close"));

const isPostgresLike = (type: string) =>
    type === "postgres" ||
    type === "greenplum" ||
    type === "redshift" ||
    type === "cockroachdb";

const isMysqlLike = (type: string) =>
    type === "mysql" || type === "mariadb" || type === "databend";

const isFileBased = (type: string) =>
    type === "sqlite" || type === "duckdb" || type === "libsql";
</script>

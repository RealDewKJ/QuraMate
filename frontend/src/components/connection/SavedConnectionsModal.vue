<template>
    <div v-if="isOpen"
        class="fixed inset-0 z-50 flex items-center justify-center bg-background/80 transition-opacity duration-100 animate-in fade-in">
        <div ref="modalRef" @keydown.capture="handleModalKeydown"
            class="fixed left-[50%] top-[50%] z-50 grid w-full max-w-lg translate-x-[-50%] translate-y-[-50%] gap-4 border bg-background p-6 shadow-lg duration-200 sm:rounded-lg md:w-full animate-in fade-in zoom-in-95 slide-in-from-left-1/2 slide-in-from-top-48">
            <div class="flex flex-col space-y-1.5 text-center sm:text-left">
                <h2 class="text-lg font-semibold leading-none tracking-tight">
                    {{ t("common.savedConnections.title") }}
                </h2>
                <p class="text-sm text-muted-foreground">
                    {{ t("common.savedConnections.description") }}
                </p>
            </div>
            <div class="space-y-2">
                <input ref="searchInputRef" v-model="searchQuery" type="text" :placeholder="t('common.savedConnections.searchPlaceholder')"
                    :aria-label="t('common.savedConnections.searchAriaLabel')"
                    class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2" />
                <p class="text-[11px] text-muted-foreground">{{ t("common.savedConnections.searchHint") }}</p>
            </div>
            <div class="sr-only" aria-live="polite" aria-atomic="true">{{ activeSelectionAnnouncement }}</div>
            <div class="sr-only" aria-live="polite" aria-atomic="true">{{ statusAnnouncementText }}</div>
            <div class="grid gap-4 py-4 max-h-[60vh] overflow-y-auto">
                <div v-if="connections.length === 0" class="text-center text-muted-foreground text-sm py-8">
                    {{ t("common.savedConnections.empty") }}
                </div>
                <div v-else-if="filteredConnections.length === 0" class="text-center text-muted-foreground text-sm py-8">
                    {{ t("common.savedConnections.noMatches") }}
                </div>
                <div v-else role="listbox" tabindex="0"
                    :aria-label="t('common.savedConnections.listAriaLabel')"
                    :aria-activedescendant="activeOptionId"
                    class="space-y-2 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 rounded-md">
                    <div v-for="(conn, index) in filteredConnections" :id="getOptionId(conn, index)" :key="conn.id" role="option"
                        :aria-selected="index === activeIndex"
                        :class="[
                            'flex items-center justify-between p-3 rounded-lg border transition-colors cursor-pointer group',
                            index === activeIndex
                                ? 'bg-accent text-accent-foreground border-primary/40'
                                : 'bg-card hover:bg-accent hover:text-accent-foreground'
                        ]"
                        @mouseenter="setActiveIndex(index)" @click="setActiveIndex(index); emit('select', conn)">
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
                                <span class="text-xs text-muted-foreground truncate">{{ getConnectionSubtitle(conn) }}</span>
                            </div>
                        </div>
                        <div class="flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button @click.stop="emit('edit', conn)"
                                class="p-2 rounded-md hover:bg-accent hover:text-accent-foreground transition-colors"
                                :title="t('common.savedConnections.edit')" :aria-label="t('common.savedConnections.edit')">
                                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24"
                                    fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                                    stroke-linejoin="round" class="lucide lucide-pencil">
                                    <path d="M17 3a2.85 2.83 0 1 1 4 4L7.5 20.5 2 22l1.5-5.5Z" />
                                    <path d="m15 5 4 4" />
                                </svg>
                            </button>
                            <button @click.stop="requestRemoveConnection(conn)"
                                class="p-2 rounded-md hover:bg-destructive hover:text-destructive-foreground transition-colors"
                                :title="t('common.savedConnections.delete')" :aria-label="t('common.savedConnections.delete')">
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
                    {{ t("common.cancel") }}
                </button>
            </div>

            <div v-if="pendingDeleteConnection"
                class="absolute inset-0 z-10 flex items-center justify-center bg-background/70 rounded-lg p-4">
                <div ref="warningDialogRef" role="dialog" aria-modal="true" aria-label="Delete saved connection warning"
                    class="w-full max-w-sm rounded-lg border border-destructive/40 bg-card p-4 shadow-lg space-y-3">
                    <div>
                        <h3 class="text-sm font-semibold text-destructive">{{ t("common.savedConnections.deleteConfirmTitle") }}</h3>
                        <p class="mt-1 text-sm text-muted-foreground">
                            {{ t("common.savedConnections.deleteConfirmPrefix") }}
                            <span class="font-medium text-foreground">{{ getConnectionLabel(pendingDeleteConnection) }}</span>
                            {{ t("common.savedConnections.deleteConfirmSuffix") }}
                        </p>
                    </div>
                    <div class="flex justify-end gap-2">
                        <button @click="cancelPendingDelete"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium border border-input bg-background hover:bg-accent h-9 px-3">
                            {{ t("common.cancel") }}
                        </button>
                        <button @click="confirmPendingDelete"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium bg-destructive text-destructive-foreground hover:bg-destructive/90 h-9 px-3">
                            {{ t("common.delete") }}
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script lang="ts" setup>
import { computed, nextTick, ref, watch } from "vue";
import { onClickOutside } from "@vueuse/core";
import { useI18n } from "vue-i18n";
import {
    type ConnectionConfig,
    getConnectionLabel,
} from "../../composables/useConnectionForm";

const props = defineProps<{
    isOpen: boolean;
    connections: ConnectionConfig[];
    statusAnnouncement?: string;
}>();

const emit = defineEmits<{
    close: [];
    select: [conn: ConnectionConfig];
    edit: [conn: ConnectionConfig];
    remove: [conn: ConnectionConfig];
}>();
const { t } = useI18n({ useScope: "global" });

const modalRef = ref<HTMLElement | null>(null);
const searchInputRef = ref<HTMLInputElement | null>(null);
const warningDialogRef = ref<HTMLDivElement | null>(null);
const searchQuery = ref("");
const activeIndex = ref(0);
const statusAnnouncementText = ref("");
const pendingDeleteConnection = ref<ConnectionConfig | null>(null);

const filteredConnections = computed(() => {
    const query = searchQuery.value.trim().toLowerCase();
    if (!query) return props.connections;
    return props.connections.filter((conn) => {
        const label = getConnectionLabel(conn).toLowerCase();
        const subtitle = getConnectionSubtitle(conn).toLowerCase();
        const dbType = (conn.type || "").toLowerCase();
        return (
            label.includes(query) ||
            subtitle.includes(query) ||
            dbType.includes(query)
        );
    });
});

const setActiveIndex = (index: number) => {
    if (filteredConnections.value.length === 0) {
        activeIndex.value = 0;
        return;
    }
    activeIndex.value = Math.max(0, Math.min(index, filteredConnections.value.length - 1));
};

const announceStatus = async (message: string) => {
    statusAnnouncementText.value = "";
    await nextTick();
    statusAnnouncementText.value = message;
};

const requestRemoveConnection = (conn: ConnectionConfig) => {
    pendingDeleteConnection.value = conn;
};

const cancelPendingDelete = async () => {
    const label = pendingDeleteConnection.value
        ? getConnectionLabel(pendingDeleteConnection.value)
        : t("common.savedConnections.selectedConnectionFallback");
    pendingDeleteConnection.value = null;
    await announceStatus(t("common.savedConnections.deleteCancelled", { label }));
};

const confirmPendingDelete = () => {
    const selected = pendingDeleteConnection.value;
    if (!selected) return;
    pendingDeleteConnection.value = null;
    emit("remove", selected);
};

const getOptionId = (conn: ConnectionConfig, index: number): string =>
    `saved-conn-option-${conn.id || index}`;

const activeOptionId = computed(() => {
    const selected = filteredConnections.value[activeIndex.value];
    if (!selected) return "";
    return getOptionId(selected, activeIndex.value);
});

const activeSelectionAnnouncement = computed(() => {
    if (props.connections.length === 0) {
        return t("common.savedConnections.noneAvailable");
    }
    if (filteredConnections.value.length === 0) {
        return t("common.savedConnections.noMatches");
    }

    const selected = filteredConnections.value[activeIndex.value];
    if (!selected) {
        return t("common.savedConnections.countFound", { count: filteredConnections.value.length });
    }

    const label = getConnectionLabel(selected);
    const subtitle = getConnectionSubtitle(selected);
    return t("common.savedConnections.selectionAnnouncement", {
        index: activeIndex.value + 1,
        total: filteredConnections.value.length,
        label,
        subtitle,
    });
});

const isTypingTarget = (target: EventTarget | null): boolean => {
    const element = target as HTMLElement | null;
    if (!element) return false;
    const tag = element.tagName;
    return element.isContentEditable || tag === "INPUT" || tag === "TEXTAREA" || tag === "SELECT";
};

const trapFocus = (event: KeyboardEvent) => {
    const container = pendingDeleteConnection.value
        ? warningDialogRef.value
        : modalRef.value;
    if (!container) return;

    const focusable = Array.from(
        container.querySelectorAll<HTMLElement>(
            'button:not([disabled]), [href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])',
        ),
    ).filter((element) => element.offsetParent !== null);

    if (focusable.length === 0) return;

    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    const current = document.activeElement as HTMLElement | null;

    if (event.shiftKey) {
        if (current === first || !container.contains(current)) {
            event.preventDefault();
            last.focus();
        }
        return;
    }

    if (current === last || !container.contains(current)) {
        event.preventDefault();
        first.focus();
    }
};

const handleModalKeydown = (event: KeyboardEvent) => {
    if (!props.isOpen) return;
    const activeElement = document.activeElement as HTMLElement | null;

    if (pendingDeleteConnection.value) {
        if (event.key === "Escape") {
            event.preventDefault();
            void cancelPendingDelete();
            return;
        }
        if (
            event.key === "Enter" &&
            !isTypingTarget(event.target) &&
            (activeElement === warningDialogRef.value ||
                activeElement === null ||
                !activeElement.closest("button"))
        ) {
            event.preventDefault();
            confirmPendingDelete();
            return;
        }
    }

    if (event.key === "Escape") {
        event.preventDefault();
        emit("close");
        return;
    }

    if (event.key === "Tab") {
        trapFocus(event);
        return;
    }

    if (filteredConnections.value.length === 0) return;

    if (event.key === "ArrowDown") {
        event.preventDefault();
        setActiveIndex(activeIndex.value + 1);
        return;
    }
    if (event.key === "ArrowUp") {
        event.preventDefault();
        setActiveIndex(activeIndex.value - 1);
        return;
    }

    if (event.key === "Enter" && !isTypingTarget(event.target)) {
        event.preventDefault();
        const selected = filteredConnections.value[activeIndex.value];
        if (selected) emit("select", selected);
        return;
    }

    if ((event.key === "Delete" || event.key === "Backspace") && !isTypingTarget(event.target)) {
        event.preventDefault();
        const selected = filteredConnections.value[activeIndex.value];
        if (selected) requestRemoveConnection(selected);
    }
};

onClickOutside(modalRef, () => emit("close"));

watch(
    () => props.isOpen,
    async (isOpen) => {
        if (!isOpen) {
            pendingDeleteConnection.value = null;
            return;
        }
        searchQuery.value = "";
        activeIndex.value = 0;
        await nextTick();
        searchInputRef.value?.focus();
    },
);

watch(filteredConnections, (nextList) => {
    if (nextList.length === 0) {
        activeIndex.value = 0;
        return;
    }
    setActiveIndex(activeIndex.value);
});

watch(
    () => props.statusAnnouncement,
    async (message) => {
        if (!props.isOpen || !message) return;
        await announceStatus(message);
    },
);

const isPostgresLike = (type: string) =>
    type === "postgres" ||
    type === "supabase" ||
    type === "greenplum" ||
    type === "redshift" ||
    type === "cockroachdb";

const isMysqlLike = (type: string) =>
    type === "mysql" || type === "mariadb" || type === "databend";

const isFileBased = (type: string) =>
    type === "sqlite" || type === "duckdb" || type === "libsql";

const getConnectionSubtitle = (conn: ConnectionConfig): string => {
    if (isFileBased(conn.type)) {
        return conn.database || "File-based database";
    }
    if (conn.host && conn.port) {
        return `${conn.host}:${conn.port}`;
    }
    return conn.host || "Network database";
};
</script>

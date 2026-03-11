<script lang="ts" setup>
import { computed } from 'vue';

interface DatabaseInfoShape {
    dbName?: string;
    version?: string;
    size?: string;
    tableCount?: number;
    viewCount?: number;
    routineCount?: number;
    engine?: string;
    category?: string;
    summary?: Record<string, string>;
    stats?: Record<string, string>;
    runtimeInfo?: Record<string, string>;
    engineDetails?: Record<string, string>;
    capabilities?: Record<string, boolean>;
}

interface Props {
    isOpen: boolean;
    isLoading: boolean;
    info: DatabaseInfoShape | null;
}

const props = defineProps<Props>();

const emit = defineEmits<{
    (e: 'close'): void;
}>();

const titleEngine = computed(() => (props.info?.engine || 'Database').toUpperCase());

const summaryRows = computed(() => {
    const fallback: Record<string, string> = {
        activeDatabase: props.info?.dbName || '',
        serverVersion: props.info?.version || '',
    };
    return buildRows(props.info?.summary || fallback);
});

const statsRows = computed(() => {
    const source = props.info?.stats || {
        size: props.info?.size || '',
        tableCount: String(props.info?.tableCount ?? ''),
        viewCount: String(props.info?.viewCount ?? ''),
        routineCount: String(props.info?.routineCount ?? ''),
    };
    return buildRows(source);
});

const runtimeRows = computed(() => buildRows(props.info?.runtimeInfo || {}));
const engineDetailRows = computed(() => buildRows(props.info?.engineDetails || {}));

const capabilityRows = computed(() => {
    const caps = props.info?.capabilities || {};
    return Object.entries(caps)
        .map(([key, value]) => ({
            key,
            label: formatLabel(key),
            value: value ? 'Yes' : 'No',
        }))
        .sort((a, b) => a.label.localeCompare(b.label));
});

function buildRows(data: Record<string, string>) {
    return Object.entries(data)
        .filter(([, value]) => value !== undefined && value !== null && String(value).trim() !== '')
        .map(([key, value]) => ({
            key,
            label: formatLabel(key),
            value: String(value),
        }))
        .sort((a, b) => a.label.localeCompare(b.label));
}

function formatLabel(input: string): string {
    return input
        .replace(/([a-z0-9])([A-Z])/g, '$1 $2')
        .replace(/[_-]+/g, ' ')
        .replace(/\s+/g, ' ')
        .trim()
        .replace(/\b\w/g, (c) => c.toUpperCase());
}
</script>

<template>
    <div v-if="isOpen" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 backdrop-blur-sm p-4 text-left">
        <div class="bg-card w-full max-w-3xl max-h-[85vh] overflow-hidden rounded-lg shadow-lg border border-border animate-in fade-in zoom-in-95 duration-200">
            <div class="flex items-center justify-between px-6 py-4 border-b border-border">
                <div>
                    <h3 class="text-lg font-semibold text-foreground tracking-tight">Database Information</h3>
                    <p class="text-xs text-muted-foreground mt-0.5">{{ titleEngine }} <span v-if="info?.category">• {{
                        info.category }}</span></p>
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

            <div v-if="isLoading" class="flex flex-col items-center justify-center py-12 gap-3">
                <div class="h-8 w-8 animate-spin rounded-full border-2 border-primary border-t-transparent"></div>
                <p class="text-sm text-muted-foreground">Loading database info...</p>
            </div>

            <div v-else-if="info" class="p-6 space-y-5 overflow-auto max-h-[calc(85vh-145px)]">
                <section v-if="summaryRows.length" class="space-y-2">
                    <h4 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Summary</h4>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-3">
                        <div v-for="item in summaryRows" :key="'summary-' + item.key" class="rounded-md border border-border p-3">
                            <p class="text-[11px] uppercase tracking-wide text-muted-foreground">{{ item.label }}</p>
                            <p class="text-sm font-medium break-all">{{ item.value }}</p>
                        </div>
                    </div>
                </section>

                <section v-if="statsRows.length" class="space-y-2">
                    <h4 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Stats</h4>
                    <div class="grid grid-cols-2 md:grid-cols-4 gap-3">
                        <div v-for="item in statsRows" :key="'stats-' + item.key" class="rounded-md border border-border p-3">
                            <p class="text-[11px] uppercase tracking-wide text-muted-foreground">{{ item.label }}</p>
                            <p class="text-sm font-semibold">{{ item.value }}</p>
                        </div>
                    </div>
                </section>

                <section v-if="capabilityRows.length" class="space-y-2">
                    <h4 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Capabilities</h4>
                    <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
                        <div v-for="item in capabilityRows" :key="'cap-' + item.key"
                            class="rounded-md border border-border px-3 py-2 flex items-center justify-between">
                            <span class="text-sm text-foreground">{{ item.label }}</span>
                            <span class="text-xs font-semibold" :class="item.value === 'Yes' ? 'text-emerald-600' : 'text-muted-foreground'">{{
                                item.value }}</span>
                        </div>
                    </div>
                </section>

                <section v-if="runtimeRows.length" class="space-y-2">
                    <h4 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Runtime</h4>
                    <div class="space-y-2">
                        <div v-for="item in runtimeRows" :key="'runtime-' + item.key"
                            class="rounded-md border border-border px-3 py-2">
                            <p class="text-[11px] uppercase tracking-wide text-muted-foreground">{{ item.label }}</p>
                            <p class="text-sm font-medium break-all">{{ item.value }}</p>
                        </div>
                    </div>
                </section>

                <section v-if="engineDetailRows.length" class="space-y-2">
                    <h4 class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">Engine Details</h4>
                    <div class="space-y-2">
                        <div v-for="item in engineDetailRows" :key="'engine-' + item.key"
                            class="rounded-md border border-border px-3 py-2">
                            <p class="text-[11px] uppercase tracking-wide text-muted-foreground">{{ item.label }}</p>
                            <p class="text-sm font-medium break-all">{{ item.value }}</p>
                        </div>
                    </div>
                </section>
            </div>

            <div class="px-6 py-4 border-t border-border flex justify-end">
                <button @click="emit('close')"
                    class="px-5 py-2 text-sm font-medium rounded-md bg-primary text-primary-foreground hover:bg-primary/90 transition-colors shadow-sm">
                    Close
                </button>
            </div>
        </div>
    </div>
</template>

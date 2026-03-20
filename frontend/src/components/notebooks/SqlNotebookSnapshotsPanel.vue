<script lang="ts" setup>
import { computed, ref, watch } from 'vue';

import type { SqlNotebookResultSnapshot } from '../../types/sqlNotebook';
import SqlNotebookSnapshotDiffViewer from './SqlNotebookSnapshotDiffViewer.vue';

const props = defineProps<{
    snapshots: SqlNotebookResultSnapshot[];
}>();

const emit = defineEmits<{
    focusCell: [cellId: string];
}>();

const baseSnapshotId = ref('');
const compareSnapshotId = ref('');

const snapshotOptions = computed(() => props.snapshots);

const baseSnapshot = computed(() => {
    return snapshotOptions.value.find((snapshot) => snapshot.id === baseSnapshotId.value) ?? null;
});

const compareSnapshot = computed(() => {
    return snapshotOptions.value.find((snapshot) => snapshot.id === compareSnapshotId.value) ?? null;
});

watch(
    () => props.snapshots,
    (snapshots) => {
        baseSnapshotId.value = snapshots[0]?.id || '';
        compareSnapshotId.value = snapshots[1]?.id || snapshots[0]?.id || '';
    },
    { immediate: true },
);

const formatSnapshotDate = (value: string) => {
    try {
        return new Intl.DateTimeFormat(undefined, {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
        }).format(new Date(value));
    } catch (_error) {
        return value;
    }
};
</script>

<template>
    <section class="space-y-3">
        <div>
            <div class="text-xs uppercase tracking-[0.12em] text-muted-foreground">Snapshots</div>
            <div class="mt-1 text-xs text-muted-foreground">A snapshot is a saved copy of a SQL cell result at one point in time so you can compare before and after runs later.</div>
        </div>

        <div v-if="snapshots.length === 0" class="rounded-lg border border-dashed border-border p-3 text-xs text-muted-foreground">
            No snapshots saved yet. Run a SQL cell and save a snapshot from the result panel.
        </div>

        <div v-else class="space-y-3">
            <div class="grid gap-3 sm:grid-cols-2">
                <label class="space-y-2">
                    <span class="text-[11px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">Base</span>
                    <select
                        v-model="baseSnapshotId"
                        class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none transition-colors focus:border-primary"
                    >
                        <option v-for="snapshot in snapshotOptions" :key="snapshot.id" :value="snapshot.id">
                            {{ snapshot.cellTitle }} • {{ formatSnapshotDate(snapshot.capturedAt) }}
                        </option>
                    </select>
                </label>
                <label class="space-y-2">
                    <span class="text-[11px] font-semibold uppercase tracking-[0.12em] text-muted-foreground">Compare</span>
                    <select
                        v-model="compareSnapshotId"
                        class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none transition-colors focus:border-primary"
                    >
                        <option v-for="snapshot in snapshotOptions" :key="snapshot.id" :value="snapshot.id">
                            {{ snapshot.cellTitle }} • {{ formatSnapshotDate(snapshot.capturedAt) }}
                        </option>
                    </select>
                </label>
            </div>

            <SqlNotebookSnapshotDiffViewer
                :base-snapshot="baseSnapshot"
                :compare-snapshot="compareSnapshot"
            />

            <div
                v-for="snapshot in snapshots"
                :key="snapshot.id"
                class="rounded-lg border border-border bg-background/80 p-3"
            >
                <div class="flex items-start justify-between gap-3">
                    <div class="min-w-0">
                        <div class="truncate text-sm font-medium">{{ snapshot.cellTitle }}</div>
                        <div class="mt-1 text-[11px] text-muted-foreground">
                            {{ snapshot.resultSets.length }} result set{{ snapshot.resultSets.length === 1 ? '' : 's' }} - {{ snapshot.totalRows }} row{{ snapshot.totalRows === 1 ? '' : 's' }}
                        </div>
                        <div class="mt-1 text-[11px] text-muted-foreground">{{ formatSnapshotDate(snapshot.capturedAt) }}</div>
                    </div>
                    <button
                        class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium transition-colors hover:bg-accent"
                        @click="emit('focusCell', snapshot.cellId)"
                    >
                        Go To Cell
                    </button>
                </div>
            </div>
        </div>
    </section>
</template>


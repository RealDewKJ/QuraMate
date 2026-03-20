<script lang="ts" setup>
import { computed, ref, watch } from "vue";

import type {
    SqlNotebookCellRunResult,
    SqlNotebookResultSet,
} from "../../types/sqlNotebook";

const props = defineProps<{
    result: SqlNotebookCellRunResult | null;
    sourceLabel?: string;
}>();

const emit = defineEmits<{
    "focus-source": [];
    rerun: [];
    "save-snapshot": [];
}>();

const collapsedResultSets = ref<Record<number, boolean>>({});
const copyFeedback = ref<string | null>(null);

const resultSummary = computed(() => {
    if (!props.result) {
        return "Run this cell to render results here.";
    }

    if (props.result.status === "running") {
        return "Running query...";
    }

    if (props.result.status === "error") {
        return props.result.errorMessage || "Query failed.";
    }

    if (props.result.status === "cancelled") {
        return "Query cancelled.";
    }

    const resultSetCount = props.result.resultSets.length;
    const totalRows = props.result.totalRows ?? 0;
    return `${resultSetCount} result set${resultSetCount === 1 ? "" : "s"} - ${totalRows} row${totalRows === 1 ? "" : "s"}`;
});

const hasTabularResults = computed(() => {
    return (props.result?.resultSets || []).some(
        (resultSet) => resultSet.columns.length > 0,
    );
});

const nonTabularMessages = computed(() => {
    return (props.result?.resultSets || [])
        .map((resultSet) => resultSet.message?.trim() || "")
        .filter(Boolean);
});

const formatTimestamp = (value?: string): string => {
    if (!value) {
        return "-";
    }

    try {
        return new Intl.DateTimeFormat(undefined, {
            month: "short",
            day: "numeric",
            hour: "2-digit",
            minute: "2-digit",
            second: "2-digit",
        }).format(new Date(value));
    } catch (_error) {
        return value;
    }
};

const toggleResultSet = (index: number) => {
    collapsedResultSets.value[index] = !collapsedResultSets.value[index];
};

const serializeValue = (value: unknown): string => {
    if (value === null || value === undefined) {
        return "NULL";
    }

    if (typeof value === "object") {
        try {
            return JSON.stringify(value);
        } catch (_error) {
            return String(value);
        }
    }

    return String(value);
};

const toTsv = (resultSet: SqlNotebookResultSet): string => {
    if (resultSet.columns.length === 0) {
        return resultSet.message || "";
    }

    const header = resultSet.columns.join("\t");
    const rows = resultSet.rows.map((row) =>
        resultSet.columns
            .map((column) => serializeValue(row[column]))
            .join("\t"),
    );
    return [header, ...rows].join("\n");
};

const toCsvCell = (value: unknown): string => {
    const normalized = serializeValue(value);
    return `"${normalized.replace(/"/g, '""')}"`;
};

const toCsv = (resultSet: SqlNotebookResultSet): string => {
    if (resultSet.columns.length === 0) {
        return toCsvCell(resultSet.message || "");
    }

    const header = resultSet.columns
        .map((column) => toCsvCell(column))
        .join(",");
    const rows = resultSet.rows.map((row) =>
        resultSet.columns.map((column) => toCsvCell(row[column])).join(","),
    );
    return [header, ...rows].join("\n");
};

const copyResultSet = async (
    resultSet: SqlNotebookResultSet,
    index: number,
) => {
    await navigator.clipboard.writeText(toTsv(resultSet));
    copyFeedback.value = `Copied result ${index + 1}`;
};

const exportResultSet = (resultSet: SqlNotebookResultSet, index: number) => {
    const blob = new Blob([toCsv(resultSet)], {
        type: "text/csv;charset=utf-8",
    });
    const url = URL.createObjectURL(blob);
    const anchor = document.createElement("a");
    anchor.href = url;
    anchor.download = `notebook-result-${index + 1}.csv`;
    document.body.appendChild(anchor);
    anchor.click();
    document.body.removeChild(anchor);
    URL.revokeObjectURL(url);
};

watch(
    () => props.result?.completedAt,
    () => {
        collapsedResultSets.value = {};
        copyFeedback.value = null;
    },
);

watch(copyFeedback, (value, _previousValue, onCleanup) => {
    if (!value) {
        return;
    }

    const timeoutId = window.setTimeout(() => {
        copyFeedback.value = null;
    }, 1800);

    onCleanup(() => {
        window.clearTimeout(timeoutId);
    });
});
</script>

<template>
    <section
        class="min-w-0 w-full max-w-full overflow-hidden rounded-xl border border-border bg-card/80 shadow-sm"
    >
        <div
            class="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3"
        >
            <div>
                <div class="text-sm font-semibold">Results</div>
                <div class="text-xs text-muted-foreground">
                    {{ resultSummary }}
                </div>
                <div
                    v-if="props.sourceLabel"
                    class="mt-1 text-[11px] text-muted-foreground"
                >
                    Source: {{ props.sourceLabel }}
                </div>
            </div>
            <div class="flex items-center gap-3">
                <div
                    v-if="copyFeedback"
                    class="text-[11px] font-medium text-primary"
                >
                    {{ copyFeedback }}
                </div>
                <div
                    v-if="props.result?.completedAt"
                    class="text-[11px] text-muted-foreground"
                >
                    {{ formatTimestamp(props.result.completedAt) }}
                </div>
                <button
                    v-if="props.sourceLabel"
                    class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                    @click="emit('focus-source')"
                >
                    Jump To Cell
                </button>
                <button
                    class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                    :disabled="!props.result"
                    @click="emit('rerun')"
                >
                    Run Again
                </button>
                <button
                    v-if="props.result && props.result.status === 'success'"
                    class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                    @click="emit('save-snapshot')"
                >
                    Save Snapshot
                </button>
            </div>
        </div>

        <div v-if="!props.result" class="p-4 text-sm text-muted-foreground">
            Results stay in the notebook canvas for the latest cell run.
        </div>

        <div
            v-else-if="props.result.status === 'running'"
            class="p-4 text-sm text-muted-foreground"
        >
            Executing SQL cell...
        </div>

        <div v-else-if="props.result.status === 'error'" class="p-4">
            <div
                class="rounded-lg border border-destructive/30 bg-destructive/5 p-3 text-sm text-destructive"
            >
                {{ props.result.errorMessage || "Query failed." }}
            </div>
        </div>

        <div
            v-else-if="props.result.resultSets.length === 0"
            class="p-4 text-sm text-muted-foreground"
        >
            Query completed without a tabular result set.
        </div>

        <div v-else-if="!hasTabularResults" class="space-y-3 p-4">
            <div
                v-for="(message, messageIndex) in nonTabularMessages"
                :key="messageIndex"
                class="rounded-lg border border-border bg-muted/20 p-3 text-sm text-foreground"
            >
                <div
                    class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground"
                >
                    Message {{ messageIndex + 1 }}
                </div>
                <div class="mt-2">
                    {{ message }}
                </div>
            </div>

            <div
                v-if="nonTabularMessages.length === 0"
                class="rounded-lg border border-border bg-muted/20 p-3 text-sm text-muted-foreground"
            >
                Query completed successfully.
            </div>
        </div>

        <div v-else class="min-w-0 w-full max-w-full space-y-4 p-4">
            <section
                v-for="(resultSet, resultSetIndex) in props.result.resultSets"
                :key="resultSetIndex"
                class="min-w-0 w-full max-w-full overflow-hidden rounded-lg border border-border"
            >
                <div
                    class="flex flex-wrap items-center justify-between gap-3 border-b border-border bg-muted/30 px-3 py-2"
                >
                    <div class="flex items-center gap-3">
                        <button
                            class="inline-flex h-7 w-7 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                            @click="toggleResultSet(resultSetIndex)"
                        >
                            <svg
                                v-if="collapsedResultSets[resultSetIndex]"
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
                            <svg
                                v-else
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
                                <path d="m18 15-6-6-6 6" />
                            </svg>
                        </button>
                        <div>
                            <div class="text-sm font-medium">
                                Result {{ resultSetIndex + 1 }}
                            </div>
                            <div class="text-xs text-muted-foreground">
                                {{ resultSet.rows.length }} row{{
                                    resultSet.rows.length === 1 ? "" : "s"
                                }}
                            </div>
                        </div>
                    </div>

                    <div class="flex items-center gap-2">
                        <button
                            class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                            @click="
                                void copyResultSet(resultSet, resultSetIndex)
                            "
                        >
                            Copy
                        </button>
                        <button
                            class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                            @click="exportResultSet(resultSet, resultSetIndex)"
                        >
                            Export CSV
                        </button>
                    </div>
                </div>

                <div v-if="!collapsedResultSets[resultSetIndex]">
                    <div
                        v-if="resultSet.message"
                        class="border-b border-border bg-muted/20 px-3 py-2 text-xs text-muted-foreground"
                    >
                        {{ resultSet.message }}
                    </div>

                    <div
                        v-if="resultSet.columns.length === 0"
                        class="p-3 text-sm text-muted-foreground"
                    >
                        {{
                            resultSet.message ||
                            "Command completed successfully."
                        }}
                    </div>

                    <div v-else class="w-full max-w-full overflow-x-auto">
                        <div class="max-h-80 min-w-0 w-max min-w-full">
                            <table
                                class="w-max min-w-full table-auto text-left text-xs"
                            >
                                <thead class="sticky top-0 bg-background">
                                    <tr class="border-b border-border">
                                        <th
                                            v-for="column in resultSet.columns"
                                            :key="column"
                                            class="whitespace-nowrap px-3 py-2 font-semibold text-foreground"
                                        >
                                            {{ column }}
                                        </th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr
                                        v-for="(
                                            row, rowIndex
                                        ) in resultSet.rows"
                                        :key="rowIndex"
                                        class="border-b border-border/60 align-top"
                                    >
                                        <td
                                            v-for="column in resultSet.columns"
                                            :key="`${rowIndex}-${column}`"
                                            class="max-w-[320px] min-w-[140px] whitespace-pre-wrap break-words px-3 py-2 font-mono text-foreground"
                                        >
                                            {{ row[column] ?? "NULL" }}
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                    </div>
                </div>
            </section>
        </div>
    </section>
</template>

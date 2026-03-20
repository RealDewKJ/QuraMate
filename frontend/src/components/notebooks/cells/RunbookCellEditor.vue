<script lang="ts" setup>
import { computed } from 'vue';

import {
    parseRunbookContent,
    stringifyRunbookContent,
    type SqlNotebookExecutionState,
    type SqlNotebookRunbookApproval,
    type SqlNotebookRunbookApprovalStatus,
} from '../../../types/sqlNotebook';

const props = defineProps<{
    title: string;
    content: string;
    executionState: SqlNotebookExecutionState;
}>();

const emit = defineEmits<{
    'update:title': [value: string];
    'update:content': [value: string];
    'update:status': [value: SqlNotebookExecutionState];
    focus: [];
}>();

const runbook = computed(() => parseRunbookContent(props.content));

const updateRunbookField = (patch: Partial<ReturnType<typeof parseRunbookContent>>) => {
    emit('update:content', stringifyRunbookContent({
        ...runbook.value,
        ...patch,
    }));
};

const checklistText = computed(() => runbook.value.checklist.join('\n'));

const createApproval = (): SqlNotebookRunbookApproval => ({
    id: `approval_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 8)}`,
    name: '',
    role: '',
    status: 'pending',
    note: '',
});

const updateApproval = (approvalId: string, patch: Partial<SqlNotebookRunbookApproval>) => {
    updateRunbookField({
        approvals: runbook.value.approvals.map((approval) => {
            if (approval.id !== approvalId) {
                return approval;
            }

            return {
                ...approval,
                ...patch,
            };
        }),
    });
};

const addApproval = () => {
    updateRunbookField({
        approvals: [...runbook.value.approvals, createApproval()],
    });
};

const removeApproval = (approvalId: string) => {
    const approvals = runbook.value.approvals.filter((approval) => approval.id !== approvalId);
    updateRunbookField({
        approvals: approvals.length > 0 ? approvals : [createApproval()],
    });
};

const approvalStatusLabel: Record<SqlNotebookRunbookApprovalStatus, string> = {
    pending: 'Pending',
    approved: 'Approved',
    blocked: 'Blocked',
};
</script>

<template>
    <div class="space-y-4" @mousedown.capture="emit('focus')" @focusin="emit('focus')">
        <div class="flex flex-wrap items-center gap-3">
            <input
                :value="props.title"
                type="text"
                placeholder="Runbook step title"
                class="min-w-0 flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm font-medium outline-none transition-colors focus:border-primary"
                @input="emit('update:title', ($event.target as HTMLInputElement).value)"
            />
            <div class="flex flex-wrap gap-2">
                <button
                    type="button"
                    class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium transition-colors hover:bg-accent"
                    @click="emit('update:status', 'idle')"
                >
                    Pending
                </button>
                <button
                    type="button"
                    class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium transition-colors hover:bg-accent"
                    @click="emit('update:status', 'running')"
                >
                    Running
                </button>
                <button
                    type="button"
                    class="inline-flex h-9 items-center justify-center rounded-md border border-emerald-500/30 bg-emerald-500/10 px-3 text-xs font-medium text-emerald-700 transition-colors hover:bg-emerald-500/15"
                    @click="emit('update:status', 'verified')"
                >
                    Verified
                </button>
                <button
                    type="button"
                    class="inline-flex h-9 items-center justify-center rounded-md border border-amber-500/30 bg-amber-500/10 px-3 text-xs font-medium text-amber-700 transition-colors hover:bg-amber-500/15"
                    @click="emit('update:status', 'skipped')"
                >
                    Skipped
                </button>
            </div>
        </div>

        <div class="grid gap-4 xl:grid-cols-[minmax(0,1fr)_minmax(0,1fr)]">
            <div class="space-y-4 rounded-xl border border-border bg-background p-4 shadow-sm">
                <div>
                    <label class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Objective</label>
                    <textarea
                        :value="runbook.objective"
                        rows="3"
                        class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus:border-primary"
                        @input="updateRunbookField({ objective: ($event.target as HTMLTextAreaElement).value })"
                    />
                </div>

                <div>
                    <label class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Checklist</label>
                    <textarea
                        :value="checklistText"
                        rows="6"
                        class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 font-mono text-sm outline-none transition-colors focus:border-primary"
                        @input="updateRunbookField({ checklist: ($event.target as HTMLTextAreaElement).value.split('\n').map((item) => item.trim()).filter(Boolean) })"
                    />
                    <p class="mt-2 text-xs text-muted-foreground">One step per line. Keep the list explicit enough for handoff.</p>
                </div>
            </div>

            <div class="space-y-4 rounded-xl border border-border bg-card p-4 shadow-sm">
                <div>
                    <label class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Expected Result</label>
                    <textarea
                        :value="runbook.expectedResult"
                        rows="3"
                        class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus:border-primary"
                        @input="updateRunbookField({ expectedResult: ($event.target as HTMLTextAreaElement).value })"
                    />
                </div>

                <div>
                    <label class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Rollback Notes</label>
                    <textarea
                        :value="runbook.rollbackNotes"
                        rows="4"
                        class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus:border-primary"
                        @input="updateRunbookField({ rollbackNotes: ($event.target as HTMLTextAreaElement).value })"
                    />
                </div>

                <div>
                    <label class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Safety Notes</label>
                    <textarea
                        :value="runbook.safetyNotes"
                        rows="4"
                        class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus:border-primary"
                        @input="updateRunbookField({ safetyNotes: ($event.target as HTMLTextAreaElement).value })"
                    />
                </div>

                <div>
                    <div class="flex items-center justify-between gap-3">
                        <label class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Approvals</label>
                        <button
                            type="button"
                            class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium transition-colors hover:bg-accent"
                            @click="addApproval"
                        >
                            Add Approver
                        </button>
                    </div>
                    <div class="mt-2 space-y-3">
                        <div
                            v-for="approval in runbook.approvals"
                            :key="approval.id"
                            class="rounded-lg border border-border bg-background/80 p-3"
                        >
                            <div class="grid gap-3 sm:grid-cols-[minmax(0,1fr)_minmax(0,1fr)_auto]">
                                <input
                                    :value="approval.name"
                                    type="text"
                                    placeholder="Approver name"
                                    class="rounded-md border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus:border-primary"
                                    @input="updateApproval(approval.id, { name: ($event.target as HTMLInputElement).value })"
                                />
                                <input
                                    :value="approval.role"
                                    type="text"
                                    placeholder="Role or team"
                                    class="rounded-md border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus:border-primary"
                                    @input="updateApproval(approval.id, { role: ($event.target as HTMLInputElement).value })"
                                />
                                <button
                                    type="button"
                                    class="inline-flex h-10 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium text-destructive transition-colors hover:bg-destructive/10"
                                    @click="removeApproval(approval.id)"
                                >
                                    Remove
                                </button>
                            </div>

                            <div class="mt-3 flex flex-wrap gap-2">
                                <button
                                    v-for="status in ['pending', 'approved', 'blocked']"
                                    :key="status"
                                    type="button"
                                    class="inline-flex h-8 items-center justify-center rounded-full border px-3 text-xs font-medium transition-colors"
                                    :class="approval.status === status
                                        ? status === 'approved'
                                            ? 'border-emerald-500/40 bg-emerald-500/10 text-emerald-700'
                                            : status === 'blocked'
                                                ? 'border-destructive/40 bg-destructive/10 text-destructive'
                                                : 'border-primary/40 bg-primary/10 text-primary'
                                        : 'border-input bg-background text-muted-foreground hover:bg-accent hover:text-foreground'"
                                    @click="updateApproval(approval.id, { status: status as SqlNotebookRunbookApprovalStatus })"
                                >
                                    {{ approvalStatusLabel[status as SqlNotebookRunbookApprovalStatus] }}
                                </button>
                            </div>

                            <textarea
                                :value="approval.note"
                                rows="2"
                                class="mt-3 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus:border-primary"
                                placeholder="Approval note, link, or condition"
                                @input="updateApproval(approval.id, { note: ($event.target as HTMLTextAreaElement).value })"
                            />
                        </div>
                    </div>
                </div>

                <div>
                    <label class="text-xs font-semibold uppercase tracking-[0.12em] text-muted-foreground">Evidence</label>
                    <textarea
                        :value="runbook.evidence"
                        rows="4"
                        class="mt-2 w-full rounded-md border border-input bg-background px-3 py-2 text-sm outline-none transition-colors focus:border-primary"
                        @input="updateRunbookField({ evidence: ($event.target as HTMLTextAreaElement).value })"
                    />
                </div>
            </div>
        </div>
    </div>
</template>

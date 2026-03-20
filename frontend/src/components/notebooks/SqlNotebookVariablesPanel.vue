<script lang="ts" setup>
import type { SqlNotebookVariable } from '../../types/sqlNotebook';

const props = defineProps<{
    variables: SqlNotebookVariable[];
    isReadOnly?: boolean;
}>();

const emit = defineEmits<{
    add: [];
    remove: [key: string];
    update: [payload: { key: string; patch: Partial<SqlNotebookVariable> }];
}>();
</script>

<template>
    <section class="space-y-3">
        <div class="flex items-center justify-between gap-3">
            <div>
                <div class="text-xs uppercase tracking-[0.12em] text-muted-foreground">Variables</div>
                <div class="mt-1 text-xs text-muted-foreground">Reusable values for SQL placeholders like <code v-pre>{{order_id}}</code>.</div>
            </div>
            <button
                class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium transition-colors hover:bg-accent"
                :disabled="props.isReadOnly"
                @click="emit('add')"
            >
                Add Variable
            </button>
        </div>

        <div v-if="props.variables.length === 0" class="rounded-lg border border-dashed border-border p-3 text-xs text-muted-foreground">
            No variables yet. Add one here or type placeholders in SQL cells and then fill them here.
        </div>

        <div v-else class="space-y-3">
            <div
                v-for="variable in props.variables"
                :key="variable.key"
                class="rounded-lg border border-border bg-background/80 p-3"
            >
                <div class="grid gap-3">
                    <div class="grid grid-cols-2 gap-3">
                        <label class="space-y-1">
                            <div class="text-[11px] font-medium uppercase tracking-[0.12em] text-muted-foreground">Key</div>
                            <input
                                :value="variable.key"
                                type="text"
                                class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:border-primary"
                                :disabled="props.isReadOnly"
                                @input="emit('update', { key: variable.key, patch: { key: ($event.target as HTMLInputElement).value.trim() } })"
                            />
                        </label>
                        <label class="space-y-1">
                            <div class="text-[11px] font-medium uppercase tracking-[0.12em] text-muted-foreground">Label</div>
                            <input
                                :value="variable.label"
                                type="text"
                                class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:border-primary"
                                :disabled="props.isReadOnly"
                                @input="emit('update', { key: variable.key, patch: { label: ($event.target as HTMLInputElement).value } })"
                            />
                        </label>
                    </div>

                    <label class="space-y-1">
                        <div class="text-[11px] font-medium uppercase tracking-[0.12em] text-muted-foreground">Value</div>
                        <input
                            :value="variable.value"
                            type="text"
                            class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:border-primary"
                            :disabled="props.isReadOnly"
                            @input="emit('update', { key: variable.key, patch: { value: ($event.target as HTMLInputElement).value } })"
                        />
                    </label>

                    <div class="grid grid-cols-[minmax(0,1fr)_auto_auto] gap-3">
                        <label class="space-y-1">
                            <div class="text-[11px] font-medium uppercase tracking-[0.12em] text-muted-foreground">Type</div>
                            <select
                                :value="variable.type"
                                class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm outline-none focus:border-primary"
                                :disabled="props.isReadOnly"
                                @change="emit('update', { key: variable.key, patch: { type: ($event.target as HTMLSelectElement).value as SqlNotebookVariable['type'] } })"
                            >
                                <option value="text">Text</option>
                                <option value="number">Number</option>
                                <option value="date">Date</option>
                            </select>
                        </label>

                        <label class="flex items-end gap-2 pb-2 text-sm text-muted-foreground">
                            <input
                                :checked="variable.required"
                                type="checkbox"
                                :disabled="props.isReadOnly"
                                @change="emit('update', { key: variable.key, patch: { required: ($event.target as HTMLInputElement).checked } })"
                            />
                            Required
                        </label>

                        <button
                            class="inline-flex h-9 items-center justify-center self-end rounded-md border border-input bg-background px-3 text-xs font-medium text-destructive transition-colors hover:bg-destructive/10"
                            :disabled="props.isReadOnly"
                            @click="emit('remove', variable.key)"
                        >
                            Remove
                        </button>
                    </div>
                </div>
            </div>
        </div>
    </section>
</template>

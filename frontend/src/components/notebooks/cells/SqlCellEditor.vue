<script lang="ts" setup>
import { defineAsyncComponent } from 'vue';

const SqlEditor = defineAsyncComponent(() => import('../../SqlEditor.vue'));

const props = defineProps<{
    title: string;
    content: string;
    tables: string[];
    getColumns: (tableName: string) => Promise<string[]>;
    editorSettings: { fontFamily: string; fontSize: number };
    isReadOnly?: boolean;
}>();

const emit = defineEmits<{
    'update:title': [value: string];
    'update:content': [value: string];
    run: [];
}>();
</script>

<template>
    <div class="space-y-3" @click.stop>
        <div class="flex items-center gap-3">
            <input
                :value="props.title"
                type="text"
                placeholder="SQL cell title"
                class="min-w-0 flex-1 rounded-md border border-input bg-background px-3 py-2 text-sm font-medium outline-none transition-colors focus:border-primary"
                @click.stop
                @input="emit('update:title', ($event.target as HTMLInputElement).value)"
            />
            <button
                class="inline-flex h-9 items-center justify-center rounded-md bg-primary px-3 text-sm font-medium text-primary-foreground transition-colors hover:bg-primary/90 disabled:opacity-50"
                :disabled="props.isReadOnly"
                @click.stop="emit('run')"
            >
                Run Cell
            </button>
        </div>

        <div class="relative z-20 overflow-visible rounded-md border border-border bg-background" @click.stop>
            <SqlEditor
                :model-value="props.content"
                :tables="props.tables"
                :get-columns="props.getColumns"
                :font-family="props.editorSettings.fontFamily"
                :font-size="props.editorSettings.fontSize"
                @update:model-value="emit('update:content', $event)"
            />
        </div>
    </div>
</template>

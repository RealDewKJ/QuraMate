<script lang="ts" setup>
import { computed } from 'vue';

interface MarkdownPreviewBlock {
    type: 'heading' | 'paragraph' | 'list' | 'code';
    level?: number;
    text?: string;
    lines?: string[];
}

const props = defineProps<{
    title: string;
    content: string;
}>();

const emit = defineEmits<{
    'update:title': [value: string];
    'update:content': [value: string];
}>();

const previewBlocks = computed<MarkdownPreviewBlock[]>(() => {
    const blocks: MarkdownPreviewBlock[] = [];
    const lines = props.content.replace(/\r\n/g, '\n').split('\n');
    let index = 0;

    while (index < lines.length) {
        const rawLine = lines[index];
        const line = rawLine.trimEnd();
        const trimmed = line.trim();

        if (!trimmed) {
            index += 1;
            continue;
        }

        if (trimmed.startsWith('```')) {
            const codeLines: string[] = [];
            index += 1;
            while (index < lines.length && !lines[index].trim().startsWith('```')) {
                codeLines.push(lines[index]);
                index += 1;
            }
            if (index < lines.length) {
                index += 1;
            }
            blocks.push({
                type: 'code',
                lines: codeLines.length > 0 ? codeLines : [''],
            });
            continue;
        }

        const headingMatch = trimmed.match(/^(#{1,4})\s+(.*)$/);
        if (headingMatch) {
            blocks.push({
                type: 'heading',
                level: headingMatch[1].length,
                text: headingMatch[2],
            });
            index += 1;
            continue;
        }

        if (/^[-*]\s+/.test(trimmed) || /^\d+\.\s+/.test(trimmed)) {
            const items: string[] = [];
            while (index < lines.length) {
                const currentTrimmed = lines[index].trim();
                if (!(/^[-*]\s+/.test(currentTrimmed) || /^\d+\.\s+/.test(currentTrimmed))) {
                    break;
                }
                items.push(currentTrimmed.replace(/^[-*]\s+/, '').replace(/^\d+\.\s+/, ''));
                index += 1;
            }
            blocks.push({
                type: 'list',
                lines: items,
            });
            continue;
        }

        const paragraphLines: string[] = [];
        while (index < lines.length && lines[index].trim()) {
            const paragraphTrimmed = lines[index].trim();
            if (
                paragraphTrimmed.startsWith('```') ||
                /^(#{1,4})\s+/.test(paragraphTrimmed) ||
                /^[-*]\s+/.test(paragraphTrimmed) ||
                /^\d+\.\s+/.test(paragraphTrimmed)
            ) {
                break;
            }
            paragraphLines.push(paragraphTrimmed);
            index += 1;
        }
        blocks.push({
            type: 'paragraph',
            lines: paragraphLines,
        });
    }

    return blocks;
});

</script>

<template>
    <div class="space-y-4">
        <input
            :value="props.title"
            type="text"
            placeholder="Notes title"
            class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm font-medium outline-none transition-colors focus:border-primary"
            @input="emit('update:title', ($event.target as HTMLInputElement).value)"
        />

        <div class="grid items-stretch gap-4 xl:grid-cols-[minmax(0,1.15fr)_minmax(320px,0.85fr)]">
            <div class="rounded-xl border border-border bg-background shadow-sm">
                <div class="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
                    <div>
                        <div class="text-sm font-semibold">Markdown Notes</div>
                        <div class="text-xs text-muted-foreground">Write context, caveats, saved table names, and reusable SQL notes.</div>
                    </div>
                    <div class="rounded-md bg-muted px-2 py-1 text-[11px] font-medium text-muted-foreground">Editor</div>
                </div>

                <textarea
                    :value="props.content"
                    rows="14"
                    placeholder="Use markdown to document goals, expected outputs, saved tables, caveats, and follow-up steps."
                    class="min-h-[280px] w-full resize-y border-0 bg-transparent px-4 py-4 font-mono text-sm leading-6 outline-none"
                    @input="emit('update:content', ($event.target as HTMLTextAreaElement).value)"
                />
            </div>

            <div class="flex min-h-[280px]">
                <div class="flex min-h-[280px] w-full flex-col rounded-xl border border-border bg-card shadow-sm">
                    <div class="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
                        <div>
                            <div class="text-sm font-semibold">Preview</div>
                            <div class="text-xs text-muted-foreground">Readable rendering for the note you are saving.</div>
                        </div>
                        <div class="rounded-md bg-muted px-2 py-1 text-[11px] font-medium text-muted-foreground">Markdown</div>
                    </div>

                    <div class="min-h-[280px] flex-1 space-y-4 overflow-auto px-4 py-4">
                        <template v-if="previewBlocks.length > 0">
                            <template v-for="(block, blockIndex) in previewBlocks" :key="blockIndex">
                                <h2
                                    v-if="block.type === 'heading' && block.level === 1"
                                    class="text-2xl font-semibold tracking-tight text-foreground"
                                >
                                    {{ block.text }}
                                </h2>
                                <h3
                                    v-else-if="block.type === 'heading' && block.level === 2"
                                    class="text-lg font-semibold text-foreground"
                                >
                                    {{ block.text }}
                                </h3>
                                <h4
                                    v-else-if="block.type === 'heading'"
                                    class="text-sm font-semibold uppercase tracking-[0.12em] text-muted-foreground"
                                >
                                    {{ block.text }}
                                </h4>
                                <p
                                    v-else-if="block.type === 'paragraph'"
                                    class="text-sm leading-7 text-foreground"
                                >
                                    {{ block.lines?.join(' ') }}
                                </p>
                                <ul
                                    v-else-if="block.type === 'list'"
                                    class="space-y-2 rounded-lg border border-border/70 bg-background/80 p-3"
                                >
                                    <li
                                        v-for="(line, lineIndex) in block.lines"
                                        :key="lineIndex"
                                        class="flex gap-3 text-sm leading-6 text-foreground"
                                    >
                                        <span class="mt-2 h-1.5 w-1.5 rounded-full bg-primary"></span>
                                        <span>{{ line }}</span>
                                    </li>
                                </ul>
                                <pre
                                    v-else
                                    class="overflow-auto rounded-lg border border-border bg-muted/70 p-3 text-xs leading-6 text-foreground"
                                ><code>{{ block.lines?.join('\n') }}</code></pre>
                            </template>
                        </template>
                        <div v-else class="rounded-lg border border-dashed border-border p-4 text-sm text-muted-foreground">
                            Start typing markdown to preview the note here.
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

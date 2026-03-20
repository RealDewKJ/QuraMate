<template>
    <div v-if="isOpen" :class="rootClass">
        <div
            v-if="!embedded"
            class="absolute inset-0 bg-black/50 backdrop-blur-sm"
            @click="$emit('close')"
        ></div>
        <div :class="panelClass">
            <div class="flex items-center justify-between px-5 py-4 border-b border-border bg-muted/20">
                <div>
                    <h3 class="text-base font-semibold">{{ t("common.aiCopilot.title") }}</h3>
                    <p class="text-xs text-muted-foreground">{{ t("common.aiCopilot.description") }}</p>
                </div>
                <button @click="$emit('close')"
                    class="h-8 w-8 inline-flex items-center justify-center rounded-md text-muted-foreground hover:bg-accent hover:text-foreground">
                    <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                        stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <path d="M18 6 6 18" />
                        <path d="m6 6 12 12" />
                    </svg>
                </button>
            </div>
            <div :class="contentClass">
                <div class="space-y-3">
                    <div class="grid gap-2">
                        <label class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Task</label>
                        <select v-model="selectedMode"
                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                            <option v-for="option in modeOptions" :key="option.value" :value="option.value">
                                {{ option.label }}
                            </option>
                        </select>
                    </div>
                    <div v-if="mode === 'backend_code'" class="grid gap-2">
                        <label class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Backend Language</label>
                        <select v-model="selectedBackendLanguage"
                            class="flex h-10 w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring">
                            <option>Node.js</option>
                            <option>Python</option>
                            <option>Java</option>
                            <option>Go</option>
                            <option>C#</option>
                        </select>
                    </div>
                    <div class="grid gap-2">
                        <label class="text-xs font-semibold uppercase tracking-wide text-muted-foreground">Instruction</label>
                        <textarea v-model="instructionText" rows="8" placeholder="Describe what you want..."
                            class="w-full rounded-md border border-input bg-background px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-ring"></textarea>
                    </div>
                    <div class="flex items-center gap-2">
                        <button @click="$emit('run')" :disabled="isLoading"
                            class="inline-flex items-center justify-center rounded-md text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 h-9 px-4 py-2 disabled:pointer-events-none disabled:opacity-50">
                            <svg v-if="isLoading" class="animate-spin mr-2 h-4 w-4" xmlns="http://www.w3.org/2000/svg"
                                fill="none" viewBox="0 0 24 24">
                                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" />
                                <path class="opacity-75" fill="currentColor" d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z" />
                            </svg>
                            {{ isLoading ? 'Running...' : 'Run Copilot' }}
                        </button>
                    </div>
                    <p v-if="error" class="text-xs text-red-500 break-all">{{ error }}</p>
                    <p v-if="latencyMs > 0" class="text-xs text-muted-foreground">Latency: {{ latencyMs }} ms</p>
                </div>
                    <div class="border border-border rounded-md bg-card min-h-[360px] flex flex-col overflow-hidden">
                    <div class="flex items-center justify-between gap-3 px-3 py-2 border-b border-border">
                        <div class="text-xs font-semibold uppercase text-muted-foreground">AI Output</div>
                        <div class="inline-flex rounded-md border border-border bg-muted/50 p-1">
                            <button
                                @click="viewMode = 'formatted'"
                                :class="[
                                    'rounded px-2 py-1 text-[11px] font-medium transition-colors',
                                    viewMode === 'formatted'
                                        ? 'bg-background text-foreground shadow-sm'
                                        : 'text-muted-foreground hover:text-foreground'
                                ]"
                            >
                                Formatted
                            </button>
                            <button
                                @click="viewMode = 'raw'"
                                :class="[
                                    'rounded px-2 py-1 text-[11px] font-medium transition-colors',
                                    viewMode === 'raw'
                                        ? 'bg-background text-foreground shadow-sm'
                                        : 'text-muted-foreground hover:text-foreground'
                                ]"
                            >
                                Raw
                            </button>
                        </div>
                    </div>
                    <div class="flex-1 overflow-auto bg-[linear-gradient(180deg,rgba(245,126,34,0.05),transparent_160px)]">
                        <div v-if="!result" class="p-4 text-xs text-muted-foreground">
                            No output yet.
                        </div>
                        <div v-else class="space-y-4 p-4">
                            <div
                                v-if="hasDetectedSql"
                                class="overflow-hidden rounded-xl border border-emerald-500/30 bg-emerald-500/5"
                            >
                                <div class="flex flex-wrap items-center justify-between gap-3 border-b border-emerald-500/20 bg-emerald-500/10 px-4 py-3">
                                    <div>
                                        <div class="text-[10px] font-semibold uppercase tracking-[0.2em] text-emerald-700 dark:text-emerald-300">
                                            Detected SQL
                                        </div>
                                        <div class="mt-1 text-sm font-medium text-foreground">
                                            Ready to copy or apply without scanning the reasoning output.
                                        </div>
                                    </div>
                                    <div class="flex flex-wrap items-center gap-2">
                                        <button
                                            @click="$emit('apply-sql')"
                                            class="inline-flex items-center justify-center rounded-md bg-emerald-600 px-3 py-2 text-xs font-medium text-white hover:bg-emerald-500"
                                        >
                                            Apply SQL
                                        </button>
                                        <button
                                            @click="copySuggestedSql"
                                            class="inline-flex items-center justify-center rounded-md border border-emerald-500/30 bg-background px-3 py-2 text-xs font-medium text-foreground hover:bg-accent"
                                        >
                                            Copy SQL
                                        </button>
                                    </div>
                                </div>
                                <div class="border-b border-emerald-500/10 px-4 py-2 text-xs text-muted-foreground">
                                    {{ sqlActionStatus }}
                                </div>
                                <pre class="max-h-52 overflow-auto px-4 py-3 text-xs leading-6 text-foreground"><code>{{ suggestedSqlPreview }}</code></pre>
                            </div>

                            <div
                                v-if="viewMode === 'raw'"
                                class="overflow-hidden rounded-lg border border-border bg-muted/50"
                            >
                                <div class="border-b border-border bg-muted px-3 py-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-muted-foreground">
                                    Raw response
                                </div>
                                <pre class="overflow-auto p-3 text-xs leading-6 text-foreground whitespace-pre-wrap"><code>{{ result }}</code></pre>
                            </div>

                            <div
                                v-else-if="hasDetectedSql && !hasFormattedReasoning"
                                class="rounded-lg border border-border/60 bg-background/80 p-3 text-xs leading-6 text-muted-foreground"
                            >
                                SQL-only response detected. Use the SQL panel above to copy or apply it directly.
                            </div>

                            <template v-else>
                                <template v-for="(block, blockIndex) in visibleOutputBlocks" :key="`${block.type}-${blockIndex}`">
                                <div
                                    v-if="block.type === 'heading'"
                                    class="rounded-lg border border-primary/20 bg-primary/5 px-3 py-2"
                                >
                                    <div class="text-[10px] font-semibold uppercase tracking-[0.2em] text-primary/80">
                                        Section
                                    </div>
                                    <div class="mt-1 text-sm font-semibold text-foreground">
                                        <template v-for="(segment, segmentIndex) in block.segments" :key="segmentIndex">
                                            <strong v-if="segment.bold">{{ segment.text }}</strong>
                                            <span v-else>{{ segment.text }}</span>
                                        </template>
                                    </div>
                                </div>

                                <div
                                    v-else-if="block.type === 'code'"
                                    class="overflow-hidden rounded-lg border border-border bg-muted/50"
                                >
                                    <div class="flex items-center justify-between border-b border-border bg-muted px-3 py-2 text-[10px] font-semibold uppercase tracking-[0.2em] text-muted-foreground">
                                        <span>{{ block.label }}</span>
                                        <span>{{ block.language }}</span>
                                    </div>
                                    <pre class="overflow-auto p-3 text-xs leading-6 text-foreground"><code>{{ block.content }}</code></pre>
                                </div>

                                <ul
                                    v-else-if="block.type === 'list'"
                                    class="space-y-2 rounded-lg border border-border/60 bg-background/70 p-3"
                                >
                                    <li
                                        v-for="(item, itemIndex) in block.items"
                                        :key="itemIndex"
                                        class="flex gap-3 text-xs leading-6 text-foreground"
                                    >
                                        <span class="mt-2 h-1.5 w-1.5 flex-none rounded-full bg-primary"></span>
                                        <span class="min-w-0">
                                            <template v-for="(segment, segmentIndex) in item.segments" :key="segmentIndex">
                                                <strong v-if="segment.bold">{{ segment.text }}</strong>
                                                <span v-else>{{ segment.text }}</span>
                                            </template>
                                        </span>
                                    </li>
                                </ul>

                                <div
                                    v-else
                                    class="rounded-lg border border-border/60 bg-background/80 p-3 text-xs leading-6 text-foreground"
                                >
                                    <template v-for="(line, lineIndex) in block.lines" :key="lineIndex">
                                        <p v-if="line.text.trim()" :class="lineIndex > 0 ? 'mt-2' : ''">
                                            <template v-for="(segment, segmentIndex) in line.segments" :key="segmentIndex">
                                                <strong v-if="segment.bold">{{ segment.text }}</strong>
                                                <span v-else>{{ segment.text }}</span>
                                            </template>
                                        </p>
                                    </template>
                                </div>
                                </template>
                            </template>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import type { AiCopilotMode, AiCopilotModeOption } from '../types/aiCopilot';

interface InlineSegment {
    text: string;
    bold: boolean;
}

interface HeadingBlock {
    type: 'heading';
    segments: InlineSegment[];
}

interface CodeBlock {
    type: 'code';
    language: string;
    label: string;
    content: string;
}

interface ListBlock {
    type: 'list';
    items: Array<{ segments: InlineSegment[] }>;
}

interface ParagraphBlock {
    type: 'paragraph';
    lines: Array<{ text: string; segments: InlineSegment[] }>;
}

type OutputBlock = HeadingBlock | CodeBlock | ListBlock | ParagraphBlock;

const props = withDefaults(defineProps<{
    isOpen: boolean;
    mode: AiCopilotMode;
    modeOptions: AiCopilotModeOption[];
    prompt: string;
    backendLanguage: string;
    isLoading: boolean;
    result: string;
    error: string;
    latencyMs: number;
    suggestedSql: string;
    embedded?: boolean;
}>(), {
    embedded: false,
});

const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'run'): void;
    (e: 'apply-sql'): void;
    (e: 'update:mode', value: AiCopilotMode): void;
    (e: 'update:prompt', value: string): void;
    (e: 'update:backend-language', value: string): void;
}>();
const { t } = useI18n({ useScope: 'global' });
const viewMode = ref<'formatted' | 'raw'>('formatted');
const sqlActionStatus = ref('Use these actions when the response contains runnable SQL.');
const rootClass = computed(() =>
    props.embedded
        ? 'flex h-full flex-col overflow-hidden bg-background'
        : 'fixed inset-0 z-[90] flex items-center justify-center p-4',
);
const panelClass = computed(() =>
    props.embedded
        ? 'flex h-full w-full flex-col overflow-hidden bg-background'
        : 'relative z-[91] flex max-h-[86vh] w-full max-w-5xl flex-col overflow-hidden rounded-xl border border-border bg-background shadow-2xl',
);
const contentClass = computed(() =>
    props.embedded
        ? 'grid flex-1 grid-cols-1 gap-4 overflow-auto p-4 lg:grid-cols-2'
        : 'grid grid-cols-1 gap-4 overflow-auto p-4 lg:grid-cols-2',
);

const selectedMode = computed({
    get: () => props.mode,
    set: (value: AiCopilotMode) => emit('update:mode', value)
});

const instructionText = computed({
    get: () => props.prompt,
    set: (value: string) => emit('update:prompt', value)
});

const selectedBackendLanguage = computed({
    get: () => props.backendLanguage,
    set: (value: string) => emit('update:backend-language', value)
});

const parseInlineSegments = (text: string): InlineSegment[] => {
    const segments: InlineSegment[] = [];
    const boldPattern = /\*\*(.+?)\*\*/g;
    let lastIndex = 0;
    let match: RegExpExecArray | null;

    while ((match = boldPattern.exec(text)) !== null) {
        if (match.index > lastIndex) {
            segments.push({
                text: text.slice(lastIndex, match.index),
                bold: false
            });
        }
        segments.push({
            text: match[1],
            bold: true
        });
        lastIndex = match.index + match[0].length;
    }

    if (lastIndex < text.length) {
        segments.push({
            text: text.slice(lastIndex),
            bold: false
        });
    }

    return segments.length > 0 ? segments : [{ text, bold: false }];
};

const normalizeListText = (text: string) => {
    return text.replace(/^(\s*)([-*]|\d+\.)\s+/, '').trim();
};

const outputBlocks = computed<OutputBlock[]>(() => {
    const source = props.result || '';
    if (!source.trim()) {
        return [];
    }

    const lines = source.replace(/\r\n/g, '\n').split('\n');
    const blocks: OutputBlock[] = [];
    let index = 0;

    while (index < lines.length) {
        const line = lines[index];
        const trimmed = line.trim();

        if (!trimmed) {
            index += 1;
            continue;
        }

        if (trimmed.startsWith('```')) {
            const language = trimmed.slice(3).trim() || 'text';
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
                language,
                label: language === 'sql' ? 'Runnable SQL' : 'Code',
                content: codeLines.join('\n').trim()
            });
            continue;
        }

        if (/^#{1,6}\s+/.test(trimmed) || /^\d+\.\s+\*\*/.test(trimmed) || /^\*\*.+\*\*$/.test(trimmed)) {
            const headingText = trimmed
                .replace(/^#{1,6}\s+/, '')
                .replace(/^\d+\.\s+/, '')
                .replace(/:$/, '');
            blocks.push({
                type: 'heading',
                segments: parseInlineSegments(headingText)
            });
            index += 1;
            continue;
        }

        if (/^(\*|-|\d+\.)\s+/.test(trimmed)) {
            const items: Array<{ segments: InlineSegment[] }> = [];
            while (index < lines.length) {
                const itemLine = lines[index].trim();
                if (!/^(\*|-|\d+\.)\s+/.test(itemLine)) {
                    break;
                }
                items.push({
                    segments: parseInlineSegments(normalizeListText(itemLine))
                });
                index += 1;
            }
            blocks.push({
                type: 'list',
                items
            });
            continue;
        }

        const paragraphLines: Array<{ text: string; segments: InlineSegment[] }> = [];
        while (index < lines.length) {
            const paragraphLine = lines[index];
            const paragraphTrimmed = paragraphLine.trim();
            if (
                !paragraphTrimmed ||
                paragraphTrimmed.startsWith('```') ||
                /^#{1,6}\s+/.test(paragraphTrimmed) ||
                /^(\*|-|\d+\.)\s+/.test(paragraphTrimmed)
            ) {
                break;
            }
            paragraphLines.push({
                text: paragraphTrimmed,
                segments: parseInlineSegments(paragraphTrimmed)
            });
            index += 1;
        }
        blocks.push({
            type: 'paragraph',
            lines: paragraphLines
        });
    }

    return blocks;
});

const suggestedSqlPreview = computed(() => props.suggestedSql.trim());
const hasDetectedSql = computed(() => suggestedSqlPreview.value.length > 0);
const visibleOutputBlocks = computed(() => {
    if (!hasDetectedSql.value) {
        return outputBlocks.value;
    }

    return outputBlocks.value.filter((block) => {
        if (block.type !== 'code') {
            return true;
        }

        return block.content.trim() !== suggestedSqlPreview.value;
    });
});
const hasFormattedReasoning = computed(() => visibleOutputBlocks.value.length > 0);

const copySuggestedSql = async () => {
    if (!hasDetectedSql.value) {
        return;
    }

    try {
        await navigator.clipboard.writeText(suggestedSqlPreview.value);
        sqlActionStatus.value = 'SQL copied to clipboard.';
    } catch (error) {
        console.error('Failed to copy suggested SQL', error);
        sqlActionStatus.value = 'Failed to copy SQL in this environment.';
    }
};

const handleEscapeKeydown = (event: KeyboardEvent) => {
    if (event.key === 'Escape' && props.isOpen && !props.embedded) {
        emit('close');
    }
};

onMounted(() => {
    window.addEventListener('keydown', handleEscapeKeydown);
});

const resetOutputPresentation = () => {
    viewMode.value = 'formatted';
    sqlActionStatus.value = hasDetectedSql.value
        ? 'Use these actions when the response contains runnable SQL.'
        : 'No runnable SQL detected yet.';
};

onBeforeUnmount(() => {
    window.removeEventListener('keydown', handleEscapeKeydown);
});

watch(
    () => [props.isOpen, props.result, props.suggestedSql],
    ([isOpen]) => {
        if (isOpen) {
            resetOutputPresentation();
        }
    },
    { immediate: true }
);
</script>

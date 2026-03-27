<script lang="ts" setup>
import { useResizeObserver } from '@vueuse/core';
import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';

import {
    buildMarkdownImageMarkup,
    buildEmbeddedImageUrl,
    extractDataUrlImages,
    fileToDataUrl,
    getImageAltText,
    isImageFile,
    makeImageId,
    parseMarkdownPreviewBlocks,
    resolveInlineImageUrl,
    type MarkdownPreviewBlock,
    type MarkdownEmbeddedImage,
} from '../../../lib/markdownImages';

const props = defineProps<{
    title: string;
    content: string;
    embeddedImages?: MarkdownEmbeddedImage[];
}>();

const emit = defineEmits<{
    'update:title': [value: string];
    'update:content': [value: string];
    'update:embedded-images': [value: MarkdownEmbeddedImage[]];
}>();

const editorRef = ref<HTMLTextAreaElement | null>(null);
const previewRef = ref<HTMLDivElement | null>(null);
const imageInputRef = ref<HTMLInputElement | null>(null);
const syncingSource = ref<'editor' | 'preview' | null>(null);
const isImageDropActive = ref(false);
const uploadStatus = ref<string | null>(null);

const getScrollRatio = (element: HTMLElement): number => {
    const scrollableHeight = element.scrollHeight - element.clientHeight;
    if (scrollableHeight <= 0) {
        return 0;
    }

    return element.scrollTop / scrollableHeight;
};

const setScrollRatio = (element: HTMLElement, ratio: number) => {
    const nextRatio = Number.isFinite(ratio) ? Math.min(Math.max(ratio, 0), 1) : 0;
    const scrollableHeight = element.scrollHeight - element.clientHeight;
    element.scrollTop = scrollableHeight > 0 ? scrollableHeight * nextRatio : 0;
};

const syncScroll = (source: 'editor' | 'preview') => {
    const editor = editorRef.value;
    const preview = previewRef.value;
    if (!editor || !preview) {
        return;
    }

    const sourceEl = source === 'editor' ? editor : preview;
    const targetEl = source === 'editor' ? preview : editor;

    syncingSource.value = source;
    setScrollRatio(targetEl, getScrollRatio(sourceEl));
    requestAnimationFrame(() => {
        if (syncingSource.value === source) {
            syncingSource.value = null;
        }
    });
};

const handleEditorScroll = () => {
    if (syncingSource.value === 'preview') {
        return;
    }

    syncScroll('editor');
};

const handlePreviewScroll = () => {
    if (syncingSource.value === 'editor') {
        return;
    }

    syncScroll('preview');
};

const refreshPreviewAlignment = async () => {
    await nextTick();

    const editor = editorRef.value;
    const preview = previewRef.value;
    if (!editor || !preview) {
        return;
    }

    setScrollRatio(preview, getScrollRatio(editor));
};

const previewBlocks = computed<MarkdownPreviewBlock[]>(() => {
    return parseMarkdownPreviewBlocks(props.content);
});

const migrateLegacyDataUrlImages = () => {
    if ((props.embeddedImages || []).length > 0 || !props.content.includes('data:image/')) {
        return;
    }

    const migrated = extractDataUrlImages(props.content);
    if (migrated.images.length === 0) {
        return;
    }

    emit('update:embedded-images', migrated.images);
    emit('update:content', migrated.content);
};

const insertMarkdownAtSelection = async (markdownText: string) => {
    const editor = editorRef.value;
    if (!editor) {
        return;
    }

    const start = editor.selectionStart ?? props.content.length;
    const end = editor.selectionEnd ?? props.content.length;
    const before = props.content.slice(0, start);
    const after = props.content.slice(end);
    const needsLeadingBreak = before.length > 0 && !before.endsWith('\n') && !markdownText.startsWith('\n');
    const needsTrailingBreak = after.length > 0 && !after.startsWith('\n') && !markdownText.endsWith('\n');
    const insertion = `${needsLeadingBreak ? '\n' : ''}${markdownText}${needsTrailingBreak ? '\n' : ''}`;
    const nextContent = `${before}${insertion}${after}`;
    const cursorPosition = before.length + insertion.length;

    emit('update:content', nextContent);
    await nextTick();
    editor.focus();
    editor.setSelectionRange(cursorPosition, cursorPosition);
    void refreshPreviewAlignment();
};

const insertImageFile = async (file: File) => {
    if (!isImageFile(file)) {
        uploadStatus.value = 'Only image files can be inserted.';
        return;
    }

    const maxBytes = 8 * 1024 * 1024;
    if (file.size > maxBytes) {
        uploadStatus.value = 'Image is too large. Please use a file smaller than 8 MB.';
        return;
    }

    uploadStatus.value = `Inserting ${file.name}...`;
    try {
        const dataUrl = await fileToDataUrl(file);
        const imageId = makeImageId();
        const embeddedImages = [...(props.embeddedImages || [])];
        embeddedImages.push({
            id: imageId,
            alt: getImageAltText(file.name),
            fileName: file.name,
            mimeType: file.type || 'image/png',
            dataUrl,
        });
        emit('update:embedded-images', embeddedImages);
        const markdown = buildMarkdownImageMarkup(getImageAltText(file.name), buildEmbeddedImageUrl(imageId));
        await insertMarkdownAtSelection(markdown);
        uploadStatus.value = null;
    } catch (_error) {
        uploadStatus.value = 'Could not read that image file.';
    }
};

const handleImageButtonClick = () => {
    imageInputRef.value?.click();
};

const handleImageInputChange = async (event: Event) => {
    const input = event.target as HTMLInputElement;
    const files = Array.from(input.files || []);
    input.value = '';
    if (files.length === 0) {
        return;
    }

    await insertImageFile(files[0]);
};

const handlePaste = async (event: ClipboardEvent) => {
    const files = Array.from(event.clipboardData?.files || []);
    const imageFile = files.find((file) => isImageFile(file));
    if (!imageFile) {
        return;
    }

    event.preventDefault();
    await insertImageFile(imageFile);
};

const handleDrop = async (event: DragEvent) => {
    const files = Array.from(event.dataTransfer?.files || []);
    const imageFile = files.find((file) => isImageFile(file));
    if (!imageFile) {
        isImageDropActive.value = false;
        return;
    }

    event.preventDefault();
    isImageDropActive.value = false;
    await insertImageFile(imageFile);
};

const handleDragOver = (event: DragEvent) => {
    const hasImageFile = Array.from(event.dataTransfer?.items || []).some((item) => item.kind === 'file' && item.type.startsWith('image/'));
    if (!hasImageFile) {
        return;
    }

    event.preventDefault();
    isImageDropActive.value = true;
};

const handleDragLeave = () => {
    isImageDropActive.value = false;
};

watch(() => props.content, () => {
    void refreshPreviewAlignment();
});

watch(() => [props.content, props.embeddedImages], () => {
    migrateLegacyDataUrlImages();
});

useResizeObserver(previewRef, () => {
    if (syncingSource.value) {
        return;
    }

    void refreshPreviewAlignment();
});

useResizeObserver(editorRef, () => {
    if (syncingSource.value) {
        return;
    }

    void refreshPreviewAlignment();
});

onBeforeUnmount(() => {
    syncingSource.value = null;
    isImageDropActive.value = false;
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
            <div class="flex min-h-[280px]">
                <div
                    class="flex min-h-[280px] w-full flex-col rounded-xl border border-border bg-background shadow-sm transition-colors"
                    :class="isImageDropActive ? 'border-primary/70 bg-primary/5' : ''"
                    @dragenter.prevent="handleDragOver"
                    @dragover="handleDragOver"
                    @dragleave="handleDragLeave"
                    @drop="handleDrop"
                >
                    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-border px-4 py-3">
                        <div class="min-w-0">
                            <div class="text-sm font-semibold">Markdown Notes</div>
                            <div class="text-xs text-muted-foreground">
                                Write context, caveats, saved table names, and reusable SQL notes. Use <code class="rounded bg-muted px-1 py-0.5 text-[11px]">![alt](https://...)</code> or paste/upload an image.
                            </div>
                            <div v-if="uploadStatus" class="mt-1 text-[11px] text-muted-foreground">
                                {{ uploadStatus }}
                            </div>
                        </div>
                        <div class="flex items-center gap-2">
                            <button
                                type="button"
                                class="inline-flex h-8 items-center justify-center rounded-md border border-input bg-background px-3 text-xs font-medium text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                                @click="handleImageButtonClick"
                            >
                                Insert Image
                            </button>
                            <div class="rounded-md bg-muted px-2 py-1 text-[11px] font-medium text-muted-foreground">Editor</div>
                        </div>
                    </div>

                    <input
                        ref="imageInputRef"
                        type="file"
                        accept="image/*"
                        class="hidden"
                        @change="handleImageInputChange"
                    />

                    <textarea
                        ref="editorRef"
                        :value="props.content"
                        rows="14"
                        wrap="soft"
                        spellcheck="false"
                        placeholder="Use markdown to document goals, expected outputs, saved tables, caveats, and follow-up steps."
                        class="h-full min-h-0 w-full flex-1 resize-y border-0 bg-transparent px-4 py-4 font-mono text-sm leading-6 outline-none whitespace-pre-wrap break-words [overflow-wrap:anywhere]"
                        @input="emit('update:content', ($event.target as HTMLTextAreaElement).value)"
                        @paste="handlePaste"
                        @scroll="handleEditorScroll"
                    />
                </div>
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

                    <div
                        ref="previewRef"
                        class="min-h-[280px] flex-1 space-y-4 overflow-auto px-4 py-4"
                        @scroll="handlePreviewScroll"
                    >
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
                                <div
                                    v-else-if="block.type === 'paragraph'"
                                    class="space-y-3 text-sm leading-7 text-foreground"
                                >
                                    <template v-for="(segment, segmentIndex) in block.segments" :key="segmentIndex">
                                        <span v-if="segment.type === 'text'">{{ segment.text }}</span>
                                        <figure v-else class="space-y-2">
                                            <img
                                                :src="resolveInlineImageUrl(segment.url || '', props.embeddedImages)"
                                                :alt="segment.alt || 'image'"
                                                class="max-w-full rounded-lg border border-border bg-background object-contain"
                                                loading="lazy"
                                            />
                                            <figcaption v-if="segment.alt" class="text-[11px] text-muted-foreground">
                                                {{ segment.alt }}
                                            </figcaption>
                                        </figure>
                                    </template>
                                </div>
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

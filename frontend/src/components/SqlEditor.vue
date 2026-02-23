<template>
    <div ref="editorContainer" class="w-full h-full min-h-[100px] border border-input rounded-md overflow-hidden"></div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount, watch, toRaw, nextTick } from 'vue';
import { useResizeObserver } from '@vueuse/core';
import * as monaco from 'monaco-editor';
import editorWorker from 'monaco-editor/esm/vs/editor/editor.worker?worker';
import jsonWorker from 'monaco-editor/esm/vs/language/json/json.worker?worker';
import cssWorker from 'monaco-editor/esm/vs/language/css/css.worker?worker';
import htmlWorker from 'monaco-editor/esm/vs/language/html/html.worker?worker';
import tsWorker from 'monaco-editor/esm/vs/language/typescript/ts.worker?worker';

// Monaco Environment Setup for Vite
self.MonacoEnvironment = {
    getWorker(_, label) {
        if (label === 'json') {
            return new jsonWorker();
        }
        if (label === 'css' || label === 'scss' || label === 'less') {
            return new cssWorker();
        }
        if (label === 'html' || label === 'handlebars' || label === 'razor') {
            return new htmlWorker();
        }
        if (label === 'typescript' || label === 'javascript') {
            return new tsWorker();
        }
        return new editorWorker();
    }
};

const props = defineProps<{
    modelValue: string;
    tables?: string[];
    readOnly?: boolean;
    theme?: string; // 'vs', 'vs-dark', 'hc-black'
    fontFamily?: string;
    fontSize?: number;
    getColumns?: (tableName: string) => Promise<string[]>;
}>();

const emit = defineEmits(['update:modelValue', 'change']);

const editorContainer = ref<HTMLElement | null>(null);
let editor: monaco.editor.IStandaloneCodeEditor | null = null;
let completionDisposable: monaco.IDisposable | null = null;
let themeObserver: MutationObserver | null = null;

// Detect current theme from document element
const detectTheme = (): string => {
    return document.documentElement.classList.contains('dark') ? 'vault-dark' : 'vs';
};

// Initialize Editor
onMounted(() => {
    if (editorContainer.value) {
        // Define custom Monaco theme to match QuraMate dark mode (v2)
        monaco.editor.defineTheme('vault-dark', {
            base: 'vs-dark',
            inherit: true,
            rules: [],
            colors: {
                'editor.background': '#161b22', // Matches GitHub Dark style background for cards
                'editor.lineHighlightBackground': '#1f242c', // Matches accent hover
            }
        });

        const currentTheme = props.theme || detectTheme();

        editor = monaco.editor.create(editorContainer.value, {
            value: props.modelValue,
            language: 'sql',
            theme: currentTheme,
            readOnly: props.readOnly || false,
            automaticLayout: false, // Using useResizeObserver instead
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            fontSize: props.fontSize || 14,
            fontFamily: props.fontFamily || 'Consolas, "Courier New", monospace',
            padding: { top: 10, bottom: 10 },
            lineNumbers: 'on',
            renderLineHighlight: 'all',
        });

        // Use useResizeObserver for better performance than automaticLayout
        useResizeObserver(editorContainer, () => {
            editor?.layout();
        });

        // Handle content changes
        editor.onDidChangeModelContent(() => {
            const value = editor?.getValue() || '';
            emit('update:modelValue', value);
            emit('change', value);
        });

        // Register Completion Item Provider
        registerCompletionProvider();

        // Observe theme changes on <html> element
        themeObserver = new MutationObserver((mutations) => {
            for (const mutation of mutations) {
                if (mutation.type === 'attributes' && mutation.attributeName === 'class') {
                    const newTheme = detectTheme();
                    monaco.editor.setTheme(newTheme);
                }
            }
        });
        themeObserver.observe(document.documentElement, {
            attributes: true,
            attributeFilter: ['class'],
        });
    }
});

// Watch for modelValue changes from parent (e.g. loading a query)
watch(() => props.modelValue, (newValue) => {
    if (editor && newValue !== editor.getValue()) {
        editor.setValue(newValue);
    }
});

// No need to watch tables for re-registration anymore, the provider closure captures latest props.tables
// The completion provider will automatically use the latest `props.tables` and `props.getColumns` due to closure.

// Watch for explicit theme prop changes
watch(() => props.theme, (newTheme) => {
    if (editor && newTheme) {
        monaco.editor.setTheme(newTheme);
    }
});

// Watch for font settings changes
watch(() => [props.fontFamily, props.fontSize], ([newFontFamily, newFontSize]) => {
    if (editor) {
        editor.updateOptions({
            fontFamily: newFontFamily as string || 'Consolas, "Courier New", monospace',
            fontSize: newFontSize as number || 14
        });
    }
});

const registerCompletionProvider = () => {
    // Only register once
    if (completionDisposable) return;

    completionDisposable = monaco.languages.registerCompletionItemProvider('sql', {
        provideCompletionItems: (model, position) => {
            // Ensure this provider only acts for THIS editor instance
            if (editor?.getModel()?.id !== model.id) return { suggestions: [] };

            const word = model.getWordUntilPosition(position);
            const range = {
                startLineNumber: position.lineNumber,
                endLineNumber: position.lineNumber,
                startColumn: word.startColumn,
                endColumn: word.endColumn,
            };

            const suggestions: monaco.languages.CompletionItem[] = [];

            // SQL Keywords
            const keywords = [
                'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'ORDER BY', 'GROUP BY', 'LIMIT',
                'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'DROP', 'ALTER', 'TABLE', 'VIEW',
                'JOIN', 'LEFT JOIN', 'RIGHT JOIN', 'INNER JOIN', 'OUTER JOIN', 'ON',
                'AS', 'DISTINCT', 'COUNT', 'SUM', 'AVG', 'MAX', 'MIN', 'HAVING',
                'TOP', 'UNION', 'ALL', 'EXISTS', 'IN', 'LIKE', 'BETWEEN', 'NULL', 'IS',
                'PROCEDURE', 'FUNCTION', 'INDEX', 'TRIGGER', 'DATABASE', 'SCHEMA'
            ];

            keywords.forEach(keyword => {
                suggestions.push({
                    label: keyword,
                    kind: monaco.languages.CompletionItemKind.Keyword,
                    insertText: keyword,
                    range: range,
                    sortText: '1_keywords'
                });
            });

            // Table Names (uses current value of props.tables)
            if (props.tables) {
                props.tables.forEach(table => {
                    suggestions.push({
                        label: table,
                        kind: monaco.languages.CompletionItemKind.Class,
                        insertText: table,
                        range: range,
                        detail: 'Table',
                        sortText: '0_tables'
                    });
                });
            }

            const fullText = model.getValue();
            const offset = model.getOffsetAt(position);
            const lastSemicolonIndex = fullText.lastIndexOf(';', offset - 1);
            const startOfStatement = lastSemicolonIndex !== -1 ? lastSemicolonIndex + 1 : 0;

            let endOfStatement = fullText.indexOf(';', offset);
            if (endOfStatement === -1) endOfStatement = fullText.length;

            const statement = fullText.substring(startOfStatement, endOfStatement);
            const relativeOffset = offset - startOfStatement;

            const fromMatch = statement.match(/FROM\s+([a-zA-Z0-9_\[\]\.]+)/i);
            const table = fromMatch ? fromMatch[1] : null;

            let suggestColumns = false;

            if (table) {
                const fromIndex = fromMatch!.index!;
                const whereMatch = statement.match(/WHERE\s+/i);
                if (whereMatch) {
                    const whereIndex = whereMatch.index!;
                    if (relativeOffset > whereIndex + whereMatch[0].length) {
                        suggestColumns = true;
                    }
                }

                if (!suggestColumns && relativeOffset < fromIndex) {
                    const selectMatch = statement.match(/SELECT\b/i);
                    if (selectMatch) {
                        const selectEnd = selectMatch.index! + selectMatch[0].length;
                        if (relativeOffset >= selectEnd) {
                            suggestColumns = true;
                        }
                    }
                }
            }

            if (table && props.getColumns && suggestColumns) {
                const cleanTable = table.replace(/[\[\]]/g, '');
                return props.getColumns(cleanTable).then(columns => {
                    const columnSuggestions = columns.map(col => ({
                        label: col,
                        kind: monaco.languages.CompletionItemKind.Field,
                        insertText: col,
                        range: range,
                        detail: `Column (${cleanTable})`,
                        sortText: '0_columns'
                    }));
                    return { suggestions: [...suggestions, ...columnSuggestions] };
                });
            }

            return { suggestions };
        }
    });
};

const getSelection = () => {
    if (editor) {
        const selection = editor.getSelection();
        if (selection && !selection.isEmpty()) {
            return editor.getModel()?.getValueInRange(selection) || '';
        }
    }
    return '';
};

defineExpose({
    getSelection
});

onBeforeUnmount(() => {
    if (themeObserver) {
        themeObserver.disconnect();
    }
    if (editor) {
        editor.dispose();
    }
    if (completionDisposable) {
        completionDisposable.dispose();
    }
});
</script>

<style scoped>
/* Optional: specific overrides if needed */
</style>

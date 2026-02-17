<template>
    <div ref="editorContainer" class="w-full h-full min-h-[100px] border border-input rounded-md overflow-hidden"></div>
</template>

<script lang="ts" setup>
import { ref, onMounted, onBeforeUnmount, watch, toRaw } from 'vue';
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
}>();

const emit = defineEmits(['update:modelValue', 'change']);

const editorContainer = ref<HTMLElement | null>(null);
let editor: monaco.editor.IStandaloneCodeEditor | null = null;
let completionDisposable: monaco.IDisposable | null = null;
let themeObserver: MutationObserver | null = null;

// Detect current theme from document element
const detectTheme = (): string => {
    return document.documentElement.classList.contains('dark') ? 'vs-dark' : 'vs';
};

// Initialize Editor
onMounted(() => {
    if (editorContainer.value) {
        const currentTheme = props.theme || detectTheme();

        editor = monaco.editor.create(editorContainer.value, {
            value: props.modelValue,
            language: 'sql',
            theme: currentTheme,
            readOnly: props.readOnly || false,
            automaticLayout: true,
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            fontSize: 14,
            fontFamily: 'Consolas, "Courier New", monospace',
            padding: { top: 10, bottom: 10 },
            lineNumbers: 'on',
            renderLineHighlight: 'all',
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

// Watch for tables changes to update autocomplete
watch(() => props.tables, () => {
    registerCompletionProvider();
}, { deep: true });

// Watch for explicit theme prop changes
watch(() => props.theme, (newTheme) => {
    if (editor && newTheme) {
        monaco.editor.setTheme(newTheme);
    }
});

const registerCompletionProvider = () => {
    // Dispose previous provider if exists
    if (completionDisposable) {
        completionDisposable.dispose();
    }

    completionDisposable = monaco.languages.registerCompletionItemProvider('sql', {
        provideCompletionItems: (model, position) => {
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
                'SELECT', 'FROM', 'WHERE', 'AND', 'OR', 'ORDER BY', 'GROUP BY', 'limit',
                'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'DROP', 'ALTER', 'TABLE', 'view',
                'JOIN', 'LEFT JOIN', 'RIGHT JOIN', 'INNER JOIN', 'OUTER JOIN', 'ON',
                'AS', 'DISTINCT', 'COUNT', 'SUM', 'AVG', 'MAX', 'MIN', 'HAVING',
                'TOP', 'UNION', 'ALL', 'EXISTS', 'IN', 'LIKE', 'BETWEEN', 'NULL', 'IS'
            ];

            keywords.forEach(keyword => {
                suggestions.push({
                    label: keyword,
                    kind: monaco.languages.CompletionItemKind.Keyword,
                    insertText: keyword,
                    range: range,
                    sortText: '1_keywords' // Lower priority
                });
            });

            // Table Names
            if (props.tables) {
                props.tables.forEach(table => {
                    suggestions.push({
                        label: table,
                        kind: monaco.languages.CompletionItemKind.Class, // Use Class icon for tables
                        insertText: table,
                        range: range,
                        detail: 'Table',
                        sortText: '0_tables' // Higher priority
                    });
                });
            }

            return { suggestions };
        }
    });
};

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

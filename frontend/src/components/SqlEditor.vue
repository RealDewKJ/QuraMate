<template>
    <div ref="editorContainer" class="sql-editor-container relative z-30 w-full h-full min-h-[100px] overflow-hidden rounded-md"></div>
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
                'editor.background': '#0a0a0a', // Matches project background
                'editor.lineHighlightBackground': '#1a1a1a', // Subtle highlight
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
            fixedOverflowWidgets: true,
            quickSuggestions: true,
            suggestOnTriggerCharacters: true,
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

            const statementBeforeCursor = statement.substring(0, relativeOffset);

            const normalizeIdentifier = (identifier: string): string => {
                const trimmed = identifier.trim();
                if (!trimmed) return '';

                const cleaned = trimmed
                    .replace(/^\[|\]$/g, '')
                    .replace(/^"|"$/g, '')
                    .replace(/^`|`$/g, '');

                const parts = cleaned
                    .split('.')
                    .map(part => part.replace(/^\[|\]$/g, '').replace(/^"|"$/g, '').replace(/^`|`$/g, '').trim())
                    .filter(Boolean);

                if (parts.length === 0) return '';
                return parts[parts.length - 1];
            };

            const parseTableAliases = (sql: string): Map<string, string> => {
                const aliasMap = new Map<string, string>();
                const sourceRegex = /\b(?:FROM|JOIN|LEFT\s+JOIN|RIGHT\s+JOIN|INNER\s+JOIN|FULL\s+JOIN|CROSS\s+JOIN|OUTER\s+JOIN|LEFT\s+OUTER\s+JOIN|RIGHT\s+OUTER\s+JOIN)\s+([a-zA-Z0-9_\.\[\]`"]+)(?:\s+(?:AS\s+)?([a-zA-Z0-9_\[\]`"]+))?/gi;
                const blocked = new Set(['ON', 'WHERE', 'GROUP', 'ORDER', 'HAVING', 'LIMIT', 'JOIN', 'LEFT', 'RIGHT', 'INNER', 'OUTER', 'FULL', 'CROSS']);
                let match: RegExpExecArray | null;

                while ((match = sourceRegex.exec(sql)) !== null) {
                    const tableName = normalizeIdentifier(match[1] || '');
                    if (!tableName) continue;

                    aliasMap.set(tableName.toLowerCase(), tableName);

                    const alias = normalizeIdentifier(match[2] || '');
                    if (alias && !blocked.has(alias.toUpperCase())) {
                        aliasMap.set(alias.toLowerCase(), tableName);
                    }
                }

                return aliasMap;
            };

            const fromMatch = statement.match(/FROM\s+([a-zA-Z0-9_\[\]\.]+)/i);
            const mainTable = fromMatch ? normalizeIdentifier(fromMatch[1]) : '';
            const aliasMatch = statementBeforeCursor.match(/([a-zA-Z0-9_\[\]`"]+)\.\s*([a-zA-Z0-9_]*)$/);
            const qualifier = aliasMatch ? normalizeIdentifier(aliasMatch[1]) : '';
            const aliases = parseTableAliases(statement);

            let suggestColumns = false;
            let targetTable = '';

            if (qualifier) {
                const resolvedTable = aliases.get(qualifier.toLowerCase());
                if (resolvedTable) {
                    suggestColumns = true;
                    targetTable = resolvedTable;
                }
            }

            if (!targetTable && mainTable && fromMatch) {
                const fromIndex = fromMatch.index!;
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

                if (suggestColumns) {
                    targetTable = mainTable;
                }
            }

            if (targetTable && props.getColumns && suggestColumns) {
                return props.getColumns(targetTable).then(columns => {
                    const columnSuggestions = columns.map(col => ({
                        label: col,
                        kind: monaco.languages.CompletionItemKind.Field,
                        insertText: col,
                        range: range,
                        detail: `Column (${targetTable})`,
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
.sql-editor-container :deep(.monaco-editor),
.sql-editor-container :deep(.overflow-guard),
.sql-editor-container :deep(.monaco-scrollable-element) {
    overflow: hidden !important;
}

.sql-editor-container :deep(.suggest-widget),
.sql-editor-container :deep(.monaco-editor-hover),
.sql-editor-container :deep(.parameter-hints-widget) {
    z-index: 9999 !important;
}
</style>

<template>
    <div
        class="er-diagram-container flex flex-col h-full bg-white dark:bg-slate-900 border border-border rounded-lg relative overflow-hidden">
        <div v-if="error" class="bg-destructive/10 border border-destructive/20 text-destructive p-4 rounded-lg m-4">
            <h3 class="font-semibold flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-alert-circle">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="12" x2="12" y1="8" y2="12" />
                    <line x1="12" x2="12.01" y1="16" y2="16" />
                </svg>
                Error generating diagram
            </h3>
            <p class="text-sm font-mono mt-1 ml-6">{{ error }}</p>
        </div>

        <div v-else-if="isLoading"
            class="absolute inset-0 z-50 flex flex-col items-center justify-center bg-background/80 backdrop-blur-sm gap-3">
            <svg class="animate-spin h-8 w-8 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none"
                viewBox="0 0 24 24">
                <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                <path class="opacity-75" fill="currentColor"
                    d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                </path>
            </svg>
            <span class="text-sm font-medium text-muted-foreground animate-pulse">Analyzing Schema...</span>
        </div>

        <div v-show="!isLoading && !error"
            class="flex-1 overflow-auto p-8 flex items-center justify-center bg-dot-pattern" ref="containerRef">
            <div class="mermaid-svg-wrapper shadow-lg bg-card rounded-lg p-4 border border-border">
                <div ref="mermaidContainer" class="mermaid"></div>
            </div>
        </div>

        <div
            class="px-4 py-2 bg-muted/80 backdrop-blur border-t border-border text-xs text-muted-foreground flex justify-between select-none">
            <div class="flex items-center gap-2">
                <span class="font-semibold text-foreground">{{ tableName }}</span>
                <span class="w-px h-3 bg-border"></span>
                <span>{{ columns.length }} Columns</span>
            </div>
            <span>{{ relationships.length }} Relationships</span>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, onMounted, watch, nextTick } from 'vue';
import mermaid from 'mermaid';

const props = defineProps<{
    tableName: string;
    columns: any[]; // Legacy support (single table)
    relationships: any[]; // ForeignKey[]
    tablesData?: Record<string, any[]>; // Multi-table support
    isDark?: boolean;
}>();

const containerRef = ref<HTMLElement | null>(null);
const mermaidContainer = ref<HTMLElement | null>(null);
const error = ref('');
const isLoading = ref(true);

const initMermaid = () => {
    mermaid.initialize({
        startOnLoad: false,
        theme: props.isDark ? 'dark' : 'default',
        securityLevel: 'loose',
        er: {
            useMaxWidth: false,
            layoutDirection: 'TB',
            diagramPadding: 20
        },
        fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace'
    });
};

const generateDiagram = async () => {
    if (!props.tableName) return;

    isLoading.value = true;
    error.value = '';

    // Wait a bit for UI to settle
    await new Promise(r => setTimeout(r, 100));

    try {
        initMermaid();

        // 1. Build Mermaid ER Syntax
        let graphDefinition = `erDiagram\n`;

        // Function to render a table block
        const renderTable = (name: string, cols: any[]) => {
            const safeName = name.replace(/[^a-zA-Z0-9_]/g, '_');
            let block = `    ${safeName} {\n`;
            cols.forEach(col => {
                const colName = typeof col === 'string' ? col : (col.name || col.column_name || 'unknown');
                const colType = typeof col === 'object' ? (col.type || col.data_type || 'string') : 'string';
                const safeCol = colName.replace(/[^a-zA-Z0-9_]/g, '_');
                block += `        ${colType.replace(/\s+/g, '_')} ${safeCol}\n`;
            });
            block += `    }\n`;
            return block;
        };

        const renderedTables = new Set<string>();

        // Render from tablesData (Multi-table mode)
        if (props.tablesData && Object.keys(props.tablesData).length > 0) {
            for (const [tblName, cols] of Object.entries(props.tablesData)) {
                graphDefinition += renderTable(tblName, cols);
                renderedTables.add(tblName);
            }
        }
        // Fallback to legacy single table mode
        else if (props.columns.length > 0) {
            graphDefinition += renderTable(props.tableName, props.columns);
            renderedTables.add(props.tableName);
        }

        // Define relationships
        const processed = new Set<string>();

        props.relationships.forEach(fk => {
            const childTable = fk.table;
            const parentTable = fk.refTable;
            const safeChild = childTable.replace(/[^a-zA-Z0-9_]/g, '_');
            const safeParent = parentTable.replace(/[^a-zA-Z0-9_]/g, '_');
            const label = fk.column === fk.refColumn ? `${fk.column}` : `${fk.column}->${fk.refColumn}`;

            // Only draw if we haven't processed this exact relationship
            const relString = `${safeParent} ||--o{ ${safeChild} : "${label}"`;

            // If we are in single-table mode, we might want to show the relationship even if the other table isn't fully defined.
            // But usually Mermaid wants the entity to exist. 
            // If we have tablesData, we assume all relevant tables are there.

            if (!processed.has(relString)) {
                graphDefinition += `    ${relString}\n`;
                processed.add(relString);
            }
        });

        // 2. Render
        if (mermaidContainer.value) {
            mermaidContainer.value.innerHTML = '';
            const { svg } = await mermaid.render(`mermaid-${Date.now()}`, graphDefinition);
            mermaidContainer.value.innerHTML = svg;
        }

    } catch (e: any) {
        console.error('Mermaid error:', e);
        error.value = `Failed to render diagram. ${e.message}`;
    } finally {
        isLoading.value = false;
    }
};

onMounted(() => {
    generateDiagram();
});

watch(() => [props.tableName, props.relationships, props.tablesData], () => {
    generateDiagram();
});
</script>

<style scoped>
.bg-dot-pattern {
    background-image: radial-gradient(#e5e7eb 1px, transparent 1px);
    background-size: 20px 20px;
}

.dark .bg-dot-pattern {
    background-image: radial-gradient(#1f2937 1px, transparent 1px);
}
</style>

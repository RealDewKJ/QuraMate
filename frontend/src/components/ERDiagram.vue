<template>
    <div class="schema-visualizer h-full flex flex-col border border-border rounded-lg overflow-hidden"
        :class="{ 'schema-visualizer--dark': isDark }">
        <div class="toolbar border-b border-border px-3 py-2 flex items-center gap-2">
            <input v-model="search" type="text" placeholder="Search table..."
                class="h-8 w-56 rounded-md border border-border bg-background px-2 text-xs outline-none focus:ring-2 focus:ring-primary/40" />
            <button class="tool-btn" @click="zoomIn">+</button>
            <button class="tool-btn" @click="zoomOut">-</button>
            <button class="tool-btn" @click="fitToView">Fit</button>
            <button class="tool-btn" :class="{ 'tool-btn--active': focusMode }" :disabled="!selectedTableName"
                @click="toggleFocusMode">
                Focus
            </button>
            <button class="tool-btn" :disabled="!selectedTableName" @click="clearSelection">Clear</button>
            <button class="tool-btn" :disabled="!hasSavedLayout" @click="resetLayout">Reset Layout</button>
            <div class="text-[11px] text-muted-foreground ml-auto">
                {{ visibleTables.length }} tables • {{ relationships.length }} links
            </div>
        </div>

        <div v-if="error" class="m-4 rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
            {{ error }}
        </div>

        <div v-else ref="viewportRef" class="viewport flex-1 relative overflow-hidden" @mousedown="startPan" @wheel.prevent="onWheel">
            <svg class="links-layer absolute inset-0" aria-hidden="true">
                <g :transform="`translate(${pan.x}, ${pan.y}) scale(${zoom})`">
                    <g v-for="(link, index) in renderedLinks" :key="`link-${index}`"
                        :class="['link-group', { 'link-group--active': link.active, 'link-group--dimmed': !link.active && hasSelection }]">
                        <path :d="link.path" class="link-path" />
                        <circle :cx="link.start.x" :cy="link.start.y" r="3" class="link-endpoint link-endpoint--fk" />
                        <circle :cx="link.end.x" :cy="link.end.y" r="3" class="link-endpoint link-endpoint--pk" />
                    </g>
                </g>
            </svg>

            <div class="canvas absolute inset-0" :style="canvasStyle">
                <article v-for="table in visibleTables" :key="table.name"
                    :class="['table-card', { 'table-card--selected': isSelected(table.name), 'table-card--dimmed': shouldDimTable(table.name) }]"
                    :style="tableStyle(table)" @click.stop="selectTable(table.name)">
                    <header class="table-header" @mousedown.stop="startTableDrag(table, $event)">
                        <span class="table-name">{{ table.name }}</span>
                        <span class="table-count">{{ table.columns.length }}</span>
                    </header>
                    <ul class="column-list">
                        <li v-for="column in table.columns" :key="`${table.name}-${column.name}`" class="column-row">
                            <span class="column-tags">
                                <span v-if="isPrimaryKey(table.name, column.name)" class="pk">PK</span>
                                <span v-if="isForeignKey(table.name, column.name)" class="fk">FK</span>
                            </span>
                            <span class="column-name">{{ column.name }}</span>
                            <span class="column-type">{{ compactType(column.type) }}</span>
                        </li>
                    </ul>
                </article>
            </div>

            <aside class="minimap" @mousedown.prevent="onMinimapPointer">
                <svg class="minimap-svg" viewBox="0 0 200 130" preserveAspectRatio="none">
                    <rect x="0" y="0" width="200" height="130" class="minimap-bg" rx="8" />
                    <rect v-for="node in miniNodes" :key="`mini-${node.name}`" :x="node.x" :y="node.y" :width="node.w"
                        :height="node.h"
                        :class="['mini-node', { 'mini-node--selected': isSelected(node.name), 'mini-node--dimmed': shouldDimTable(node.name) } ]"
                        rx="2" />
                    <rect v-if="miniViewport" :x="miniViewport.x" :y="miniViewport.y" :width="miniViewport.w"
                        :height="miniViewport.h" class="mini-viewport" rx="3" />
                </svg>
            </aside>
        </div>
    </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, onUnmounted, reactive, ref, watch } from 'vue';

type SchemaColumn = { name: string; type: string };
type ForeignKey = { table: string; column: string; refTable: string; refColumn: string };
type TableNode = { name: string; columns: SchemaColumn[]; x: number; y: number; width: number; height: number };

type LinkPoint = { x: number; y: number };
type RenderedLink = { path: string; start: LinkPoint; end: LinkPoint; active: boolean };

type Bounds = { minX: number; minY: number; maxX: number; maxY: number; width: number; height: number };

const props = defineProps<{
    tableName: string;
    columns: any[];
    relationships: ForeignKey[];
    tablesData?: Record<string, any[]>;
    isDark?: boolean;
}>();

const viewportRef = ref<HTMLElement | null>(null);
const error = ref('');
const search = ref('');
const zoom = ref(1);
const pan = reactive({ x: 40, y: 40 });
const isDragging = ref(false);
const selectedTableName = ref('');
const focusMode = ref(false);
const manualPositions = ref<Record<string, { x: number; y: number }>>({});
const hasSavedLayout = ref(false);

let dragStartX = 0;
let dragStartY = 0;
let startPanX = 0;
let startPanY = 0;
let dragTableName = '';
let tableDragStartX = 0;
let tableDragStartY = 0;
let tableStartX = 0;
let tableStartY = 0;
let isTableDragging = false;

const TABLE_WIDTH = 300;
const TABLE_HEADER_HEIGHT = 38;
const TABLE_ROW_HEIGHT = 26;
const TABLE_GAP_X = 120;
const TABLE_GAP_Y = 70;

const normalizedTables = computed<TableNode[]>(() => {
    const rows: Array<[string, any[]]> =
        props.tablesData && Object.keys(props.tablesData).length > 0
            ? Object.entries(props.tablesData)
            : (props.tableName ? [[props.tableName, props.columns || []]] : []);

    const toColumn = (raw: any): SchemaColumn => ({
        name: (raw?.name || raw?.column_name || raw?.COLUMN_NAME || 'unknown').toString(),
        type: (raw?.type || raw?.data_type || raw?.DATA_TYPE || raw?.Type || 'text').toString(),
    });

    return rows.map(([name, cols]) => {
        const columns = (cols || []).map(toColumn);
        const height = TABLE_HEADER_HEIGHT + (Math.max(columns.length, 1) * TABLE_ROW_HEIGHT) + 10;
        return { name, columns, x: 0, y: 0, width: TABLE_WIDTH, height };
    });
});

const relationSetByTable = computed(() => {
    const map = new Map<string, Set<string>>();
    normalizedTables.value.forEach((node) => map.set(node.name, new Set()));

    props.relationships.forEach((rel) => {
        if (!map.has(rel.table) || !map.has(rel.refTable)) return;
        map.get(rel.table)?.add(rel.refTable);
        map.get(rel.refTable)?.add(rel.table);
    });

    return map;
});

const layoutStorageKey = computed(() => {
    const names = normalizedTables.value.map((n) => n.name).sort().join('|');
    return `quramate:er-layout:v1:${props.tableName}:${names}`;
});

const autoPositionedTables = computed<TableNode[]>(() => {
    const nodes = normalizedTables.value.map((n) => ({ ...n }));
    const byName = new Map(nodes.map((n) => [n.name, n]));
    if (nodes.length === 0) return nodes;

    const root = byName.has(props.tableName) ? props.tableName : nodes[0].name;
    const visited = new Set<string>();
    const levelMap = new Map<string, number>();
    const queue: string[] = [root];
    visited.add(root);
    levelMap.set(root, 0);

    while (queue.length) {
        const current = queue.shift() as string;
        const level = levelMap.get(current) || 0;
        relationSetByTable.value.get(current)?.forEach((next) => {
            if (visited.has(next)) return;
            visited.add(next);
            levelMap.set(next, level + 1);
            queue.push(next);
        });
    }

    let maxLevel = 0;
    nodes.forEach((n) => {
        if (!levelMap.has(n.name)) {
            maxLevel += 1;
            levelMap.set(n.name, maxLevel + 1);
        } else {
            maxLevel = Math.max(maxLevel, levelMap.get(n.name) || 0);
        }
    });

    const levelBuckets = new Map<number, TableNode[]>();
    nodes.forEach((node) => {
        const level = levelMap.get(node.name) || 0;
        if (!levelBuckets.has(level)) levelBuckets.set(level, []);
        levelBuckets.get(level)?.push(node);
    });

    for (const [level, bucket] of levelBuckets.entries()) {
        bucket.sort((a, b) => a.name.localeCompare(b.name));
        let currentY = 0;
        bucket.forEach((node) => {
            node.x = level * (TABLE_WIDTH + TABLE_GAP_X);
            node.y = currentY;
            currentY += node.height + TABLE_GAP_Y;
        });
    }

    return nodes;
});

const positionedTables = computed<TableNode[]>(() => {
    return autoPositionedTables.value.map((table) => {
        const override = manualPositions.value[table.name];
        if (!override) return table;
        return { ...table, x: override.x, y: override.y };
    });
});

const focusSet = computed(() => {
    if (!selectedTableName.value) return new Set<string>();

    const set = new Set<string>([selectedTableName.value]);
    relationSetByTable.value.get(selectedTableName.value)?.forEach((related) => set.add(related));
    return set;
});

const filteredBySearch = computed(() => {
    const q = search.value.trim().toLowerCase();
    if (!q) return positionedTables.value;
    return positionedTables.value.filter((t) => t.name.toLowerCase().includes(q));
});

const visibleTables = computed(() => {
    if (!focusMode.value || focusSet.value.size === 0) return filteredBySearch.value;
    return filteredBySearch.value.filter((t) => focusSet.value.has(t.name));
});

const hasSelection = computed(() => !!selectedTableName.value);

const renderedLinks = computed<RenderedLink[]>(() => {
    const lookup = new Map(visibleTables.value.map((t) => [t.name, t]));
    return props.relationships
        .filter((rel) => lookup.has(rel.table) && lookup.has(rel.refTable))
        .map((rel) => {
            const child = lookup.get(rel.table) as TableNode;
            const parent = lookup.get(rel.refTable) as TableNode;

            const childColIdx = Math.max(child.columns.findIndex((c) => c.name === rel.column), 0);
            const parentColIdx = Math.max(parent.columns.findIndex((c) => c.name === rel.refColumn), 0);

            const fromLeft = parent.x < child.x;
            const sx = child.x + (fromLeft ? 0 : child.width);
            const sy = child.y + TABLE_HEADER_HEIGHT + (childColIdx + 0.5) * TABLE_ROW_HEIGHT;
            const tx = parent.x + (fromLeft ? parent.width : 0);
            const ty = parent.y + TABLE_HEADER_HEIGHT + (parentColIdx + 0.5) * TABLE_ROW_HEIGHT;

            const bend = Math.max(Math.abs(tx - sx) * 0.45, 60);
            const c1x = sx + (fromLeft ? -bend : bend);
            const c2x = tx + (fromLeft ? bend : -bend);

            const active = !selectedTableName.value
                || rel.table === selectedTableName.value
                || rel.refTable === selectedTableName.value;

            return {
                path: `M ${sx} ${sy} C ${c1x} ${sy}, ${c2x} ${ty}, ${tx} ${ty}`,
                start: { x: sx, y: sy },
                end: { x: tx, y: ty },
                active,
            };
        });
});

const canvasStyle = computed(() => ({
    transform: `translate(${pan.x}px, ${pan.y}px) scale(${zoom.value})`,
    transformOrigin: '0 0',
}));

const contentBounds = computed<Bounds | null>(() => {
    if (visibleTables.value.length === 0) return null;

    const minX = Math.min(...visibleTables.value.map((t) => t.x));
    const minY = Math.min(...visibleTables.value.map((t) => t.y));
    const maxX = Math.max(...visibleTables.value.map((t) => t.x + t.width));
    const maxY = Math.max(...visibleTables.value.map((t) => t.y + t.height));

    return {
        minX,
        minY,
        maxX,
        maxY,
        width: Math.max(maxX - minX, 1),
        height: Math.max(maxY - minY, 1),
    };
});

const miniScale = computed(() => {
    if (!contentBounds.value) return { sx: 1, sy: 1 };
    return {
        sx: 200 / contentBounds.value.width,
        sy: 130 / contentBounds.value.height,
    };
});

const miniNodes = computed(() => {
    if (!contentBounds.value) return [];
    return visibleTables.value.map((table) => ({
        name: table.name,
        x: (table.x - contentBounds.value!.minX) * miniScale.value.sx,
        y: (table.y - contentBounds.value!.minY) * miniScale.value.sy,
        w: Math.max(table.width * miniScale.value.sx, 2),
        h: Math.max(table.height * miniScale.value.sy, 2),
    }));
});

const miniViewport = computed(() => {
    const viewport = viewportRef.value;
    if (!viewport || !contentBounds.value) return null;

    const worldLeft = (-pan.x) / zoom.value;
    const worldTop = (-pan.y) / zoom.value;
    const worldWidth = viewport.clientWidth / zoom.value;
    const worldHeight = viewport.clientHeight / zoom.value;

    const x = (worldLeft - contentBounds.value.minX) * miniScale.value.sx;
    const y = (worldTop - contentBounds.value.minY) * miniScale.value.sy;
    const w = worldWidth * miniScale.value.sx;
    const h = worldHeight * miniScale.value.sy;

    return {
        x: Math.max(0, Math.min(200, x)),
        y: Math.max(0, Math.min(130, y)),
        w: Math.max(8, Math.min(200, w)),
        h: Math.max(8, Math.min(130, h)),
    };
});

const tableStyle = (table: TableNode) => ({
    left: `${table.x}px`,
    top: `${table.y}px`,
    width: `${table.width}px`,
});

const compactType = (value: string) => value.replace(/\s+/g, ' ').trim().toLowerCase();

const isPrimaryKey = (tableName: string, columnName: string) => {
    return props.relationships.some((rel) => rel.refTable === tableName && rel.refColumn === columnName);
};

const isForeignKey = (tableName: string, columnName: string) => {
    return props.relationships.some((rel) => rel.table === tableName && rel.column === columnName);
};

const isSelected = (tableName: string) => selectedTableName.value === tableName;

const isRelatedToSelected = (tableName: string) => {
    if (!selectedTableName.value) return false;
    return relationSetByTable.value.get(selectedTableName.value)?.has(tableName) || false;
};

const shouldDimTable = (tableName: string) => {
    if (!selectedTableName.value) return false;
    if (tableName === selectedTableName.value) return false;
    return !isRelatedToSelected(tableName);
};

const selectTable = (tableName: string) => {
    selectedTableName.value = tableName;
};

const clearSelection = () => {
    selectedTableName.value = '';
    focusMode.value = false;
};

const toggleFocusMode = () => {
    if (!selectedTableName.value) return;
    focusMode.value = !focusMode.value;
};

const loadLayout = () => {
    if (typeof window === 'undefined') return;
    try {
        const raw = window.localStorage.getItem(layoutStorageKey.value);
        if (!raw) {
            manualPositions.value = {};
            hasSavedLayout.value = false;
            return;
        }

        const parsed = JSON.parse(raw) as Record<string, { x: number; y: number }>;
        if (!parsed || typeof parsed !== 'object') {
            manualPositions.value = {};
            hasSavedLayout.value = false;
            return;
        }

        const validEntries = Object.entries(parsed).filter(([name, point]) => {
            return !!name && typeof point?.x === 'number' && typeof point?.y === 'number';
        });

        manualPositions.value = Object.fromEntries(validEntries);
        hasSavedLayout.value = validEntries.length > 0;
    } catch {
        manualPositions.value = {};
        hasSavedLayout.value = false;
    }
};

const persistLayout = () => {
    if (typeof window === 'undefined') return;

    const entries = Object.entries(manualPositions.value);
    hasSavedLayout.value = entries.length > 0;

    if (entries.length === 0) {
        window.localStorage.removeItem(layoutStorageKey.value);
        return;
    }

    window.localStorage.setItem(layoutStorageKey.value, JSON.stringify(manualPositions.value));
};

const resetLayout = async () => {
    manualPositions.value = {};
    persistLayout();
    await fitToView();
};

const zoomAt = (clientX: number, clientY: number, nextZoom: number) => {
    const viewport = viewportRef.value;
    if (!viewport) return;

    const rect = viewport.getBoundingClientRect();
    const cursorX = clientX - rect.left;
    const cursorY = clientY - rect.top;

    const worldX = (cursorX - pan.x) / zoom.value;
    const worldY = (cursorY - pan.y) / zoom.value;

    zoom.value = Math.min(2.2, Math.max(0.35, nextZoom));
    pan.x = cursorX - worldX * zoom.value;
    pan.y = cursorY - worldY * zoom.value;
};

const zoomIn = () => {
    const rect = viewportRef.value?.getBoundingClientRect();
    if (!rect) return;
    zoomAt(rect.left + rect.width / 2, rect.top + rect.height / 2, zoom.value + 0.1);
};

const zoomOut = () => {
    const rect = viewportRef.value?.getBoundingClientRect();
    if (!rect) return;
    zoomAt(rect.left + rect.width / 2, rect.top + rect.height / 2, zoom.value - 0.1);
};

const onWheel = (event: WheelEvent) => {
    const delta = event.deltaY > 0 ? -0.08 : 0.08;
    zoomAt(event.clientX, event.clientY, zoom.value + delta);
};

const startPan = (event: MouseEvent) => {
    if ((event.target as HTMLElement).closest('.table-card, .tool-btn, input, .minimap')) return;
    isDragging.value = true;
    dragStartX = event.clientX;
    dragStartY = event.clientY;
    startPanX = pan.x;
    startPanY = pan.y;
};

const startTableDrag = (table: TableNode, event: MouseEvent) => {
    if (event.button !== 0) return;
    isTableDragging = true;
    dragTableName = table.name;
    tableDragStartX = event.clientX;
    tableDragStartY = event.clientY;
    tableStartX = table.x;
    tableStartY = table.y;
};

const onMouseMove = (event: MouseEvent) => {
    if (isTableDragging && dragTableName) {
        const nextX = tableStartX + (event.clientX - tableDragStartX) / zoom.value;
        const nextY = tableStartY + (event.clientY - tableDragStartY) / zoom.value;
        manualPositions.value = {
            ...manualPositions.value,
            [dragTableName]: { x: nextX, y: nextY },
        };
        return;
    }

    if (!isDragging.value) return;
    pan.x = startPanX + (event.clientX - dragStartX);
    pan.y = startPanY + (event.clientY - dragStartY);
};

const onMouseUp = () => {
    if (isTableDragging) {
        isTableDragging = false;
        dragTableName = '';
        persistLayout();
    }
    isDragging.value = false;
};

const centerWorldPoint = (worldX: number, worldY: number) => {
    const viewport = viewportRef.value;
    if (!viewport) return;

    pan.x = viewport.clientWidth / 2 - worldX * zoom.value;
    pan.y = viewport.clientHeight / 2 - worldY * zoom.value;
};

const onMinimapPointer = (event: MouseEvent) => {
    if (!contentBounds.value) return;
    const target = event.currentTarget as HTMLElement;
    const rect = target.getBoundingClientRect();

    const ratioX = (event.clientX - rect.left) / rect.width;
    const ratioY = (event.clientY - rect.top) / rect.height;

    const worldX = contentBounds.value.minX + ratioX * contentBounds.value.width;
    const worldY = contentBounds.value.minY + ratioY * contentBounds.value.height;

    centerWorldPoint(worldX, worldY);
};

const fitToView = async () => {
    await nextTick();
    const viewport = viewportRef.value;
    if (!viewport || !contentBounds.value) return;

    const margin = 80;
    const scaleX = (viewport.clientWidth - margin) / contentBounds.value.width;
    const scaleY = (viewport.clientHeight - margin) / contentBounds.value.height;
    zoom.value = Math.min(1.4, Math.max(0.35, Math.min(scaleX, scaleY)));

    pan.x = (viewport.clientWidth - contentBounds.value.width * zoom.value) / 2 - contentBounds.value.minX * zoom.value;
    pan.y = (viewport.clientHeight - contentBounds.value.height * zoom.value) / 2 - contentBounds.value.minY * zoom.value;
};

watch([() => props.tableName, () => props.tablesData, () => props.relationships], async () => {
    error.value = '';
    selectedTableName.value = props.tableName || '';
    focusMode.value = false;
    loadLayout();
    await fitToView();
}, { deep: true });

watch([search, focusMode], async () => {
    await fitToView();
});

watch(layoutStorageKey, () => {
    loadLayout();
});

onMounted(async () => {
    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);

    if (!props.tableName) {
        error.value = 'No table selected.';
        return;
    }

    selectedTableName.value = props.tableName;
    loadLayout();
    await fitToView();
});

onUnmounted(() => {
    window.removeEventListener('mousemove', onMouseMove);
    window.removeEventListener('mouseup', onMouseUp);
});
</script>

<style scoped>
.schema-visualizer {
    background: linear-gradient(180deg, #f8fafc 0%, #f1f5f9 100%);
}

.schema-visualizer--dark {
    background: linear-gradient(180deg, #0f172a 0%, #111827 100%);
}

.toolbar {
    backdrop-filter: blur(6px);
    background: color-mix(in srgb, var(--background) 86%, transparent);
}

.tool-btn {
    height: 30px;
    min-width: 30px;
    border-radius: 6px;
    border: 1px solid hsl(var(--border));
    background: hsl(var(--background));
    font-size: 12px;
    font-weight: 600;
    padding: 0 10px;
}

.tool-btn:hover {
    background: hsl(var(--accent));
}

.tool-btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
}

.tool-btn--active {
    border-color: color-mix(in srgb, hsl(var(--primary)) 65%, hsl(var(--border)));
    background: color-mix(in srgb, hsl(var(--primary)) 20%, hsl(var(--background)));
}

.viewport {
    cursor: grab;
}

.viewport:active {
    cursor: grabbing;
}

.links-layer {
    pointer-events: none;
}

.link-group {
    opacity: 0.9;
}

.link-group--active {
    opacity: 1;
}

.link-group--dimmed {
    opacity: 0.18;
}

.link-path {
    stroke: color-mix(in srgb, hsl(var(--primary)) 65%, transparent);
    stroke-width: 2;
    fill: none;
}

.link-endpoint {
    stroke-width: 1;
    stroke: color-mix(in srgb, hsl(var(--background)) 70%, transparent);
}

.link-endpoint--fk {
    fill: #2563eb;
}

.link-endpoint--pk {
    fill: #16a34a;
}

.canvas {
    position: absolute;
    inset: 0;
    will-change: transform;
}

.table-card {
    position: absolute;
    border-radius: 12px;
    border: 1px solid hsl(var(--border));
    background: color-mix(in srgb, hsl(var(--background)) 90%, #ffffff 10%);
    overflow: hidden;
    box-shadow: 0 8px 24px rgba(15, 23, 42, 0.08);
    transition: opacity 120ms ease, box-shadow 120ms ease, border-color 120ms ease;
}

.table-card--selected {
    border-color: color-mix(in srgb, hsl(var(--primary)) 75%, hsl(var(--border)));
    box-shadow: 0 10px 26px rgba(59, 130, 246, 0.2);
}

.table-card--dimmed {
    opacity: 0.4;
}

.schema-visualizer--dark .table-card {
    background: color-mix(in srgb, #0f172a 82%, #111827 18%);
    box-shadow: 0 12px 30px rgba(0, 0, 0, 0.32);
}

.table-header {
    height: 38px;
    padding: 0 10px;
    border-bottom: 1px solid hsl(var(--border));
    display: flex;
    align-items: center;
    justify-content: space-between;
    background: color-mix(in srgb, hsl(var(--primary)) 16%, transparent);
    cursor: grab;
    user-select: none;
}

.table-header:active {
    cursor: grabbing;
}

.table-name {
    font-size: 12px;
    font-weight: 700;
    letter-spacing: 0.02em;
}

.table-count {
    font-size: 11px;
    opacity: 0.65;
}

.column-list {
    list-style: none;
    margin: 0;
    padding: 6px;
}

.column-row {
    display: grid;
    grid-template-columns: 64px 1fr auto;
    align-items: center;
    gap: 8px;
    font-size: 11px;
    min-height: 26px;
    padding: 0 4px;
    border-radius: 6px;
}

.column-row:hover {
    background: color-mix(in srgb, hsl(var(--accent)) 70%, transparent);
}

.column-tags {
    display: flex;
    gap: 4px;
}

.pk,
.fk {
    border-radius: 4px;
    border: 1px solid hsl(var(--border));
    padding: 1px 4px;
    font-size: 9px;
    font-weight: 700;
}

.pk {
    background: color-mix(in srgb, #16a34a 22%, transparent);
}

.fk {
    background: color-mix(in srgb, #2563eb 18%, transparent);
}

.column-name {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    color: hsl(var(--foreground));
}

.column-type {
    font-size: 10px;
    opacity: 0.65;
    white-space: nowrap;
}

.minimap {
    position: absolute;
    right: 12px;
    bottom: 12px;
    width: 200px;
    height: 130px;
    border-radius: 8px;
    overflow: hidden;
    border: 1px solid hsl(var(--border));
    background: color-mix(in srgb, hsl(var(--background)) 85%, transparent);
    box-shadow: 0 6px 14px rgba(0, 0, 0, 0.14);
}

.minimap-svg {
    width: 100%;
    height: 100%;
}

.minimap-bg {
    fill: color-mix(in srgb, hsl(var(--background)) 92%, transparent);
}

.mini-node {
    fill: color-mix(in srgb, hsl(var(--foreground)) 20%, transparent);
}

.mini-node--selected {
    fill: color-mix(in srgb, hsl(var(--primary)) 70%, transparent);
}

.mini-node--dimmed {
    opacity: 0.35;
}

.mini-viewport {
    fill: transparent;
    stroke-width: 1.2;
    stroke: color-mix(in srgb, hsl(var(--primary)) 75%, #ffffff 25%);
}

@media (max-width: 768px) {
    .toolbar input {
        width: 120px;
    }

    .table-card {
        width: 260px !important;
    }

    .column-row {
        grid-template-columns: 58px 1fr auto;
    }

    .minimap {
        width: 150px;
        height: 100px;
    }
}
</style>

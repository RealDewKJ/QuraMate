<template>
    <div class="flex flex-col h-full bg-background text-foreground">
        <!-- Header -->
        <div class="flex items-center justify-between p-4 border-b border-border bg-card">
            <h2 class="text-lg font-semibold flex items-center gap-2">
                <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none"
                    stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                    class="lucide lucide-pencil-ruler text-primary">
                    <path d="M2 22h20" />
                    <path d="M12 6 2 16v6h6l10-10" />
                    <path d="m9 9 5 5" />
                </svg>
                <template v-if="!isCreateMode">Design: {{ tableName }}</template>
                <template v-else>
                    Create Table: <input v-model="inputTableName" placeholder="Table Name"
                        class="ml-2 bg-transparent border-b border-border focus:border-primary focus:outline-none text-lg font-semibold px-1" />
                </template>
            </h2>
            <div class="flex gap-2">
                <button @click="$emit('close')"
                    class="px-4 py-2 text-sm font-medium hover:bg-accent hover:text-accent-foreground rounded-md transition-colors">Cancel</button>
                <button @click="saveChanges" :disabled="!hasChanges || isSaving"
                    class="px-4 py-2 text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 rounded-md transition-colors disabled:opacity-50 flex items-center gap-2">
                    <svg v-if="isSaving" class="animate-spin h-4 w-4" xmlns="http://www.w3.org/2000/svg" fill="none"
                        viewBox="0 0 24 24">
                        <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4">
                        </circle>
                        <path class="opacity-75" fill="currentColor"
                            d="M4 12a8 8 0 0 1 8-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 0 1 4 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z">
                        </path>
                    </svg>
                    Save Changes
                </button>
            </div>
        </div>

        <!-- Tabs -->
        <div class="flex items-center border-b border-border bg-muted/20 px-4 pt-1 gap-1">
            <button v-for="tab in ['Columns', 'Indexes', 'Foreign Keys']" :key="tab" @click="activeTab = tab"
                class="px-4 py-2 text-sm font-medium rounded-t-md transition-all border-l border-r border-t border-transparent select-none"
                :class="activeTab === tab ? 'bg-background text-foreground border-border shadow-sm mb-[-1px]' : 'text-muted-foreground hover:text-foreground hover:bg-background/50'">
                {{ tab }}
            </button>
        </div>

        <!-- Content -->
        <div class="flex-1 overflow-auto p-4 bg-background">
            <!-- Loading State -->
            <div v-if="isLoading" class="flex items-center justify-center h-full text-muted-foreground">
                <svg class="animate-spin h-8 w-8 text-primary" xmlns="http://www.w3.org/2000/svg" fill="none"
                    viewBox="0 0 24 24">
                    <circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle>
                    <path class="opacity-75" fill="currentColor" d="M12 2A10 10 0 0 0 2 12h4a6 6 0 0 1 6-6V2z"></path>
                </svg>
            </div>

            <div v-else>
                <!-- Columns Tab -->
                <div v-if="activeTab === 'Columns'" class="space-y-4">
                    <div class="border border-border rounded-lg overflow-hidden bg-card">
                        <table class="w-full text-sm text-left">
                            <thead class="bg-muted text-muted-foreground text-xs uppercase font-medium">
                                <tr>
                                    <th class="p-3 border-b border-border">Name</th>
                                    <th class="p-3 border-b border-border">Type</th>
                                    <th class="p-3 border-b border-border text-center">Nullable</th>
                                    <th class="p-3 border-b border-border text-center">PK</th>
                                    <th class="p-3 border-b border-border text-center">Auto Inc</th>
                                    <th class="p-3 border-b border-border">Default</th>
                                    <th class="p-3 border-b border-border">Actions</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-border">
                                <tr v-for="(col, idx) in localColumns" :key="idx"
                                    class="transition-colors hover:bg-muted/30"
                                    :class="{ 'bg-destructive/10 hover:bg-destructive/20': col.status === 'deleted', 'bg-green-500/10 hover:bg-green-500/20': col.status === 'new' }">
                                    <td class="p-2">
                                        <input v-model="col.name"
                                            class="w-full bg-transparent px-2 py-1 rounded border border-transparent focus:border-primary focus:outline-none"
                                            :disabled="col.status === 'deleted'" placeholder="Column Name" />
                                    </td>
                                    <td class="p-2">
                                        <div class="flex items-center gap-1">
                                            <select v-model="col.baseType"
                                                class="bg-transparent px-1 py-1 rounded border border-transparent focus:border-primary focus:outline-none text-xs"
                                                :disabled="col.status === 'deleted'">
                                                <option v-for="t in commonTypes" :key="t" :value="t">{{ t }}</option>
                                                <option v-if="!commonTypes.includes(col.baseType)"
                                                    :value="col.baseType">{{
                                                        col.baseType }}</option>
                                            </select>
                                            <input v-if="typesWithLength.includes(col.baseType)" v-model="col.length"
                                                class="w-12 bg-transparent px-1 py-1 rounded border border-border focus:border-primary focus:outline-none text-xs"
                                                :disabled="col.status === 'deleted'" placeholder="255" />
                                        </div>
                                    </td>
                                    <td class="p-2 text-center">
                                        <input type="checkbox" v-model="col.nullable"
                                            :disabled="col.status === 'deleted'"
                                            class="rounded border-input text-primary focus:ring-primary" />
                                    </td>
                                    <td class="p-2 text-center">
                                        <input type="checkbox" v-model="col.primaryKey"
                                            :disabled="col.status === 'deleted'"
                                            class="rounded border-input text-primary focus:ring-primary" />
                                    </td>
                                    <td class="p-2 text-center">
                                        <input type="checkbox" v-model="col.autoIncrement"
                                            :disabled="col.status === 'deleted'"
                                            class="rounded border-input text-primary focus:ring-primary" />
                                    </td>
                                    <td class="p-2">
                                        <input v-model="col.defaultValue"
                                            class="w-full bg-transparent px-2 py-1 rounded border border-transparent focus:border-primary focus:outline-none"
                                            :disabled="col.status === 'deleted'" placeholder="NULL" />
                                    </td>
                                    <td class="p-2">
                                        <button @click="markColumnDeleted(idx)" v-if="col.status !== 'deleted'"
                                            class="text-destructive hover:text-destructive/80 transition-colors p-1 rounded hover:bg-destructive/10"
                                            title="Delete Column">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                                stroke-linecap="round" stroke-linejoin="round"
                                                class="lucide lucide-trash-2">
                                                <path d="M3 6h18" />
                                                <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                                                <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                                                <line x1="10" x2="10" y1="11" y2="17" />
                                                <line x1="14" x2="14" y1="11" y2="17" />
                                            </svg>
                                        </button>
                                        <button @click="restoreColumn(idx)" v-else
                                            class="text-primary hover:text-primary/80 transition-colors p-1 rounded hover:bg-primary/10"
                                            title="Restore Column">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                                stroke-linecap="round" stroke-linejoin="round"
                                                class="lucide lucide-undo-2">
                                                <path d="M9 14 4 9l5-5" />
                                                <path d="M4 9h10.5a5.5 5.5 0 0 1 5.5 5.5v0a5.5 5.5 0 0 1-5.5 5.5H11" />
                                            </svg>
                                        </button>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                    <button @click="addColumn"
                        class="px-4 py-2 text-sm font-medium bg-secondary text-secondary-foreground rounded-md hover:bg-secondary/90 transition-colors flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-plus">
                            <path d="M5 12h14" />
                            <path d="M12 5v14" />
                        </svg>
                        Add Column
                    </button>
                </div>

                <!-- Indexes Tab -->
                <div v-if="activeTab === 'Indexes'" class="space-y-4">
                    <div class="border border-border rounded-lg overflow-hidden bg-card">
                        <table class="w-full text-sm text-left">
                            <thead class="bg-muted text-muted-foreground text-xs uppercase font-medium">
                                <tr>
                                    <th class="p-3 border-b border-border">Name</th>
                                    <th class="p-3 border-b border-border">Columns</th>
                                    <th class="p-3 border-b border-border text-center">Unique</th>
                                    <th class="p-3 border-b border-border text-center">Primary</th>
                                    <th class="p-3 border-b border-border">Actions</th>
                                </tr>
                            </thead>
                            <tbody class="divide-y divide-border">
                                <tr v-for="(idx, i) in localIndexes" :key="i"
                                    class="transition-colors hover:bg-muted/30"
                                    :class="{ 'bg-destructive/10 hover:bg-destructive/20': idx.status === 'deleted', 'bg-green-500/10 hover:bg-green-500/20': idx.status === 'new' }">
                                    <td class="p-2">
                                        <input v-model="idx.name"
                                            class="w-full bg-transparent px-2 py-1 rounded border border-transparent focus:border-primary focus:outline-none"
                                            :disabled="idx.status === 'deleted'" placeholder="Index Name" />
                                    </td>
                                    <td class="p-2">
                                        <input v-model="idx.columnsStr"
                                            class="w-full bg-transparent px-2 py-1 rounded border border-transparent focus:border-primary focus:outline-none"
                                            :disabled="idx.status === 'deleted'" placeholder="col1, col2"
                                            @change="updateIndexColumns(idx)" />
                                    </td>
                                    <td class="p-2 text-center">
                                        <input type="checkbox" v-model="idx.unique" :disabled="idx.status === 'deleted'"
                                            class="rounded border-input text-primary focus:ring-primary" />
                                    </td>
                                    <td class="p-2 text-center">
                                        <input type="checkbox" v-model="idx.primary"
                                            :disabled="idx.status === 'deleted'"
                                            class="rounded border-input text-primary focus:ring-primary" />
                                    </td>
                                    <td class="p-2">
                                        <button @click="markIndexDeleted(i)" v-if="idx.status !== 'deleted'"
                                            class="text-destructive hover:text-destructive/80 transition-colors p-1 rounded hover:bg-destructive/10"
                                            title="Delete Index">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                                stroke-linecap="round" stroke-linejoin="round"
                                                class="lucide lucide-trash-2">
                                                <path d="M3 6h18" />
                                                <path d="M19 6v14c0 1-1 2-2 2H7c-1 0-2-1-2-2V6" />
                                                <path d="M8 6V4c0-1 1-2 2-2h4c1 0 2 1 2 2v2" />
                                                <line x1="10" x2="10" y1="11" y2="17" />
                                                <line x1="14" x2="14" y1="11" y2="17" />
                                            </svg>
                                        </button>
                                        <button @click="restoreIndex(i)" v-else
                                            class="text-primary hover:text-primary/80 transition-colors p-1 rounded hover:bg-primary/10"
                                            title="Restore Index">
                                            <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16"
                                                viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"
                                                stroke-linecap="round" stroke-linejoin="round"
                                                class="lucide lucide-undo-2">
                                                <path d="M9 14 4 9l5-5" />
                                                <path d="M4 9h10.5a5.5 5.5 0 0 1 5.5 5.5v0a5.5 5.5 0 0 1-5.5 5.5H11" />
                                            </svg>
                                        </button>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>
                    <button @click="addIndex"
                        class="px-4 py-2 text-sm font-medium bg-secondary text-secondary-foreground rounded-md hover:bg-secondary/90 transition-colors flex items-center gap-2">
                        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none"
                            stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
                            class="lucide lucide-plus">
                            <path d="M5 12h14" />
                            <path d="M12 5v14" />
                        </svg>
                        Add Index
                    </button>
                </div>

                <!-- Foreign Keys Tab -->
                <div v-if="activeTab === 'Foreign Keys'" class="space-y-4">
                    <div class="p-8 text-center text-muted-foreground">
                        <p>Foreign Key editing coming soon.</p>
                    </div>
                </div>

            </div>
        </div>
        <Toast ref="toastRef" />
    </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue';
import Toast from './Toast.vue';

const toastRef = ref(null);
const props = defineProps({
    tableName: String,
    connectionId: String,
    dbType: String
});

const emit = defineEmits(['close', 'refresh', 'success']);

const inputTableName = ref(props.tableName || '');
const isCreateMode = computed(() => !props.tableName);

const activeTab = ref('Columns');
const isLoading = ref(true);
const isSaving = ref(false);

const commonTypes = [
    'VARCHAR', 'NVARCHAR', 'TEXT', 'INT', 'BIGINT', 'SMALLINT', 'TINYINT',
    'BOOLEAN', 'BIT', 'DATE', 'DATETIME', 'TIMESTAMP', 'DECIMAL', 'NUMERIC',
    'REAL', 'DOUBLE', 'FLOAT', 'BLOB', 'JSON'
];

const typesWithLength = ['VARCHAR', 'NVARCHAR', 'CHAR', 'NCHAR', 'DECIMAL', 'NUMERIC', 'VARBINARY', 'BINARY'];

function parseType(fullType) {
    if (!fullType) return { base: 'VARCHAR', length: '255' };
    const match = fullType.match(/^([^(]+)(?:\(([^)]+)\))?$/);
    if (match) {
        return {
            base: match[1].trim().toUpperCase(),
            length: match[2] || ''
        };
    }
    return { base: fullType.toUpperCase(), length: '' };
}

const originalColumns = ref([]);
const localColumns = ref([]);
const originalIndexes = ref([]);
const localIndexes = ref([]);


onMounted(async () => {
    await fetchData();
});

async function fetchData() {
    if (isCreateMode.value) {
        isLoading.value = false;
        return;
    }
    isLoading.value = true;
    try {
        if (window.go && window.go.main && window.go.main.App) {
            // Fetch Columns
            const cols = await window.go.main.App.GetTableDefinition(props.connectionId, props.tableName);
            originalColumns.value = JSON.parse(JSON.stringify(cols));
            // Store originalName to track renames
            localColumns.value = cols.map(c => {
                const { base, length } = parseType(c.type);
                return {
                    ...c,
                    originalName: c.name,
                    status: 'existing',
                    baseType: base,
                    length: length
                };
            });

            // Fetch Indexes
            const indexes = await window.go.main.App.GetTableIndexes(props.connectionId, props.tableName);
            originalIndexes.value = JSON.parse(JSON.stringify(indexes));
            localIndexes.value = indexes.map(i => ({ ...i, columnsStr: i.columns.join(', '), originalName: i.name, status: 'existing' }));
        }
    } catch (e) {
        console.error("Failed to load table definition", e);
    } finally {
        isLoading.value = false;
    }
}

// Column Actions
function addColumn() {
    localColumns.value.push({
        name: 'new_column',
        type: 'VARCHAR(255)',
        baseType: 'VARCHAR',
        length: '255',
        nullable: true,
        primaryKey: false,
        autoIncrement: false,
        defaultValue: null,
        status: 'new'
    });
}
function markColumnDeleted(index) {
    if (localColumns.value[index].status === 'new') {
        localColumns.value.splice(index, 1);
    } else {
        localColumns.value[index].status = 'deleted';
    }
}
function restoreColumn(index) {
    localColumns.value[index].status = 'existing';
}

// Index Actions
function addIndex() {
    localIndexes.value.push({
        name: `idx_${props.tableName}_new`,
        columns: [],
        columnsStr: '',
        unique: false,
        primary: false,
        status: 'new'
    });
}
function markIndexDeleted(index) {
    if (localIndexes.value[index].status === 'new') {
        localIndexes.value.splice(index, 1);
    } else {
        localIndexes.value[index].status = 'deleted';
    }
}
function restoreIndex(index) {
    localIndexes.value[index].status = 'existing';
}
function updateIndexColumns(idx) {
    idx.columns = idx.columnsStr.split(',').map(s => s.trim()).filter(s => s);
}

// Change Detection
const hasChanges = computed(() => {
    // In Create Mode, we need at least a table name and one valid column
    if (isCreateMode.value) {
        return !!inputTableName.value && localColumns.value.some(c => c.status === 'new' && c.name && c.type);
    }

    // Check Columns
    for (const col of localColumns.value) {
        if (col.status === 'new' || col.status === 'deleted') return true;
        if (col.status === 'existing') {
            // Check for modifications
            if (col.name !== col.originalName) return true;
            // Compare other fields
            const orig = originalColumns.value.find(c => c.name === col.originalName);
            if (orig) {
                if (col.baseType !== parseType(orig.type).base) return true;
                if (col.length !== parseType(orig.type).length) return true;
                if (col.nullable !== orig.nullable) return true;
                if (col.defaultValue != orig.defaultValue) return true;
                if (col.primaryKey !== orig.primaryKey) return true;
                if (col.autoIncrement !== orig.autoIncrement) return true;
            }
        }
    }

    // Check Indexes
    for (const idx of localIndexes.value) {
        if (idx.status === 'new' || idx.status === 'deleted') return true;
        if (idx.status === 'existing') {
            if (idx.name !== idx.originalName) return true;
            const orig = originalIndexes.value.find(i => i.name === idx.originalName);
            if (orig) {
                if (idx.unique !== orig.unique) return true;
                if (idx.primary !== orig.primary) return true;
                if (JSON.stringify(idx.columns) !== JSON.stringify(orig.columns)) return true;
            }
        }
    }

    return false;
});

async function saveChanges() {
    isSaving.value = true;
    try {
        // Construct full type string for each column
        localColumns.value.forEach(col => {
            if (typesWithLength.includes(col.baseType)) {
                let len = col.length;
                if (!len || len.trim() === '') {
                    len = '255'; // Default to 255 if empty
                }
                col.type = `${col.baseType}(${len})`;
            } else {
                col.type = col.baseType;
            }
        });

        if (isCreateMode.value) {
            const columnsToCreate = localColumns.value
                .filter(c => c.status === 'new')
                .map(c => ({
                    name: c.name,
                    type: c.type,
                    nullable: c.nullable,
                    defaultValue: c.defaultValue,
                    primaryKey: c.primaryKey,
                    autoIncrement: c.autoIncrement
                }));

            if (window.go && window.go.main && window.go.main.App) {
                const result = await window.go.main.App.CreateTable(props.connectionId, inputTableName.value, columnsToCreate);
                if (result === "Success") {
                    emit('refresh');
                    emit('success', 'Table created successfully!');
                    emit('close');
                } else {
                    toastRef.value?.error('Error creating table: ' + result);
                }
            }
            return;
        }

        const changes = {
            renameTable: "",
            addColumns: [],
            dropColumns: [],
            alterColumns: [],
            addIndexes: [],
            dropIndexes: [],
            addFKs: [],
            dropFKs: []
        };

        // Process Columns
        for (const col of localColumns.value) {
            if (col.status === 'new') {
                changes.addColumns.push({
                    name: col.name,
                    type: col.type,
                    nullable: col.nullable,
                    defaultValue: col.defaultValue,
                    primaryKey: col.primaryKey,
                    autoIncrement: col.autoIncrement
                });
            } else if (col.status === 'deleted') {
                changes.dropColumns.push(col.originalName);
            } else if (col.status === 'existing') {
                const orig = originalColumns.value.find(c => c.name === col.originalName);
                if (orig) {
                    const isModified =
                        col.name !== orig.name ||
                        col.type !== orig.type ||
                        col.nullable !== orig.nullable ||
                        col.defaultValue != orig.defaultValue ||
                        col.primaryKey !== orig.primaryKey ||
                        col.autoIncrement !== orig.autoIncrement;

                    if (isModified) {
                        changes.alterColumns.push({
                            oldName: col.originalName,
                            newDefinition: {
                                name: col.name,
                                type: col.type,
                                nullable: col.nullable,
                                defaultValue: col.defaultValue,
                                primaryKey: col.primaryKey,
                                autoIncrement: col.autoIncrement
                            }
                        });
                    }
                }
            }
        }

        // Process Indexes
        for (const idx of localIndexes.value) {
            if (idx.status === 'new') {
                changes.addIndexes.push({
                    name: idx.name,
                    columns: idx.columns,
                    unique: idx.unique,
                    primary: idx.primary
                });
            } else if (idx.status === 'deleted') {
                changes.dropIndexes.push(idx.originalName);
            } else if (idx.status === 'existing') {
                const orig = originalIndexes.value.find(i => i.name === idx.originalName);
                if (orig) {
                    const isModified =
                        idx.name !== orig.name || // Rename usually means drop/add for indexes
                        idx.unique !== orig.unique ||
                        idx.primary !== orig.primary ||
                        JSON.stringify(idx.columns) !== JSON.stringify(orig.columns);

                    if (isModified) {
                        // Treat modification as Drop + Add
                        changes.dropIndexes.push(idx.originalName);
                        changes.addIndexes.push({
                            name: idx.name,
                            columns: idx.columns,
                            unique: idx.unique,
                            primary: idx.primary
                        });
                    }
                }
            }
        }

        console.log("Saving changes:", changes);

        if (window.go && window.go.main && window.go.main.App) {
            const result = await window.go.main.App.AlterTable(props.connectionId, props.tableName, changes);
            if (result === "Success") {
                // Refresh data
                await fetchData();
                emit('refresh');
                emit('success', 'Changes saved successfully!');
            } else {
                toastRef.value?.error('Error saving: ' + result);
            }
        }

    } catch (e) {
        console.error(e);
        toastRef.value?.error('Error saving changes: ' + e);
    } finally {
        isSaving.value = false;
    }
}

// Auto-fill default length when type changes
watch(localColumns, (newCols) => {
    newCols.forEach(col => {
        if (typesWithLength.includes(col.baseType) && (!col.length || col.length.trim() === '')) {
            if (col.baseType === 'VARCHAR' || col.baseType === 'NVARCHAR') {
                col.length = '255';
            }
        }
    });
}, { deep: true });
</script>

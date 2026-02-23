import { ref, computed, nextTick, Ref } from 'vue';
import { GetTableDefinition, ExecuteTransientQuery, UpdateRecord } from '../../wailsjs/go/main/App';
import { QueryTab } from '../types/dashboard';

export function useRecordOperations(
    connectionId: string,
    isReadOnly: boolean,
    activeTab: Ref<QueryTab | undefined>,
    runQuery: () => void,
    toast: any
) {
    // Update Confirmation State
    const updateConfirmation = ref<{
        isOpen: boolean;
        tableName: string;
        column: string;
        originalValue: any;
        newValue: any;
        rowIndex: number;
        item: any;
    } | null>(null);

    // Insert Row Modal State
    const insertRowModal = ref<{
        isOpen: boolean;
        tableName: string;
        columns: string[];
        columnDefs: Record<string, { type: string; nullable: boolean; autoIncrement: boolean; primaryKey: boolean }>;
        values: Record<string, string>;
        nullColumns: Record<string, boolean>;
        isInserting: boolean;
        error: string;
    } | null>(null);

    const activeResultSet = computed(() => {
        if (!activeTab.value || !activeTab.value.resultSets || activeTab.value.resultSets.length === 0) return null;
        return activeTab.value.resultSets[0];
    });

    const getInputTypeForColumn = (sqlType: string): string => {
        const t = sqlType.toLowerCase();
        if (t.includes('int') || t === 'number' || t === 'decimal' || t === 'numeric' || t === 'float' || t === 'real' || t === 'double') return 'number';
        if (t.includes('date') || t.includes('time') || t === 'timestamp') return 'datetime-local';
        if (t === 'boolean' || t === 'bit' || t === 'bool') return 'checkbox';
        return 'text';
    };

    const getInputType = (col: string): string => {
        if (!insertRowModal.value?.columnDefs[col]) return 'text';
        return getInputTypeForColumn(insertRowModal.value.columnDefs[col].type);
    };

    const getNumberStep = (col: string): string => {
        if (!insertRowModal.value?.columnDefs[col]) return 'any';
        const t = insertRowModal.value.columnDefs[col].type.toLowerCase();
        if (t.includes('int') || t === 'serial' || t === 'bigserial' || t === 'smallserial') return '1';
        return 'any';
    };

    const isEditable = (col: string) => {
        if (isReadOnly) return false;
        if (!activeTab.value || !activeTab.value.tableName || activeTab.value.primaryKeys.length === 0) return false;
        if (activeTab.value.isDesignView) return false;
        if (activeTab.value.primaryKeys.includes(col)) return false;
        return true;
    };

    const initiateQuickUpdate = (newValue: any, item: any, col: string) => {
        if (!activeTab.value || !activeTab.value.tableName) return;

        if (!isEditable(col)) {
            toast?.error("This column cannot be edited (Primary Key or Read Only).");
            return;
        }

        const originalValue = item.data[col];
        updateConfirmation.value = {
            isOpen: true,
            tableName: activeTab.value.tableName,
            column: col,
            originalValue: originalValue,
            newValue: newValue,
            rowIndex: item.index,
            item: item
        };
    };

    const confirmUpdate = async () => {
        if (!updateConfirmation.value || !activeTab.value) return;

        const { tableName, column, newValue, item } = updateConfirmation.value;
        const col = column;

        const conditions: Record<string, any> = {};
        for (const pk of activeTab.value.primaryKeys) {
            conditions[pk] = item.data[pk];
        }

        const updates: Record<string, any> = {};
        updates[col] = newValue;

        try {
            const result = await UpdateRecord(connectionId, tableName, updates, conditions);
            if (result === "Success") {
                item.data[col] = newValue;
                if (activeResultSet.value) {
                    const realIndex = activeResultSet.value.rows.findIndex(r => {
                        for (const pk of activeTab.value!.primaryKeys) {
                            if (r[pk] !== conditions[pk]) return false;
                        }
                        return true;
                    });
                    if (realIndex !== -1) {
                        activeResultSet.value.rows[realIndex][col] = newValue;
                    }
                }
            } else {
                toast?.error('Update failed: ' + result);
            }
        } catch (e: any) {
            toast?.error('Update error: ' + e);
        } finally {
            if (activeTab.value) {
                activeTab.value.editingCell = null;
            }
            updateConfirmation.value = null;
        }
    };

    const openInsertRowModal = async () => {
        if (!activeTab.value || !activeTab.value.tableName || !activeTab.value.resultSets?.[0]?.columns) return;

        const columns = activeTab.value.resultSets[0].columns;
        const pks = activeTab.value.primaryKeys || [];
        const values: Record<string, string> = {};
        const nullColumns: Record<string, boolean> = {};
        const columnDefs: Record<string, { type: string; nullable: boolean; autoIncrement: boolean; primaryKey: boolean }> = {};

        try {
            const defs = await GetTableDefinition(connectionId, activeTab.value.tableName);
            if (defs && defs.length > 0) {
                for (const def of defs) {
                    columnDefs[def.name] = {
                        type: def.type,
                        nullable: def.nullable,
                        autoIncrement: def.autoIncrement,
                        primaryKey: def.primaryKey
                    };
                }
            }
        } catch (e) {
            console.warn('Failed to fetch column definitions', e);
        }

        for (const col of columns) {
            const def = columnDefs[col];
            nullColumns[col] = (def?.autoIncrement || false) || pks.includes(col);
            if (def) {
                const inputType = getInputTypeForColumn(def.type);
                values[col] = inputType === 'checkbox' ? '0' : '';
            } else {
                values[col] = '';
            }
        }

        insertRowModal.value = {
            isOpen: true,
            tableName: activeTab.value.tableName,
            columns: columns,
            columnDefs: columnDefs,
            values: values,
            nullColumns: nullColumns,
            isInserting: false,
            error: ''
        };

        nextTick(() => {
            const firstInput = document.querySelector('.insert-row-input') as HTMLInputElement;
            if (firstInput) firstInput.focus();
        });
    };

    const toggleInsertNull = (col: string) => {
        if (!insertRowModal.value) return;
        insertRowModal.value.nullColumns[col] = !insertRowModal.value.nullColumns[col];
        if (insertRowModal.value.nullColumns[col]) {
            insertRowModal.value.values[col] = '';
        }
    };

    const confirmInsertRow = async () => {
        if (!insertRowModal.value || !activeTab.value) return;

        insertRowModal.value.isInserting = true;
        insertRowModal.value.error = '';

        const { tableName, columns, values, nullColumns, columnDefs } = insertRowModal.value;
        const insertCols: string[] = [];
        const insertVals: string[] = [];

        for (const col of columns) {
            if (nullColumns[col]) {
                insertCols.push(col);
                insertVals.push('NULL');
            } else {
                insertCols.push(col);
                const val = values[col];
                const def = columnDefs[col];
                const inputType = def ? getInputTypeForColumn(def.type) : 'text';

                if (inputType === 'number' || inputType === 'checkbox') {
                    insertVals.push(val || '0');
                } else {
                    const escaped = val.replace(/'/g, "''");
                    insertVals.push(`'${escaped}'`);
                }
            }
        }

        const sql = `INSERT INTO ${tableName} (${insertCols.join(', ')}) VALUES (${insertVals.join(', ')})`;

        try {
            const res = await ExecuteTransientQuery(connectionId, sql);
            if (res.error) {
                insertRowModal.value.isInserting = false;
                insertRowModal.value.error = res.error;
                return;
            }

            insertRowModal.value = null;
            toast?.success('Row inserted successfully!');
            runQuery();
        } catch (e: any) {
            if (insertRowModal.value) {
                insertRowModal.value.isInserting = false;
                insertRowModal.value.error = e.toString();
            }
        }
    };

    return {
        updateConfirmation,
        insertRowModal,
        isEditable,
        initiateQuickUpdate,
        confirmUpdate,
        openInsertRowModal,
        toggleInsertNull,
        confirmInsertRow,
        getInputType,
        getNumberStep,
        cancelInsertRow: () => { insertRowModal.value = null; },
        cancelUpdate: () => { 
            if (activeTab.value) activeTab.value.editingCell = null;
            updateConfirmation.value = null; 
        }
    };
}

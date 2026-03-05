import { computed, nextTick, ref, toValue } from 'vue';
import type { MaybeRefOrGetter, Ref } from 'vue';
import { ExecuteTransientQuery, GetTableDefinition, InsertRowsBatch, UpdateRecord } from '../../wailsjs/go/main/App';
import { QueryTab } from '../types/dashboard';

interface ColumnDefMeta {
    type: string;
    nullable: boolean;
    autoIncrement: boolean;
    primaryKey: boolean;
    defaultValue?: any;
}

interface RowValidationResult {
    isValid: boolean;
    error: string;
    payload: Record<string, any>;
}

interface PastePreviewRow {
    id: number;
    sourceRowNumber: number;
    values: Record<string, string>;
    include: boolean;
    isValid: boolean;
    error: string;
    payload: Record<string, any>;
    hasUnmappedValue: boolean;
}

interface PastePreviewModalState {
    isOpen: boolean;
    tableName: string;
    columns: string[];
    columnDefs: Record<string, ColumnDefMeta>;
    rows: PastePreviewRow[];
    isInserting: boolean;
    error: string;
}

export function useRecordOperations(
    connectionId: MaybeRefOrGetter<string>,
    isReadOnly: boolean,
    activeTab: Ref<QueryTab | undefined>,
    runQuery: () => void,
    toast: any
) {
    const updateConfirmation = ref<{
        isOpen: boolean;
        tableName: string;
        column: string;
        originalValue: any;
        newValue: any;
        rowIndex: number;
        item: any;
    } | null>(null);

    const insertRowModal = ref<{
        isOpen: boolean;
        tableName: string;
        columns: string[];
        columnDefs: Record<string, ColumnDefMeta>;
        values: Record<string, string>;
        nullColumns: Record<string, boolean>;
        isInserting: boolean;
        error: string;
    } | null>(null);

    const pastePreviewModal = ref<PastePreviewModalState | null>(null);
    const isPastingRows = ref(false);

    const activeResultSet = computed(() => {
        if (!activeTab.value || !activeTab.value.resultSets || activeTab.value.resultSets.length === 0) return null;
        return activeTab.value.resultSets[0];
    });

    const normalizeColumnName = (value: string) => value.trim().toLowerCase();

    const getInputTypeForColumn = (sqlType: string): string => {
        const t = sqlType.toLowerCase();
        if (t.includes('int') || t === 'number' || t === 'decimal' || t === 'numeric' || t === 'float' || t === 'real' || t === 'double') return 'number';
        if (t.includes('date') || t.includes('time') || t === 'timestamp') return 'datetime-local';
        if (t === 'boolean' || t === 'bit' || t === 'bool') return 'checkbox';
        return 'text';
    };

    const getInputType = (col: string): string => {
        const fromInsert = insertRowModal.value?.columnDefs[col];
        if (fromInsert) return getInputTypeForColumn(fromInsert.type);

        const fromPaste = pastePreviewModal.value?.columnDefs[col];
        if (fromPaste) return getInputTypeForColumn(fromPaste.type);

        return 'text';
    };

    const getNumberStep = (col: string): string => {
        const fromInsert = insertRowModal.value?.columnDefs[col];
        const fromPaste = pastePreviewModal.value?.columnDefs[col];
        const def = fromInsert || fromPaste;

        if (!def) return 'any';
        const t = def.type.toLowerCase();
        if (t.includes('int') || t === 'serial' || t === 'bigserial' || t === 'smallserial') return '1';
        return 'any';
    };

    const parseBooleanValue = (raw: string): number | null => {
        const normalized = raw.trim().toLowerCase();
        if (['1', 'true', 'yes', 'y'].includes(normalized)) return 1;
        if (['0', 'false', 'no', 'n'].includes(normalized)) return 0;
        return null;
    };

    const hasDefaultValue = (value: any): boolean => {
        if (value === null || value === undefined) return false;
        const asString = String(value).trim().toLowerCase();
        return asString !== '' && asString !== 'null';
    };

    const isColumnRequired = (def?: ColumnDefMeta): boolean => {
        if (!def) return false;
        if (def.autoIncrement || def.primaryKey || def.nullable) return false;
        return !hasDefaultValue(def.defaultValue);
    };

    const getColumnDefinitions = async (tableName: string): Promise<Record<string, ColumnDefMeta>> => {
        const id = toValue(connectionId);
        if (!id) return {};

        const columnDefs: Record<string, ColumnDefMeta> = {};

        try {
            const defs = await GetTableDefinition(id, tableName);
            if (defs && defs.length > 0) {
                for (const def of defs) {
                    columnDefs[def.name] = {
                        type: def.type,
                        nullable: def.nullable,
                        autoIncrement: def.autoIncrement,
                        primaryKey: def.primaryKey,
                        defaultValue: def.defaultValue
                    };
                }
            }
        } catch (e) {
            console.warn('Failed to fetch column definitions', e);
        }

        return columnDefs;
    };

    const parseClipboardRows = (clipboardText: string): string[][] => {
        const normalized = clipboardText.replace(/\r\n/g, '\n').replace(/\r/g, '\n');
        return normalized
            .split('\n')
            .filter((line) => line.length > 0)
            .map((line) => line.split('\t'));
    };

    const validatePreviewRow = (
        rowValues: Record<string, string>,
        columns: string[],
        columnDefs: Record<string, ColumnDefMeta>,
        sourceRowNumber: number,
        hasUnmappedValue: boolean
    ): RowValidationResult => {
        if (hasUnmappedValue) {
            return {
                isValid: false,
                error: `Row ${sourceRowNumber}: contains values with no matching table column`,
                payload: {}
            };
        }

        const payload: Record<string, any> = {};

        for (const col of columns) {
            const rawValue = rowValues[col] ?? '';
            const trimmed = rawValue.trim();
            const def = columnDefs[col];

            if (trimmed === '') {
                continue;
            }

            if (/^null$/i.test(trimmed)) {
                if (!def?.nullable) {
                    return {
                        isValid: false,
                        error: `Row ${sourceRowNumber}: column ${col} does not allow NULL`,
                        payload: {}
                    };
                }
                payload[col] = null;
                continue;
            }

            const inputType = def ? getInputTypeForColumn(def.type) : 'text';

            if (inputType === 'number') {
                const parsed = Number(trimmed);
                if (!Number.isFinite(parsed)) {
                    return {
                        isValid: false,
                        error: `Row ${sourceRowNumber}: column ${col} expects a number`,
                        payload: {}
                    };
                }
                payload[col] = parsed;
                continue;
            }

            if (inputType === 'checkbox') {
                const parsed = parseBooleanValue(trimmed);
                if (parsed === null) {
                    return {
                        isValid: false,
                        error: `Row ${sourceRowNumber}: column ${col} expects boolean (true/false or 1/0)`,
                        payload: {}
                    };
                }
                payload[col] = parsed;
                continue;
            }

            if (inputType === 'datetime-local' && trimmed.includes('T')) {
                payload[col] = trimmed.replace('T', ' ');
                continue;
            }

            payload[col] = trimmed;
        }

        for (const col of columns) {
            if (!isColumnRequired(columnDefs[col])) continue;
            if (!Object.prototype.hasOwnProperty.call(payload, col)) {
                return {
                    isValid: false,
                    error: `Row ${sourceRowNumber}: required column ${col} is missing`,
                    payload: {}
                };
            }
        }

        if (Object.keys(payload).length === 0) {
            return {
                isValid: false,
                error: `Row ${sourceRowNumber}: no insertable values`,
                payload: {}
            };
        }

        return {
            isValid: true,
            error: '',
            payload
        };
    };

    const revalidatePreviewRow = (row: PastePreviewRow, columns: string[], columnDefs: Record<string, ColumnDefMeta>) => {
        const validation = validatePreviewRow(row.values, columns, columnDefs, row.sourceRowNumber, row.hasUnmappedValue);
        row.isValid = validation.isValid;
        row.error = validation.error;
        row.payload = validation.payload;
        if (!row.isValid && row.include) {
            row.include = false;
        }
    };

    const buildPastePreviewRows = (
        parsedRows: string[][],
        orderedColumns: string[],
        columnDefs: Record<string, ColumnDefMeta>
    ): { columns: string[]; rows: PastePreviewRow[] } => {
        const normalizedLookup = new Map<string, string>();
        for (const col of orderedColumns) {
            normalizedLookup.set(normalizeColumnName(col), col);
        }

        const insertableColumns = orderedColumns.filter((col) => !columnDefs[col]?.autoIncrement);

        const firstRow = parsedRows[0] || [];
        const headerMatches = firstRow.filter((cell) => normalizedLookup.has(normalizeColumnName(cell))).length;
        const hasHeader = firstRow.length > 0 && headerMatches > 0 && (headerMatches / firstRow.length) >= 0.6;

        const mappedColumns = hasHeader
            ? firstRow.map((cell) => normalizedLookup.get(normalizeColumnName(cell)) || '')
            : insertableColumns;

        const dataRows = hasHeader ? parsedRows.slice(1) : parsedRows;

        const rows: PastePreviewRow[] = [];

        dataRows.forEach((cells, idx) => {
            const sourceRowNumber = hasHeader ? idx + 2 : idx + 1;
            const values: Record<string, string> = {};

            for (const col of insertableColumns) {
                values[col] = '';
            }

            let hasUnmappedValue = false;

            for (let i = 0; i < cells.length; i++) {
                const raw = cells[i] ?? '';
                const mapped = mappedColumns[i];
                const trimmed = raw.trim();

                if (!mapped) {
                    if (trimmed !== '') hasUnmappedValue = true;
                    continue;
                }

                if (!Object.prototype.hasOwnProperty.call(values, mapped)) {
                    continue;
                }

                values[mapped] = raw;
            }

            const validation = validatePreviewRow(values, insertableColumns, columnDefs, sourceRowNumber, hasUnmappedValue);
            const include = validation.isValid;

            rows.push({
                id: idx + 1,
                sourceRowNumber,
                values,
                include,
                isValid: validation.isValid,
                error: validation.error,
                payload: validation.payload,
                hasUnmappedValue
            });
        });

        return {
            columns: insertableColumns,
            rows
        };
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
            toast.value?.error('This column cannot be edited (Primary Key or Read Only).');
            return;
        }

        const originalValue = item.data[col];
        updateConfirmation.value = {
            isOpen: true,
            tableName: activeTab.value.tableName,
            column: col,
            originalValue,
            newValue,
            rowIndex: item.index,
            item
        };
    };

    const confirmUpdate = async () => {
        if (!updateConfirmation.value || !activeTab.value) return;
        const id = toValue(connectionId);
        if (!id) return;

        const { tableName, column, newValue, item } = updateConfirmation.value;
        const col = column;

        const conditions: Record<string, any> = {};
        for (const pk of activeTab.value.primaryKeys) {
            conditions[pk] = item.data[pk];
        }

        const updates: Record<string, any> = {};
        let processedValue = newValue;

        if (typeof processedValue === 'string' && /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}/.test(processedValue)) {
            const originalValue = updateConfirmation.value.originalValue;
            if (typeof originalValue === 'string') {
                if (originalValue.includes('Z')) {
                    if (!processedValue.includes('Z')) processedValue += 'Z';
                } else if (!originalValue.includes('T') && originalValue.includes(' ')) {
                    processedValue = processedValue.replace('T', ' ');
                }
            } else {
                processedValue = processedValue.replace('T', ' ');
            }
        }
        updates[col] = processedValue;

        try {
            const result = await UpdateRecord(id, tableName, updates, conditions);
            if (result === 'Success') {
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
                toast.value?.error('Update failed: ' + result);
            }
        } catch (e: any) {
            toast.value?.error('Update error: ' + e);
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
        const columnDefs = await getColumnDefinitions(activeTab.value.tableName);

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
            columns,
            columnDefs,
            values,
            nullColumns,
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
        const id = toValue(connectionId);
        if (!id) return;

        insertRowModal.value.isInserting = true;
        insertRowModal.value.error = '';

        const { tableName, columns, values, nullColumns, columnDefs } = insertRowModal.value;
        const insertCols: string[] = [];
        const insertVals: string[] = [];

        for (const col of columns) {
            if (!nullColumns[col]) {
                insertCols.push(col);
                const val = values[col];
                const def = columnDefs[col];
                const inputType = def ? getInputTypeForColumn(def.type) : 'text';

                let processedVal = val;
                if (inputType === 'datetime-local' && processedVal && processedVal.includes('T')) {
                    processedVal = processedVal.replace('T', ' ');
                }

                if (inputType === 'number' || inputType === 'checkbox') {
                    insertVals.push(processedVal || '0');
                } else {
                    const escaped = processedVal.replace(/'/g, "''");
                    insertVals.push(`'${escaped}'`);
                }
            }
        }

        const sql = `INSERT INTO ${tableName} (${insertCols.join(', ')}) VALUES (${insertVals.join(', ')})`;

        try {
            const res = await ExecuteTransientQuery(id, sql);
            if (res.error) {
                insertRowModal.value.isInserting = false;
                insertRowModal.value.error = res.error;
                return;
            }

            insertRowModal.value = null;
            toast.value?.success('Row inserted successfully!');
            runQuery();
        } catch (e: any) {
            if (insertRowModal.value) {
                insertRowModal.value.isInserting = false;
                insertRowModal.value.error = e.toString();
            }
        }
    };

    const pasteRowsFromClipboard = async () => {
        if (isReadOnly) {
            toast.value?.error('Cannot paste rows in read-only mode.');
            return;
        }

        if (!activeTab.value || !activeTab.value.tableName || !activeTab.value.resultSets?.[0]?.columns) {
            toast.value?.error('Open a table result before pasting rows.');
            return;
        }

        if (isPastingRows.value) return;

        if (!navigator.clipboard?.readText) {
            toast.value?.error('Clipboard read is not supported in this environment.');
            return;
        }

        isPastingRows.value = true;

        try {
            const text = await navigator.clipboard.readText();
            if (!text || !text.trim()) {
                toast.value?.error('Clipboard is empty.');
                return;
            }

            const parsedRows = parseClipboardRows(text);
            if (parsedRows.length === 0) {
                toast.value?.error('No rows detected in clipboard data.');
                return;
            }

            const tableName = activeTab.value.tableName;
            const orderedColumns = activeTab.value.resultSets[0].columns.filter((col: string) => !!col);
            const columnDefs = await getColumnDefinitions(tableName);
            const preview = buildPastePreviewRows(parsedRows, orderedColumns, columnDefs);

            if (preview.rows.length === 0) {
                toast.value?.error('No rows available for preview.');
                return;
            }

            pastePreviewModal.value = {
                isOpen: true,
                tableName,
                columns: preview.columns,
                columnDefs,
                rows: preview.rows,
                isInserting: false,
                error: ''
            };
        } catch (e: any) {
            toast.value?.error('Paste preview failed: ' + e.toString());
        } finally {
            isPastingRows.value = false;
        }
    };

    const updatePastePreviewValue = (payload: { rowId: number; col: string; value: any }) => {
        if (!pastePreviewModal.value) return;

        const row = pastePreviewModal.value.rows.find((item) => item.id === payload.rowId);
        if (!row) return;

        row.values[payload.col] = String(payload.value ?? '');
        revalidatePreviewRow(row, pastePreviewModal.value.columns, pastePreviewModal.value.columnDefs);
    };

    const togglePastePreviewRow = (rowId: number) => {
        if (!pastePreviewModal.value) return;
        const row = pastePreviewModal.value.rows.find((item) => item.id === rowId);
        if (!row) return;

        if (!row.include && !row.isValid) {
            toast.value?.error(`Row ${row.sourceRowNumber} is invalid. Please edit or keep it skipped.`);
            return;
        }

        row.include = !row.include;
    };


    const autoFixPastePreviewRows = () => {
        if (!pastePreviewModal.value) return;
        const { rows, columns, columnDefs } = pastePreviewModal.value;

        for (const row of rows) {
            for (const col of columns) {
                const rawValue = row.values[col] ?? '';
                const trimmed = String(rawValue).trim();
                const inputType = getInputTypeForColumn(columnDefs[col]?.type || '');

                if (trimmed === '') {
                    row.values[col] = '';
                    continue;
                }

                if (inputType === 'number') {
                    row.values[col] = trimmed.replace(/,/g, '');
                    continue;
                }

                if (inputType === 'checkbox') {
                    const parsedBool = parseBooleanValue(trimmed);
                    row.values[col] = parsedBool !== null ? String(parsedBool) : trimmed;
                    continue;
                }

                if (inputType === 'datetime-local') {
                    row.values[col] = trimmed.replace(' ', 'T');
                    continue;
                }

                row.values[col] = trimmed;
            }

            revalidatePreviewRow(row, columns, columnDefs);
            if (row.isValid) {
                row.include = true;
            }
        }
    };

    const confirmPastePreviewInsert = async () => {
        if (!pastePreviewModal.value) return;
        const id = toValue(connectionId);
        if (!id) return;

        const modal = pastePreviewModal.value;
        modal.error = '';

        const selectedRows = modal.rows.filter((row) => row.include);
        if (selectedRows.length === 0) {
            modal.error = 'Select at least one row to insert.';
            return;
        }

        const invalidSelected = selectedRows.filter((row) => !row.isValid);
        if (invalidSelected.length > 0) {
            modal.error = `There are ${invalidSelected.length} selected invalid row(s). Please fix or uncheck them.`;
            return;
        }

        modal.isInserting = true;

        try {
            const rowsPayload = selectedRows.map((row) => row.payload);
            const result = await InsertRowsBatch(id, modal.tableName, rowsPayload);

            if (result.error) {
                modal.error = result.error;
                return;
            }

            const skippedByUser = modal.rows.length - selectedRows.length;
            const skippedTotal = skippedByUser + (result.skipped || 0);
            if (skippedTotal > 0) {
                toast.value?.success(`Inserted ${result.inserted} row(s), skipped ${skippedTotal} row(s).`);
            } else {
                toast.value?.success(`Inserted ${result.inserted} row(s) from clipboard.`);
            }

            pastePreviewModal.value = null;
            runQuery();
        } catch (e: any) {
            modal.error = e.toString();
        } finally {
            if (pastePreviewModal.value) {
                pastePreviewModal.value.isInserting = false;
            }
        }
    };

    const cancelPastePreview = () => {
        pastePreviewModal.value = null;
    };

    return {
        updateConfirmation,
        insertRowModal,
        pastePreviewModal,
        isEditable,
        initiateQuickUpdate,
        confirmUpdate,
        openInsertRowModal,
        toggleInsertNull,
        confirmInsertRow,
        pasteRowsFromClipboard,
        updatePastePreviewValue,
        togglePastePreviewRow,
        autoFixPastePreviewRows,
        confirmPastePreviewInsert,
        cancelPastePreview,
        getInputType,
        getNumberStep,
        cancelInsertRow: () => { insertRowModal.value = null; },
        cancelUpdate: () => {
            if (activeTab.value) activeTab.value.editingCell = null;
            updateConfirmation.value = null;
        }
    };
}




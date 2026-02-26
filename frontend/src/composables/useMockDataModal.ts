import { reactive } from 'vue';
import type { Ref } from 'vue';

import type { QueryTab } from '../types/dashboard';

interface UseMockDataModalOptions {
    activeTab: Ref<QueryTab | undefined>;
    connectionId: Ref<string>;
    generateId: () => string;
    getTableDefinition: (connectionId: string, tableName: string) => Promise<any[]>;
    executeQuery: (connectionId: string, query: string, requestId: string) => Promise<{ error?: string }>;
    getEscapedTableName: (name: string) => string;
    getEscapedIdentifier: (name: string) => string;
    generateMockSqlValue: (columnDef: any, rowIndex: number) => string;
    onRefreshCurrentTable: () => void;
    onRefreshRowCount: (tableName: string) => Promise<void>;
    onError?: (message: string) => void;
    onSuccess?: (message: string) => void;
}

export function useMockDataModal(options: UseMockDataModalOptions) {
    const mockDataModal = reactive({
        isOpen: false,
        tableName: '',
        rowCount: 10
    });

    const mockDataConfirm = reactive({
        isOpen: false,
        tableName: '',
        rowCount: 10,
        isLoading: false
    });

    const openMockDataModal = () => {
        if (!options.activeTab.value?.tableName) return;
        mockDataModal.tableName = options.activeTab.value.tableName;
        mockDataModal.rowCount = 10;
        mockDataModal.isOpen = true;
    };

    const openMockDataConfirm = () => {
        const rows = Number(mockDataModal.rowCount);
        if (!Number.isFinite(rows) || rows < 1 || rows > 500) {
            options.onError?.('Please enter rows between 1 and 500.');
            return;
        }
        mockDataModal.isOpen = false;
        mockDataConfirm.tableName = mockDataModal.tableName;
        mockDataConfirm.rowCount = Math.floor(rows);
        mockDataConfirm.isOpen = true;
    };

    const confirmMockDataInsert = async () => {
        const tableName = mockDataConfirm.tableName;
        const rowCount = mockDataConfirm.rowCount;
        if (!tableName || rowCount < 1) return;

        mockDataConfirm.isLoading = true;
        try {
            const colDefs = await options.getTableDefinition(options.connectionId.value, tableName);
            const insertable = (colDefs || []).filter((c: any) => !c.autoIncrement);

            if (insertable.length === 0) {
                options.onError?.('No insertable columns found for this table.');
                return;
            }

            const escapedTable = options.getEscapedTableName(tableName);
            const columnList = insertable.map((c: any) => options.getEscapedIdentifier(c.name)).join(', ');
            const valueRows = Array.from({ length: rowCount }, (_, idx) => {
                const values = insertable.map((c: any) => options.generateMockSqlValue(c, idx + 1)).join(', ');
                return `(${values})`;
            }).join(',\n');

            const sql = `INSERT INTO ${escapedTable} (${columnList}) VALUES\n${valueRows};`;
            const res = await options.executeQuery(options.connectionId.value, sql, options.generateId());
            if (res.error) {
                options.onError?.(`Mock insert failed: ${res.error}`);
                return;
            }

            options.onSuccess?.(`Inserted ${rowCount} mock rows into ${tableName}.`);
            mockDataConfirm.isOpen = false;

            if (options.activeTab.value?.tableName === tableName) {
                options.onRefreshCurrentTable();
            } else {
                await options.onRefreshRowCount(tableName);
            }
        } catch (e: any) {
            console.error('Mock insert failed', e);
            options.onError?.(`Mock insert failed: ${String(e)}`);
        } finally {
            mockDataConfirm.isLoading = false;
        }
    };

    return {
        mockDataModal,
        mockDataConfirm,
        openMockDataModal,
        openMockDataConfirm,
        confirmMockDataInsert,
    };
}

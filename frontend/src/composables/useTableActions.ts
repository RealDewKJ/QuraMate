import { reactive } from 'vue';
import type { Ref } from 'vue';

interface UseTableActionsOptions {
    connectionId: Ref<string>;
    dbType: Ref<string>;
    isReadOnly: Ref<boolean>;
    getTargetTable: () => string;
    closeContextMenu: () => void;
    generateId: () => string;
    executeQuery: (connectionId: string, query: string, requestId: string) => Promise<{ error?: string }>;
    getEscapedTableName: (tableName: string) => string;
    loadTables: () => Promise<void>;
    checkRowCount: (tableName: string) => Promise<void>;
    onSuccess?: (message: string) => void;
    onError?: (message: string) => void;
}

export function useTableActions(options: UseTableActionsOptions) {
    const tableActionConfirmation = reactive({
        isOpen: false,
        action: 'TRUNCATE' as 'TRUNCATE' | 'DROP',
        tableName: '',
        isLoading: false
    });

    const handleTruncateTable = () => {
        const tableName = options.getTargetTable();
        if (!tableName) return;

        if (options.isReadOnly.value) {
            options.onError?.('Read-only mode is enabled.');
            options.closeContextMenu();
            return;
        }

        tableActionConfirmation.action = 'TRUNCATE';
        tableActionConfirmation.tableName = tableName;
        tableActionConfirmation.isOpen = true;
        options.closeContextMenu();
    };

    const handleDropTable = () => {
        const tableName = options.getTargetTable();
        if (!tableName) return;

        if (options.isReadOnly.value) {
            options.onError?.('Read-only mode is enabled.');
            options.closeContextMenu();
            return;
        }

        tableActionConfirmation.action = 'DROP';
        tableActionConfirmation.tableName = tableName;
        tableActionConfirmation.isOpen = true;
        options.closeContextMenu();
    };

    const confirmTableAction = async () => {
        if (!tableActionConfirmation.tableName) return;

        tableActionConfirmation.isLoading = true;
        const tableName = tableActionConfirmation.tableName;
        const type = (options.dbType.value || '').toLowerCase();
        const escapedTableName = options.getEscapedTableName(tableName);

        const query = tableActionConfirmation.action === 'TRUNCATE'
            ? ((type.includes('sqlite') || type.includes('libsql'))
                ? `DELETE FROM ${escapedTableName};`
                : `TRUNCATE TABLE ${escapedTableName};`)
            : `DROP TABLE ${escapedTableName};`;

        try {
            const res = await options.executeQuery(options.connectionId.value, query, options.generateId());
            if (res.error) {
                const actionText = tableActionConfirmation.action === 'TRUNCATE' ? 'truncate' : 'drop';
                options.onError?.(`Failed to ${actionText} table: ${res.error}`);
                return;
            }

            if (tableActionConfirmation.action === 'TRUNCATE') {
                options.onSuccess?.(`Table "${tableName}" truncated.`);
                await options.checkRowCount(tableName);
            } else {
                options.onSuccess?.(`Table "${tableName}" dropped.`);
            }

            await options.loadTables();
            tableActionConfirmation.isOpen = false;
        } catch (e) {
            console.error('Failed to process table action', e);
            options.onError?.('Failed to execute table action.');
        } finally {
            tableActionConfirmation.isLoading = false;
        }
    };

    return {
        tableActionConfirmation,
        handleTruncateTable,
        handleDropTable,
        confirmTableAction,
    };
}

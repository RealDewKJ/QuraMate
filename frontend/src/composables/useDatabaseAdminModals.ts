import { reactive } from 'vue';
import type { Ref } from 'vue';

import { DropDatabase, ExportDatabase, GetDatabaseInfo, SelectFolder } from '../../wailsjs/go/main/App';

interface DatabaseInfoShape {
    dbName?: string;
    summary?: Record<string, string>;
}

interface UseDatabaseAdminModalsOptions {
    connectionId: Ref<string>;
    closeContextMenu: () => void;
    onSuccess?: (message: string) => void;
    onError?: (message: string) => void;
    onDatabaseDropped?: () => void;
}

export function useDatabaseAdminModals(options: UseDatabaseAdminModalsOptions) {
    const dbInfoModal = reactive({
        isOpen: false,
        info: null as DatabaseInfoShape | null,
        isLoading: false
    });

    const dropDbConfirmation = reactive({
        isOpen: false,
        dbName: '',
        isLoading: false
    });

    const exportDbModal = reactive({
        isOpen: false,
        folderPath: '',
        format: 'SQL',
        isLoading: false
    });

    const handleBackupExport = async () => {
        options.closeContextMenu();
        const folder = await SelectFolder();
        if (folder) {
            exportDbModal.isOpen = true;
            exportDbModal.folderPath = folder;
        }
    };

    const confirmExportDb = async () => {
        exportDbModal.isLoading = true;
        try {
            const res = await ExportDatabase(options.connectionId.value, exportDbModal.format, exportDbModal.folderPath);
            if (res === 'Success') {
                options.onSuccess?.('Backup successful!');
                exportDbModal.isOpen = false;
            } else {
                options.onError?.(res);
            }
        } catch (e) {
            options.onError?.('Export failed: ' + e);
        } finally {
            exportDbModal.isLoading = false;
        }
    };

    const handleDatabaseInfo = async () => {
        options.closeContextMenu();
        dbInfoModal.isOpen = true;
        dbInfoModal.isLoading = true;
        try {
            dbInfoModal.info = await GetDatabaseInfo(options.connectionId.value);
        } catch (e) {
            options.onError?.('Failed to fetch database info');
        } finally {
            dbInfoModal.isLoading = false;
        }
    };

    const handleDropDatabase = async () => {
        options.closeContextMenu();
        try {
            const info = await GetDatabaseInfo(options.connectionId.value) as DatabaseInfoShape;
            dropDbConfirmation.dbName = (info.dbName || info.summary?.activeDatabase || '').trim();
            dropDbConfirmation.isOpen = true;
        } catch (e) {
            options.onError?.('Could not fetch database name');
        }
    };

    const confirmDropDatabase = async () => {
        dropDbConfirmation.isLoading = true;
        try {
            const res = await DropDatabase(options.connectionId.value, dropDbConfirmation.dbName);
            if (!res) {
                options.onSuccess?.('Database dropped successfully.');
                dropDbConfirmation.isOpen = false;
                options.onDatabaseDropped?.();
            } else {
                options.onError?.(res);
            }
        } catch (e) {
            options.onError?.('Failed to drop database: ' + e);
        } finally {
            dropDbConfirmation.isLoading = false;
        }
    };

    return {
        dbInfoModal,
        dropDbConfirmation,
        exportDbModal,
        handleBackupExport,
        confirmExportDb,
        handleDatabaseInfo,
        handleDropDatabase,
        confirmDropDatabase,
    };
}

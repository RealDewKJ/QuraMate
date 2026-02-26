import { reactive } from 'vue';

interface UseDashboardContextMenusOptions {
    isInsideResults: (target: HTMLElement) => boolean;
    onOutsideSelection: () => void;
}

export function useDashboardContextMenus(options: UseDashboardContextMenusOptions) {
    const contextMenu = reactive({
        showDb: false,
        showFolder: false,
        show: false,
        showRow: false,
        showView: false,
        showRoutine: false,
        position: { x: 0, y: 0 },
        targetTable: '',
        targetRow: null as any,
        targetColumn: '',
        targetFolder: '',
        targetView: '',
        targetRoutine: '',
        targetRoutineType: 'PROCEDURE' as 'PROCEDURE' | 'FUNCTION'
    });

    const closeContextMenu = () => {
        contextMenu.showDb = false;
        contextMenu.showFolder = false;
        contextMenu.show = false;
        contextMenu.showRow = false;
        contextMenu.showView = false;
        contextMenu.showRoutine = false;
    };

    const openDbContextMenu = (event: MouseEvent) => {
        closeContextMenu();
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showDb = true;
    };

    const openFolderContextMenu = (event: MouseEvent, folderName: string) => {
        closeContextMenu();
        contextMenu.targetFolder = folderName;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showFolder = true;
    };

    const openContextMenu = (event: MouseEvent, table: string) => {
        closeContextMenu();
        contextMenu.targetTable = table;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.show = true;
    };

    const handleRowContextMenu = (event: MouseEvent, row: any, col: string) => {
        closeContextMenu();
        contextMenu.targetRow = row;
        contextMenu.targetColumn = col;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showRow = true;
    };

    const openViewContextMenu = (event: MouseEvent, view: string) => {
        closeContextMenu();
        contextMenu.targetView = view;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showView = true;
    };

    const openRoutineContextMenu = (event: MouseEvent, routine: string, type: 'PROCEDURE' | 'FUNCTION') => {
        closeContextMenu();
        contextMenu.targetRoutine = routine;
        contextMenu.targetRoutineType = type;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showRoutine = true;
    };

    const handleGlobalClick = (event: MouseEvent) => {
        const target = event.target as HTMLElement;

        const isInsideResults = options.isInsideResults(target);
        const isInsideContextMenu = target.closest('.context-menu-fixed');

        if (!isInsideResults && !isInsideContextMenu) {
            options.onOutsideSelection();
        }

        closeContextMenu();
    };

    return {
        contextMenu,
        openDbContextMenu,
        openFolderContextMenu,
        openContextMenu,
        handleRowContextMenu,
        openViewContextMenu,
        openRoutineContextMenu,
        closeContextMenu,
        handleGlobalClick,
    };
}

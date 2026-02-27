import { reactive, nextTick } from 'vue';

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
        targetTable: '', targetRow: null as any,
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

    const adjustPosition = async () => {
        await nextTick();
        const menuEls = document.querySelectorAll('.context-menu-fixed');
        if (menuEls.length > 0) {
            // Get the last opened one (or the visible one)
            const menuEl = Array.from(menuEls).find(el => (el as HTMLElement).offsetParent !== null) as HTMLElement || menuEls[menuEls.length - 1] as HTMLElement;
            if (menuEl) {
                const rect = menuEl.getBoundingClientRect();
                let newX = contextMenu.position.x;
                let newY = contextMenu.position.y;
                
                if (rect.right > window.innerWidth) {
                    newX = window.innerWidth - rect.width - 5;
                }
                if (rect.bottom > window.innerHeight) {
                    newY = window.innerHeight - rect.height - 5;
                }
                
                contextMenu.position.x = Math.max(0, newX);
                contextMenu.position.y = Math.max(0, newY);
            }
        }
    };

    const openDbContextMenu = (event: MouseEvent) => {
        closeContextMenu();
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showDb = true;
        adjustPosition();
    };

    const openFolderContextMenu = (event: MouseEvent, folderName: string) => {
        closeContextMenu();
        contextMenu.targetFolder = folderName;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showFolder = true;
        adjustPosition();
    };

    const openContextMenu = (event: MouseEvent, table: string) => {
        closeContextMenu();
        contextMenu.targetTable = table;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.show = true;
        adjustPosition();
    };

    const handleRowContextMenu = (event: MouseEvent, row: any, col: string) => {
        closeContextMenu();
        contextMenu.targetRow = row;
        contextMenu.targetColumn = col;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showRow = true;
        adjustPosition();
    };

    const openViewContextMenu = (event: MouseEvent, view: string) => {
        closeContextMenu();
        contextMenu.targetView = view;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showView = true;
        adjustPosition();
    };

    const openRoutineContextMenu = (event: MouseEvent, routine: string, type: 'PROCEDURE' | 'FUNCTION') => {
        closeContextMenu();
        contextMenu.targetRoutine = routine;
        contextMenu.targetRoutineType = type;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showRoutine = true;
        adjustPosition();
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

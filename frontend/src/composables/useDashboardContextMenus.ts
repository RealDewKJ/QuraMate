import { nextTick, onBeforeUnmount, onMounted, reactive, watch } from 'vue';

interface UseDashboardContextMenusOptions {
    isInsideResults: (target: HTMLElement) => boolean;
    onOutsideSelection: () => void;
}

export interface DashboardContextMenuState {
    showDb: boolean;
    showFolder: boolean;
    show: boolean;
    showRow: boolean;
    showHeader: boolean;
    showView: boolean;
    showRoutine: boolean;
    position: { x: number; y: number };
    targetTable: string;
    targetRow: Record<string, unknown> | null;
    targetColumn: string;
    targetResultSetIndex: number | null;
    targetFolder: string;
    targetView: string;
    targetRoutine: string;
    targetRoutineType: 'PROCEDURE' | 'FUNCTION';
    targetRowIndex: string | number | null;
}

export function useDashboardContextMenus(options: UseDashboardContextMenusOptions) {
    const viewportPadding = 8;
    const contextMenu = reactive<DashboardContextMenuState>({
        showDb: false,
        showFolder: false,
        show: false,
        showRow: false,
        showHeader: false,
        showView: false,
        showRoutine: false,
        position: { x: 0, y: 0 },
        targetTable: '', targetRow: null,
        targetColumn: '',
        targetResultSetIndex: null,
        targetFolder: '',
        targetView: '',
        targetRoutine: '',
        targetRoutineType: 'PROCEDURE' as 'PROCEDURE' | 'FUNCTION',
        targetRowIndex: null as string | number | null
    });

    const closeContextMenu = () => {
        contextMenu.showDb = false;
        contextMenu.showFolder = false;
        contextMenu.show = false;
        contextMenu.showRow = false;
        contextMenu.showHeader = false;
        contextMenu.showView = false;
        contextMenu.showRoutine = false;
    };

    const adjustPosition = async () => {
        await nextTick();
        const menuEls = document.querySelectorAll('.context-menu-fixed');
        if (menuEls.length > 0) {
            const menuEl = Array.from(menuEls).find(el => (el as HTMLElement).offsetParent !== null) as HTMLElement || menuEls[menuEls.length - 1] as HTMLElement;
            if (menuEl) {
                const rect = menuEl.getBoundingClientRect();
                const maxX = Math.max(viewportPadding, window.innerWidth - rect.width - viewportPadding);
                const maxY = Math.max(viewportPadding, window.innerHeight - rect.height - viewportPadding);

                contextMenu.position.x = Math.min(Math.max(viewportPadding, contextMenu.position.x), maxX);
                contextMenu.position.y = Math.min(Math.max(viewportPadding, contextMenu.position.y), maxY);
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

    const handleRowContextMenu = (event: MouseEvent, row: any, col: string, rowIndex?: string | number) => {
        closeContextMenu();
        contextMenu.targetRow = row;
        contextMenu.targetColumn = col;
        contextMenu.targetResultSetIndex = null;
        contextMenu.targetRowIndex = rowIndex ?? null;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showRow = true;
        adjustPosition();
    };

    const openHeaderContextMenu = (event: MouseEvent, col: string, resultSetIndex: number) => {
        closeContextMenu();
        contextMenu.targetRow = null;
        contextMenu.targetColumn = col;
        contextMenu.targetResultSetIndex = resultSetIndex;
        contextMenu.targetRowIndex = null;
        contextMenu.position = { x: event.clientX, y: event.clientY };
        contextMenu.showHeader = true;
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

        if (!isInsideContextMenu) {
            closeContextMenu();
        }
    };

    const hasOpenMenu = () =>
        contextMenu.showDb
        || contextMenu.showFolder
        || contextMenu.show
        || contextMenu.showRow
        || contextMenu.showHeader
        || contextMenu.showView
        || contextMenu.showRoutine;

    const handleViewportChange = () => {
        if (hasOpenMenu()) {
            adjustPosition();
        }
    };

    watch(
        () => [
            contextMenu.showDb,
            contextMenu.showFolder,
            contextMenu.show,
            contextMenu.showRow,
            contextMenu.showHeader,
            contextMenu.showView,
            contextMenu.showRoutine,
        ],
        (states) => {
            if (states.some(Boolean)) {
                adjustPosition();
            }
        }
    );

    onMounted(() => {
        window.addEventListener('resize', handleViewportChange);
        window.addEventListener('scroll', handleViewportChange, true);
    });

    onBeforeUnmount(() => {
        window.removeEventListener('resize', handleViewportChange);
        window.removeEventListener('scroll', handleViewportChange, true);
    });

    return {
        contextMenu,
        openDbContextMenu,
        openFolderContextMenu,
        openContextMenu,
        handleRowContextMenu,
        openHeaderContextMenu,
        openViewContextMenu,
        openRoutineContextMenu,
        closeContextMenu,
        handleGlobalClick,
    };
}

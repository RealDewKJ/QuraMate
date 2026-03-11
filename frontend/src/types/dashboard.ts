import { ResultSet } from './database';

export interface CellEdit {
    rowId: any;
    col: string;
    value: any;
    resultSetIndex?: number;
}

export interface QueryTab {
    id: string;
    name: string;
    tableName?: string;
    query: string;
    resultSets: ResultSet[];
    primaryKeys: string[];
    filters: Record<string, string>;
    sortColumn?: string;
    sortDirection: 'asc' | 'desc' | null;
    error: string;
    isLoading: boolean;
    isExplaining?: boolean;
    explanation?: string;
    isAiExplaining?: boolean;
    aiExplanation?: string;
    queryExecuted: boolean;
    executionTime?: number;
    editingCell?: CellEdit | null;
    isDesignView?: boolean;
    isERView?: boolean;
    relationships?: any[];
    tablesData?: Record<string, { name: string, type: string }[]>;
    activeQueryIds: string[];
    resultViewTab: 'data' | 'messages' | 'analysis';
    completionTime?: string;
    totalRowCount?: number;
    isPartialStats?: boolean;
    fetchTime?: number;
    editorHeight: number;
    columnWidths: Record<string, number>;
    isRoutine?: boolean;
    routineName?: string;
    routineType?: 'PROCEDURE' | 'FUNCTION';
    sqlFilePath?: string;
    isActivityMonitorView?: boolean;
}

export interface ContextMenuState {
    show: boolean;
    showRow: boolean;
    showView: boolean;
    showRoutine: boolean;
    position: { x: number; y: number };
    targetTable: string;
    targetRow: any;
    targetColumn: string;
    targetView: string;
    targetRoutine: string;
    targetRoutineType: 'PROCEDURE' | 'FUNCTION';
}

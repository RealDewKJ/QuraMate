import { markRaw, reactive } from 'vue';
import type { Ref } from 'vue';

import { ExecuteQuery, ExecuteQueryStream, CancelQuery, SaveQueryHistory } from '../../wailsjs/go/main/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

import type { QueryTab } from '../types/dashboard';
import type { ColumnMetadata } from '../types/database';

export interface UseQueryExecutionOptions {
    connectionId: Ref<string>;
    connectionName: Ref<string | undefined>;
    dbType: Ref<string>;
    activeTab: Ref<QueryTab | undefined>;
    safeModeEnabled: Ref<boolean>;
    generateId: () => string;
    getSelectedQuery: () => string;
}

export function useQueryExecution(options: UseQueryExecutionOptions) {
    const activeStreamCleanups = new Set<() => void>();

    const safeModeConfirmation = reactive({
        isOpen: false,
        queryToRun: ''
    });

    const calculateInitialWidth = (meta: ColumnMetadata): number => {
        if (!meta) return 150;

        const type = (meta.type || '').toUpperCase();
        const length = meta.length || 0;

        if (type.includes('BLOB') || length === -1 || length > 1000 || type === 'TEXT' || type === 'LONGTEXT') {
            return 300;
        }

        if (type.includes('CHAR') || type.includes('TEXT') || type.includes('STRING')) {
            if (length > 0) {
                return Math.min(300, Math.max(120, length * 8.5 + 32));
            }
            return 200;
        }

        if (type.includes('INT') || type.includes('DECIMAL') || type.includes('NUMERIC') || type.includes('BIT')) {
            return 120;
        }

        if (type.includes('DATE') || type.includes('TIME') || type.includes('TIMESTAMP')) {
            return 180;
        }

        return 150;
    };

    const splitSqlStatements = (sql: string): string[] => {
        const statements: string[] = [];
        let current = '';
        let inSingleQuote = false;
        let inDoubleQuote = false;
        let inBacktick = false;
        let inLineComment = false;
        let inBlockComment = false;

        for (let i = 0; i < sql.length; i++) {
            const ch = sql[i];
            const next = sql[i + 1];

            if (inLineComment) {
                current += ch;
                if (ch === '\n') inLineComment = false;
                continue;
            }
            if (inBlockComment) {
                current += ch;
                if (ch === '*' && next === '/') {
                    current += '/';
                    i++;
                    inBlockComment = false;
                }
                continue;
            }

            if (!inSingleQuote && !inDoubleQuote && !inBacktick) {
                if (ch === '-' && next === '-') {
                    current += ch;
                    inLineComment = true;
                    continue;
                }
                if (ch === '/' && next === '*') {
                    current += ch;
                    inBlockComment = true;
                    continue;
                }
            }

            if (ch === '\'' && !inDoubleQuote && !inBacktick) {
                inSingleQuote = !inSingleQuote;
                current += ch;
                continue;
            }
            if (ch === '"' && !inSingleQuote && !inBacktick) {
                inDoubleQuote = !inDoubleQuote;
                current += ch;
                continue;
            }
            if (ch === '`' && !inSingleQuote && !inDoubleQuote) {
                inBacktick = !inBacktick;
                current += ch;
                continue;
            }

            if (ch === ';' && !inSingleQuote && !inDoubleQuote && !inBacktick) {
                const trimmed = current.trim();
                if (trimmed) statements.push(trimmed);
                current = '';
                continue;
            }

            current += ch;
        }

        const tail = current.trim();
        if (tail) statements.push(tail);

        return statements;
    };

    const mapRowsToObjects = (columns: string[], rows: any[]) => {
        return (rows || []).map((row: any) => {
            if (Array.isArray(row)) {
                return Object.fromEntries(columns.map((col: string, i: number) => [col, row[i]]));
            }
            return row;
        });
    };

    const runQuery = async (forceBypassSafeMode: boolean = false) => {
        if (!options.activeTab.value) return;

        options.activeTab.value.error = '';
        options.activeTab.value.resultSets = [];
        options.activeTab.value.filters = {};
        options.activeTab.value.queryExecuted = false;
        options.activeTab.value.isLoading = true;
        options.activeTab.value.isExplaining = false;
        options.activeTab.value.explanation = undefined;
        options.activeTab.value.isAiExplaining = false;
        options.activeTab.value.aiExplanation = undefined;
        options.activeTab.value.executionTime = undefined;
        options.activeTab.value.fetchTime = undefined;
        options.activeTab.value.editingCell = null;
        options.activeTab.value.totalRowCount = undefined;
        options.activeTab.value.isPartialStats = false;

        const startTime = performance.now();
        const reqId = options.generateId();
        const tab = options.activeTab.value;
        tab.activeQueryIds.push(reqId);

        let queryToRun = tab.query;
        const selection = options.getSelectedQuery();
        if (selection && selection.trim()) {
            queryToRun = selection;
        }

        if (!forceBypassSafeMode && options.safeModeEnabled.value) {
            let normalizedQuery = queryToRun.replace(/\/\*[\s\S]*?\*\//g, '');
            normalizedQuery = normalizedQuery.replace(/--.*$/gm, '');
            normalizedQuery = normalizedQuery.trim();

            const isRisky = /^(update|delete)\b/i.test(normalizedQuery) && !/\bwhere\b/i.test(normalizedQuery);
            if (isRisky) {
                safeModeConfirmation.queryToRun = queryToRun;
                safeModeConfirmation.isOpen = true;
                tab.isLoading = false;
                tab.activeQueryIds = tab.activeQueryIds.filter((id: string) => id !== reqId);
                return;
            }
        }

        const statements = splitSqlStatements(queryToRun);
        if (statements.length > 1) {

            try {
                if (queryToRun && queryToRun.trim().length > 0) {
                    SaveQueryHistory(queryToRun, options.connectionName.value || options.dbType.value || '');
                }

                for (const statement of statements) {
                    const res = await ExecuteQuery(options.connectionId.value, statement, reqId);
                    if (res.error) {
                        throw new Error(res.error);
                    }

                    const resultSets = Array.isArray(res.resultSets) ? res.resultSets : [];
                    for (const rs of resultSets) {
                        const columns = Array.isArray(rs?.columns) ? rs.columns : [];
                        const rows = mapRowsToObjects(columns, Array.isArray(rs?.rows) ? rs.rows : []);
                        const columnTypes = Array.isArray((rs as any)?.columnTypes) ? (rs as any).columnTypes : [];

                        tab.resultSets.push({
                            columns,
                            columnTypes,
                            rows: markRaw(rows),
                        });

                        columns.forEach((col: string, i: number) => {
                            if (!tab.columnWidths[col]) {
                                tab.columnWidths[col] = calculateInitialWidth(columnTypes[i]);
                            }
                        });
                    }
                }

                tab.queryExecuted = true;
                const hasDataWithRows = tab.resultSets.some((rs: any) => rs.columns && rs.columns.length > 0 && rs.rows && rs.rows.length > 0);
                tab.resultViewTab = hasDataWithRows ? 'data' : 'messages';
                tab.executionTime = Math.round(performance.now() - startTime);
                tab.completionTime = new Date().toLocaleString();
                tab.isLoading = false;
                tab.error = '';
            } catch (e: any) {
                tab.error = e?.toString ? e.toString() : String(e);
                tab.queryExecuted = true;
                tab.resultViewTab = 'messages';
                tab.executionTime = Math.round(performance.now() - startTime);
                tab.completionTime = new Date().toLocaleString();
                tab.isLoading = false;
            } finally {
                tab.activeQueryIds = tab.activeQueryIds.filter((id: string) => id !== reqId);
            }
            return;
        }

        let firstBatchReceived = false;
        let cleanedUp = false;

        const cleanup = () => {
            if (cleanedUp) return;
            cleanedUp = true;
            EventsOff('query:batch:' + reqId);
            EventsOff('query:done:' + reqId);
            EventsOff('query:error:' + reqId);
            EventsOff('query:stats:' + reqId);
            tab.activeQueryIds = tab.activeQueryIds.filter((id: string) => id !== reqId);
            activeStreamCleanups.delete(cleanup);
        };
        activeStreamCleanups.add(cleanup);

        EventsOn('query:stats:' + reqId, (stats: any) => {
            if (tab.activeQueryIds.includes(reqId)) {
                if (stats.phase === 'execution') {
                    tab.executionTime = stats.time;
                } else {
                    if (stats.rows >= 0) {
                        tab.totalRowCount = stats.rows;
                    } else {
                        tab.totalRowCount = undefined;
                    }

                    if (stats.time !== undefined) tab.executionTime = stats.time;
                    if (stats.fetchTime !== undefined) tab.fetchTime = stats.fetchTime;

                    tab.isPartialStats = stats.partial;
                }
            }
        });

        EventsOn('query:batch:' + reqId, (batch: any) => {
            if (!firstBatchReceived) {
                firstBatchReceived = true;
                if (tab.executionTime === undefined) {
                    tab.executionTime = Math.round(performance.now() - startTime);
                }
            }

            const rsIdx = batch.resultSetIdx;
            const columns = batch.columns || [];
            const columnTypes = batch.columnTypes || [];
            const batchRows = batch.rows || [];

            const mappedRows = batchRows.map((row: any[]) =>
                Object.fromEntries(columns.map((col: string, i: number) => [col, row[i]]))
            );

            while (tab.resultSets.length <= rsIdx) {
                tab.resultSets.push({ columns: [], columnTypes: [], rows: [] });
            }

            const rs = tab.resultSets[rsIdx];
            if (columns.length > 0 && rs.columns.length === 0) {
                rs.columns = columns;
                rs.columnTypes = columnTypes;

                columns.forEach((col: string, i: number) => {
                    if (!tab.columnWidths[col]) {
                        const meta = columnTypes[i];
                        tab.columnWidths[col] = calculateInitialWidth(meta);
                    }
                });
            }
            rs.rows = markRaw(rs.rows.concat(mappedRows));

            if (!tab.queryExecuted) {
                tab.queryExecuted = true;
                tab.resultViewTab = (columns.length > 0 && mappedRows.length > 0) ? 'data' : 'messages';
            }
        });

        EventsOn('query:done:' + reqId, () => {
            tab.isLoading = false;
            tab.completionTime = new Date().toLocaleString();
            if (!tab.queryExecuted) {
                tab.queryExecuted = true;
                const hasDataWithRows = tab.resultSets.some((rs: any) => rs.columns && rs.columns.length > 0 && rs.rows && rs.rows.length > 0);
                tab.resultViewTab = hasDataWithRows ? 'data' : 'messages';
            }
            if (!firstBatchReceived) {
                if (tab.executionTime === undefined) {
                    tab.executionTime = Math.round(performance.now() - startTime);
                }
            }
            cleanup();
        });

        EventsOn('query:error:' + reqId, (errMsg: string) => {
            tab.error = errMsg;
            tab.isLoading = false;
            tab.queryExecuted = true;
            tab.executionTime = Math.round(performance.now() - startTime);
            tab.completionTime = new Date().toLocaleString();
            tab.resultViewTab = 'messages';
            cleanup();
        });

        try {
            if (queryToRun && queryToRun.trim().length > 0) {
                SaveQueryHistory(queryToRun, options.connectionName.value || options.dbType.value || '');
            }
            const err = await ExecuteQueryStream(options.connectionId.value, queryToRun, reqId);
            if (err) {
                tab.error = err;
                tab.isLoading = false;
                tab.queryExecuted = true;
                tab.resultViewTab = 'messages';
                cleanup();
            }
        } catch (e: any) {
            tab.error = e.toString();
            tab.isLoading = false;
            tab.executionTime = Math.round(performance.now() - startTime);
            cleanup();
        }
    };

    const stopQuery = async () => {
        if (!options.activeTab.value) return;
        const tab = options.activeTab.value;

        try {
            const ids = [...tab.activeQueryIds];
            await Promise.all(ids.map((id) => CancelQuery(id)));
        } catch (e) {
            console.error('Error stopping query:', e);
        }
    };

    const confirmSafeModeQuery = () => {
        safeModeConfirmation.isOpen = false;
        void runQuery(true);
    };

    const cancelSafeModeQuery = () => {
        safeModeConfirmation.isOpen = false;
        if (options.activeTab.value) {
            options.activeTab.value.isLoading = false;
        }
    };

    const cleanupAllStreams = () => {
        activeStreamCleanups.forEach((cleanup) => cleanup());
        activeStreamCleanups.clear();
    };

    return {
        safeModeConfirmation,
        runQuery,
        stopQuery,
        confirmSafeModeQuery,
        cancelSafeModeQuery,
        cleanupAllStreams,
    };
}

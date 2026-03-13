import { markRaw, reactive } from 'vue';
import type { Ref } from 'vue';

import { CancelQuery, ExecuteQuery, ExecuteQueryStream, LogClientEvent, SaveQueryHistory } from '../../wailsjs/go/app/App';
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime';

import type { QueryTab } from '../types/dashboard';
import type { ColumnMetadata } from '../types/database';

export interface UseQueryExecutionOptions {
    connectionId: Ref<string>;
    connectionName: Ref<string | undefined>;
    dbType: Ref<string>;
    activeTab: Ref<QueryTab | undefined>;
    safeModeEnabled: Ref<boolean>;
    perfLoggingEnabled: Ref<boolean>;
    queryHistoryEnabled: Ref<boolean>;
    queryHistoryRetentionDays: Ref<number>;
    generateId: () => string;
    getSelectedQuery: () => string;
}

export function useQueryExecution(options: UseQueryExecutionOptions) {
    const activeStreamCleanups = new Set<() => void>();
    const logPerf = (event: 'success' | 'error' | 'cancelled', payload: Record<string, unknown>) => {
        if (!options.perfLoggingEnabled.value) {
            return;
        }

        const serializedPayload = JSON.stringify(payload);
        const compactPayload = serializedPayload.length > 1500
            ? `${serializedPayload.slice(0, 1500)}...`
            : serializedPayload;
        const message = `[QueryPerf] ${event} ${compactPayload}`;

        void LogClientEvent('INFO', message).catch((err: unknown) => {
            console.warn('Failed to write query perf event to app logs', err);
        });
        console.info(message);
    };

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
    const createUniqueColumns = (columns: string[]): string[] => {
        const seen = new Map<string, number>();
        return columns.map((column) => {
            const base = String(column ?? '');
            const count = (seen.get(base) || 0) + 1;
            seen.set(base, count);
            return count === 1 ? base : `${base} (${count})`;
        });
    };
    const mapRowsToObjects = (columns: string[], rows: any[]) => {
        const mapped = (rows || []).map((row: any) => {
            if (Array.isArray(row)) {
                return markRaw(Object.fromEntries(columns.map((col: string, i: number) => [col, row[i]])));
            }
            return markRaw(row);
        });
        return markRaw(mapped);
    };

    const runQuery = async (forceBypassSafeMode: boolean = false) => {
        if (!options.activeTab.value) return;

        options.activeTab.value.error = '';
        options.activeTab.value.resultSets = markRaw([]);
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
        
        try {
            if (queryToRun && queryToRun.trim().length > 0) {
                SaveQueryHistory(
                    queryToRun,
                    options.connectionName.value || options.dbType.value || '',
                    options.queryHistoryEnabled.value,
                    options.queryHistoryRetentionDays.value
                )
                    .then((result: { success?: boolean; error?: string }) => {
                        if (result && result.success === false) {
                            console.warn('Save query history failed:', result.error || 'unknown error');
                        }
                    })
                    .catch((err: unknown) => {
                        console.warn('Save query history failed:', err);
                    });
            }

            let totalExecutionTime = 0;
            let totalFetchTime = 0;
            let totalRows = 0;
            let isFirstStatementDataReceived = false;

            for (let idx = 0; idx < statements.length; idx++) {
                if (!tab.isLoading) break; // Abort if cancelled by user
                
                const statement = statements[idx];
                const statementReqId = statements.length > 1 ? `${reqId}-${idx}` : reqId;
                
                if (statements.length > 1) {
                    tab.activeQueryIds.push(statementReqId);
                }

                await new Promise<void>((resolve, reject) => {
                    let firstBatchReceived = false;
                    let cleanedUp = false;
                    const statementStartTime = performance.now();
                    const rsOffset = tab.resultSets.length;

                    const cleanup = () => {
                        if (cleanedUp) return;
                        cleanedUp = true;
                        EventsOff('query:batch:' + statementReqId);
                        EventsOff('query:done:' + statementReqId);
                        EventsOff('query:error:' + statementReqId);
                        EventsOff('query:stats:' + statementReqId);
                        tab.activeQueryIds = tab.activeQueryIds.filter((id: string) => id !== statementReqId);
                        activeStreamCleanups.delete(cleanup);
                    };
                    activeStreamCleanups.add(cleanup);

                    EventsOn('query:stats:' + statementReqId, (stats: any) => {
                        if (tab.activeQueryIds.includes(statementReqId) || statements.length === 1) {
                            if (stats.phase === 'execution') {
                                tab.executionTime = totalExecutionTime + stats.time;
                            } else {
                                if (stats.rows >= 0) {
                                    tab.totalRowCount = totalRows + stats.rows;
                                }

                                if (stats.time !== undefined) {
                                    tab.executionTime = totalExecutionTime + stats.time;
                                }
                                if (stats.fetchTime !== undefined) {
                                    tab.fetchTime = totalFetchTime + stats.fetchTime;
                                }

                                tab.isPartialStats = stats.partial;
                            }
                        }
                    });

                    EventsOn('query:batch:' + statementReqId, (batch: any) => {
                        if (!firstBatchReceived) {
                            firstBatchReceived = true;
                            if (tab.executionTime === undefined) {
                                tab.executionTime = totalExecutionTime + Math.round(performance.now() - statementStartTime);
                            }
                        }

                        const rsIdx = rsOffset + batch.resultSetIdx;
                        const columns = batch.columns || [];
                        const uniqueColumns = createUniqueColumns(columns);
                        const columnTypes = batch.columnTypes || [];
                        const batchRows = batch.rows || [];

                        const mappedRows = mapRowsToObjects(uniqueColumns, batchRows);

                        while (tab.resultSets.length <= rsIdx) {
                            tab.resultSets.push(markRaw({ columns: markRaw([]), columnTypes: markRaw([]), rows: markRaw([]) }));
                        }

                        const rs = tab.resultSets[rsIdx];
                        if (columns.length > 0 && rs.columns.length === 0) {
                            rs.columns = markRaw(uniqueColumns.slice());
                            rs.columnTypes = markRaw(columnTypes.slice());

                            uniqueColumns.forEach((col: string, i: number) => {
                                if (!tab.columnWidths[col]) {
                                    const meta = columnTypes[i];
                                    tab.columnWidths[col] = calculateInitialWidth(meta);
                                }
                            });
                        }
                        rs.rows = markRaw((rs.rows || []).concat(mappedRows));

                        if (!tab.queryExecuted || (!isFirstStatementDataReceived && mappedRows.length > 0)) {
                            tab.queryExecuted = true;
                            if (mappedRows.length > 0) {
                                isFirstStatementDataReceived = true;
                            }
                            tab.resultViewTab = (columns.length > 0 && isFirstStatementDataReceived) ? 'data' : 'messages';
                        }
                    });

                    EventsOn('query:done:' + statementReqId, () => {
                        if (!tab.queryExecuted) {
                            tab.queryExecuted = true;
                            const hasDataWithRows = tab.resultSets.some((rs: any) => rs.columns && rs.columns.length > 0 && rs.rows && rs.rows.length > 0);
                            tab.resultViewTab = hasDataWithRows ? 'data' : 'messages';
                        }
                        if (!firstBatchReceived) {
                            if (tab.executionTime === undefined || statements.length > 1) {
                                tab.executionTime = totalExecutionTime + Math.round(performance.now() - statementStartTime);
                            }
                        }
                        
                        // Commit stats to running totals
                        totalExecutionTime = tab.executionTime || 0;
                        totalFetchTime = tab.fetchTime || 0;
                        totalRows = tab.totalRowCount || 0;

                        cleanup();
                        resolve();
                    });

                    EventsOn('query:error:' + statementReqId, (errMsg: string) => {
                        cleanup();
                        reject(new Error(errMsg));
                    });

                    ExecuteQueryStream(options.connectionId.value, statement, statementReqId).then((err: string) => {
                        if (err) {
                            cleanup();
                            reject(new Error(err));
                        }
                    }).catch((e: unknown) => {
                        cleanup();
                        reject(e);
                    });
                });
            }

            // All statements completed successfully
            tab.isLoading = false;
            tab.completionTime = new Date().toLocaleString();
            if (!tab.queryExecuted) {
                tab.queryExecuted = true;
                const hasDataWithRows = tab.resultSets.some((rs: any) => rs.columns && rs.columns.length > 0 && rs.rows && rs.rows.length > 0);
                tab.resultViewTab = hasDataWithRows ? 'data' : 'messages';
            }

        } catch (e: any) {
            tab.error = e.toString().replace(/^Error:\s*/, '');
            tab.isLoading = false;
            tab.executionTime = Math.round(performance.now() - startTime);
            tab.completionTime = new Date().toLocaleString();
            tab.queryExecuted = true;
            tab.resultViewTab = 'data';
            const elapsedMs = Math.round(performance.now() - startTime);
            const resultSetCount = tab.resultSets?.length || 0;
            const totalRowsFromSets = (tab.resultSets || []).reduce((sum: number, rs: any) => sum + (rs.rows?.length || 0), 0);
            logPerf('error', {
                connection: options.connectionName.value || options.dbType.value || '',
                statements: statements.length,
                resultSets: resultSetCount,
                rows: totalRowsFromSets,
                elapsedMs,
                executionMs: tab.executionTime,
                fetchMs: tab.fetchTime,
                error: tab.error,
            });
        } finally {
            const elapsedMs = Math.round(performance.now() - startTime);
            const resultSetCount = tab.resultSets?.length || 0;
            const totalRowsFromSets = (tab.resultSets || []).reduce((sum: number, rs: any) => sum + (rs.rows?.length || 0), 0);
            const hasError = !!tab.error;
            if (!hasError) {
                const event: 'success' | 'cancelled' = tab.queryExecuted ? 'success' : 'cancelled';
                logPerf(event, {
                    connection: options.connectionName.value || options.dbType.value || '',
                    statements: statements.length,
                    resultSets: resultSetCount,
                    rows: totalRowsFromSets,
                    elapsedMs,
                    executionMs: tab.executionTime,
                    fetchMs: tab.fetchTime,
                    totalRowCount: tab.totalRowCount,
                    partialStats: tab.isPartialStats,
                });
            }
            tab.activeQueryIds = tab.activeQueryIds.filter((id: string) => !id.startsWith(reqId));
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


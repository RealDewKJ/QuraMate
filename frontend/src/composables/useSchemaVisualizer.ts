import type { QueryTab } from '../types/dashboard';

type ExecuteQueryResult = {
    resultSets?: any[];
    error?: string;
};

type SchemaVisualizerDeps = {
    connectionId: string;
    dbType: string;
    executeQuery: (connectionId: string, query: string, queryId: string) => Promise<ExecuteQueryResult>;
    getTables: (connectionId: string) => Promise<string[]>;
    getForeignKeys: (connectionId: string, tableName: string) => Promise<any[]>;
    generateId: () => string;
};

export const useSchemaVisualizer = (deps: SchemaVisualizerDeps) => {
    const generateDatabaseERDiagram = async (schemaTab: QueryTab): Promise<void> => {
        schemaTab.name = 'Schema Visualizer';
        schemaTab.isERView = true;
        schemaTab.tableName = '';
        schemaTab.relationships = [];
        schemaTab.tablesData = {};
        schemaTab.error = '';
        schemaTab.isLoading = true;

        try {
            const type = (deps.dbType || '').toLowerCase();
            let allTables: string[] = [];
            const debug = {
                connectionId: deps.connectionId,
                dbType: type,
                tableSources: {
                    sysTables: 0,
                    informationSchemaTables: 0,
                    getTables: 0,
                },
                columnSources: {
                    sysColumnsRows: 0,
                    informationSchemaColumnsRows: 0,
                    perTableFallbackAttempts: 0,
                    perTableFallbackHits: 0,
                    perTableFallbackErrors: 0,
                },
                fkSources: {
                    attempts: 0,
                    errors: 0,
                },
                samplePerTableFallbackErrors: [] as string[],
                sampleForeignKeyErrors: [] as string[],
                unmatchedColumnTableKeys: new Set<string>(),
            };

            if (type.includes('mssql') || type.includes('sqlserver')) {
                const tableSet = new Set<string>();
                const collectTableNames = (resultSets: any[] | undefined): string[] => {
                    if (!resultSets || resultSets.length === 0) {
                        return [];
                    }
                    const rs = resultSets[0];
                    const rows = (rs.rows || []).map((row: any) => {
                        if (Array.isArray(row)) {
                            return Object.fromEntries((rs.columns || []).map((col: string, i: number) => [col, row[i]]));
                        }
                        return row && typeof row === 'object' ? row : {};
                    });
                    return rows
                        .map((r: any) => (r.full_name || r.FULL_NAME || r.table_name || r.TABLE_NAME || '').toString())
                        .filter(Boolean);
                };

                const reqId = deps.generateId();
                schemaTab.activeQueryIds.push(reqId);
                try {
                    const tableRes = await deps.executeQuery(
                        deps.connectionId,
                        "SELECT CONCAT(s.name, '.', t.name) AS full_name FROM sys.tables t INNER JOIN sys.schemas s ON s.schema_id = t.schema_id WHERE t.is_ms_shipped = 0 ORDER BY s.name, t.name",
                        reqId
                    );
                    if (!tableRes.error) {
                        const names = collectTableNames(tableRes.resultSets);
                        debug.tableSources.sysTables = names.length;
                        names.forEach((name) => tableSet.add(name));
                    }
                } finally {
                    schemaTab.activeQueryIds = schemaTab.activeQueryIds.filter((id) => id !== reqId);
                }

                const reqId2 = deps.generateId();
                schemaTab.activeQueryIds.push(reqId2);
                try {
                    const tableRes2 = await deps.executeQuery(
                        deps.connectionId,
                        "SELECT CONCAT(TABLE_SCHEMA, '.', TABLE_NAME) AS full_name FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_TYPE = 'BASE TABLE' ORDER BY TABLE_SCHEMA, TABLE_NAME",
                        reqId2
                    );
                    if (!tableRes2.error) {
                        const names = collectTableNames(tableRes2.resultSets);
                        debug.tableSources.informationSchemaTables = names.length;
                        names.forEach((name) => tableSet.add(name));
                    }
                } finally {
                    schemaTab.activeQueryIds = schemaTab.activeQueryIds.filter((id) => id !== reqId2);
                }

                const fetchedTables = await deps.getTables(deps.connectionId);
                debug.tableSources.getTables = (fetchedTables || []).filter(Boolean).length;
                (fetchedTables || []).filter(Boolean).forEach((name: string) => tableSet.add(name));

                const mergedTables = Array.from(tableSet).filter(Boolean);
                const qualifiedByShort = new Set(
                    mergedTables
                        .filter((name) => name.includes('.'))
                        .map((name) => name.split('.').slice(-1)[0].toLowerCase())
                );
                allTables = mergedTables
                    .filter((name) => name.includes('.') || !qualifiedByShort.has(name.toLowerCase()))
                    .sort((a, b) => a.localeCompare(b));
            } else {
                const fetchedTables = await deps.getTables(deps.connectionId);
                allTables = (fetchedTables || [])
                    .filter(Boolean)
                    .sort((a: string, b: string) => a.localeCompare(b));
            }

            if (allTables.length === 0) {
                schemaTab.error = 'No tables found in this database.';
                return;
            }

            schemaTab.tableName = allTables[0];

            const escapeSqlLiteral = (value: string) => value.replace(/'/g, "''");
            const getSchemaQuery = (tbl: string) => {
                if (type.includes('mssql') || type.includes('sqlserver')) {
                    const dot = tbl.indexOf('.');
                    if (dot > 0) {
                        const schemaName = tbl.slice(0, dot);
                        const tableName = tbl.slice(dot + 1);
                        const escapedSchema = schemaName.replace(/]/g, ']]');
                        const escapedTable = tableName.replace(/]/g, ']]');
                        return `SELECT c.name AS COLUMN_NAME, typ.name AS DATA_TYPE
FROM sys.columns c
INNER JOIN sys.types typ ON c.user_type_id = typ.user_type_id
WHERE c.object_id = OBJECT_ID(N'[${escapedSchema}].[${escapedTable}]')
ORDER BY c.column_id`;
                    }
                    return `SELECT COLUMN_NAME, DATA_TYPE FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = '${tbl}'`;
                } else if (type.includes('postgres') || type.includes('greenplum') || type.includes('redshift') || type.includes('cockroach')) {
                    return `SELECT column_name, data_type FROM information_schema.columns WHERE table_name = '${tbl}'`;
                } else if (type.includes('mysql') || type.includes('maria') || type.includes('databend')) {
                    return `DESCRIBE ${tbl}`;
                } else if (type.includes('sqlite') || type.includes('libsql')) {
                    return `SELECT name AS column_name, type AS data_type FROM pragma_table_info('${escapeSqlLiteral(tbl)}')`;
                }
                return `SELECT * FROM ${tbl} LIMIT 1`;
            };

            const fkLists: any[][] = [];
            for (const tbl of allTables) {
                debug.fkSources.attempts += 1;
                try {
                    const normalizedTable = type.includes('mssql') && tbl.includes('.')
                        ? tbl.split('.').slice(1).join('.')
                        : tbl;
                    const fks = (await deps.getForeignKeys(deps.connectionId, normalizedTable)) || [];
                    fkLists.push(fks);
                } catch (e: any) {
                    debug.fkSources.errors += 1;
                    if (debug.sampleForeignKeyErrors.length < 20) {
                        const msg = e?.toString?.() || String(e);
                        debug.sampleForeignKeyErrors.push(`${tbl}: ${msg}`);
                    }
                    fkLists.push([]);
                }
            }

            const fkMap = new Map<string, any>();
            fkLists.flat().forEach((fk: any) => {
                const key = [fk.table, fk.column, fk.refTable, fk.refColumn, fk.constraint || ''].join('|');
                if (!fkMap.has(key)) {
                    fkMap.set(key, fk);
                }
            });
            schemaTab.relationships = Array.from(fkMap.values());

            const tablesData: Record<string, { name: string; type: string }[]> = {};
            const tableNameLookup = new Map<string, string[]>();
            allTables.forEach((tbl) => {
                tablesData[tbl] = [];
                const shortName = tbl.includes('.') ? tbl.split('.').slice(-1)[0] : tbl;
                const shortKey = shortName.toLowerCase();
                const bucket = tableNameLookup.get(shortKey) || [];
                bucket.push(tbl);
                tableNameLookup.set(shortKey, bucket);
            });

            if (type.includes('mssql') || type.includes('sqlserver')) {
                const resolveTableKey = (key: string): string | null => {
                    if (tablesData[key]) {
                        return key;
                    }

                    const lowerKey = key.toLowerCase();
                    const caseInsensitiveExact = Object.keys(tablesData).find((k) => k.toLowerCase() === lowerKey);
                    if (caseInsensitiveExact) {
                        return caseInsensitiveExact;
                    }

                    const shortName = key.includes('.') ? key.split('.').slice(-1)[0] : key;
                    const candidates = tableNameLookup.get(shortName.toLowerCase()) || [];
                    if (candidates.length === 1) {
                        return candidates[0];
                    }

                    if (key.includes('.')) {
                        const schema = key.split('.')[0].toLowerCase();
                        const schemaMatch = candidates.find((c) => c.includes('.') && c.split('.')[0].toLowerCase() === schema);
                        if (schemaMatch) {
                            return schemaMatch;
                        }
                    }

                    if (candidates.length > 1) {
                        return [...candidates].sort((a, b) => a.localeCompare(b))[0];
                    }
                    return null;
                };

                const normalizeRows = (resultSet: any) => {
                    return (resultSet.rows || []).map((row: any) => {
                        if (Array.isArray(row)) {
                            return Object.fromEntries((resultSet.columns || []).map((col: string, i: number) => [col, row[i]]));
                        }
                        return row && typeof row === 'object' ? row : {};
                    });
                };

                const applyColumnRows = (rows: any[]) => {
                    rows.forEach((row: any) => {
                        const key = (row.table_name || row.TABLE_NAME || '').toString();
                        const columnName = (row.column_name || row.COLUMN_NAME || '').toString();
                        if (!key || !columnName) {
                            return;
                        }

                        const targetKey = resolveTableKey(key);
                        if (!targetKey) {
                            debug.unmatchedColumnTableKeys.add(key);
                            return;
                        }

                        tablesData[targetKey].push({
                            name: columnName,
                            type: (row.data_type || row.DATA_TYPE || 'string').toString(),
                        });
                    });
                };

                const reqId3 = deps.generateId();
                schemaTab.activeQueryIds.push(reqId3);
                try {
                    const res3 = await deps.executeQuery(
                        deps.connectionId,
                        "SELECT CONCAT(s.name, '.', t.name) AS table_name, c.name AS column_name, typ.name AS data_type FROM sys.tables t INNER JOIN sys.schemas s ON s.schema_id = t.schema_id LEFT JOIN sys.columns c ON c.object_id = t.object_id LEFT JOIN sys.types typ ON c.user_type_id = typ.user_type_id WHERE t.is_ms_shipped = 0 ORDER BY s.name, t.name, c.column_id",
                        reqId3
                    );
                    if (!res3.error && res3.resultSets && res3.resultSets.length > 0) {
                        const rows = normalizeRows(res3.resultSets[0]);
                        debug.columnSources.sysColumnsRows = rows.length;
                        applyColumnRows(rows);
                    }
                } finally {
                    schemaTab.activeQueryIds = schemaTab.activeQueryIds.filter((id) => id !== reqId3);
                }

                let missingTablesCount = Object.values(tablesData).filter((cols) => cols.length === 0).length;
                if (missingTablesCount > 0) {
                    const reqId4 = deps.generateId();
                    schemaTab.activeQueryIds.push(reqId4);
                    try {
                        const res4 = await deps.executeQuery(
                            deps.connectionId,
                            "SELECT CONCAT(TABLE_SCHEMA, '.', TABLE_NAME) AS table_name, COLUMN_NAME AS column_name, DATA_TYPE AS data_type FROM INFORMATION_SCHEMA.COLUMNS ORDER BY TABLE_SCHEMA, TABLE_NAME, ORDINAL_POSITION",
                            reqId4
                        );
                        if (!res4.error && res4.resultSets && res4.resultSets.length > 0) {
                            const rows = normalizeRows(res4.resultSets[0]);
                            debug.columnSources.informationSchemaColumnsRows = rows.length;
                            applyColumnRows(rows);
                        }
                    } finally {
                        schemaTab.activeQueryIds = schemaTab.activeQueryIds.filter((id) => id !== reqId4);
                    }
                    missingTablesCount = Object.values(tablesData).filter((cols) => cols.length === 0).length;
                }

                if (missingTablesCount > 0) {
                    const missingTables = allTables.filter((tbl) => (tablesData[tbl] || []).length === 0);
                    debug.columnSources.perTableFallbackAttempts = missingTables.length;

                    for (const tbl of missingTables) {
                        const reqId = deps.generateId();
                        schemaTab.activeQueryIds.push(reqId);
                        try {
                            const res = await deps.executeQuery(deps.connectionId, getSchemaQuery(tbl), reqId);
                            if (res.error || !res.resultSets || res.resultSets.length === 0) {
                                debug.columnSources.perTableFallbackErrors += 1;
                                if (res.error && debug.samplePerTableFallbackErrors.length < 20) {
                                    debug.samplePerTableFallbackErrors.push(`${tbl}: ${res.error}`);
                                }
                                continue;
                            }
                            const rs = res.resultSets[0];
                            const rows = (rs.rows || []).map((row: any) => {
                                if (Array.isArray(row)) {
                                    return Object.fromEntries((rs.columns || []).map((col: string, i: number) => [col, row[i]]));
                                }
                                return row && typeof row === 'object' ? row : {};
                            });
                            tablesData[tbl] = rows.map((col: any) => ({
                                name: col.COLUMN_NAME || col.column_name || col.Field || col.field || col.name || col.Name || col.column || 'unknown',
                                type: col.DATA_TYPE || col.data_type || col.Type || col.type || col.dataType || 'string',
                            }));
                            if (tablesData[tbl].length > 0) {
                                debug.columnSources.perTableFallbackHits += 1;
                            }
                        } catch (e: any) {
                            debug.columnSources.perTableFallbackErrors += 1;
                            if (debug.samplePerTableFallbackErrors.length < 20) {
                                const msg = e?.toString?.() || String(e);
                                debug.samplePerTableFallbackErrors.push(`${tbl}: ${msg}`);
                            }
                        } finally {
                            schemaTab.activeQueryIds = schemaTab.activeQueryIds.filter((id) => id !== reqId);
                        }
                    }
                }
            } else {
                const schemaPromises = allTables.map(async (tbl) => {
                    const reqId = deps.generateId();
                    schemaTab.activeQueryIds.push(reqId);
                    try {
                        const res = await deps.executeQuery(deps.connectionId, getSchemaQuery(tbl), reqId);
                        if (res.error || !res.resultSets || res.resultSets.length === 0) {
                            return;
                        }

                        const rs = res.resultSets[0];
                        const rows = (rs.rows || []).map((row: any) => {
                            if (Array.isArray(row)) {
                                return Object.fromEntries((rs.columns || []).map((col: string, i: number) => [col, row[i]]));
                            }
                            return row && typeof row === 'object' ? row : {};
                        });

                        tablesData[tbl] = rows.map((col: any) => ({
                            name: col.COLUMN_NAME || col.column_name || col.Field || col.field || col.name || col.Name || col.column || 'unknown',
                            type: col.DATA_TYPE || col.data_type || col.Type || col.type || col.dataType || 'string',
                        }));
                    } finally {
                        schemaTab.activeQueryIds = schemaTab.activeQueryIds.filter((id) => id !== reqId);
                    }
                });

                await Promise.all(schemaPromises);
            }

            schemaTab.tablesData = tablesData;

            const totalLoadedColumns = Object.values(tablesData).reduce((sum, cols) => sum + cols.length, 0);
            const tablesWithColumns = Object.values(tablesData).filter((cols) => cols.length > 0).length;
            const tablesWithoutColumns = Object.keys(tablesData).filter((name) => (tablesData[name] || []).length === 0);
            console.info('[SchemaVisualizerDebug]', {
                connectionId: debug.connectionId,
                dbType: debug.dbType,
                totalTables: allTables.length,
                tableSources: debug.tableSources,
                totalLoadedColumns,
                tablesWithColumns,
                tablesWithoutColumnsCount: tablesWithoutColumns.length,
                sampleTablesWithoutColumns: tablesWithoutColumns.slice(0, 20),
                columnSources: debug.columnSources,
                samplePerTableFallbackErrors: debug.samplePerTableFallbackErrors,
                fkSources: debug.fkSources,
                sampleForeignKeyErrors: debug.sampleForeignKeyErrors,
                unmatchedColumnTableKeysCount: debug.unmatchedColumnTableKeys.size,
                sampleUnmatchedColumnTableKeys: Array.from(debug.unmatchedColumnTableKeys).slice(0, 20),
            });
            if (totalLoadedColumns === 0) {
                schemaTab.error = 'Could not load table columns. Please verify the connection is still active and your DB user can read table metadata.';
            }
        } catch (e: any) {
            schemaTab.error = e?.toString?.() || 'Failed to generate schema visualizer data.';
        } finally {
            schemaTab.isLoading = false;
        }
    };

    return {
        generateDatabaseERDiagram,
    };
};

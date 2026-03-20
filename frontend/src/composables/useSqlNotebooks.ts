import { computed, ref, toValue, watch, type MaybeRefOrGetter } from 'vue';

import { LoadSetting, SaveSetting } from '../../wailsjs/go/app/App';
import {
    createSqlNotebook,
    createSqlNotebookCell,
    type SqlNotebookExecutionState,
    type SqlNotebook,
    type SqlNotebookCell,
    type SqlNotebookCellType,
    type SqlNotebookMetadata,
    type SqlNotebookShareBundle,
    type SqlNotebookTemplatePreset,
    type SqlNotebookResultSnapshot,
    type SqlNotebookVariable,
} from '../types/sqlNotebook';

interface UseSqlNotebooksOptions {
    storageKey: MaybeRefOrGetter<string>;
    dbType: MaybeRefOrGetter<string>;
    connectionName: MaybeRefOrGetter<string | undefined>;
}

interface PersistedSqlNotebookState {
    version: number;
    activeNotebookId: string | null;
    activeCellIdByNotebook: Record<string, string | null>;
    notebooks: SqlNotebook[];
}

const SQL_NOTEBOOKS_STORAGE_VERSION = 1;

const normalizeSearchText = (value: string): string => {
    return value
        .toLowerCase()
        .replace(/[`"'()[\]{}.,;:_/-]+/g, ' ')
        .replace(/\s+/g, ' ')
        .trim();
};

const fuzzyIncludes = (source: string, keyword: string): boolean => {
    if (!keyword) {
        return true;
    }

    if (source.includes(keyword)) {
        return true;
    }

    const sourceTokens = source.split(' ').filter(Boolean);
    const keywordTokens = keyword.split(' ').filter(Boolean);
    if (
        keywordTokens.length > 0 &&
        keywordTokens.every((token) =>
            sourceTokens.some((sourceToken) => sourceToken.includes(token) || fuzzyIncludes(sourceToken, token)),
        )
    ) {
        return true;
    }

    let sourceIndex = 0;
    for (const char of keyword.replace(/\s+/g, '')) {
        sourceIndex = source.indexOf(char, sourceIndex);
        if (sourceIndex === -1) {
            return false;
        }
        sourceIndex += 1;
    }

    return true;
};

const extractPotentialTableNames = (content: string): string[] => {
    const matches = content.match(/\b(?:from|join|update|into|table)\s+([a-zA-Z0-9_.[\]"]+)/gi) || [];
    return matches
        .map((match) => match.replace(/\b(?:from|join|update|into|table)\s+/i, '').trim())
        .filter(Boolean);
};

const buildNotebookSearchIndex = (notebook: SqlNotebook): string => {
    const parts = [
        notebook.title,
        notebook.description,
        ...notebook.tags,
        notebook.metadata.environment,
        notebook.metadata.purpose,
        notebook.metadata.owner,
        notebook.connectionScope.connectionName,
        notebook.connectionScope.dbType,
    ];

    for (const variable of notebook.variables) {
        parts.push(variable.key, variable.label, variable.value);
    }

    for (const cell of notebook.cells) {
        parts.push(cell.title, cell.content);
        if (cell.type === 'sql') {
            parts.push(...extractPotentialTableNames(cell.content));
        }
    }

    return normalizeSearchText(parts.join(' '));
};

const normalizeCell = (cell: Partial<SqlNotebookCell> | null | undefined): SqlNotebookCell | null => {
    if (!cell || typeof cell.id !== 'string' || typeof cell.type !== 'string') {
        return null;
    }

    if (cell.type !== 'sql' && cell.type !== 'markdown' && cell.type !== 'runbook') {
        return null;
    }

    const executionState: SqlNotebookExecutionState =
        cell.executionState === 'running'
            || cell.executionState === 'success'
            || cell.executionState === 'error'
            || cell.executionState === 'verified'
            || cell.executionState === 'skipped'
            ? cell.executionState
            : 'idle';

    return {
        id: cell.id,
        type: cell.type,
        title: typeof cell.title === 'string' && cell.title.trim().length > 0
            ? cell.title
            : cell.type === 'sql'
                ? 'SQL Cell'
                : cell.type === 'runbook'
                    ? 'Runbook Step'
                    : 'Notes',
        content: typeof cell.content === 'string' ? cell.content : '',
        collapsed: !!cell.collapsed,
        executionState,
        lastRunAt: typeof cell.lastRunAt === 'string' ? cell.lastRunAt : undefined,
    };
};

const normalizeNotebook = (
    notebook: Partial<SqlNotebook> | null | undefined,
    fallbackDbType: string,
    fallbackConnectionName: string,
): SqlNotebook | null => {
    if (!notebook || typeof notebook.id !== 'string') {
        return null;
    }

    const cells = Array.isArray(notebook.cells)
        ? notebook.cells
            .map((cell) => normalizeCell(cell))
            .filter((cell): cell is SqlNotebookCell => !!cell)
        : [];

    const now = new Date().toISOString();

    return {
        id: notebook.id,
        title: typeof notebook.title === 'string' && notebook.title.trim().length > 0
            ? notebook.title
            : 'Untitled Notebook',
        description: typeof notebook.description === 'string'
            ? notebook.description
            : '',
        tags: Array.isArray(notebook.tags)
            ? notebook.tags.filter((tag): tag is string => typeof tag === 'string')
            : [],
        variables: Array.isArray(notebook.variables)
            ? notebook.variables
                .map((variable) => {
                    if (!variable || typeof variable.key !== 'string') {
                        return null;
                    }

                    return {
                        key: variable.key,
                        label: typeof variable.label === 'string' && variable.label.trim().length > 0
                            ? variable.label
                            : variable.key,
                        value: typeof variable.value === 'string' ? variable.value : '',
                        required: variable.required !== false,
                        type: variable.type === 'number' || variable.type === 'date' ? variable.type : 'text',
                    } satisfies SqlNotebookVariable;
                })
                .filter((variable): variable is SqlNotebookVariable => !!variable)
            : [],
        metadata: {
            environment: typeof notebook.metadata?.environment === 'string' ? notebook.metadata.environment : '',
            purpose: typeof notebook.metadata?.purpose === 'string' ? notebook.metadata.purpose : '',
            owner: typeof notebook.metadata?.owner === 'string' ? notebook.metadata.owner : '',
        },
        isFavorite: notebook.isFavorite === true,
        isTemplate: notebook.isTemplate === true,
        snapshots: Array.isArray(notebook.snapshots)
            ? notebook.snapshots
                .map((snapshot) => {
                    if (!snapshot || typeof snapshot.id !== 'string' || typeof snapshot.cellId !== 'string') {
                        return null;
                    }

                    return {
                        id: snapshot.id,
                        cellId: snapshot.cellId,
                        cellTitle: typeof snapshot.cellTitle === 'string' ? snapshot.cellTitle : 'SQL Cell',
                        capturedAt: typeof snapshot.capturedAt === 'string' ? snapshot.capturedAt : now,
                        totalRows: typeof snapshot.totalRows === 'number' ? snapshot.totalRows : 0,
                        resultSets: Array.isArray(snapshot.resultSets) ? snapshot.resultSets : [],
                    } satisfies SqlNotebookResultSnapshot;
                })
                .filter((snapshot): snapshot is SqlNotebookResultSnapshot => !!snapshot)
            : [],
        connectionScope: {
            dbType: notebook.connectionScope?.dbType || fallbackDbType,
            connectionName: notebook.connectionScope?.connectionName || fallbackConnectionName,
        },
        cells: cells.length > 0 ? cells : [createSqlNotebookCell('markdown'), createSqlNotebookCell('sql')],
        createdAt: typeof notebook.createdAt === 'string' ? notebook.createdAt : now,
        updatedAt: typeof notebook.updatedAt === 'string' ? notebook.updatedAt : now,
        lastOpenedAt: typeof notebook.lastOpenedAt === 'string' ? notebook.lastOpenedAt : now,
    };
};

export function useSqlNotebooks(options: UseSqlNotebooksOptions) {
    const notebooks = ref<SqlNotebook[]>([]);
    const activeNotebookId = ref<string | null>(null);
    const activeCellIdByNotebook = ref<Record<string, string | null>>({});
    const searchQuery = ref('');
    const isLoading = ref(false);
    const isRestoring = ref(false);
    const isSaving = ref(false);
    const isDirty = ref(false);
    const lastSavedAt = ref<string | null>(null);

    const filteredNotebooks = computed(() => {
        const keyword = normalizeSearchText(searchQuery.value);
        const visibleNotebooks = keyword
            ? notebooks.value.filter((notebook) => {
                const searchIndex = buildNotebookSearchIndex(notebook);
                return fuzzyIncludes(searchIndex, keyword);
            })
            : notebooks.value;

        return [...visibleNotebooks].sort((left, right) => {
            if (left.isFavorite !== right.isFavorite) {
                return left.isFavorite ? -1 : 1;
            }

            if (left.isTemplate !== right.isTemplate) {
                return left.isTemplate ? -1 : 1;
            }

            return right.updatedAt.localeCompare(left.updatedAt);
        });
    });

    const activeNotebook = computed(() => {
        return notebooks.value.find((notebook) => notebook.id === activeNotebookId.value) ?? null;
    });

    const activeCellId = computed(() => {
        const notebookId = activeNotebookId.value;
        if (!notebookId) {
            return null;
        }

        return activeCellIdByNotebook.value[notebookId] ?? null;
    });

    const persist = async () => {
        const storageKey = toValue(options.storageKey);
        if (!storageKey) {
            return;
        }

        const payload: PersistedSqlNotebookState = {
            version: SQL_NOTEBOOKS_STORAGE_VERSION,
            activeNotebookId: activeNotebookId.value,
            activeCellIdByNotebook: activeCellIdByNotebook.value,
            notebooks: notebooks.value,
        };

        try {
            isSaving.value = true;
            await SaveSetting(storageKey, JSON.stringify(payload));
            isDirty.value = false;
            lastSavedAt.value = new Date().toISOString();
        } catch (error) {
            console.error('Failed to persist SQL notebooks', error);
        } finally {
            isSaving.value = false;
        }
    };

    const markDirty = () => {
        if (isRestoring.value) {
            return;
        }

        isDirty.value = true;
    };

    const load = async () => {
        const storageKey = toValue(options.storageKey);
        const dbType = toValue(options.dbType) || 'unknown';
        const connectionName = toValue(options.connectionName) || 'Current Connection';

        isLoading.value = true;
        isRestoring.value = true;

        try {
            const raw = await LoadSetting(storageKey);
            if (!raw) {
                notebooks.value = [];
                activeNotebookId.value = null;
                activeCellIdByNotebook.value = {};
                isDirty.value = false;
                return;
            }

            const parsed = JSON.parse(raw) as Partial<PersistedSqlNotebookState>;
            if (parsed.version !== SQL_NOTEBOOKS_STORAGE_VERSION || !Array.isArray(parsed.notebooks)) {
                notebooks.value = [];
                activeNotebookId.value = null;
                activeCellIdByNotebook.value = {};
                isDirty.value = false;
                return;
            }

            const restoredNotebooks = parsed.notebooks
                .map((notebook) => normalizeNotebook(notebook, dbType, connectionName))
                .filter((notebook): notebook is SqlNotebook => !!notebook);

            if (restoredNotebooks.length === 0) {
                notebooks.value = [];
                activeNotebookId.value = null;
                activeCellIdByNotebook.value = {};
                isDirty.value = false;
                return;
            }

            notebooks.value = restoredNotebooks;
            activeNotebookId.value = restoredNotebooks.some((notebook) => notebook.id === parsed.activeNotebookId)
                ? parsed.activeNotebookId || restoredNotebooks[0].id
                : restoredNotebooks[0].id;
            activeCellIdByNotebook.value = Object.fromEntries(
                restoredNotebooks.map((notebook) => {
                    const persistedCellId = parsed.activeCellIdByNotebook?.[notebook.id] ?? null;
                    const hasCell = persistedCellId
                        ? notebook.cells.some((cell) => cell.id === persistedCellId)
                        : false;
                    return [notebook.id, hasCell ? persistedCellId : notebook.cells[0]?.id ?? null];
                }),
            );
            isDirty.value = false;
        } catch (error) {
            console.error('Failed to load SQL notebooks', error);
            notebooks.value = [];
            activeNotebookId.value = null;
            activeCellIdByNotebook.value = {};
            isDirty.value = false;
        } finally {
            isLoading.value = false;
            isRestoring.value = false;
        }
    };

    const touchNotebook = (notebookId: string) => {
        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook) {
            return;
        }

        const now = new Date().toISOString();
        notebook.updatedAt = now;
        notebook.lastOpenedAt = now;
    };

    const createNotebookItem = () => {
        const notebook = createSqlNotebook({
            dbType: toValue(options.dbType) || 'unknown',
            connectionName: toValue(options.connectionName) || 'Current Connection',
        });

        notebooks.value = [notebook, ...notebooks.value];
        activeNotebookId.value = notebook.id;
        activeCellIdByNotebook.value = {
            ...activeCellIdByNotebook.value,
            [notebook.id]: notebook.cells[0]?.id ?? null,
        };
        markDirty();
        return notebook;
    };

    const selectNotebook = (notebookId: string) => {
        activeNotebookId.value = notebookId;
        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook) {
            return;
        }

        activeCellIdByNotebook.value = {
            ...activeCellIdByNotebook.value,
            [notebookId]: activeCellIdByNotebook.value[notebookId] ?? notebook.cells[0]?.id ?? null,
        };
    };

    const setActiveCell = (cellId: string | null, notebookId = activeNotebookId.value) => {
        if (!notebookId) {
            return;
        }

        if ((activeCellIdByNotebook.value[notebookId] ?? null) === cellId) {
            return;
        }

        activeCellIdByNotebook.value = {
            ...activeCellIdByNotebook.value,
            [notebookId]: cellId,
        };
    };

    const updateNotebook = (notebookId: string, patch: Partial<Pick<SqlNotebook, 'title' | 'description' | 'tags' | 'variables' | 'metadata' | 'isFavorite' | 'isTemplate' | 'snapshots'>>) => {
        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook) {
            return;
        }

        if (typeof patch.title === 'string') {
            notebook.title = patch.title;
        }
        if (typeof patch.description === 'string') {
            notebook.description = patch.description;
        }
        if (Array.isArray(patch.tags)) {
            notebook.tags = patch.tags;
        }
        if (Array.isArray(patch.variables)) {
            notebook.variables = patch.variables;
        }
        if (patch.metadata) {
            notebook.metadata = {
                ...notebook.metadata,
                ...patch.metadata,
            };
        }
        if (typeof patch.isFavorite === 'boolean') {
            notebook.isFavorite = patch.isFavorite;
        }
        if (typeof patch.isTemplate === 'boolean') {
            notebook.isTemplate = patch.isTemplate;
        }
        if (Array.isArray(patch.snapshots)) {
            notebook.snapshots = patch.snapshots;
        }

        touchNotebook(notebookId);
        markDirty();
    };

    const deleteNotebook = (notebookId: string) => {
        const nextNotebooks = notebooks.value.filter((notebook) => notebook.id !== notebookId);
        notebooks.value = nextNotebooks;
        const nextActiveCellIds = { ...activeCellIdByNotebook.value };
        delete nextActiveCellIds[notebookId];
        activeCellIdByNotebook.value = nextActiveCellIds;

        if (nextNotebooks.length === 0) {
            activeNotebookId.value = null;
            markDirty();
            return;
        }

        if (activeNotebookId.value === notebookId) {
            activeNotebookId.value = nextNotebooks[0].id;
        }

        markDirty();
    };

    const addCell = (type: SqlNotebookCellType, notebookId = activeNotebookId.value) => {
        if (!notebookId) {
            return null;
        }

        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook) {
            return null;
        }

        const newCell = createSqlNotebookCell(type);
        notebook.cells.push(newCell);
        activeCellIdByNotebook.value = {
            ...activeCellIdByNotebook.value,
            [notebookId]: newCell.id,
        };
        touchNotebook(notebookId);
        markDirty();
        return newCell;
    };

    const updateCell = (cellId: string, patch: Partial<SqlNotebookCell>, notebookId = activeNotebookId.value) => {
        if (!notebookId) {
            return;
        }

        const notebook = notebooks.value.find((item) => item.id === notebookId);
        const cell = notebook?.cells.find((item) => item.id === cellId);
        if (!notebook || !cell) {
            return;
        }

        if (typeof patch.title === 'string') {
            cell.title = patch.title;
        }
        if (typeof patch.content === 'string') {
            cell.content = patch.content;
        }
        if (typeof patch.collapsed === 'boolean') {
            cell.collapsed = patch.collapsed;
        }
        if (patch.executionState) {
            cell.executionState = patch.executionState;
        }
        if (typeof patch.lastRunAt === 'string') {
            cell.lastRunAt = patch.lastRunAt;
        }

        touchNotebook(notebookId);
        markDirty();
    };

    const setCellExecutionState = (
        cellId: string,
        patch: Partial<Pick<SqlNotebookCell, 'executionState' | 'lastRunAt'>>,
        notebookId = activeNotebookId.value,
    ) => {
        if (!notebookId) {
            return;
        }

        const notebook = notebooks.value.find((item) => item.id === notebookId);
        const cell = notebook?.cells.find((item) => item.id === cellId);
        if (!notebook || !cell) {
            return;
        }

        if (patch.executionState) {
            cell.executionState = patch.executionState;
        }
        if (typeof patch.lastRunAt === 'string') {
            cell.lastRunAt = patch.lastRunAt;
        }
    };

    const removeCell = (cellId: string, notebookId = activeNotebookId.value) => {
        if (!notebookId) {
            return;
        }

        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook || notebook.cells.length <= 1) {
            return;
        }

        notebook.cells = notebook.cells.filter((cell) => cell.id !== cellId);
        if (activeCellIdByNotebook.value[notebookId] === cellId) {
            activeCellIdByNotebook.value = {
                ...activeCellIdByNotebook.value,
                [notebookId]: notebook.cells[0]?.id ?? null,
            };
        }
        touchNotebook(notebookId);
        markDirty();
    };

    const duplicateCell = (cellId: string, notebookId = activeNotebookId.value) => {
        if (!notebookId) {
            return null;
        }

        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook) {
            return null;
        }

        const index = notebook.cells.findIndex((cell) => cell.id === cellId);
        if (index === -1) {
            return null;
        }

        const sourceCell = notebook.cells[index];
        const duplicatedCell: SqlNotebookCell = {
            ...sourceCell,
            id: createSqlNotebookCell(sourceCell.type).id,
            title: `${sourceCell.title} Copy`,
            executionState: 'idle',
            lastRunAt: undefined,
        };

        notebook.cells.splice(index + 1, 0, duplicatedCell);
        activeCellIdByNotebook.value = {
            ...activeCellIdByNotebook.value,
            [notebookId]: duplicatedCell.id,
        };
        touchNotebook(notebookId);
        markDirty();
        return duplicatedCell;
    };

    const duplicateNotebook = (notebookId = activeNotebookId.value) => {
        if (!notebookId) {
            return null;
        }

        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook) {
            return null;
        }

        const duplicatedNotebook = createSqlNotebook({
            ...notebook.connectionScope,
        });
        duplicatedNotebook.title = `${notebook.title} Copy`;
        duplicatedNotebook.description = notebook.description;
        duplicatedNotebook.tags = [...notebook.tags];
        duplicatedNotebook.variables = notebook.variables.map((variable) => ({ ...variable }));
        duplicatedNotebook.metadata = { ...notebook.metadata };
        duplicatedNotebook.isFavorite = false;
        duplicatedNotebook.isTemplate = notebook.isTemplate;
        duplicatedNotebook.snapshots = notebook.snapshots.map((snapshot) => ({
            ...snapshot,
            resultSets: snapshot.resultSets.map((resultSet) => ({
                ...resultSet,
                rows: resultSet.rows.map((row) => ({ ...row })),
            })),
        }));
        duplicatedNotebook.cells = notebook.cells.map((cell) => ({
            ...cell,
            id: createSqlNotebookCell(cell.type).id,
            executionState: 'idle',
            lastRunAt: undefined,
        }));

        notebooks.value = [duplicatedNotebook, ...notebooks.value];
        activeNotebookId.value = duplicatedNotebook.id;
        activeCellIdByNotebook.value = {
            ...activeCellIdByNotebook.value,
            [duplicatedNotebook.id]: duplicatedNotebook.cells[0]?.id ?? null,
        };
        markDirty();
        return duplicatedNotebook;
    };

    const createNotebookFromTemplate = (templateId: string, title?: string) => {
        const template = notebooks.value.find((item) => item.id === templateId && item.isTemplate);
        if (!template) {
            return null;
        }

        const notebook = createSqlNotebook({
            dbType: toValue(options.dbType) || template.connectionScope.dbType,
            connectionName: toValue(options.connectionName) || template.connectionScope.connectionName,
        });

        notebook.title = title?.trim() || `${template.title} Run`;
        notebook.description = template.description;
        notebook.tags = [...template.tags];
        notebook.variables = template.variables.map((variable) => ({ ...variable }));
        notebook.metadata = { ...template.metadata };
        notebook.isFavorite = false;
        notebook.isTemplate = false;
        notebook.snapshots = [];
        notebook.cells = template.cells.map((cell) => ({
            ...cell,
            id: createSqlNotebookCell(cell.type).id,
            executionState: 'idle',
            lastRunAt: undefined,
        }));

        notebooks.value = [notebook, ...notebooks.value];
        activeNotebookId.value = notebook.id;
        activeCellIdByNotebook.value = {
            ...activeCellIdByNotebook.value,
            [notebook.id]: notebook.cells[0]?.id ?? null,
        };
        markDirty();
        return notebook;
    };

    const createNotebookFromPreset = (preset: SqlNotebookTemplatePreset, title?: string) => {
        const notebook = createSqlNotebook({
            dbType: toValue(options.dbType) || 'unknown',
            connectionName: toValue(options.connectionName) || 'Current Connection',
        });

        notebook.title = title?.trim() || preset.title;
        notebook.description = preset.description;
        notebook.tags = [...preset.tags];
        notebook.variables = preset.variables.map((variable) => ({ ...variable }));
        notebook.metadata = {
            environment: preset.metadata.environment || '',
            purpose: preset.metadata.purpose || '',
            owner: preset.metadata.owner || '',
        };
        notebook.cells = preset.cells.map((cell) => ({
            ...createSqlNotebookCell(cell.type),
            title: cell.title,
            content: cell.content,
        }));

        notebooks.value = [notebook, ...notebooks.value];
        activeNotebookId.value = notebook.id;
        activeCellIdByNotebook.value = {
            ...activeCellIdByNotebook.value,
            [notebook.id]: notebook.cells[0]?.id ?? null,
        };
        markDirty();
        return notebook;
    };

    const importSharedNotebook = (bundle: SqlNotebookShareBundle) => {
        const notebook = createSqlNotebook({
            dbType: toValue(options.dbType) || bundle.source.dbType || 'unknown',
            connectionName: toValue(options.connectionName) || 'Imported Shared Notebook',
        });

        const importedNotebook = bundle.notebook;
        notebook.title = importedNotebook.title.trim() || 'Imported Notebook';
        notebook.description = importedNotebook.description;
        notebook.tags = importedNotebook.tags
            .filter((tag): tag is string => typeof tag === 'string' && tag.trim().length > 0)
            .map((tag) => tag.trim());
        notebook.variables = importedNotebook.variables
            .map((variable) => {
                if (!variable || typeof variable.key !== 'string' || variable.key.trim().length === 0) {
                    return null;
                }

                return {
                    key: variable.key.trim(),
                    label: typeof variable.label === 'string' && variable.label.trim().length > 0
                        ? variable.label.trim()
                        : variable.key.trim(),
                    value: typeof variable.value === 'string' ? variable.value : '',
                    required: variable.required !== false,
                    type: variable.type === 'number' || variable.type === 'date' ? variable.type : 'text',
                } satisfies SqlNotebookVariable;
            })
            .filter((variable): variable is SqlNotebookVariable => !!variable);
        notebook.metadata = {
            environment: typeof importedNotebook.metadata?.environment === 'string' ? importedNotebook.metadata.environment : '',
            purpose: typeof importedNotebook.metadata?.purpose === 'string' ? importedNotebook.metadata.purpose : '',
            owner: typeof importedNotebook.metadata?.owner === 'string' ? importedNotebook.metadata.owner : '',
        } satisfies SqlNotebookMetadata;
        notebook.isFavorite = false;
        notebook.isTemplate = false;
        notebook.snapshots = [];
        const importedCells: SqlNotebookCell[] = [];
        for (const cell of importedNotebook.cells) {
            if (!cell || typeof cell.type !== 'string') {
                continue;
            }

            const cellType: SqlNotebookCellType | null = cell.type === 'sql' || cell.type === 'markdown' || cell.type === 'runbook'
                ? cell.type
                : null;
            if (!cellType) {
                continue;
            }

            importedCells.push({
                ...createSqlNotebookCell(cellType),
                title: typeof cell.title === 'string' && cell.title.trim().length > 0
                    ? cell.title.trim()
                    : cellType === 'sql'
                        ? 'SQL Cell'
                        : cellType === 'runbook'
                            ? 'Runbook Step'
                            : 'Notes',
                content: typeof cell.content === 'string' ? cell.content : '',
                collapsed: !!cell.collapsed,
                executionState: 'idle',
                lastRunAt: undefined,
            });
        }

        notebook.cells = importedCells;

        if (notebook.cells.length === 0) {
            notebook.cells = [createSqlNotebookCell('markdown'), createSqlNotebookCell('sql')];
        }

        notebooks.value = [notebook, ...notebooks.value];
        activeNotebookId.value = notebook.id;
        activeCellIdByNotebook.value = {
            ...activeCellIdByNotebook.value,
            [notebook.id]: notebook.cells[0]?.id ?? null,
        };
        markDirty();
        return notebook;
    };

    const moveCell = (cellId: string, direction: 'up' | 'down', notebookId = activeNotebookId.value) => {
        if (!notebookId) {
            return;
        }

        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook) {
            return;
        }

        const currentIndex = notebook.cells.findIndex((cell) => cell.id === cellId);
        if (currentIndex === -1) {
            return;
        }

        const targetIndex = direction === 'up' ? currentIndex - 1 : currentIndex + 1;
        if (targetIndex < 0 || targetIndex >= notebook.cells.length) {
            return;
        }

        const [cell] = notebook.cells.splice(currentIndex, 1);
        notebook.cells.splice(targetIndex, 0, cell);
        touchNotebook(notebookId);
        markDirty();
    };

    const upsertVariable = (variable: SqlNotebookVariable, notebookId = activeNotebookId.value) => {
        if (!notebookId) {
            return;
        }

        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook) {
            return;
        }

        const existingIndex = notebook.variables.findIndex((item) => item.key === variable.key);
        if (existingIndex >= 0) {
            notebook.variables.splice(existingIndex, 1, variable);
        } else {
            notebook.variables.push(variable);
        }

        touchNotebook(notebookId);
        markDirty();
    };

    const removeVariable = (key: string, notebookId = activeNotebookId.value) => {
        if (!notebookId) {
            return;
        }

        const notebook = notebooks.value.find((item) => item.id === notebookId);
        if (!notebook) {
            return;
        }

        notebook.variables = notebook.variables.filter((variable) => variable.key !== key);
        touchNotebook(notebookId);
        markDirty();
    };

    watch(
        [
            () => toValue(options.storageKey),
            () => toValue(options.dbType),
            () => toValue(options.connectionName),
        ],
        () => {
            void load();
        },
        { immediate: true },
    );

    return {
        notebooks,
        filteredNotebooks,
        activeNotebookId,
        activeNotebook,
        activeCellId,
        searchQuery,
        isLoading,
        isSaving,
        isDirty,
        lastSavedAt,
        load,
        saveNow: persist,
        createNotebook: createNotebookItem,
        selectNotebook,
        setActiveCell,
        updateNotebook,
        deleteNotebook,
        addCell,
        updateCell,
        setCellExecutionState,
        removeCell,
        duplicateCell,
        duplicateNotebook,
        createNotebookFromTemplate,
        createNotebookFromPreset,
        importSharedNotebook,
        moveCell,
        upsertVariable,
        removeVariable,
    };
}

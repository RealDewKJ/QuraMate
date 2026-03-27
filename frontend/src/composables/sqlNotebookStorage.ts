import { LoadSetting, SaveSetting } from '../../wailsjs/go/app/App';
import { extractDataUrlImages } from '../lib/markdownImages';
import {
    createSqlNotebook,
    createSqlNotebookCell,
    type SqlNotebookEmbeddedImage,
    type SqlNotebook,
    type SqlNotebookCell,
    type SqlNotebookMetadata,
    type SqlNotebookResultSnapshot,
    type SqlNotebookVariable,
} from '../types/sqlNotebook';

interface PersistedSqlNotebookState {
    version: number;
    activeNotebookId: string | null;
    notebooks: SqlNotebook[];
}

const SQL_NOTEBOOKS_STORAGE_VERSION = 1;

const normalizeCell = (cell: Partial<SqlNotebookCell> | null | undefined): SqlNotebookCell | null => {
    if (!cell || typeof cell.id !== 'string' || typeof cell.type !== 'string') {
        return null;
    }

    if (cell.type !== 'sql' && cell.type !== 'markdown') {
        return null;
    }

    const normalizedContent = typeof cell.content === 'string' ? cell.content : '';
    const embeddedImages = Array.isArray(cell.embeddedImages)
        ? cell.embeddedImages
            .map((image): SqlNotebookEmbeddedImage | null => {
                if (!image || typeof image.id !== 'string' || typeof image.dataUrl !== 'string') {
                    return null;
                }

                return {
                    id: image.id,
                    alt: typeof image.alt === 'string' && image.alt.trim().length > 0 ? image.alt.trim() : 'image',
                    fileName: typeof image.fileName === 'string' && image.fileName.trim().length > 0 ? image.fileName : 'image',
                    mimeType: typeof image.mimeType === 'string' && image.mimeType.trim().length > 0 ? image.mimeType : 'image/png',
                    dataUrl: image.dataUrl,
                };
            })
            .filter((image): image is SqlNotebookEmbeddedImage => !!image)
        : [];
    const migrated = extractDataUrlImages(normalizedContent);

    return {
        id: cell.id,
        type: cell.type,
        title: typeof cell.title === 'string' && cell.title.trim().length > 0
            ? cell.title
            : cell.type === 'sql'
                ? 'SQL Cell'
                : 'Notes',
        content: migrated.content,
        collapsed: !!cell.collapsed,
        executionState: cell.executionState === 'running' || cell.executionState === 'success' || cell.executionState === 'error'
            ? cell.executionState
            : 'idle',
        lastRunAt: typeof cell.lastRunAt === 'string' ? cell.lastRunAt : undefined,
        embeddedImages: [...embeddedImages, ...migrated.images],
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
        description: typeof notebook.description === 'string' ? notebook.description : '',
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
        } satisfies SqlNotebookMetadata,
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

const buildEmptyState = (): PersistedSqlNotebookState => ({
    version: SQL_NOTEBOOKS_STORAGE_VERSION,
    activeNotebookId: null,
    notebooks: [],
});

export const loadSqlNotebookState = async (
    storageKey: string,
    dbType: string,
    connectionName: string,
): Promise<PersistedSqlNotebookState> => {
    try {
        const raw = await LoadSetting(storageKey);
        if (!raw) {
            return buildEmptyState();
        }

        const parsed = JSON.parse(raw) as Partial<PersistedSqlNotebookState>;
        if (parsed.version !== SQL_NOTEBOOKS_STORAGE_VERSION || !Array.isArray(parsed.notebooks)) {
            return buildEmptyState();
        }

        const notebooks = parsed.notebooks
            .map((notebook) => normalizeNotebook(notebook, dbType, connectionName))
            .filter((notebook): notebook is SqlNotebook => !!notebook);

        if (notebooks.length === 0) {
            return buildEmptyState();
        }

        return {
            version: SQL_NOTEBOOKS_STORAGE_VERSION,
            activeNotebookId: notebooks.some((notebook) => notebook.id === parsed.activeNotebookId)
                ? parsed.activeNotebookId || notebooks[0].id
                : notebooks[0].id,
            notebooks,
        };
    } catch (error) {
        console.error('Failed to load SQL notebook storage', error);
        return buildEmptyState();
    }
};

export const saveSqlNotebookState = async (
    storageKey: string,
    state: PersistedSqlNotebookState,
): Promise<void> => {
    await SaveSetting(storageKey, JSON.stringify(state));
};

export const appendSqlToNotebook = async (options: {
    storageKey: string;
    dbType: string;
    connectionName: string;
    notebookId?: string | null;
    notebookTitle?: string;
    sqlTitle: string;
    sql: string;
    description?: string;
}): Promise<SqlNotebook> => {
    const state = await loadSqlNotebookState(options.storageKey, options.dbType, options.connectionName);
    const now = new Date().toISOString();

    let notebook = options.notebookId
        ? state.notebooks.find((item) => item.id === options.notebookId) || null
        : null;

    if (!notebook) {
        notebook = createSqlNotebook({
            dbType: options.dbType,
            connectionName: options.connectionName,
        });
        if (options.notebookTitle?.trim()) {
            notebook.title = options.notebookTitle.trim();
        }
        state.notebooks = [notebook, ...state.notebooks];
        state.activeNotebookId = notebook.id;
    }

    const newCell = createSqlNotebookCell('sql');
    newCell.title = options.sqlTitle.trim() || 'Saved Query';
    newCell.content = options.sql;
    newCell.executionState = 'idle';
    notebook.cells.push(newCell);

    if (typeof options.description === 'string') {
        const nextDescription = options.description.trim();
        if (nextDescription) {
            notebook.description = nextDescription;
        }
    }

    notebook.updatedAt = now;
    notebook.lastOpenedAt = now;

    await saveSqlNotebookState(options.storageKey, state);
    return notebook;
};

export type { PersistedSqlNotebookState };

<script lang="ts" setup>
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';

import { ExecuteQuery, SelectExportFile, WriteTextFile } from '../../../wailsjs/go/app/App';
import type { app as wailsApp } from '../../../wailsjs/go/models';
import type { ResultSet } from '../../types/database';
import {
    isSqlNotebookShareBundle,
    SQL_NOTEBOOK_SHARE_BUNDLE_KIND,
    SQL_NOTEBOOK_SHARE_BUNDLE_VERSION,
} from '../../types/sqlNotebook';
import type {
    SqlNotebookCell,
    SqlNotebookCellRunResult,
    SqlNotebookLiveSession,
    SqlNotebookLiveSessionStatus,
    SqlNotebookResultSnapshot,
    SqlNotebookShareBundle,
} from '../../types/sqlNotebook';
import SqlNotebookCellDeleteModal from './SqlNotebookCellDeleteModal.vue';
import SqlNotebookDeleteModal from './SqlNotebookDeleteModal.vue';
import SqlNotebookEditor from './SqlNotebookEditor.vue';
import SqlNotebookExecutionGuardrailModal from './SqlNotebookExecutionGuardrailModal.vue';
import SqlNotebookImportCodeModal from './SqlNotebookImportCodeModal.vue';
import SqlNotebookSnapshotsPanel from './SqlNotebookSnapshotsPanel.vue';
import SqlNotebookShareModal from './SqlNotebookShareModal.vue';
import SqlNotebookUnsavedChangesModal from './SqlNotebookUnsavedChangesModal.vue';
import SqlNotebooksSidebar from './SqlNotebooksSidebar.vue';

import { useSqlNotebooks } from '../../composables/useSqlNotebooks';
import { DEFAULT_SQL_NOTEBOOK_SHARE_RELAY_URL } from '../../constants/sqlNotebookShare';

const props = defineProps<{
    tabId: string;
    connectionId: string;
    storageKey: string;
    dbType: string;
    connectionName?: string;
    tables: string[];
    getColumns: (tableName: string) => Promise<string[]>;
    editorSettings: { fontFamily: string; fontSize: number };
    isReadOnly?: boolean;
}>();

const {
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
    saveNow,
    createNotebook,
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
    importSharedNotebook,
    moveCell,
} = useSqlNotebooks({
    storageKey: computed(() => props.storageKey),
    dbType: computed(() => props.dbType),
    connectionName: computed(() => props.connectionName),
});

const notebookPendingDeleteId = ref<string | null>(null);
const notebookPendingSwitchId = ref<string | null>(null);
const cellPendingDeleteId = ref<string | null>(null);
const isShareModalOpen = ref(false);
const isExportingShare = ref(false);
const isImportCodeModalOpen = ref(false);
const isImportingShareCode = ref(false);
const isSnapshotsPanelCollapsed = ref(true);
const shareScope = ref<'notebook' | 'sql'>('notebook');
const shareView = ref<'summary' | 'live'>('summary');
const shareCellId = ref('');
const shareCode = ref('');
const shareCodeExpiresAt = ref('');
const shareModalError = ref('');
const liveSessionCache = ref<Record<string, { code: string; sessionId: string; expiresAt: string }>>({});
const liveSession = ref<SqlNotebookLiveSession | null>(null);
const lastLiveNotebookSignature = ref('');
const importShareCode = ref('');
const importShareCodeError = ref('');
const guardrailPendingRun = ref<{ cellId: string; sql: string; preview: string } | null>(null);
const sidebarFilters = ref({
    favorites: false,
    environment: null as string | null,
});
const resultsByCellId = ref<Record<string, SqlNotebookCellRunResult>>({});
let liveSessionSocket: WebSocket | null = null;
let liveSessionSyncTimer: ReturnType<typeof setTimeout> | null = null;
const isApplyingLiveSessionUpdate = ref(false);

interface SelectSqlNotebookDetail {
    notebookId?: string;
}

interface SqlNotebookStorageUpdatedDetail {
    storageKey?: string;
    notebookId?: string;
}

type NotebookSaveHandler = () => Promise<boolean>;

interface NotebookWindowRegistry extends Window {
    __quraMateDirtyNotebookTabs?: Set<string>;
    __quraMateNotebookSaveHandlers?: Map<string, NotebookSaveHandler>;
}

const getDirtyNotebookTabs = (): Set<string> => {
    const notebookWindow = window as NotebookWindowRegistry;
    if (!notebookWindow.__quraMateDirtyNotebookTabs) {
        notebookWindow.__quraMateDirtyNotebookTabs = new Set<string>();
    }
    return notebookWindow.__quraMateDirtyNotebookTabs;
};

const getNotebookSaveHandlers = (): Map<string, NotebookSaveHandler> => {
    const notebookWindow = window as NotebookWindowRegistry;
    if (!notebookWindow.__quraMateNotebookSaveHandlers) {
        notebookWindow.__quraMateNotebookSaveHandlers = new Map<string, NotebookSaveHandler>();
    }
    return notebookWindow.__quraMateNotebookSaveHandlers;
};

const activeNotebookSnapshots = computed(() => {
    return [...(activeNotebook.value?.snapshots || [])].sort((left, right) => right.capturedAt.localeCompare(left.capturedAt));
});

const hasSnapshots = computed(() => activeNotebookSnapshots.value.length > 0);

const sidebarEnvironmentOptions = computed(() => {
    return Array.from(
        new Set(
            notebooks.value
                .map((notebook) => notebook.metadata.environment.trim())
                .filter(Boolean),
        ),
    ).sort((left, right) => left.localeCompare(right));
});

const sidebarNotebooks = computed(() => {
    return filteredNotebooks.value.filter((notebook) => {
        if (sidebarFilters.value.favorites && !notebook.isFavorite) {
            return false;
        }

        if (sidebarFilters.value.environment && notebook.metadata.environment !== sidebarFilters.value.environment) {
            return false;
        }

        return true;
    });
});

const activeNotebookSqlCells = computed(() => {
    return activeNotebook.value?.cells.filter((cell) => cell.type === 'sql') ?? [];
});

const guardrailEnvironmentLabel = computed(() => {
    return activeNotebook.value?.metadata.environment || 'protected environment';
});

const isProtectedEnvironment = computed(() => {
    const environment = activeNotebook.value?.metadata.environment.toLowerCase().trim() || '';
    return ['prod', 'production', 'live', 'staging'].some((keyword) => environment.includes(keyword));
});

const activeShareCell = computed(() => {
    return activeNotebookSqlCells.value.find((cell) => cell.id === shareCellId.value) ?? activeNotebookSqlCells.value[0] ?? null;
});

const shareTargetLabel = computed(() => {
    return shareScope.value === 'sql' ? 'SQL Cell' : 'Notebook';
});

const shareTargetDescription = computed(() => {
    return shareScope.value === 'sql'
        ? 'Copy a quick summary or start SessionShare from this SQL Notebook workspace.'
        : 'Copy a quick summary or start SessionShare for the full notebook.';
});

const canUseLiveSession = computed(() => shareScope.value === 'notebook' && !!activeNotebook.value);

const activeLiveSession = computed(() => {
    const currentSession = liveSession.value;
    if (!currentSession || !activeNotebook.value || currentSession.notebookId !== activeNotebook.value.id) {
        return null;
    }

    return currentSession;
});

const notebookShareRelayUrl = DEFAULT_SQL_NOTEBOOK_SHARE_RELAY_URL.trim();

const serializeShareBundleForReuse = (bundle: SqlNotebookShareBundle): string => {
    const normalizedBundle = {
        ...bundle,
        exportedAt: '',
    };

    return JSON.stringify(normalizedBundle);
};

const buildShareConnectionLabel = (dbType: string) => {
    return `${dbType.toUpperCase()} connection (name redacted)`;
};

const buildShareBundle = (
    notebook: NonNullable<typeof activeNotebook.value>,
    scope: 'notebook' | 'sql',
    cell: SqlNotebookCell | null,
): SqlNotebookShareBundle => {
    const sharedCells = scope === 'sql' && cell
        ? [
            {
                ...cell,
                executionState: 'idle',
                lastRunAt: undefined,
            } satisfies SqlNotebookCell,
        ]
        : notebook.cells.map((item) => ({
            ...item,
            executionState: item.type === 'sql' ? 'idle' : item.executionState,
            lastRunAt: undefined,
        } satisfies SqlNotebookCell));

    return {
        kind: SQL_NOTEBOOK_SHARE_BUNDLE_KIND,
        version: SQL_NOTEBOOK_SHARE_BUNDLE_VERSION,
        exportedAt: new Date().toISOString(),
        redacted: true,
        source: {
            dbType: notebook.connectionScope.dbType,
            connectionNameRedacted: true,
            scope,
        },
        notebook: {
            title: scope === 'sql' && cell ? `${notebook.title} - ${cell.title}` : notebook.title,
            description: scope === 'sql' && cell
                ? `Imported from ${notebook.title}. Connection details were redacted before sharing.`
                : notebook.description,
            tags: [...notebook.tags],
            variables: [],
            metadata: { ...notebook.metadata },
            cells: sharedCells,
        },
    };
};

const shareSummaryText = computed(() => {
    const notebook = activeNotebook.value;
    if (!notebook) {
        return '';
    }

    if (shareScope.value === 'sql' && activeShareCell.value) {
        const cell = activeShareCell.value;
        const result = resultsByCellId.value[cell.id];
        const lines = [
            `# ${cell.title}`,
            '',
            `Notebook: ${notebook.title}`,
            `Connection: ${buildShareConnectionLabel(notebook.connectionScope.dbType)}`,
        ];
        if (cell.lastRunAt) {
            lines.push(`Last Run: ${formatRunTimestamp(cell.lastRunAt)}`);
        }
        if (result?.status === 'success') {
            lines.push(`Latest Result: ${result.totalRows ?? 0} rows across ${result.resultSets.length} result set(s)`);
        }
        lines.push('', '```sql', cell.content.trim(), '```');
        return lines.join('\n');
    }

    const lines = [
        `# ${notebook.title}`,
        '',
        notebook.description || 'Shared from QuraMate SQL Notebooks.',
        '',
        `Connection: ${buildShareConnectionLabel(notebook.connectionScope.dbType)}`,
    ];

    if (notebook.metadata.environment) {
        lines.push(`Environment: ${notebook.metadata.environment}`);
    }
    if (notebook.metadata.owner) {
        lines.push(`Owner: ${notebook.metadata.owner}`);
    }
    if (notebook.tags.length > 0) {
        lines.push(`Tags: ${notebook.tags.join(', ')}`);
    }

    for (const cell of notebook.cells) {
        lines.push('', `## ${cell.title}`, `Type: ${cell.type}`);
        if (cell.type === 'sql') {
            lines.push('', '```sql', cell.content.trim(), '```');
        } else {
            lines.push('', cell.content.trim());
        }
    }

    return lines.join('\n').trim();
});

const shareBundle = computed(() => {
    const notebook = activeNotebook.value;
    if (!notebook) {
        return null;
    }

    return buildShareBundle(notebook, shareScope.value, activeShareCell.value);
});

const sharePreviewText = computed(() => {
    if (shareView.value === 'live') {
        return activeLiveSession.value?.code
            ? `${activeLiveSession.value.code}\n\nAnyone with this code can join the SessionShare while it stays active.`
            : 'Start a SessionShare to sync this notebook with connected teammates.';
    }

    return shareSummaryText.value;
});

const formatRunTimestamp = (value?: string): string => {
    if (!value) {
        return '-';
    }

    try {
        return new Intl.DateTimeFormat(undefined, {
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit',
            second: '2-digit',
        }).format(new Date(value));
    } catch (_error) {
        return value;
    }
};

const toSocketUrl = (value: string): string => {
    return value.replace(/^https:/i, 'wss:').replace(/^http:/i, 'ws:');
};

const createLiveSessionBundle = (): SqlNotebookShareBundle | null => {
    const notebook = activeNotebook.value;
    if (!notebook) {
        return null;
    }

    return buildShareBundle(notebook, 'notebook', null);
};

const buildLiveNotebookSyncState = (notebook: NonNullable<typeof activeNotebook.value>) => {
    return {
        title: notebook.title,
        description: notebook.description,
        tags: [...notebook.tags],
        metadata: { ...notebook.metadata },
        cells: notebook.cells.map((cell) => ({
            id: cell.id,
            type: cell.type,
            title: cell.title,
            content: cell.content,
            collapsed: cell.collapsed,
            executionState: cell.executionState,
            lastRunAt: cell.lastRunAt,
        })),
    };
};

const buildLiveNotebookSignature = (notebook: NonNullable<typeof activeNotebook.value>) => {
    return JSON.stringify(buildLiveNotebookSyncState(notebook));
};

const applySharedNotebookBundleToLocal = (notebookId: string, bundle: SqlNotebookShareBundle) => {
    const notebook = notebooks.value.find((item) => item.id === notebookId);
    if (!notebook) {
        return;
    }

    const incoming = bundle.notebook;
    notebook.title = incoming.title || notebook.title;
    notebook.description = incoming.description || '';
    notebook.tags = [...incoming.tags];
    notebook.metadata = {
        environment: incoming.metadata.environment || '',
        purpose: incoming.metadata.purpose || '',
        owner: incoming.metadata.owner || '',
    };
    notebook.cells = incoming.cells.map((cell) => ({
        ...cell,
        executionState: 'idle',
        lastRunAt: undefined,
    }));
    notebook.updatedAt = new Date().toISOString();
};

const applyLiveNotebookPatchToLocal = (
    notebookId: string,
    patch: Partial<ReturnType<typeof buildLiveNotebookSyncState>>,
) => {
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
        notebook.tags = [...patch.tags];
    }
    if (patch.metadata) {
        notebook.metadata = {
            environment: patch.metadata.environment || '',
            purpose: patch.metadata.purpose || '',
            owner: patch.metadata.owner || '',
        };
    }
    if (Array.isArray(patch.cells)) {
        notebook.cells = patch.cells.map((cell) => ({
            ...cell,
            executionState: cell.executionState || 'idle',
            lastRunAt: cell.lastRunAt,
        }));
    }
    notebook.updatedAt = new Date().toISOString();
};

const disconnectLiveSession = (status: SqlNotebookLiveSessionStatus = 'ended', errorMessage = '') => {
    if (liveSessionSocket) {
        liveSessionSocket.close();
        liveSessionSocket = null;
    }

    if (liveSessionSyncTimer) {
        clearTimeout(liveSessionSyncTimer);
        liveSessionSyncTimer = null;
    }

    if (liveSession.value) {
        liveSession.value = {
            ...liveSession.value,
            status,
            errorMessage: errorMessage || liveSession.value.errorMessage,
        };
    }
};

const queueLiveSessionSnapshotBroadcast = () => {
    if (!activeLiveSession.value || !liveSessionSocket || liveSessionSocket.readyState !== WebSocket.OPEN || isApplyingLiveSessionUpdate.value) {
        return;
    }

    if (liveSessionSyncTimer) {
        clearTimeout(liveSessionSyncTimer);
    }

    liveSessionSyncTimer = setTimeout(() => {
        if (!activeLiveSession.value || !liveSessionSocket || liveSessionSocket.readyState !== WebSocket.OPEN) {
            return;
        }

        const notebook = activeNotebook.value;
        if (!notebook) {
            return;
        }

        const signature = buildLiveNotebookSignature(notebook);
        if (signature === lastLiveNotebookSignature.value) {
            return;
        }
        lastLiveNotebookSignature.value = signature;

        liveSessionSocket.send(JSON.stringify({
            type: 'patch',
            payload: {
                notebook: buildLiveNotebookSyncState(notebook),
            },
        }));
    }, 300);
};

const connectLiveSessionSocket = (session: SqlNotebookLiveSession) => {
    if (!notebookShareRelayUrl || notebookShareRelayUrl.includes('YOUR-SUBDOMAIN')) {
        return;
    }

    if (liveSessionSocket) {
        liveSessionSocket.close();
        liveSessionSocket = null;
    }

    const socketUrl = `${toSocketUrl(notebookShareRelayUrl)}/api/notebook-live-sessions/${session.code}/socket`;
    const socket = new WebSocket(socketUrl);
    liveSessionSocket = socket;

    socket.addEventListener('open', () => {
        if (!liveSession.value || liveSession.value.code !== session.code) {
            return;
        }

        liveSession.value = {
            ...liveSession.value,
            status: 'connected',
            errorMessage: '',
        };

        if (activeNotebook.value) {
            lastLiveNotebookSignature.value = '';
        }
        queueLiveSessionSnapshotBroadcast();
    });

    socket.addEventListener('message', (event) => {
        if (!liveSession.value || liveSession.value.code !== session.code) {
            return;
        }

        try {
            const parsed = JSON.parse(String(event.data)) as {
                type?: string;
                payload?: {
                    peers?: number;
                    bundle?: SqlNotebookShareBundle;
                    notebook?: Partial<ReturnType<typeof buildLiveNotebookSyncState>>;
                };
            };

            if (parsed.type === 'presence') {
                liveSession.value = {
                    ...liveSession.value,
                    peerCount: typeof parsed.payload?.peers === 'number' ? parsed.payload.peers : liveSession.value.peerCount,
                    lastMessageAt: new Date().toISOString(),
                };
                return;
            }

            if (parsed.type === 'session-expired') {
                disconnectLiveSession('expired', 'This live session expired.');
                return;
            }

            if (parsed.type === 'session-ended') {
                disconnectLiveSession('ended', 'The host ended this live session.');
                return;
            }

            if (parsed.type === 'patch' && parsed.payload?.notebook && activeLiveSession.value) {
                isApplyingLiveSessionUpdate.value = true;
                try {
                    applyLiveNotebookPatchToLocal(activeLiveSession.value.notebookId, parsed.payload.notebook);
                    const notebook = notebooks.value.find((item) => item.id === activeLiveSession.value?.notebookId);
                    if (notebook) {
                        lastLiveNotebookSignature.value = buildLiveNotebookSignature(notebook);
                    }
                } finally {
                    isApplyingLiveSessionUpdate.value = false;
                }

                liveSession.value = {
                    ...liveSession.value,
                    lastMessageAt: new Date().toISOString(),
                };
                void saveNow();
                return;
            }

            if (parsed.type === 'snapshot' && parsed.payload?.bundle && activeLiveSession.value) {
                isApplyingLiveSessionUpdate.value = true;
                try {
                    applySharedNotebookBundleToLocal(activeLiveSession.value.notebookId, parsed.payload.bundle);
                    const notebook = notebooks.value.find((item) => item.id === activeLiveSession.value?.notebookId);
                    if (notebook) {
                        lastLiveNotebookSignature.value = buildLiveNotebookSignature(notebook);
                    }
                } finally {
                    isApplyingLiveSessionUpdate.value = false;
                }

                liveSession.value = {
                    ...liveSession.value,
                    lastMessageAt: new Date().toISOString(),
                };
                void saveNow();
            }
        } catch (error) {
            console.error('Failed to process live session message', error);
        }
    });

    socket.addEventListener('close', () => {
        if (!liveSession.value || liveSession.value.code !== session.code || liveSession.value.status === 'ended' || liveSession.value.status === 'expired') {
            return;
        }

        liveSession.value = {
            ...liveSession.value,
            status: 'error',
            errorMessage: 'Live session disconnected.',
        };
    });

    socket.addEventListener('error', () => {
        if (!liveSession.value || liveSession.value.code !== session.code) {
            return;
        }

        liveSession.value = {
            ...liveSession.value,
            status: 'error',
            errorMessage: 'Could not connect to the live session.',
        };
    });
};

const focusCell = async (cellId: string | null, options?: { smooth?: boolean; focusInner?: boolean }) => {
    if (!cellId) {
        return;
    }

    setActiveCell(cellId);
    await nextTick();
    const element = document.getElementById(`sql-notebook-cell-${cellId}`);
    if (!element) {
        return;
    }

    if (options?.focusInner !== false) {
        const focusTarget = element.querySelector('input, textarea, button, [contenteditable="true"]') as HTMLElement | null;
        (focusTarget ?? element).focus({ preventScroll: true });
    }
    element.scrollIntoView({
        behavior: options?.smooth === false ? 'auto' : 'smooth',
        block: 'center',
    });
};

const saveSnapshot = (cellId: string) => {
    const notebook = activeNotebook.value;
    const result = resultsByCellId.value[cellId];
    const cell = notebook?.cells.find((item) => item.id === cellId);
    if (!notebook || !cell || !result || result.status !== 'success') {
        return;
    }

    const snapshot: SqlNotebookResultSnapshot = {
        id: `snapshot_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`,
        cellId,
        cellTitle: cell.title,
        capturedAt: new Date().toISOString(),
        totalRows: result.totalRows ?? 0,
        resultSets: result.resultSets.map((resultSet) => ({
            ...resultSet,
            rows: resultSet.rows.map((row) => ({ ...row })),
        })),
    };

    updateNotebook(notebook.id, {
        snapshots: [snapshot, ...notebook.snapshots],
    });
};

const looksDestructiveSql = (sql: string) => {
    return /\b(delete|truncate|drop|alter|update|insert|create|replace|merge)\b/i.test(sql);
};

const createNotebookQueryId = (): string => {
    return `notebook_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`;
};

const mapResultSet = (resultSet: ResultSet): SqlNotebookCellRunResult['resultSets'][number] => {
    const columns = Array.isArray(resultSet.columns) ? resultSet.columns : [];
    const rows = Array.isArray(resultSet.rows)
        ? resultSet.rows.map((row) => {
            if (Array.isArray(row)) {
                return Object.fromEntries(columns.map((column, index) => [column, row[index]]));
            }
            if (row && typeof row === 'object') {
                return row as Record<string, unknown>;
            }
            return {};
        })
        : [];

    return {
        columns,
        rows,
        message: resultSet.message,
    };
};

const setCellRunResult = (cellId: string, result: SqlNotebookCellRunResult) => {
    resultsByCellId.value = {
        ...resultsByCellId.value,
        [cellId]: result,
    };
};

const executeSqlCell = async (cellId: string, sqlOverride?: string) => {
    const notebook = activeNotebook.value;
    const cell = notebook?.cells.find((item) => item.id === cellId);
    if (!notebook || !cell || cell.type !== 'sql') {
        return;
    }

    setActiveCell(cell.id);
    const resolvedSql = sqlOverride || cell.content;

    const startedAt = new Date().toISOString();
    setCellRunResult(cell.id, {
        status: 'running',
        resultSets: [],
        startedAt,
    });
    setCellExecutionState(cell.id, {
        executionState: 'running',
    }, notebook.id);

    try {
        const response = await ExecuteQuery(props.connectionId, resolvedSql, createNotebookQueryId()) as wailsApp.QueryResult;
        const completedAt = new Date().toISOString();

        if (response.error) {
            setCellRunResult(cell.id, {
                status: 'error',
                resultSets: [],
                startedAt,
                completedAt,
                errorMessage: response.error,
            });
            setCellExecutionState(cell.id, {
                executionState: 'error',
                lastRunAt: completedAt,
            }, notebook.id);
            return;
        }

        const resultSets = (response.resultSets || []).map((resultSet) => mapResultSet(resultSet as unknown as ResultSet));
        const totalRows = resultSets.reduce((total, resultSet) => total + resultSet.rows.length, 0);

        setCellRunResult(cell.id, {
            status: 'success',
            resultSets,
            startedAt,
            completedAt,
            totalRows,
        });
        setCellExecutionState(cell.id, {
            executionState: 'success',
            lastRunAt: completedAt,
        }, notebook.id);
    } catch (error) {
        const completedAt = new Date().toISOString();
        setCellRunResult(cell.id, {
            status: 'error',
            resultSets: [],
            startedAt,
            completedAt,
            errorMessage: error instanceof Error ? error.message : String(error),
        });
        setCellExecutionState(cell.id, {
            executionState: 'error',
            lastRunAt: completedAt,
        }, notebook.id);
    }
};

const handleRunCell = async (cellId: string) => {
    const notebook = activeNotebook.value;
    const cell = notebook?.cells.find((item) => item.id === cellId);
    if (!notebook || !cell || cell.type !== 'sql') {
        return;
    }

    const resolvedSql = cell.content;

    if (isProtectedEnvironment.value && looksDestructiveSql(resolvedSql)) {
        guardrailPendingRun.value = {
            cellId,
            sql: resolvedSql,
            preview: resolvedSql.slice(0, 1200),
        };
        return;
    }

    await executeSqlCell(cellId, resolvedSql);
};

const notebookPendingDelete = computed(() => {
    const notebookId = notebookPendingDeleteId.value;
    if (!notebookId) {
        return null;
    }

    return (
        filteredNotebooks.value.find((notebook) => notebook.id === notebookId)
        ?? (activeNotebook.value?.id === notebookId ? activeNotebook.value : null)
        ?? null
    );
});

const cellPendingDelete = computed(() => {
    const cellId = cellPendingDeleteId.value;
    if (!cellId || !activeNotebook.value) {
        return null;
    }

    return activeNotebook.value.cells.find((cell) => cell.id === cellId) ?? null;
});

const notebookPendingSwitch = computed(() => {
    const notebookId = notebookPendingSwitchId.value;
    if (!notebookId) {
        return null;
    }

    return filteredNotebooks.value.find((notebook) => notebook.id === notebookId)
        ?? null;
});

const activeNotebookTitle = computed(() => {
    return activeNotebook.value?.title || 'this notebook';
});

const pendingNotebookTitle = computed(() => {
    return notebookPendingSwitch.value?.title || 'the selected notebook';
});

const openDeleteNotebookWarning = (notebookId: string) => {
    notebookPendingDeleteId.value = notebookId;
};

const closeDeleteNotebookWarning = () => {
    notebookPendingDeleteId.value = null;
};

const closeUnsavedChangesWarning = () => {
    notebookPendingSwitchId.value = null;
};

const openDeleteCellWarning = (cellId: string) => {
    cellPendingDeleteId.value = cellId;
};

const closeDeleteCellWarning = () => {
    cellPendingDeleteId.value = null;
};

const openShareNotebook = () => {
    shareScope.value = 'notebook';
    shareView.value = 'summary';
    shareCode.value = '';
    shareCodeExpiresAt.value = '';
    shareModalError.value = '';
    shareCellId.value = activeNotebookSqlCells.value[0]?.id || '';
    isShareModalOpen.value = true;
};

const openShareSqlCell = (cellId: string) => {
    shareScope.value = 'sql';
    shareView.value = 'summary';
    shareCode.value = '';
    shareCodeExpiresAt.value = '';
    shareModalError.value = '';
    shareCellId.value = cellId;
    isShareModalOpen.value = true;
};

const closeShareModal = () => {
    isShareModalOpen.value = false;
    shareModalError.value = '';
};

const openImportCodeModal = () => {
    importShareCode.value = '';
    importShareCodeError.value = '';
    isImportCodeModalOpen.value = true;
};

const closeImportCodeModal = () => {
    isImportCodeModalOpen.value = false;
    isImportingShareCode.value = false;
    importShareCodeError.value = '';
};

const closeExecutionGuardrail = () => {
    guardrailPendingRun.value = null;
};

const handleCreateNotebook = async () => {
    createNotebook();
    await saveNow();
};

const handleImportNotebook = () => {
    openImportCodeModal();
};

const handleEndLiveSession = () => {
    if (liveSessionSocket && liveSessionSocket.readyState === WebSocket.OPEN) {
        liveSessionSocket.send(JSON.stringify({
            type: 'session-ended',
            payload: {
                code: activeLiveSession.value?.code || '',
            },
        }));
    }
    disconnectLiveSession('ended');
};

const handleDuplicateNotebook = async (notebookId: string) => {
    duplicateNotebook(notebookId);
    await saveNow();
};

const copyShareContent = async () => {
    if (!sharePreviewText.value.trim()) {
        return;
    }

    try {
        await navigator.clipboard.writeText(sharePreviewText.value);
    } catch (error) {
        console.error('Failed to copy shared notebook content', error);
    }
};

const exportShareContent = async () => {
    if (shareView.value === 'live') {
        if (!canUseLiveSession.value || !activeNotebook.value) {
                shareModalError.value = 'SessionShare is available only for full notebooks.';
            return;
        }

        const bundle = createLiveSessionBundle();
        if (!bundle) {
            return;
        }

        isExportingShare.value = true;
        shareModalError.value = '';
        try {
            if (!notebookShareRelayUrl || notebookShareRelayUrl.includes('YOUR-SUBDOMAIN')) {
                shareModalError.value = 'Share service URL is not configured in this app build yet.';
                return;
            }

            const response = await fetch(`${notebookShareRelayUrl}/api/notebook-live-sessions`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    payload: JSON.parse(serializeShareBundleForReuse(bundle)),
                }),
            });

            const result = await response.json() as { code?: string; sessionId?: string; expiresAt?: string; error?: string };
            if (!response.ok || !result.code || !result.sessionId || !result.expiresAt) {
                shareModalError.value = result.error || 'Could not start SessionShare.';
                return;
            }

            const session: SqlNotebookLiveSession = {
                code: result.code,
                sessionId: result.sessionId,
                notebookId: activeNotebook.value.id,
                status: 'hosting',
                role: 'host',
                expiresAt: result.expiresAt,
                peerCount: 1,
            };
            liveSession.value = session;
            liveSessionCache.value = {
                ...liveSessionCache.value,
                [activeNotebook.value.id]: {
                    code: result.code,
                    sessionId: result.sessionId,
                    expiresAt: result.expiresAt,
                },
            };
            shareCode.value = result.code;
            shareCodeExpiresAt.value = formatRunTimestamp(result.expiresAt);
            connectLiveSessionSocket(session);
        } catch (error) {
            console.error('Failed to create live notebook session', error);
            shareModalError.value = error instanceof Error ? error.message : String(error);
        } finally {
            isExportingShare.value = false;
        }
        return;
    }

    if (!sharePreviewText.value.trim()) {
        return;
    }

    const baseNameSource = shareScope.value === 'sql'
        ? activeShareCell.value?.title || 'sql-cell'
        : activeNotebook.value?.title || 'sql-notebook';
    const normalizedBaseName = baseNameSource
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, '-')
        .replace(/^-+|-+$/g, '') || 'sql-notebook';
    const exportFileName = `${normalizedBaseName}.md`;

    isExportingShare.value = true;
    try {
        const filePath = await SelectExportFile(exportFileName);
        if (!filePath) {
            return;
        }

        const writeResult = await WriteTextFile(filePath, sharePreviewText.value);
        if (writeResult && writeResult !== 'Success') {
            throw new Error(writeResult);
        }
    } catch (error) {
        console.error('Failed to export shared notebook content', error);
    } finally {
        isExportingShare.value = false;
    }
};

const confirmImportShareCode = async () => {
    const code = importShareCode.value.trim().toUpperCase();
    if (!code) {
        importShareCodeError.value = 'Enter a SessionShare code first.';
        return;
    }
    if (!notebookShareRelayUrl || notebookShareRelayUrl.includes('YOUR-SUBDOMAIN')) {
        importShareCodeError.value = 'Share service URL is not configured in this app build yet.';
        return;
    }

    isImportingShareCode.value = true;
    importShareCodeError.value = '';
    try {
        const response = await fetch(`${notebookShareRelayUrl}/api/notebook-live-sessions/${code}/join`, {
            method: 'POST',
        });
        const result = await response.json() as { sessionId?: string; code?: string; expiresAt?: string; payload?: unknown; error?: string };
        if (!response.ok || !result.sessionId || !result.code || !result.expiresAt || !result.payload) {
            importShareCodeError.value = result.error || 'SessionShare not found.';
            return;
        }

        if (!isSqlNotebookShareBundle(result.payload)) {
            importShareCodeError.value = 'This SessionShare payload is invalid or expired.';
            return;
        }

        const importedNotebook = importSharedNotebook(result.payload);
        liveSession.value = {
            code: result.code,
            sessionId: result.sessionId,
            notebookId: importedNotebook.id,
            status: 'joining',
            role: 'peer',
            expiresAt: result.expiresAt,
            peerCount: 1,
        };
        liveSessionCache.value = {
            ...liveSessionCache.value,
            [importedNotebook.id]: {
                code: result.code,
                sessionId: result.sessionId,
                expiresAt: result.expiresAt,
            },
        };
        await saveNow();
        connectLiveSessionSocket(liveSession.value);
        closeImportCodeModal();
    } catch (error) {
        console.error('Failed to join SessionShare', error);
        importShareCodeError.value = 'Could not join this SessionShare.';
    } finally {
        isImportingShareCode.value = false;
    }
};

const confirmDeleteNotebook = async () => {
    if (!notebookPendingDeleteId.value) {
        return;
    }

    const notebookId = notebookPendingDeleteId.value;
    const notebook = filteredNotebooks.value.find((item) => item.id === notebookId)
        ?? activeNotebook.value;
    const nextResults = { ...resultsByCellId.value };
    for (const cell of notebook?.cells || []) {
        delete nextResults[cell.id];
    }

    deleteNotebook(notebookId);
    resultsByCellId.value = nextResults;
    notebookPendingDeleteId.value = null;
    await saveNow();
};

const handleSave = async () => {
    if (props.isReadOnly || !isDirty.value || isSaving.value) {
        return !isDirty.value;
    }

    await saveNow();
    return !isDirty.value;
};

const handleDeleteCell = (cellId: string) => {
    removeCell(cellId);
    if (!resultsByCellId.value[cellId]) {
        return;
    }

    const nextResults = { ...resultsByCellId.value };
    delete nextResults[cellId];
    resultsByCellId.value = nextResults;
};

const confirmDeleteCell = () => {
    if (!cellPendingDeleteId.value) {
        return;
    }

    handleDeleteCell(cellPendingDeleteId.value);
    cellPendingDeleteId.value = null;
};

const handleActivateCell = (cellId: string) => {
    setActiveCell(cellId);
};

const toggleFavorite = (notebookId: string) => {
    const notebook = filteredNotebooks.value.find((item) => item.id === notebookId)
        ?? activeNotebook.value;
    if (!notebook) {
        return;
    }

    updateNotebook(notebookId, {
        isFavorite: !notebook.isFavorite,
    });
    void saveNow();
};

const handleUpdateCellExecutionState = (payload: { cellId: string; value: 'idle' | 'running' | 'success' | 'error' | 'verified' | 'skipped' }) => {
    updateCell(payload.cellId, {
        executionState: payload.value,
        lastRunAt: new Date().toISOString(),
    });
};

const handleSelectNotebook = async (notebookId: string) => {
    if (notebookId === activeNotebookId.value) {
        return;
    }

    if (isDirty.value) {
        notebookPendingSwitchId.value = notebookId;
        return;
    }

    selectNotebook(notebookId);
};

const confirmSwitchNotebookWithSave = async () => {
    const notebookId = notebookPendingSwitchId.value;
    if (!notebookId) {
        return;
    }

    const didSave = await handleSave();
    if (!didSave) {
        return;
    }

    selectNotebook(notebookId);
    notebookPendingSwitchId.value = null;
};

const confirmSwitchNotebookDiscardChanges = () => {
    const notebookId = notebookPendingSwitchId.value;
    if (!notebookId) {
        return;
    }

    void load().then(() => {
        selectNotebook(notebookId);
        notebookPendingSwitchId.value = null;
    });
};

const confirmGuardrailRun = async () => {
    if (!guardrailPendingRun.value) {
        return;
    }

    const pendingRun = guardrailPendingRun.value;
    guardrailPendingRun.value = null;
    await executeSqlCell(pendingRun.cellId, pendingRun.sql);
};

const handleWindowKeydown = (event: KeyboardEvent) => {
    if (props.isReadOnly) {
        return;
    }

    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
        event.preventDefault();
        void handleSave();
    }
};

const handleBeforeUnload = (event: BeforeUnloadEvent) => {
    if (!isDirty.value) {
        return;
    }

    event.preventDefault();
    event.returnValue = '';
};

const handleSelectNotebookEvent = (event: Event) => {
    const detail = (event as CustomEvent<SelectSqlNotebookDetail>).detail;
    if (!detail?.notebookId) {
        return;
    }

    void handleSelectNotebook(detail.notebookId);
};

const handleSqlNotebookStorageUpdated = (event: Event) => {
    const detail = (event as CustomEvent<SqlNotebookStorageUpdatedDetail>).detail;
    if (!detail?.storageKey || detail.storageKey !== props.storageKey) {
        return;
    }

    resultsByCellId.value = {};
    void load().then(() => {
        if (detail.notebookId) {
            selectNotebook(detail.notebookId);
        }
    });
};

onMounted(() => {
    getNotebookSaveHandlers().set(props.tabId, handleSave);
    window.addEventListener('keydown', handleWindowKeydown);
    window.addEventListener('beforeunload', handleBeforeUnload);
    window.addEventListener('quramate:select-sql-notebook', handleSelectNotebookEvent as EventListener);
    window.addEventListener('quramate:sql-notebook-storage-updated', handleSqlNotebookStorageUpdated as EventListener);
});

onUnmounted(() => {
    disconnectLiveSession('ended');
    getDirtyNotebookTabs().delete(props.tabId);
    getNotebookSaveHandlers().delete(props.tabId);
    window.removeEventListener('keydown', handleWindowKeydown);
    window.removeEventListener('beforeunload', handleBeforeUnload);
    window.removeEventListener('quramate:select-sql-notebook', handleSelectNotebookEvent as EventListener);
    window.removeEventListener('quramate:sql-notebook-storage-updated', handleSqlNotebookStorageUpdated as EventListener);
});

watch(
    [() => props.tabId, isDirty],
    ([tabId, dirty], [previousTabId]) => {
        const dirtyTabs = getDirtyNotebookTabs();
        if (previousTabId && previousTabId !== tabId) {
            dirtyTabs.delete(previousTabId);
        }

        if (dirty) {
            dirtyTabs.add(tabId);
        } else {
            dirtyTabs.delete(tabId);
        }
    },
    { immediate: true },
);

watch(
    () => props.tabId,
    (tabId, previousTabId) => {
        const handlers = getNotebookSaveHandlers();
        if (previousTabId && previousTabId !== tabId) {
            handlers.delete(previousTabId);
        }
        handlers.set(tabId, handleSave);
    },
    { immediate: true },
);

watch(
    [() => props.storageKey, () => props.connectionId],
    () => {
        resultsByCellId.value = {};
    },
);

watch(
    () => activeNotebookId.value,
    () => {
        void focusCell(activeCellId.value, { smooth: false });
    },
    { immediate: true },
);

watch(
    () => activeCellId.value,
    (cellId) => {
        if (!cellId) {
            return;
        }

        void focusCell(cellId, { smooth: false });
    },
);

watch(
    hasSnapshots,
    (value) => {
        isSnapshotsPanelCollapsed.value = !value;
    },
    { immediate: true },
);

watch(
    [isShareModalOpen, shareView, activeNotebookId],
    ([isOpen, view, notebookId]) => {
        if (!isOpen || view !== 'live' || !notebookId) {
            return;
        }

        const cached = liveSessionCache.value[notebookId];
        if (!cached) {
            if (!activeLiveSession.value) {
                shareCode.value = '';
                shareCodeExpiresAt.value = '';
            }
            return;
        }

        const expiresAt = new Date(cached.expiresAt);
        if (Number.isNaN(expiresAt.getTime()) || expiresAt.getTime() <= Date.now()) {
            const nextCache = { ...liveSessionCache.value };
            delete nextCache[notebookId];
            liveSessionCache.value = nextCache;
            return;
        }

        shareCode.value = cached.code;
        shareCodeExpiresAt.value = formatRunTimestamp(cached.expiresAt);
    },
    { immediate: true },
);

watch(
    activeNotebook,
    (notebook) => {
        if (!notebook || !activeLiveSession.value || activeLiveSession.value.notebookId !== notebook.id || isApplyingLiveSessionUpdate.value) {
            return;
        }

        queueLiveSessionSnapshotBroadcast();
    },
    { deep: true },
);
</script>

<template>
    <div class="flex h-full min-h-0 min-w-0 bg-background">
        <SqlNotebooksSidebar
            :notebooks="sidebarNotebooks"
            :active-notebook-id="activeNotebookId"
            :search-query="searchQuery"
            :is-loading="isLoading"
            :active-filters="sidebarFilters"
            :environment-options="sidebarEnvironmentOptions"
            @update:search-query="searchQuery = $event"
            @update:filters="sidebarFilters = $event"
            @select-notebook="handleSelectNotebook"
            @create-notebook="handleCreateNotebook"
            @import-notebook="handleImportNotebook"
            @toggle-favorite="toggleFavorite"
            @duplicate-notebook="handleDuplicateNotebook"
            @request-delete-notebook="openDeleteNotebookWarning"
        />

        <div class="flex min-h-0 min-w-0 flex-1 overflow-hidden bg-muted/10">
            <div class="flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden">
                <div
                    v-if="activeLiveSession"
                    class="mx-4 mt-4 flex flex-wrap items-center justify-between gap-3 rounded-xl border border-primary/20 bg-primary/5 px-4 py-3 text-sm"
                >
                    <div class="space-y-1">
                        <div class="font-medium text-foreground">
                            SessionShare {{ activeLiveSession.code }}
                        </div>
                        <div class="text-xs text-muted-foreground">
                            Status: {{ activeLiveSession.status }} | {{ activeLiveSession.peerCount }} peer{{ activeLiveSession.peerCount === 1 ? '' : 's' }} | Expires {{ formatRunTimestamp(activeLiveSession.expiresAt) }}
                        </div>
                        <div v-if="activeLiveSession.errorMessage" class="text-xs text-destructive">
                            {{ activeLiveSession.errorMessage }}
                        </div>
                    </div>
                    <button
                        class="inline-flex h-9 items-center justify-center rounded-md border border-input bg-background px-3 text-sm font-medium transition-colors hover:bg-accent"
                        @click="handleEndLiveSession"
                    >
                        End SessionShare
                    </button>
                </div>

                <SqlNotebookEditor
                    :notebook="activeNotebook"
                    :results-by-cell-id="resultsByCellId"
                    :active-cell-id="activeCellId"
                    :tables="props.tables"
                    :get-columns="props.getColumns"
                    :editor-settings="props.editorSettings"
                    :is-dirty="isDirty"
                    :is-saving="isSaving"
                    :last-saved-at="lastSavedAt"
                    :is-read-only="props.isReadOnly"
                    @update-title="updateNotebook($event.notebookId, { title: $event.value })"
                    @update-description="updateNotebook($event.notebookId, { description: $event.value })"
                    @update-tags="updateNotebook(activeNotebook?.id || '', { tags: $event })"
                    @update-metadata="activeNotebook && updateNotebook(activeNotebook.id, { metadata: { ...activeNotebook.metadata, ...$event } })"
                    @toggle-favorite="activeNotebook && toggleFavorite(activeNotebook.id)"
                    @add-cell="addCell($event)"
                    @share="openShareNotebook"
                    @update-cell-title="updateCell($event.cellId, { title: $event.value })"
                    @update-cell-content="updateCell($event.cellId, { content: $event.value })"
                    @update-cell-collapsed="updateCell($event.cellId, { collapsed: $event.value })"
                    @update-cell-execution-state="handleUpdateCellExecutionState"
                    @delete-cell="openDeleteCellWarning"
                    @duplicate-cell="duplicateCell"
                    @move-cell="moveCell($event.cellId, $event.direction)"
                    @run-cell="handleRunCell"
                    @save-snapshot="saveSnapshot"
                    @share-cell="openShareSqlCell"
                    @activate-cell="handleActivateCell"
                    @save="handleSave"
                />
            </div>

            <aside
                class="hidden min-h-0 shrink-0 border-l border-border bg-card xl:flex xl:flex-col"
                :class="isSnapshotsPanelCollapsed ? 'w-16' : 'w-80'"
            >
                <div
                    class="flex items-center justify-between border-b border-border px-3 py-4"
                    :class="isSnapshotsPanelCollapsed ? 'justify-center px-2' : 'gap-3'"
                >
                    <div v-if="!isSnapshotsPanelCollapsed" class="min-w-0">
                        <h3 class="text-sm font-semibold">Snapshots</h3>
                        <p class="mt-1 text-xs text-muted-foreground">
                            Saved result states for before and after comparisons.
                        </p>
                    </div>
                    <button
                        class="inline-flex h-8 w-8 shrink-0 items-center justify-center rounded-md border border-input bg-background text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
                        :title="isSnapshotsPanelCollapsed ? 'Open snapshots panel' : 'Collapse snapshots panel'"
                        @click="isSnapshotsPanelCollapsed = !isSnapshotsPanelCollapsed"
                    >
                        <svg
                            xmlns="http://www.w3.org/2000/svg"
                            width="16"
                            height="16"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            stroke-width="2"
                            stroke-linecap="round"
                            stroke-linejoin="round"
                            :class="isSnapshotsPanelCollapsed ? 'rotate-180' : ''"
                        >
                            <path d="m9 18 6-6-6-6" />
                        </svg>
                    </button>
                </div>

                <div v-if="!isSnapshotsPanelCollapsed" class="min-h-0 flex-1 overflow-y-auto p-4">
                    <div class="rounded-lg border border-border bg-background/80 p-3">
                        <SqlNotebookSnapshotsPanel
                            :snapshots="activeNotebookSnapshots"
                            @focus-cell="handleActivateCell"
                        />
                    </div>
                </div>
            </aside>
        </div>

        <SqlNotebookDeleteModal
            :is-open="!!notebookPendingDelete"
            :notebook-title="notebookPendingDelete?.title || 'this notebook'"
            @close="closeDeleteNotebookWarning"
            @confirm="confirmDeleteNotebook"
        />

        <SqlNotebookCellDeleteModal
            :is-open="!!cellPendingDelete"
            :cell-title="cellPendingDelete?.title || 'this cell'"
            :cell-type-label="cellPendingDelete?.type === 'sql' ? 'SQL' : cellPendingDelete?.type === 'runbook' ? 'Runbook' : 'Notes'"
            @close="closeDeleteCellWarning"
            @confirm="confirmDeleteCell"
        />

        <SqlNotebookShareModal
            :is-open="isShareModalOpen"
            :target-label="shareTargetLabel"
            :target-description="shareTargetDescription"
            :selected-view="shareView"
            :allow-live-session="canUseLiveSession"
            :has-live-code="!!activeLiveSession?.code"
            :preview-text="sharePreviewText"
            :expires-at="shareCodeExpiresAt"
            :live-status="activeLiveSession?.status"
            :live-peer-count="activeLiveSession?.peerCount"
            :error-message="shareModalError"
            :is-exporting="isExportingShare"
            @close="closeShareModal"
            @copy="copyShareContent"
            @export="exportShareContent"
            @disconnect-live="handleEndLiveSession"
            @update:view="shareView = $event"
        />

        <SqlNotebookImportCodeModal
            :is-open="isImportCodeModalOpen"
            :code="importShareCode"
            :is-importing="isImportingShareCode"
            :error-message="importShareCodeError"
            @close="closeImportCodeModal"
            @import="confirmImportShareCode"
            @update:code="importShareCode = $event"
        />

        <SqlNotebookExecutionGuardrailModal
            :is-open="!!guardrailPendingRun"
            :environment-label="guardrailEnvironmentLabel"
            :notebook-title="activeNotebook?.title || 'this notebook'"
            :cell-title="activeNotebook?.cells.find((cell) => cell.id === guardrailPendingRun?.cellId)?.title || 'SQL Cell'"
            :query-preview="guardrailPendingRun?.preview || ''"
            @close="closeExecutionGuardrail"
            @confirm="confirmGuardrailRun"
        />

        <SqlNotebookUnsavedChangesModal
            :is-open="!!notebookPendingSwitch"
            :current-notebook-title="activeNotebookTitle"
            :next-notebook-title="pendingNotebookTitle"
            :is-saving="isSaving"
            @close="closeUnsavedChangesWarning"
            @discard="confirmSwitchNotebookDiscardChanges"
            @save="confirmSwitchNotebookWithSave"
        />
    </div>
</template>


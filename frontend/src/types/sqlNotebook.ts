export type SqlNotebookCellType = 'sql' | 'markdown' | 'runbook';

export type SqlNotebookExecutionState = 'idle' | 'running' | 'success' | 'error' | 'verified' | 'skipped';

export type SqlNotebookVariableType = 'text' | 'number' | 'date';

export interface SqlNotebookEmbeddedImage {
    id: string;
    alt: string;
    fileName: string;
    mimeType: string;
    dataUrl: string;
}

export interface SqlNotebookCell {
    id: string;
    type: SqlNotebookCellType;
    title: string;
    content: string;
    collapsed: boolean;
    executionState: SqlNotebookExecutionState;
    lastRunAt?: string;
    embeddedImages?: SqlNotebookEmbeddedImage[];
}

export interface SqlNotebookVariable {
    key: string;
    label: string;
    value: string;
    required: boolean;
    type: SqlNotebookVariableType;
}

export interface SqlNotebookMetadata {
    environment: string;
    purpose: string;
    owner: string;
}

export interface SqlNotebookResultSet {
    columns: string[];
    rows: Array<Record<string, unknown>>;
    message?: string;
}

export interface SqlNotebookCellRunResult {
    status: 'idle' | 'running' | 'success' | 'error' | 'cancelled';
    resultSets: SqlNotebookResultSet[];
    errorMessage?: string;
    startedAt?: string;
    completedAt?: string;
    totalRows?: number;
}

export interface SqlNotebookResultSnapshot {
    id: string;
    cellId: string;
    cellTitle: string;
    capturedAt: string;
    totalRows: number;
    resultSets: SqlNotebookResultSet[];
}

export interface SqlNotebookRunbookContent {
    objective: string;
    checklist: string[];
    expectedResult: string;
    rollbackNotes: string;
    safetyNotes: string;
    approvals: SqlNotebookRunbookApproval[];
    evidence: string;
}

export type SqlNotebookRunbookApprovalStatus = 'pending' | 'approved' | 'blocked';

export interface SqlNotebookRunbookApproval {
    id: string;
    name: string;
    role: string;
    status: SqlNotebookRunbookApprovalStatus;
    note: string;
}

export interface SqlNotebookTemplatePreset {
    id: string;
    title: string;
    description: string;
    category: 'operations' | 'release' | 'analysis';
    metadata: Partial<SqlNotebookMetadata>;
    tags: string[];
    variables: SqlNotebookVariable[];
    cells: Array<{
        type: SqlNotebookCellType;
        title: string;
        content: string;
        embeddedImages?: SqlNotebookEmbeddedImage[];
    }>;
}

export interface SqlNotebookConnectionScope {
    dbType: string;
    connectionName: string;
}

export type SqlNotebookShareScope = 'notebook' | 'sql';

export type SqlNotebookShareView = 'summary' | 'live';

export type SqlNotebookLiveSessionStatus = 'idle' | 'hosting' | 'joining' | 'connected' | 'expired' | 'ended' | 'error';

export interface SqlNotebook {
    id: string;
    title: string;
    description: string;
    tags: string[];
    variables: SqlNotebookVariable[];
    metadata: SqlNotebookMetadata;
    isFavorite: boolean;
    isTemplate: boolean;
    snapshots: SqlNotebookResultSnapshot[];
    connectionScope: SqlNotebookConnectionScope;
    cells: SqlNotebookCell[];
    createdAt: string;
    updatedAt: string;
    lastOpenedAt: string;
}

export interface SqlNotebookShareBundleSource {
    dbType: string;
    connectionNameRedacted: true;
    scope: SqlNotebookShareScope;
}

export interface SqlNotebookShareBundleNotebook {
    title: string;
    description: string;
    tags: string[];
    variables: SqlNotebookVariable[];
    metadata: SqlNotebookMetadata;
    cells: SqlNotebookCell[];
}

export interface SqlNotebookShareBundle {
    kind: typeof SQL_NOTEBOOK_SHARE_BUNDLE_KIND;
    version: typeof SQL_NOTEBOOK_SHARE_BUNDLE_VERSION;
    exportedAt: string;
    redacted: true;
    source: SqlNotebookShareBundleSource;
    notebook: SqlNotebookShareBundleNotebook;
}

export interface SqlNotebookLiveSession {
    code: string;
    sessionId: string;
    notebookId: string;
    status: SqlNotebookLiveSessionStatus;
    role: 'host' | 'peer';
    expiresAt: string;
    peerCount: number;
    lastMessageAt?: string;
    errorMessage?: string;
}

export const SQL_NOTEBOOK_SHARE_BUNDLE_KIND = 'sql-notebook-share-bundle';
export const SQL_NOTEBOOK_SHARE_BUNDLE_VERSION = 1;

export const createDefaultRunbookContent = (): SqlNotebookRunbookContent => ({
    objective: 'Describe the operational goal for this step.',
    checklist: [
        'Confirm the target connection and environment.',
        'Review expected impact before continuing.',
        'Capture the result after execution.',
    ],
    expectedResult: 'Document what success looks like after this step.',
    rollbackNotes: 'Add rollback or fallback instructions if this step fails.',
    safetyNotes: 'List approvals, maintenance windows, or safety checks required before running SQL.',
    approvals: [
        {
            id: createId('approval'),
            name: 'Database owner',
            role: 'Required approver',
            status: 'pending',
            note: 'Confirm impact and maintenance window before execution.',
        },
    ],
    evidence: 'Capture links, screenshots, row counts, or notes that prove the step was completed safely.',
});

const normalizeApprovalStatus = (value: unknown): SqlNotebookRunbookApprovalStatus => {
    return value === 'approved' || value === 'blocked' ? value : 'pending';
};

const normalizeRunbookApprovals = (value: unknown): SqlNotebookRunbookApproval[] => {
    const defaultApprovals = createDefaultRunbookContent().approvals;
    if (!Array.isArray(value)) {
        return defaultApprovals;
    }

    const normalized = value
        .map((item, index) => {
            if (typeof item === 'string') {
                const name = item.trim();
                if (!name) {
                    return null;
                }

                return {
                    id: createId(`approval_${index}`),
                    name,
                    role: '',
                    status: 'pending',
                    note: '',
                } satisfies SqlNotebookRunbookApproval;
            }

            if (!item || typeof item !== 'object') {
                return null;
            }

            const approval = item as Partial<SqlNotebookRunbookApproval>;
            const name = typeof approval.name === 'string' ? approval.name.trim() : '';
            const role = typeof approval.role === 'string' ? approval.role : '';
            if (!name && !role) {
                return null;
            }

            return {
                id: typeof approval.id === 'string' && approval.id.trim().length > 0 ? approval.id : createId(`approval_${index}`),
                name: name || role || `Approver ${index + 1}`,
                role,
                status: normalizeApprovalStatus(approval.status),
                note: typeof approval.note === 'string' ? approval.note : '',
            } satisfies SqlNotebookRunbookApproval;
        })
        .filter((item): item is SqlNotebookRunbookApproval => !!item);

    return normalized.length > 0 ? normalized : defaultApprovals;
};

export const parseRunbookContent = (content: string): SqlNotebookRunbookContent => {
    try {
        const parsed = JSON.parse(content) as Partial<SqlNotebookRunbookContent>;
        return {
            objective: typeof parsed.objective === 'string' ? parsed.objective : createDefaultRunbookContent().objective,
            checklist: Array.isArray(parsed.checklist)
                ? parsed.checklist.filter((item): item is string => typeof item === 'string' && item.trim().length > 0)
                : createDefaultRunbookContent().checklist,
            expectedResult: typeof parsed.expectedResult === 'string' ? parsed.expectedResult : createDefaultRunbookContent().expectedResult,
            rollbackNotes: typeof parsed.rollbackNotes === 'string' ? parsed.rollbackNotes : createDefaultRunbookContent().rollbackNotes,
            safetyNotes: typeof parsed.safetyNotes === 'string' ? parsed.safetyNotes : createDefaultRunbookContent().safetyNotes,
            approvals: normalizeRunbookApprovals(parsed.approvals),
            evidence: typeof parsed.evidence === 'string' ? parsed.evidence : createDefaultRunbookContent().evidence,
        };
    } catch (_error) {
        return createDefaultRunbookContent();
    }
};

export const stringifyRunbookContent = (content: SqlNotebookRunbookContent): string => {
    return JSON.stringify(content);
};

const createId = (prefix: string): string => {
    return `${prefix}_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`;
};

export const createSqlNotebookCell = (type: SqlNotebookCellType): SqlNotebookCell => {
    const isSql = type === 'sql';

    return {
        id: createId('cell'),
        type,
        title: isSql ? 'SQL Cell' : type === 'runbook' ? 'Runbook Step' : 'Notes',
        content: isSql
            ? '-- Write your SQL here\nSELECT 1;'
            : type === 'runbook'
                ? stringifyRunbookContent(createDefaultRunbookContent())
                : 'Document the goal, assumptions, and next steps for this notebook.',
        collapsed: false,
        executionState: type === 'runbook' ? 'idle' : 'idle',
        embeddedImages: type === 'markdown' ? [] : undefined,
    };
};

export const SQL_NOTEBOOK_TEMPLATE_PRESETS: SqlNotebookTemplatePreset[] = [
    {
        id: 'incident-investigation',
        title: 'Incident Investigation',
        description: 'Collect context, run focused diagnostics, and capture rollback notes for a live issue.',
        category: 'operations',
        metadata: { purpose: 'Incident response', environment: 'production' },
        tags: ['incident', 'diagnostics'],
        variables: [
            { key: 'entity_id', label: 'Entity ID', value: '', required: true, type: 'text' },
        ],
        cells: [
            {
                type: 'runbook',
                title: 'Triage Checklist',
                content: stringifyRunbookContent({
                    objective: 'Confirm the incident scope before running any production query.',
                    checklist: [
                        'Confirm impacted service and timeframe.',
                        'Check maintenance window and change freeze status.',
                        'Notify stakeholders before running write operations.',
                    ],
                    expectedResult: 'A clear scope and approved investigation window.',
                    rollbackNotes: 'Pause and escalate if approval is missing.',
                    safetyNotes: 'Production write queries require extra confirmation.',
                    approvals: [
                        {
                            id: createId('approval'),
                            name: 'Incident commander',
                            role: 'Incident lead',
                            status: 'pending',
                            note: 'Approve scope before any production action.',
                        },
                        {
                            id: createId('approval'),
                            name: 'Database owner',
                            role: 'Write-query approver',
                            status: 'pending',
                            note: 'Required before any data-changing SQL.',
                        },
                    ],
                    evidence: 'Paste impacted record samples and incident ticket updates here.',
                }),
            },
            {
                type: 'sql',
                title: 'Load Impacted Records',
                content: 'SELECT *\nFROM impacted_records\nWHERE entity_id = {{entity_id}}\nORDER BY updated_at DESC;',
            },
            {
                type: 'markdown',
                title: 'Investigation Notes',
                content: 'Capture findings, owner updates, and decisions made during the incident.',
            },
        ],
    },
    {
        id: 'release-readiness',
        title: 'Release Readiness',
        description: 'Track deployment checks, validation queries, and rollback guidance before release.',
        category: 'release',
        metadata: { purpose: 'Release checklist', environment: 'staging' },
        tags: ['release', 'checklist'],
        variables: [],
        cells: [
            {
                type: 'runbook',
                title: 'Pre-release Checklist',
                content: stringifyRunbookContent({
                    objective: 'Verify the environment is ready for deployment.',
                    checklist: [
                        'Confirm backup status.',
                        'Validate migration order.',
                        'Assign rollback owner.',
                    ],
                    expectedResult: 'Release can proceed with documented approvals.',
                    rollbackNotes: 'Use previous artifact and rollback migration if validation fails.',
                    safetyNotes: 'Run validation SQL before any write migration.',
                    approvals: [
                        {
                            id: createId('approval'),
                            name: 'Release manager',
                            role: 'Release approver',
                            status: 'pending',
                            note: 'Confirm release window and rollback owner.',
                        },
                        {
                            id: createId('approval'),
                            name: 'Migration owner',
                            role: 'Database approver',
                            status: 'pending',
                            note: 'Approve migration order before deployment.',
                        },
                    ],
                    evidence: 'Capture release ticket, migration checksum, and smoke-test output.',
                }),
            },
            {
                type: 'sql',
                title: 'Schema Validation',
                content: 'SELECT table_name, column_name\nFROM information_schema.columns\nWHERE table_schema NOT IN (\'information_schema\', \'pg_catalog\')\nORDER BY table_name, ordinal_position;',
            },
        ],
    },
    {
        id: 'table-audit',
        title: 'Table Audit',
        description: 'Inspect row counts, freshness, and anomalies for a target table quickly.',
        category: 'analysis',
        metadata: { purpose: 'Audit', environment: 'dev' },
        tags: ['audit', 'quality'],
        variables: [
            { key: 'table_name', label: 'Table Name', value: '', required: true, type: 'text' },
        ],
        cells: [
            {
                type: 'markdown',
                title: 'Audit Scope',
                content: 'Document why this table is being audited and what anomaly you are checking.',
            },
            {
                type: 'sql',
                title: 'Row Count',
                content: 'SELECT COUNT(*) AS total_rows\nFROM {{table_name}};',
            },
            {
                type: 'sql',
                title: 'Latest Activity',
                content: 'SELECT *\nFROM {{table_name}}\nORDER BY 1 DESC\nLIMIT 50;',
            },
        ],
    },
];

export const createSqlNotebook = (scope: SqlNotebookConnectionScope): SqlNotebook => {
    const now = new Date().toISOString();

    return {
        id: createId('notebook'),
        title: 'Untitled Notebook',
        description: 'Capture reusable SQL workflows, notes, and operational context.',
        tags: [],
        variables: [],
        metadata: {
            environment: '',
            purpose: '',
            owner: '',
        },
        isFavorite: false,
        isTemplate: false,
        snapshots: [],
        connectionScope: scope,
        cells: [
            createSqlNotebookCell('markdown'),
            createSqlNotebookCell('sql'),
        ],
        createdAt: now,
        updatedAt: now,
        lastOpenedAt: now,
    };
};

export const isSqlNotebookShareBundle = (value: unknown): value is SqlNotebookShareBundle => {
    if (!value || typeof value !== 'object') {
        return false;
    }

    const bundle = value as Partial<SqlNotebookShareBundle>;
    return bundle.kind === SQL_NOTEBOOK_SHARE_BUNDLE_KIND
        && bundle.version === SQL_NOTEBOOK_SHARE_BUNDLE_VERSION
        && bundle.redacted === true
        && !!bundle.source
        && typeof bundle.source.dbType === 'string'
        && bundle.source.connectionNameRedacted === true
        && (bundle.source.scope === 'notebook' || bundle.source.scope === 'sql')
        && !!bundle.notebook
        && typeof bundle.notebook.title === 'string'
        && typeof bundle.notebook.description === 'string'
        && Array.isArray(bundle.notebook.tags)
        && Array.isArray(bundle.notebook.variables)
        && !!bundle.notebook.metadata
        && Array.isArray(bundle.notebook.cells);
};

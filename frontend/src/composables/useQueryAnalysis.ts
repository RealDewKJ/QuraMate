import type { Ref } from 'vue';

import { ExplainQuery } from '../../wailsjs/go/main/App';
import { useAiCopilot } from './useAiCopilot';
import type { QueryTab } from '../types/dashboard';
import type { AIMessage } from '../lib/ai/client';

interface AiCompletionOptions {
    temperature?: number;
    maxTokens?: number;
}

interface UseQueryAnalysisOptions {
    activeTab: Ref<QueryTab | undefined>;
    connectionId: Ref<string>;
    connectionName: Ref<string | undefined>;
    dbType: Ref<string>;
    tables: Ref<string[]>;
    getSelectedQuery: () => string;
    fetchTableColumns: (tableName: string) => Promise<string[]>;
    loadHistory: (scope: string) => Promise<{ query: string }[]>;
    complete: (messages: AIMessage[], options?: AiCompletionOptions) => Promise<{ text: string }>;
    onToastError?: (message: string) => void;
    onToastSuccess?: (message: string) => void;
}

export function useQueryAnalysis(options: UseQueryAnalysisOptions) {
    const getCurrentQueryForAnalysis = () => {
        const tab = options.activeTab.value;
        if (!tab) return '';
        const selection = options.getSelectedQuery();
        if (selection && selection.trim()) {
            return selection;
        }
        return tab.query;
    };

    const collectSchemaContext = async (maxTables: number = 12) => {
        const tableNames = (options.tables.value || []).slice(0, maxTables);
        if (tableNames.length === 0) return 'No schema available.';

        const schemaRows = await Promise.all(tableNames.map(async (tableName) => {
            try {
                const columns = await options.fetchTableColumns(tableName);
                return `${tableName}(${(columns || []).slice(0, 20).join(', ')})`;
            } catch {
                return `${tableName}(...)`;
            }
        }));

        return schemaRows.join('\n');
    };

    const collectResultSampleContext = () => {
        const tab = options.activeTab.value;
        if (!tab || !tab.resultSets || tab.resultSets.length === 0) return 'No result sample.';
        const rs = tab.resultSets[0];
        if (!rs || !rs.rows || rs.rows.length === 0) return 'No result sample.';

        const rows = rs.rows.slice(0, 5);
        return JSON.stringify(rows, null, 2);
    };

    const {
        aiCopilot,
        aiCopilotModeOptions,
        openAiCopilot,
        setAiCopilotMode,
        runAiCopilot,
        applyAiSqlToEditor
    } = useAiCopilot({
        getRuntimeContext: async () => ({
            dbType: options.dbType.value || 'SQL',
            historyScope: options.connectionName.value || options.dbType.value || '',
            currentQuery: getCurrentQueryForAnalysis(),
            currentError: options.activeTab.value?.error || '',
            currentPlan: options.activeTab.value?.explanation || '',
            resultSample: collectResultSampleContext(),
            schemaContext: await collectSchemaContext()
        }),
        loadHistory: (scope) => options.loadHistory(scope),
        complete: (messages, completionOptions) => options.complete(messages, completionOptions),
        onApplySql: (sql) => {
            if (!options.activeTab.value) return;
            options.activeTab.value.query = sql;
        },
        onAppliedSql: () => {
            options.onToastSuccess?.('Applied AI SQL to editor');
        }
    });

    const analyzeQuery = async () => {
        const tab = options.activeTab.value;
        if (!tab || !tab.query.trim()) return;

        tab.isExplaining = true;
        tab.error = '';
        tab.explanation = 'Analyzing...';
        tab.resultViewTab = 'analysis';
        tab.queryExecuted = true;

        const queryToAnalyze = getCurrentQueryForAnalysis();

        try {
            const plan = await ExplainQuery(options.connectionId.value, queryToAnalyze);
            tab.explanation = plan;
        } catch (err: any) {
            tab.error = 'Failed to analyze query: ' + err.toString();
            tab.explanation = undefined;
        } finally {
            tab.isExplaining = false;
        }
    };

    const explainWithAI = async () => {
        const tab = options.activeTab.value;
        if (!tab || !tab.query.trim()) return;

        const queryToAnalyze = getCurrentQueryForAnalysis();
        tab.isAiExplaining = true;
        tab.aiExplanation = 'Analyzing with AI...';
        tab.resultViewTab = 'analysis';
        tab.queryExecuted = true;

        try {
            const dbType = options.dbType.value || 'SQL';
            const contextPlan = tab.explanation ? `\n\nQuery execution plan:\n${tab.explanation}` : '';
            const result = await options.complete(
                [
                    {
                        role: 'system',
                        content: `You are a senior ${dbType} performance engineer. Explain SQL clearly and safely. Keep response concise with sections: Summary, Performance Risks, Suggested Rewrite.`
                    },
                    {
                        role: 'user',
                        content: `Analyze this ${dbType} SQL query:\n\n${queryToAnalyze}${contextPlan}`
                    }
                ],
                { temperature: 0.2, maxTokens: 700 }
            );
            tab.aiExplanation = result.text || 'AI returned an empty response.';
        } catch (err: any) {
            const errorMessage = err?.message || String(err);
            tab.aiExplanation = `Failed to explain with AI: ${errorMessage}`;
            options.onToastError?.(`AI explain failed: ${errorMessage}`);
        } finally {
            tab.isAiExplaining = false;
        }
    };

    return {
        aiCopilot,
        aiCopilotModeOptions,
        openAiCopilot,
        setAiCopilotMode,
        runAiCopilot,
        applyAiSqlToEditor,
        analyzeQuery,
        explainWithAI,
    };
}

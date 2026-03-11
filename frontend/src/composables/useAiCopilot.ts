import { reactive, watch } from 'vue';
import type { AIMessage } from '../lib/ai/client';
import type { AiCopilotMode, AiCopilotModeOption } from '../types/aiCopilot';

interface AiCopilotState {
    isOpen: boolean;
    mode: AiCopilotMode;
    prompt: string;
    backendLanguage: string;
    isLoading: boolean;
    result: string;
    error: string;
    latencyMs: number;
    suggestedSQL: string;
}

interface AiCopilotRuntimeContext {
    dbType: string;
    historyScope: string;
    currentQuery: string;
    currentError: string;
    currentPlan: string;
    resultSample: string;
    schemaContext: string;
}

interface AiCopilotHistoryItem {
    query: string;
}

interface AiCompletionResponse {
    text: string;
}

interface AiCompletionOptions {
    temperature?: number;
    maxTokens?: number;
}

interface UseAiCopilotOptions {
    getRuntimeContext: () => Promise<AiCopilotRuntimeContext> | AiCopilotRuntimeContext;
    loadHistory: (scope: string) => Promise<AiCopilotHistoryItem[]>;
    complete: (messages: AIMessage[], options?: AiCompletionOptions) => Promise<AiCompletionResponse>;
    onApplySql: (sql: string) => void;
    onAppliedSql?: () => void;
}

const defaultModePrompts: Record<AiCopilotMode, string> = {
    text_to_sql: 'Retrieve users who signed up in the last 30 days, along with each user\'s order count.',
    context_completion: 'Complete the current SQL query by joining all relevant related tables.',
    refine_query: 'Refine the existing query to filter sales greater than 500 and sort by the most recent date.',
    sql_explainer: 'Explain this SQL step by step so a non-technical team can understand it.',
    fix_error: 'Fix this SQL error and provide a query that runs successfully.',
    optimize: 'Optimize this query for better performance, and include reasons plus recommended indexes.',
    mock_data: 'Generate 20 realistic mock data rows for this table.',
    schema_insights: 'Recommend schema improvements to support faster query performance.',
    analyze_plan: 'Analyze and explain this execution plan, and suggest ways to resolve bottlenecks.',
    backend_code: 'Generate backend code from this SQL query.',
    summary: 'Summarize key insights from this query result.'
};

const modeOptions: AiCopilotModeOption[] = [
    { value: 'text_to_sql', label: 'Natural Language to SQL' },
    { value: 'context_completion', label: 'Context-Aware Completion' },
    { value: 'refine_query', label: 'Refine Query via Chat' },
    { value: 'sql_explainer', label: 'SQL Explainer (Step-by-step)' },
    { value: 'fix_error', label: 'Fix Query Errors' },
    { value: 'optimize', label: 'Performance Optimization' },
    { value: 'mock_data', label: 'Generate Mock Data' },
    { value: 'schema_insights', label: 'Schema Insights & Recommendations' },
    { value: 'analyze_plan', label: 'Analyze Query Plan' },
    { value: 'backend_code', label: 'Generate Backend Code' },
    { value: 'summary', label: 'Summary of Results' }
];

const extractFirstSqlBlock = (text: string): string => {
    const sqlBlockMatch = text.match(/```sql\s*([\s\S]*?)```/i);
    if (sqlBlockMatch && sqlBlockMatch[1]) return sqlBlockMatch[1].trim();

    const genericBlockMatch = text.match(/```[\w-]*\s*([\s\S]*?)```/i);
    if (genericBlockMatch && genericBlockMatch[1]) {
        const candidate = genericBlockMatch[1].trim();
        if (/^(select|with|insert|update|delete|create|alter|drop|truncate|explain)\b/i.test(candidate)) {
            return candidate;
        }
    }

    const sqlLineMatch = text.match(/(select|with|insert|update|delete|create|alter|drop|truncate|explain)[\s\S]*/i);
    return sqlLineMatch ? sqlLineMatch[0].trim() : '';
};

const isSqlOnlyMode = (mode: AiCopilotMode): boolean => {
    return [
        'text_to_sql',
        'context_completion',
        'refine_query',
        'fix_error',
        'optimize',
        'mock_data'
    ].includes(mode);
};

const normalizeSqlOnlyResponse = (text: string): string => {
    const fromBlock = extractFirstSqlBlock(text);
    if (fromBlock) return fromBlock;
    return '';
};

export const useAiCopilot = (options: UseAiCopilotOptions) => {
    const aiCopilot = reactive<AiCopilotState>({
        isOpen: false,
        mode: 'text_to_sql',
        prompt: '',
        backendLanguage: 'Node.js',
        isLoading: false,
        result: '',
        error: '',
        latencyMs: 0,
        suggestedSQL: ''
    });

    const open = (mode?: AiCopilotMode) => {
        if (mode) aiCopilot.mode = mode;
        aiCopilot.isOpen = true;
        aiCopilot.error = '';
        if (!aiCopilot.prompt.trim()) {
            aiCopilot.prompt = defaultModePrompts[aiCopilot.mode];
        }
    };

    const setMode = (mode: AiCopilotMode) => {
        aiCopilot.mode = mode;
    };

    watch(() => aiCopilot.mode, (mode) => {
        if (!aiCopilot.prompt.trim()) {
            aiCopilot.prompt = defaultModePrompts[mode];
        }
    });

    const run = async () => {
        if (!aiCopilot.prompt.trim()) {
            aiCopilot.error = 'Please enter an instruction.';
            return;
        }

        aiCopilot.isLoading = true;
        aiCopilot.error = '';
        aiCopilot.result = '';
        aiCopilot.suggestedSQL = '';
        aiCopilot.latencyMs = 0;

        const startedAt = performance.now();

        try {
            const context = await options.getRuntimeContext();
            const history = await options.loadHistory(context.historyScope);
            const recentHistory = (history || []).slice(0, 8).map((item) => item.query).join('\n---\n');

            const taskLabel = modeOptions.find((m) => m.value === aiCopilot.mode)?.label || aiCopilot.mode;
            const sqlOnlyMode = isSqlOnlyMode(aiCopilot.mode);
            const systemPrompt = sqlOnlyMode
                ? `You are a senior SQL copilot.
Return ONLY runnable SQL text.
Do NOT include analysis, explanations, bullet points, markdown, or code fences.
Use only table/column names from provided schema and context.`
                : `You are a senior database copilot for SQL work.
Return concise, actionable output.
If task expects runnable SQL, include exactly one final SQL block in \`\`\`sql ... \`\`\`.
If task is backend code, return one complete code block in ${aiCopilot.backendLanguage}.
Always use provided schema and context strictly.`;

            const backendPreferenceSection = aiCopilot.mode === 'backend_code'
                ? `Preferred Backend Language: ${aiCopilot.backendLanguage}`
                : 'Preferred Backend Language: (not required for this task)';

            const userPrompt = `Task: ${taskLabel}
DB Type: ${context.dbType}
Instruction: ${aiCopilot.prompt}
${backendPreferenceSection}

Current SQL:
${context.currentQuery || '(empty)'}

Current Error:
${context.currentError || '(none)'}

Execution Plan:
${context.currentPlan || '(none)'}

Recent Query History:
${recentHistory || '(none)'}

Schema Snapshot:
${context.schemaContext}

Result Sample (JSON):
${context.resultSample}`;

            const response = await options.complete(
                [
                    { role: 'system', content: systemPrompt },
                    { role: 'user', content: userPrompt }
                ],
                { temperature: 0.2, maxTokens: 1400 }
            );

            if (sqlOnlyMode) {
                const normalizedSql = normalizeSqlOnlyResponse(response.text);
                if (!normalizedSql) {
                    throw new Error('AI did not return runnable SQL. Please try again with more specific instruction.');
                }
                aiCopilot.result = normalizedSql;
                aiCopilot.suggestedSQL = normalizedSql;
            } else {
                aiCopilot.result = response.text;
                aiCopilot.suggestedSQL = extractFirstSqlBlock(response.text);
            }

            aiCopilot.latencyMs = Math.round(performance.now() - startedAt);
        } catch (err: any) {
            const message = err?.message || String(err);
            const details = err?.details ? `\n${JSON.stringify(err.details, null, 2)}` : '';
            aiCopilot.error = `${message}${details}`;
            aiCopilot.latencyMs = Math.round(performance.now() - startedAt);
        } finally {
            aiCopilot.isLoading = false;
        }
    };

    const applySuggestedSql = () => {
        if (!aiCopilot.suggestedSQL) return;
        options.onApplySql(aiCopilot.suggestedSQL);
        aiCopilot.isOpen = false;
        options.onAppliedSql?.();
    };

    return {
        aiCopilot,
        aiCopilotModeOptions: modeOptions,
        openAiCopilot: open,
        setAiCopilotMode: setMode,
        runAiCopilot: run,
        applyAiSqlToEditor: applySuggestedSql
    };
};


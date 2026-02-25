export type AiCopilotMode =
    | 'text_to_sql'
    | 'context_completion'
    | 'refine_query'
    | 'sql_explainer'
    | 'fix_error'
    | 'optimize'
    | 'mock_data'
    | 'schema_insights'
    | 'analyze_plan'
    | 'backend_code'
    | 'summary';

export interface AiCopilotModeOption {
    value: AiCopilotMode;
    label: string;
}
